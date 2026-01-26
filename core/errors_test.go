package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigError(t *testing.T) {
	t.Run("with field", func(t *testing.T) {
		err := &ConfigError{
			Field:   "client_id",
			Message: "is required",
		}

		assert.Equal(t, "config error [client_id]: is required", err.Error())
	})

	t.Run("without field", func(t *testing.T) {
		err := &ConfigError{
			Message: "invalid configuration",
		}

		assert.Equal(t, "config error: invalid configuration", err.Error())
	})
}

func TestNewConfigFieldError(t *testing.T) {
	err := NewConfigFieldError("redirect_url", "cannot be empty")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "redirect_url")
	assert.Contains(t, err.Error(), "cannot be empty")

	// Should be castable to ConfigError
	var configErr *ConfigError
	assert.ErrorAs(t, err, &configErr)
	assert.Equal(t, "redirect_url", configErr.Field)
}
