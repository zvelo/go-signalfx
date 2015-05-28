package sfxconfig

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

const (
	// ClientVersion is the version of this package, sent as part of the default
	// user agent
	ClientVersion = "0.1.0"

	// DefaultMaxConnections is the maximum idle (keep-alive) connections to
	// maintain with the signalfx server.
	DefaultMaxConnections = 20

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
	MaxConnections        uint32
	TimeoutDuration       time.Duration
	URL                   string
	AuthToken             string
	UserAgent             string
	DefaultSource         string
	TLSInsecureSkipVerify bool
	DimensionSources      []string
}

// Clone makes a deep copy of a Config
func (c *Config) Clone() *Config {
	ret := *c
	copy(ret.DimensionSources, c.DimensionSources)
	return &ret
}

// Transport returns an http.Transport configured according to Config
func (c *Config) Transport() *http.Transport {
	return &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: c.TLSInsecureSkipVerify},
		MaxIdleConnsPerHost:   int(c.MaxConnections),
		ResponseHeaderTimeout: c.TimeoutDuration,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, c.TimeoutDuration)
		},
	}
}

// New generates a new Config with default values
func New() *Config {
	return &Config{
		MaxConnections:  DefaultMaxConnections,
		TimeoutDuration: DefaultTimeoutDuration,
		URL:             DefaultURL,
		UserAgent:       DefaultUserAgent,
	}
}
