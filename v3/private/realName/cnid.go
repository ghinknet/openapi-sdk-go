package realName

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	v3 "github.com/ghinknet/openapi-sdk-go/v3"
	"github.com/ghinknet/openapi-sdk-go/v3/client"
)

// IsValidID checks whether the ID is a valid Chinese Mainland ID
func IsValidID(idNumber string) bool {
	if len(idNumber) != 18 {
		return false
	}

	for _, c := range idNumber[:17] {
		if c < '0' || c > '9' {
			return false
		}
	}

	lastChar := strings.ToUpper(string(idNumber[17]))
	if lastChar != "X" && (lastChar[0] < '0' || lastChar[0] > '9') {
		return false
	}

	year, err := strconv.Atoi(idNumber[6:10])
	if err != nil {
		return false
	}

	month, err := strconv.Atoi(idNumber[10:12])
	if err != nil || month < 1 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(idNumber[12:14])
	if err != nil {
		return false
	}

	if !IsValidDate(year, month, day) {
		return false
	}

	factors := [17]int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checksumDict := map[int]string{
		0:  "1",
		1:  "0",
		2:  "X",
		3:  "9",
		4:  "8",
		5:  "7",
		6:  "6",
		7:  "5",
		8:  "4",
		9:  "3",
		10: "2",
	}

	total := 0
	for i, char := range idNumber[:17] {
		num, _ := strconv.Atoi(string(char))
		total += num * factors[i]
	}

	remainder := total % 11
	correctChecksum := checksumDict[remainder]

	return lastChar == correctChecksum
}

// IsValidDate checks whether the date is a valid date
func IsValidDate(year, month, day int) bool {
	_, err := time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", year, month, day))
	return err == nil
}

// VerifyCNID verifies whether the provided CNID is valid
func VerifyCNID(c *client.Client, id string, name string) (ok bool, err error) {
	// Check CNID format valid
	if !IsValidID(id) {
		return false, nil
	}

	// Build payload
	payload := v3.MapAny{
		"id":   id,
		"name": name,
	}

	// Send request
	result := c.Send(
		fmt.Sprintf("%s%s/cnid", c.GetEndpoint(), Endpoint),
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
	if !result.OK() {
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
