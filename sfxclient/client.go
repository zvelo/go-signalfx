package sfxclient

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/zvelo/go-signalfx/sfxconfig"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// A Client is used to send datapoints to SignalFx
type Client struct {
	config *sfxconfig.Config
	tr     *http.Transport
	client *http.Client
}

// New returns a new Client
func New(config *sfxconfig.Config) *Client {
	tr := config.Transport()

	return &Client{
		config: config.Clone(),
		tr:     tr,
		client: &http.Client{Transport: tr},
	}
}

// Submit forwards raw datapoints to SignalFx
func (c *Client) Submit(ctx context.Context, dps *sfxproto.DataPoints) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	jsonBytes, err := dps.Marshal(c.config)
	if err != nil {
		return ErrMarshal(err)
	}

	req, _ := http.NewRequest("POST", c.config.URL, bytes.NewBuffer(jsonBytes))
	req.Header = http.Header{
		"User-Agent":   {c.config.UserAgent},
		"X-SF-TOKEN":   {c.config.AuthToken},
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
		c.tr.CancelRequest(req)
		<-done // wait for the request to be cancelled
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
