package client

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v3 "github.com/ghinknet/openapi-sdk-go/v3"
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
		Code int    `json:"code"`
		Msg  string `json:"msg"`
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

	// Copy retry delay
	retryDelay := s.client.retryDelay

	for attempt := 0; attempt < s.client.maxRetries; attempt++ {
		if result := func() *Result {
			// Construct client
			client := &http.Client{
				Timeout: time.Duration(s.client.timeout) * time.Second,
			}

			// Add headers
			s.request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.client.token))
			s.request.Header.Add("User-Agent", v3.UserAgent)

			// Send request
			s.client.Logger.Debug(nil, fmt.Sprintf(
				"send request to %s, method %s with token (attempt %d)", s.request.URL, s.request.Method, attempt+1,
			))
			res, err := client.Do(s.request)
			if err != nil {
				s.client.Logger.Debug(nil, fmt.Sprintf("request failed: %v, retrying...", err))
				return nil // Retry on network errors
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(res.Body)

			// Handler http code error
			if res.StatusCode != http.StatusOK {
				s.client.Logger.Debug(nil, fmt.Sprintf("received HTTP status %d, retrying...", res.StatusCode))
				return nil // Retry on non-200 status codes
			}

			// Get request result
			body, err := io.ReadAll(res.Body)
			if err != nil {
				s.client.Logger.Debug(nil, fmt.Sprintf("failed to read response body: %v, retrying...", err))
				return nil // Retry on body read errors
			}

			// Parse result
			parsed := s.parse(body)

			// Output log
			var bodyRaw any
			if err = s.client.unmarshal(body, &bodyRaw); err != nil {
				s.client.Logger.Debug(nil, fmt.Sprintf("failed to unmarshal response body: %v, retrying...", err))
				return nil // Retry on unmarshal errors
			}
			s.client.Logger.Debug(nil, fmt.Sprintf(
				"openAPI response httpCode %d, apiCode %d, responseBody %s",
				res.StatusCode, parsed.Code, fmt.Sprint(bodyRaw),
			))

			// Check failed reason
			if parsed.Code == 801 {
				s.client.Logger.Debug(nil, "permission denied, maybe token expired, try to renew")

				// Sleep to prevent too many requests
				time.Sleep(time.Duration(retryDelay) * time.Second)

				if s.client.exponentialBackoff {
					retryDelay *= 2 // Exponential backoff
				}

				if err = applyToken(s.client); err != nil {
					return &Result{
						client: s.client,
						Err:    err,
					}
				}

				return nil // Retry after token renewal
			}

			// Return parsed result
			return parsed
		}(); result != nil {
			return result
		}

		// Wait before retrying
		if attempt < s.client.maxRetries-1 {
			s.client.Logger.Debug(nil, fmt.Sprintf("retrying in %v...", retryDelay))

			time.Sleep(time.Duration(retryDelay) * time.Second)

			if s.client.exponentialBackoff {
				retryDelay *= 2 // Exponential backoff
			}
		}
	}

	// If all retries failed, return an error
	return &Result{
		client: s.client,
		Err:    fmt.Errorf("request failed after %d retries", s.client.maxRetries),
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

	// Copy retry delay
	retryDelay := s.client.retryDelay

	for attempt := 0; attempt < s.client.maxRetries; attempt++ {
		if result := func() *Result {
			// Construct client
			client := &http.Client{
				Timeout: time.Duration(s.client.timeout) * time.Second,
			}

			// Add headers
			s.request.Header.Add("Authorization", fmt.Sprintf("Basic %s:%s", s.client.secretID, s.client.secretKey))
			s.request.Header.Add("User-Agent", v3.UserAgent)

			// Send request
			s.client.Logger.Debug(nil, fmt.Sprintf(
				"send request to %s, method %s with key (attempt %d)", s.request.URL, s.request.Method, attempt+1,
			))
			res, err := client.Do(s.request)
			if err != nil {
				s.client.Logger.Debug(nil, fmt.Sprintf("request failed: %v, retrying...", err))
				return nil // Retry on network errors
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(res.Body)

			// Handler http code error
			if res.StatusCode != http.StatusOK {
				s.client.Logger.Debug(nil, fmt.Sprintf("received HTTP status %d, retrying...", res.StatusCode))
				return nil // Retry on non-200 status codes
			}

			// Get request result
			body, err := io.ReadAll(res.Body)
			if err != nil {
				s.client.Logger.Debug(nil, fmt.Sprintf("failed to read response body: %v, retrying...", err))
				return nil // Retry on body read errors
			}

			// Parse result
			parsed := s.parse(body)

			// Output log
			var bodyRaw any
			if err = s.client.unmarshal(body, &bodyRaw); err != nil {
				s.client.Logger.Debug(nil, fmt.Sprintf("failed to unmarshal response body: %v, retrying...", err))
				return nil // Retry on unmarshal errors
			}
			s.client.Logger.Debug(nil, fmt.Sprintf(
				"openAPI response httpCode %d, apiCode %d, responseBody %s",
				res.StatusCode, parsed.Code, fmt.Sprint(bodyRaw),
			))

			// Check failed reason
			if parsed.Code == 801 {
				s.client.Logger.Debug(nil, "permission denied")

				// Sleep to prevent too many requests
				time.Sleep(time.Duration(retryDelay) * time.Second)

				if s.client.exponentialBackoff {
					retryDelay *= 2 // Exponential backoff
				}

				return nil // Retry after token renewal
			}

			// Return parsed result
			return parsed
		}(); result != nil {
			return result
		}

		// Wait before retrying
		if attempt < s.client.maxRetries-1 {
			s.client.Logger.Debug(nil, fmt.Sprintf("retrying in %v...", retryDelay))

			time.Sleep(time.Duration(retryDelay) * time.Second)

			if s.client.exponentialBackoff {
				retryDelay *= 2 // Exponential backoff
			}
		}
	}

	// If all retries failed, return an error
	return &Result{
		client: s.client,
		Err:    fmt.Errorf("request failed after %d retries", s.client.maxRetries),
	}
}

// OK returns a bool value stands for the success or not of the request
func (r *Result) OK() bool {
	return r.Code == 200
}

// Unmarshal can unmarshal a request data body to customized struct
func (r *Result) Unmarshal(v any) error {
	return r.client.unmarshal(r.Body, v)
}
