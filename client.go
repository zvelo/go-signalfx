package signalfx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zvelo/go-signalfx/sfxconfig"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

const (
	// tokenHeaderName is the header key for the auth token in the HTTP request
	tokenHeaderName = "X-SF-TOKEN"

	contentType = "application/x-protobuf"
)

// A Client is used to send datapoints to SignalFx
type Client struct {
	config *sfxconfig.Config
	tr     *http.Transport
	client *http.Client
}

// New returns a new Client
func New(config *sfxconfig.Config) *Client {
	// TODO(jrubin) validate the config?

	tr := config.Transport()

	return &Client{
		config: config,
		tr:     tr,
		client: &http.Client{Transport: tr},
	}
}

// Submit forwards datapoints to SignalFx
func (c *Client) Submit(ctx context.Context, msg *sfxproto.DataPointUploadMessage) error {
	c.config.Lock()
	endpoint := c.config.URL
	userAgent := c.config.UserAgent
	authToken := c.config.AuthToken
	c.config.Unlock()

	jsonBytes, err := msg.Marshal(c.config)
	if err != nil {
		return newError("Unable to marshal object", err)
	}

	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBytes))
	req.Header = http.Header{
		tokenHeaderName: {authToken},
		"Content-Type":  {contentType},
		"User-Agent":    {userAgent},
		"Connection":    {"Keep-Alive"},
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
		return ctx.Err()
	case <-done:
		if err != nil {
			return newError("Unable to POST request", err)
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return newError("Unable to verify response body", err)
	}

	if resp.StatusCode != 200 {
		return newError(string(respBody), fmt.Errorf("invalid status code: %d", resp.StatusCode))
	}

	var body string
	err = json.Unmarshal(respBody, &body)
	if err != nil {
		return newError(string(respBody), err)
	}

	if body != "OK" {
		return newError(body, errors.New("body decode error"))
	}

	return nil
}
