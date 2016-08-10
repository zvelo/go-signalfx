package signalfx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
	"zvelo.io/go-signalfx/sfxproto"
)

const (
	// TokenHeader is the header on which SignalFx looks for the api token
	TokenHeader = "X-SF-TOKEN"
)

// A Client is used to send datapoints to SignalFx
type Client struct {
	config *Config
	tr     http.RoundTripper
	client *http.Client
}

// NewClient returns a new Client. config is copied, so future changes to the
// external config object are not reflected within the client.
func NewClient(config *Config) *Client {
	tr := config.Transport()

	return &Client{
		config: config.Clone(),
		tr:     tr,
		client: &http.Client{Transport: tr},
	}
}

// Submit forwards raw datapoints to SignalFx
func (c *Client) Submit(ctx context.Context, pdps *sfxproto.DataPoints) error {
	if ctx == nil {
		ctx = context.Background()
	} else if ctx.Err() != nil {
		return ErrContext(ctx.Err())
	}

	jsonBytes, err := pdps.Marshal()
	if err != nil {
		return ErrMarshal(err)
	}

	req, _ := http.NewRequest("POST", c.config.URL, bytes.NewBuffer(jsonBytes))
	req.Header = http.Header{
		TokenHeader:    {c.config.AuthToken},
		"User-Agent":   {c.config.UserAgent},
		"Connection":   {"Keep-Alive"},
		"Content-Type": {"application/x-protobuf"},
	}

	var resp *http.Response
	done := make(chan interface{}, 1)

	go func() {
		resp, err = c.client.Do(req)
		done <- true
	}()

	select {
	case <-ctx.Done():
		if tr, ok := c.tr.(*http.Transport); ok {
			tr.CancelRequest(req)
			<-done // wait for the request to be canceled
		} else {
			if c.config.Logger != nil {
				fmt.Fprintf(c.config.Logger, "tried to cancel non-cancellable transport %T", tr)
			}
		}
		return ErrContext(ctx.Err())
	case <-done:
		if err != nil {
			return ErrPost(err)
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrResponse(err)
	}

	if resp.StatusCode != 200 {
		return &ErrStatus{respBody, resp.StatusCode}
	}

	var body string
	if err = json.Unmarshal(respBody, &body); err != nil {
		return &ErrJSON{respBody}
	}

	if body != "OK" {
		return &ErrInvalidBody{body}
	}

	return nil
}
