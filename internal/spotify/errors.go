package spotify

import (
	"fmt"
	"time"
)

type APIError struct {
	StatusCode int           `json:"status_code,omitempty"`
	Code       string        `json:"error"`
	Message    string        `json:"message"`
	RetryAfter time.Duration `json:"-"`
}

func (e *APIError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("%s: %s (retry after %s)", e.Code, e.Message, e.RetryAfter)
	}
	return e.Code + ": " + e.Message
}
