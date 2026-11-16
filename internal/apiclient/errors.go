package apiclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int
	Body       string
	Endpoint   string
	Method     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("MCS API %s %s returned status %d: %s", e.Method, e.Endpoint, e.StatusCode, e.Body)
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

func IsConflict(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 409
	}
	return false
}

// Detail returns the `detail` message of an API error body, e.g. why MCS answered 409, or ""
// when err is not an APIError or its body carries no detail.
func Detail(err error) string {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return ""
	}
	var body struct {
		Detail string `json:"detail"`
	}
	if json.Unmarshal([]byte(apiErr.Body), &body) != nil {
		return ""
	}
	return strings.TrimSpace(body.Detail)
}
