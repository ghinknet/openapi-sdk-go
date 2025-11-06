package client

import (
	"time"

	"github.com/bytedance/sonic"
)

// Client provides basic struct for client object
type Client struct {
	SecretID  string
	SecretKey string
	Token     string
	Expiry    time.Time
	Logger    Logger
	Marshal   func(any) ([]byte, error)
	Unmarshal func([]byte, any) error
}

// Option provides a basic option type
type Option func(*Client)

// WithLogger sets default logger to custom
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithMarshal sets default marshal lib
func WithMarshal(marshal func(any) ([]byte, error)) Option {
	return func(c *Client) {
		c.Marshal = marshal
	}
}

// WithUnmarshal sets default unmarshal lib
func WithUnmarshal(unmarshal func([]byte, any) error) Option {
	return func(c *Client) {
		c.Unmarshal = unmarshal
	}
}

// NewClient creates a new client to use service of Ghink Open API
func NewClient(secretID string, secretKey string, options ...Option) (*Client, error) {
	// Create client
	client := new(Client)

	// Load default logger
	client.Logger = NewLogger()

	// Load default marshal and unmarshal lib
	client.Marshal = sonic.Marshal
	client.Unmarshal = sonic.Unmarshal

	// Load options
	for _, f := range options {
		f(client)
	}

	return client, nil
}
