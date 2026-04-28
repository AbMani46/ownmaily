package mailer

import (
	"encoding/json"
	"fmt"
)

// NewMailer constructs a Mailer from provider name and raw JSON credentials.
// Unknown or empty providers fall back to LogMailer so startup never fails.
func NewMailer(provider, credentials string) (Mailer, error) {
	switch provider {
	case "resend":
		var creds struct {
			APIKey string `json:"api_key"`
		}
		if err := json.Unmarshal([]byte(credentials), &creds); err != nil {
			return nil, fmt.Errorf("resend: parse credentials: %w", err)
		}
		if creds.APIKey == "" {
			return nil, fmt.Errorf("resend: api_key is empty")
		}
		return NewResendMailer(creds.APIKey), nil

	case "mailgun":
		var creds struct {
			APIKey string `json:"api_key"`
			Domain string `json:"domain"`
		}
		if err := json.Unmarshal([]byte(credentials), &creds); err != nil {
			return nil, fmt.Errorf("mailgun: invalid credentials: %w", err)
		}
		if creds.APIKey == "" || creds.Domain == "" {
			return nil, fmt.Errorf("mailgun: api_key and domain are required")
		}
		return NewMailgunMailer(creds.APIKey, creds.Domain), nil

	case "ses":
		return nil, fmt.Errorf("provider not yet implemented")

	default:
		return &LogMailer{}, nil
	}
}
