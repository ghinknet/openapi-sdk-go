package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Result provides a basic struct to return result
type Result struct {
	client *Client
	Code   int
	Msg    string
	Body   []byte
	Err    error
}

// Sender provides a basic struct to send request
type Sender struct {
	client  *Client
	request *http.Request
	err     error
}

// Send provides a sender to send request
func (c *Client) Send(url string, method string, payload any) *Sender {
	// Marshal payload
	jsonPayload, err := c.Marshal(payload)
	if err != nil {
		return &Sender{
			client: c,
			err:    err,
		}
	}

	// Build http request
	req, err := http.NewRequest(method, url, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return &Sender{
			client: c,
			err:    err,
		}
	}

	// Set content-type
	req.Header.Add("Content-Type", "application/json")

	// Return sender
	return &Sender{
		client:  c,
		request: req,
		err:     nil,
	}
}

// parse returns parsed body data
func (s *Sender) parse(body []byte) *Result {
	var result struct {
		Code int    `json:"Code"`
		Msg  string `json:"Msg"`
		Data any    `json:"data"`
	}

	// Unmarshal body
	if err := s.client.Unmarshal(body, &result); err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}

	// Remarshal data part
	dataBody, err := s.client.Marshal(result.Data)
	if err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}

	// Return full result
	return &Result{
		client: s.client,
		Code:   result.Code,
		Msg:    result.Msg,
		Body:   dataBody,
	}
}

// WithToken sends a request with token to authorize
func (s *Sender) WithToken() *Result {
	// Handle error
	if s.err != nil {
		return &Result{
			client: s.client,
			Err:    s.err,
		}
	}

	// Construct client
	client := &http.Client{}

	// Add header
	s.request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.client.Token))

	// Send request
	res, err := client.Do(s.request)
	if err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	// Get request result
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}

	// Return parse result
	return s.parse(body)
}

// WithKey sends a request with SecretID and SecretKey to authorize
func (s *Sender) WithKey() *Result {
	// Handle error
	if s.err != nil {
		return &Result{
			client: s.client,
			Err:    s.err,
		}
	}

	// Construct client
	client := &http.Client{}

	// Add header
	s.request.Header.Add("Authorization", fmt.Sprintf("Basic %s:%s", s.client.SecretID, s.client.SecretKey))

	// Send request
	res, err := client.Do(s.request)
	if err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	// Get request result
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}

	// Return parse result
	return s.parse(body)
}

// Ok returns a bool value stands for the success or not of the request
func (r *Result) Ok() bool {
	return r.Code == 200
}

// Unmarshal can unmarshal a request data body to customized struct
func (r *Result) Unmarshal(v any) error {
	return r.client.Unmarshal(r.Body, v)
}
