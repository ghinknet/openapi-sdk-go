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
	// Process payload
	var finalPayload io.Reader = nil
	if payload != nil {
		// Marshal payload
		jsonPayload, err := c.marshal(payload)
		if err != nil {
			return &Sender{
				client: c,
				err:    err,
			}
		}
		finalPayload = strings.NewReader(string(jsonPayload))
	}

	// Build http request
	req, err := http.NewRequest(method, url, finalPayload)
	if err != nil {
		return &Sender{
			client: c,
			err:    err,
		}
	}

	// Set content-type
	if method == http.MethodPost {
		req.Header.Add("Content-Type", "application/json")
	}

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

	// unmarshal body
	if err := s.client.unmarshal(body, &result); err != nil {
		return &Result{
			client: s.client,
			Err:    err,
		}
	}

	// Remarshal data part
	dataBody, err := s.client.marshal(result.Data)
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
	for {
		// Construct client
		client := &http.Client{}

		// Add header
		s.request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.client.token))

		// Send request
		s.client.Logger.Debug(nil, fmt.Sprintf(
			"send request to %s, method %s with token", s.request.URL, s.request.Method,
		))
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

		// Handler http code error
		if res.StatusCode != http.StatusOK {
			return &Result{
				client: s.client,
				Code:   res.StatusCode,
			}
		}

		// Get request result
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return &Result{
				client: s.client,
				Err:    err,
			}
		}

		// Parse result
		parsed := s.parse(body)

		// Output log
		var bodyRaw any
		err = s.client.unmarshal(body, &bodyRaw)
		if err != nil {
			return &Result{
				client: s.client,
				Err:    err,
			}
		}
		s.client.Logger.Debug(nil, fmt.Sprintf(
			"openAPI response httpCode %d, apiCode %d, responseBody %s",
			res.StatusCode, parsed.Code, fmt.Sprint(bodyRaw),
		))

		// Check failed reason
		if parsed.Code == 801 {
			s.client.Logger.Debug(nil, "token expired, try to renew")
			err = applyToken(s.client)
			if err != nil {
				return &Result{
					client: s.client,
					Err:    err,
				}
			}
			continue
		}

		// Return parsed result
		return parsed
	}
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

	for {
		// Construct client
		client := &http.Client{}

		// Add header
		s.request.Header.Add("Authorization", fmt.Sprintf("Basic %s:%s", s.client.SecretID, s.client.SecretKey))

		// Send request
		s.client.Logger.Debug(nil, fmt.Sprintf(
			"send request to %s, method %s with key", s.request.URL, s.request.Method,
		))
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

		// Handler http code error
		if res.StatusCode != http.StatusOK {
			return &Result{
				client: s.client,
				Code:   res.StatusCode,
			}
		}

		// Get request result
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return &Result{
				client: s.client,
				Err:    err,
			}
		}

		// Parse result
		parsed := s.parse(body)

		// Output log
		var bodyRaw any
		err = s.client.unmarshal(body, &bodyRaw)
		if err != nil {
			return &Result{
				client: s.client,
				Err:    err,
			}
		}
		s.client.Logger.Debug(nil, fmt.Sprintf(
			"openAPI response httpCode %d, apiCode %d, responseBody %s",
			res.StatusCode, parsed.Code, fmt.Sprint(bodyRaw),
		))

		// Check failed reason
		if parsed.Code == 801 {
			s.client.Logger.Debug(nil, "token expired, try to renew")
			err = applyToken(s.client)
			if err != nil {
				return &Result{
					client: s.client,
					Err:    err,
				}
			}
			continue
		}

		// Return parsed result
		return parsed
	}
}

// Ok returns a bool value stands for the success or not of the request
func (r *Result) Ok() bool {
	return r.Code == 200
}

// Unmarshal can unmarshal a request data body to customized struct
func (r *Result) Unmarshal(v any) error {
	return r.client.unmarshal(r.Body, v)
}
