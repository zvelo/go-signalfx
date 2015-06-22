package signalfx

import (
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	// ClientVersion is the version of this package, sent as part of the default
	// user agent
	ClientVersion = "0.1.0"

	// DefaultMaxIdleConnections is the maximum idle (keep-alive) connections to
	// maintain with the signalfx server.
	DefaultMaxIdleConnections = 2

	// DefaultTimeoutDuration is the timeout used for connecting to the signalfx
	// server, including name resolution, as well as weaiting for headers
	DefaultTimeoutDuration = 60 * time.Second

	// DefaultURL is the URL used to send datapoints to signalfx
	DefaultURL = "https://ingest.signalfx.com/v2/datapoint"

	// DefaultUserAgent is the user agent sent to signalfx
	DefaultUserAgent = "go-signalfx/" + ClientVersion
)

// Config is used to configure a Client. It should be created with New to have
// default values automatically set.
type Config struct {
	MaxIdleConnections    uint32
	TimeoutDuration       time.Duration
	URL                   string
	AuthToken             string
	UserAgent             string
	TLSInsecureSkipVerify bool
}

// Clone makes a deep copy of a Config
func (c *Config) Clone() *Config {
	// there are no fields, presently, that require any spacial handling
	ret := *c
	return &ret
}

// Transport returns an http.Transport configured according to Config
func (c *Config) Transport() *http.Transport {
	return &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: c.TLSInsecureSkipVerify},
		MaxIdleConnsPerHost:   int(c.MaxIdleConnections),
		ResponseHeaderTimeout: c.TimeoutDuration,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, c.TimeoutDuration)
		},
	}
}

// NewConfig generates a new Config with default values. If $SFX_API_TOKEN is
// set in the environment, it will be used as the AuthToken.
func NewConfig() *Config {
	return &Config{
		MaxIdleConnections: DefaultMaxIdleConnections,
		TimeoutDuration:    DefaultTimeoutDuration,
		URL:                DefaultURL,
		UserAgent:          DefaultUserAgent,
		AuthToken:          os.Getenv("SFX_API_TOKEN"),
	}
}
