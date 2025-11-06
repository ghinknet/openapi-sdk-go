package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	v3 "github.com/ghinknet/openapi-sdk-go/v3"
)

// Client provides basic struct for client object
type Client struct {
	SecretID    string
	SecretKey   string
	enableToken bool
	token       string
	expiry      time.Time
	logger      Logger
	marshal     func(any) ([]byte, error)
	unmarshal   func([]byte, any) error
}

// Option provides a basic option type
type Option func(*Client)

// WithLogger sets default logger to custom
func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
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

// EnableToken enables token as authorization
func EnableToken(enableToken bool) Option {
	return func(c *Client) {
		c.enableToken = enableToken
	}
}

// NewClient creates a new client to use service of Ghink Open API
func NewClient(secretID string, secretKey string, options ...Option) (*Client, error) {
	// Create client
	client := new(Client)

	// Load default logger
	client.logger = NewLogger()

	// Load default marshal and unmarshal lib
	client.marshal = json.Marshal
	client.unmarshal = json.Unmarshal

	// Load options
	for _, f := range options {
		f(client)
	}

	// Save keys
	client.SecretID = secretID
	client.SecretKey = secretKey

	// Try to get token
	if client.enableToken {
		// Send request
		result := client.Send(
			fmt.Sprintf("%s/openAPI/token", v3.Endpoint),
			http.MethodGet,
			nil,
		).WithKey()
		if result.Err != nil {
			client.logger.Error(nil, fmt.Sprintf(
				"failed to get token, sender error: %s", result.Err.Error(),
			))
			return nil, result.Err
		}

		// Check status code
		if !result.Ok() {
			client.logger.Error(nil, fmt.Sprintf(
				"failed to get token, upstream failed: code: %d, msg: %s", result.Code, result.Msg,
			))
			return nil, fmt.Errorf("failed to get token, upstream failed: code: %d, msg: %s", result.Code, result.Msg)
		}

		// Build token struct
		var token struct {
			Token string `json:"token"`
		}

		// Unmarshal token data
		if err := result.Unmarshal(&token); err != nil {
			client.logger.Error(nil, fmt.Sprintf(
				"failed to get token, unmarshal error: %s", result.Err.Error(),
			))
			return nil, err
		}

		client.token = token.Token
	}

	return client, nil
}
