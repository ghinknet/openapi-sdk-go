package client

import (
	"fmt"
	"net/http"

	"github.com/ghinknet/json"
	v3 "github.com/ghinknet/openapi-sdk-go/v3"
)

// Client provides basic struct for client object
type Client struct {
	endpoint           string
	secretID           string
	secretKey          string
	enableToken        bool
	token              string
	timeout            int
	maxRetries         int
	retryDelay         int
	exponentialBackoff bool
	marshal            func(any) ([]byte, error)
	unmarshal          func([]byte, any) error
	Logger             Logger
}

// Option provides a basic option type
type Option func(*Client)

// WithLogger sets default logger to custom
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.Logger = logger
	}
}

// WithEndpoint sets default endpoint
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithMarshal sets default marshal lib
func WithMarshal(marshal func(any) ([]byte, error)) Option {
	return func(c *Client) {
		c.marshal = marshal
	}
}

// WithUnmarshal sets default unmarshal lib
func WithUnmarshal(unmarshal func([]byte, any) error) Option {
	return func(c *Client) {
		c.unmarshal = unmarshal
	}
}

// WithTimeout sets timeout for request
func WithTimeout(timeout int) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMaxRetries sets max retries for request
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay sets retry delay for request
func WithRetryDelay(retryDelay int) Option {
	return func(c *Client) {
		c.retryDelay = retryDelay
	}
}

// WithExponentialBackoff sets exponential backoff for request
func WithExponentialBackoff(exponentialBackoff bool) Option {
	return func(c *Client) {
		c.exponentialBackoff = exponentialBackoff
	}
}

// EnableToken enables token as authorization
func EnableToken(enableToken bool) Option {
	return func(c *Client) {
		c.enableToken = enableToken
	}
}

// GetEndpoint returns endpoint
func (c *Client) GetEndpoint() string {
	return c.endpoint
}

// applyToken applies a new token
func applyToken(c *Client) error {
	// Send request
	result := c.Send(
		fmt.Sprintf("%s/openAPI/token", c.endpoint),
		http.MethodGet,
		nil,
	).WithKey()
	if result.Err != nil {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to get token, sender error: %s", result.Err.Error(),
		))
		return result.Err
	}

	// Check status code
	if !result.OK() {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to get token, upstream failed: code: %d, msg: %s", result.Code, result.Msg,
		))
		return fmt.Errorf("failed to get token, upstream failed: code: %d, msg: %s", result.Code, result.Msg)
	}

	// Build token struct
	var token struct {
		Token string `json:"token"`
	}

	// Unmarshal token data
	if err := result.Unmarshal(&token); err != nil {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to get token, unmarshal error: %s", result.Err.Error(),
		))
		return err
	}

	// Save token
	c.token = token.Token
	return nil
}

// NewClient creates a new client to use service of Ghink Open API
func NewClient(secretID string, secretKey string, options ...Option) (*Client, error) {
	// Create client
	client := new(Client)

	// Load default logger
	client.Logger = NewLogger()

	// Load default endpoint
	client.endpoint = v3.Endpoint

	// Load default marshal and unmarshal lib
	client.marshal = json.Marshal
	client.unmarshal = json.Unmarshal

	// Load default maxRetries and retryDelay
	client.timeout = 3
	client.maxRetries = 5
	client.retryDelay = 1
	client.exponentialBackoff = true

	// Enable token in default
	client.enableToken = true

	// Load options
	for _, f := range options {
		f(client)
	}

	// Save keys
	client.secretID = secretID
	client.secretKey = secretKey

	// Try to get token
	if client.enableToken {
		if err := applyToken(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}
