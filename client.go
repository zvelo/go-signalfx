package signalfx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/zvelo/go-signalfx/sfconfig"
	"github.com/zvelo/go-signalfx/sfproto"
	"golang.org/x/net/context"
)

// A Client is used to send datapoints to SignalFx
type Client struct {
	config *sfconfig.Config
	tr     *http.Transport
	client *http.Client
}

// New returns a new Client
func New(config *sfconfig.Config) *Client {
	tr := config.Transport()

	return &Client{
		config: config.Clone(),
		tr:     tr,
		client: &http.Client{Transport: tr},
	}
}

// Submit forwards datapoints to SignalFx
func (c *Client) Submit(ctx context.Context, dp sfproto.DataPoints) error {
	jsonBytes, err := dp.Marshal(c.config)
	if err != nil {
		return newError("Unable to marshal object", err)
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
	if err = json.Unmarshal(respBody, &body); err != nil {
		return newError(string(respBody), err)
	}

	if body != "OK" {
		return newError(body, errors.New("body decode error"))
	}

	return nil
}
