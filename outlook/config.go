package outlook

import (
	"fmt"

	"github.com/danielrivera/mailbridge-go/core"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

// Config holds the configuration for connecting to Microsoft Graph API (Outlook).
type Config struct {
	ClientID     string   // The application (client) ID from Microsoft Entra ID app registration
	ClientSecret string   // The client secret from Microsoft Entra ID app registration
	TenantID     string   // The directory (tenant) ID. Use "consumers" for personal Microsoft accounts, "organizations" for work/school accounts, "common" for both, or your specific tenant ID
	RedirectURL  string   // The redirect URL configured in Microsoft Entra ID app registration
	Scopes       []string // The Microsoft Graph API scopes (default: Mail.Read, Mail.ReadWrite, offline_access)
}

// Validate checks if the configuration is valid.
// Returns core.ConfigError if required fields are missing.
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return &core.ConfigError{Field: "ClientID", Message: "ClientID is required"}
	}
	if c.ClientSecret == "" {
		return &core.ConfigError{Field: "ClientSecret", Message: "ClientSecret is required"}
	}
	if c.TenantID == "" {
		return &core.ConfigError{Field: "TenantID", Message: "TenantID is required"}
	}
	if c.RedirectURL == "" {
		return &core.ConfigError{Field: "RedirectURL", Message: "RedirectURL is required"}
	}
	return nil
}

// ToOAuth2Config converts Config to oauth2.Config.
func (c *Config) ToOAuth2Config() *oauth2.Config {
	scopes := c.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}

	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Scopes:       scopes,
		Endpoint:     microsoft.AzureADEndpoint(c.TenantID),
	}
}

// DefaultScopes returns the default Microsoft Graph API scopes for email operations.
func DefaultScopes() []string {
	return []string{
		"Mail.Read",
		"Mail.ReadWrite",
		"offline_access",
	}
}

// String returns a string representation of the Config (hides sensitive data).
func (c *Config) String() string {
	return fmt.Sprintf("Config{ClientID: %s, TenantID: %s, RedirectURL: %s, Scopes: %v}",
		maskString(c.ClientID),
		c.TenantID,
		c.RedirectURL,
		c.Scopes,
	)
}

// maskString masks a string for logging (shows first 4 chars, rest as asterisks).
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}
