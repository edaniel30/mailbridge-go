package core

import (
	"errors"
	"fmt"
)

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("config error [%s]: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

// NewConfigFieldError creates a new ConfigError for a specific field.
func NewConfigFieldError(field, message string) error {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}

// ErrNotConnected is returned when an operation is attempted on a disconnected client
var ErrNotConnected = errors.New("client not connected")
