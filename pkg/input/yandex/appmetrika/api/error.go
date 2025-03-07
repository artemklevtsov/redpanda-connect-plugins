package api

import (
	"fmt"
	"log/slog"
)

// APIError API error object.
// Source: https://yandex.ru/dev/metrika/doc/api2/management/concept/errors.html#errors__resp
type APIError struct {
	Code    int    `json:"code"`    // Code is the error code.
	Message string `json:"message"` // Message is a human-readable error message.
	Reasons []struct {
		ErrorType string `json:"error_type"` // ErrorType specifies the type of error.
		Message   string `json:"message"`    // Message is a more detailed error message.
		Location  string `json:"location"`   // Location specifies where the error occurred.
	} `json:"errors"` // Reasons is a list of error details.
}

// Error returns a string representation of the API error.
func (e APIError) Error() string {
	return fmt.Sprintf("Yandex.AppMetriika API error %d: %s", e.Code, e.Message)
}

// LogValue returns a slog.Value representation of the API error.
func (e *APIError) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("code", e.Code),
		slog.String("message", e.Message),
	)
}
