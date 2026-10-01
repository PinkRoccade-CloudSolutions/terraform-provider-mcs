package apiclient

import (
	"errors"
	"fmt"
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
