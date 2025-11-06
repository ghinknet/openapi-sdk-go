package realName

import (
	"fmt"
	"net/http"

	v3 "github.com/ghinknet/openapi-sdk-go/v3"
	"github.com/ghinknet/openapi-sdk-go/v3/client"
)

// VerifyCNID verifies whether the provided CNID is valid
func VerifyCNID(c *client.Client, id string, name string) (ok bool, err error) {
	// Build payload
	payload := v3.MapString{
		"id":   id,
		"name": name,
	}

	// Send request
	result := c.Send(
		fmt.Sprintf("%s%s/cnid", v3.Endpoint, Endpoint),
		http.MethodPost,
		payload,
	).WithToken()
	if result.Err != nil {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to verify CNID, sender error: %s", result.Err.Error(),
		))
		return false, result.Err
	}

	// Check status code
	if !result.Ok() {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to verify CNID, upstream failed: code: %d, msg: %s", result.Code, result.Msg,
		))
		return false, fmt.Errorf("failed to verify CNID, upstream failed: code: %d, msg: %s", result.Code, result.Msg)
	}

	// Build verify result struct
	var Ok struct {
		Ok bool `json:"ok"`
	}

	// Unmarshal token data
	if err = result.Unmarshal(&Ok); err != nil {
		c.Logger.Error(nil, fmt.Sprintf(
			"failed to verify CNID, unmarshal error: %s", result.Err.Error(),
		))
		return false, err
	}

	return Ok.Ok, nil
}
