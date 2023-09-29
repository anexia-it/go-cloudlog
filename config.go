package cloudlog

import (
	"net/http"
	"os"
)

// Config is the go-cloudlog config
type Config struct {
	Hostname string
	Encoder  EventEncoder
	Client   Client
	BaseURL  string
}

// Client is the HTTP Client managing the connection to Cloudlog
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewDefaultConfig returns a new config with default values
func NewDefaultConfig() *Config {
	hostname, _ := os.Hostname()
	c := &Config{
		Hostname: hostname,
		Encoder:  NewAutomaticEventEncoder(),
		Client:   &http.Client{},
	}
	return c
}
