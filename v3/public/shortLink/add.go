package shortLink

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.gh.ink/openapi/sdk/20260422/v3"
	"go.gh.ink/openapi/sdk/20260422/v3/client"
)

// Add a short link
func Add(c *client.Client, link string, validity *time.Time) (ok string, err error) {
	// Build payload
	payload := openapi.MapAny{
		"link":     link,
		"validity": validity.Unix(),
	}

	// Send request
	result := c.Send(
		strings.Join([]string{c.GetEndpoint(), Endpoint, "/add"}, ""),
		http.MethodPost,
		payload,
	).WithToken()
	if result.Err != nil {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to add short link, sender error: %s", result.Err.Error(),
		))
		return "", result.Err
	}

	// Check status code
	if !result.OK() {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to add short link, upstream failed: code: %d, msg: %s", result.Code, result.Msg,
		))
		return "", fmt.Errorf("failed to add short link, upstream failed: code: %d, msg: %s", result.Code, result.Msg)
	}

	// Build verify result struct
	var Link struct {
		LinkID string `json:"linkID"`
	}

	// Unmarshal token data
	if err = result.Unmarshal(&Link); err != nil {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to add short link, unmarshal error: %s", result.Err.Error(),
		))
		return "", err
	}

	return Link.LinkID, nil
}
