package shortLink

import (
	"fmt"
	"net/http"
	"time"

	v3 "github.com/ghinknet/openapi-sdk-go/v3"
	"github.com/ghinknet/openapi-sdk-go/v3/client"
)

// Add a short link
func Add(c *client.Client, link string, validity *time.Time) (ok string, err error) {
	// Build payload
	payload := v3.MapAny{
		"link":     link,
		"validity": validity.Unix(),
	}

	// Send request
	result := c.Send(
		fmt.Sprintf("%s%s/add", c.GetEndpoint(), Endpoint),
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
