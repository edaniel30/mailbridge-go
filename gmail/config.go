package gmail

import (
	"github.com/danielrivera/mailbridge-go/core"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Config holds Gmail provider configuration
type Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes,omitempty"`
}

// DefaultScopes returns the default Gmail API scopes
func DefaultScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/gmail.readonly",
		"https://www.googleapis.com/auth/gmail.send",
		"https://www.googleapis.com/auth/gmail.modify",
		"https://www.googleapis.com/auth/gmail.labels",
	}
}

// Validate validates the Gmail configuration
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return core.NewConfigFieldError("client_id", "is required")
	}
	if c.ClientSecret == "" {
		return core.NewConfigFieldError("client_secret", "is required")
	}
	if c.RedirectURL == "" {
		return core.NewConfigFieldError("redirect_url", "is required")
	}
	if len(c.Scopes) == 0 {
		c.Scopes = DefaultScopes()
	}
	return nil
}

// ToOAuth2Config converts Gmail config to oauth2.Config
func (c *Config) ToOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       c.Scopes,
		Endpoint:     google.Endpoint,
	}
}
