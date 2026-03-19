package youtrack

import (
	"encoding/json"
	"fmt"
)

// APIError represents an HTTP error response from the YouTrack API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API %d: %s", e.StatusCode, e.Body)
}

// Message returns the human-readable error description if the body is a
// YouTrack JSON error response, otherwise returns the raw body.
func (e *APIError) Message() string {
	var parsed struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if json.Unmarshal([]byte(e.Body), &parsed) == nil && parsed.Description != "" {
		return parsed.Description
	}
	return e.Body
}
