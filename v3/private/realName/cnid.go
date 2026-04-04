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
	runeNumber := []rune(idNumber)

	if len(runeNumber) != 18 {
		return false
	}

	for _, c := range runeNumber[:17] {
		if c < '0' || c > '9' {
			return false
		}
	}

	lastChar := strings.ToUpper(string(runeNumber[17]))
	if lastChar != "X" && (lastChar[0] < '0' || lastChar[0] > '9') {
		return false
	}

	year, err := strconv.Atoi(string(runeNumber[6:10]))
	if err != nil {
		return false
	}

	month, err := strconv.Atoi(string(runeNumber[10:12]))
	if err != nil || month < 1 || month > 12 {
		return false
	}

	day, err := strconv.Atoi(string(runeNumber[12:14]))
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
	for i, char := range runeNumber[:17] {
		num, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}
		total += num * factors[i]
	}

	correctChecksum := checksumDict[total%11]

	return lastChar == correctChecksum
}

// IsValidDate checks whether the date is a valid date
func IsValidDate(year, month, day int) bool {
	_, err := time.Parse("2006-01-02", fmt.Sprintf("%04d-%02d-%02d", year, month, day))
	return err == nil
}

// VerifyCNID verifies whether the provided CNID is valid
func VerifyCNID(c *client.Client, id string, name string) (ok bool, err error) {
	// Pre-process ID
	id = strings.ToLower(id)

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
		strings.Join([]string{c.GetEndpoint(), Endpoint, "/cnid"}, ""),
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
