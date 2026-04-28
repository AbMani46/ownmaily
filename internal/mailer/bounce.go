package mailer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BounceEvent struct {
	Email     string
	Type      string // "hard" or "soft"
	Reason    string
	Timestamp time.Time
}

type resendWebhookPayload struct {
	Type string `json:"type"`
	Data struct {
		EmailID    string   `json:"email_id"`
		From       string   `json:"from"`
		To         []string `json:"to"`
		BounceType string   `json:"bounce_type"`
	} `json:"data"`
}

// ParseMailgunBounce parses a Mailgun form-encoded webhook request into a BounceEvent.
// Returns (nil, nil) for event types that don't map to a bounce or complaint.
func ParseMailgunBounce(r *http.Request) (*BounceEvent, error) {
	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("mailgun: parse form: %w", err)
	}

	event := r.FormValue("event")
	recipient := r.FormValue("recipient")

	switch event {
	case "bounced":
		severity := r.FormValue("severity")
		bounceType := "hard"
		if severity == "soft" {
			bounceType = "soft"
		}
		return &BounceEvent{
			Email:     recipient,
			Type:      bounceType,
			Reason:    "hard_bounce",
			Timestamp: time.Now(),
		}, nil

	case "complained":
		return &BounceEvent{
			Email:     recipient,
			Type:      "hard",
			Reason:    "complained",
			Timestamp: time.Now(),
		}, nil

	default:
		return nil, nil
	}
}

// ParseSESBounce parses an SNS Notification body from SES into BounceEvents.
// Returns nil, nil for SNS message types other than "Notification".
func ParseSESBounce(body []byte) ([]*BounceEvent, error) {
	var envelope struct {
		Type    string `json:"Type"`
		Message string `json:"Message"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("ses: parse sns envelope: %w", err)
	}

	if envelope.Type != "Notification" {
		return nil, nil
	}

	var notification struct {
		NotificationType string `json:"notificationType"`
		Bounce           struct {
			BounceType       string `json:"bounceType"`
			BouncedRecipients []struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"bouncedRecipients"`
		} `json:"bounce"`
		Complaint struct {
			ComplainedRecipients []struct {
				EmailAddress string `json:"emailAddress"`
			} `json:"complainedRecipients"`
		} `json:"complaint"`
	}
	if err := json.Unmarshal([]byte(envelope.Message), &notification); err != nil {
		return nil, fmt.Errorf("ses: parse notification message: %w", err)
	}

	var events []*BounceEvent
	switch notification.NotificationType {
	case "Bounce":
		bounceType := "hard"
		if notification.Bounce.BounceType == "Transient" {
			bounceType = "soft"
		}
		for _, r := range notification.Bounce.BouncedRecipients {
			events = append(events, &BounceEvent{
				Email:     r.EmailAddress,
				Type:      bounceType,
				Reason:    "hard_bounce",
				Timestamp: time.Now(),
			})
		}
	case "Complaint":
		for _, r := range notification.Complaint.ComplainedRecipients {
			events = append(events, &BounceEvent{
				Email:     r.EmailAddress,
				Type:      "hard",
				Reason:    "complained",
				Timestamp: time.Now(),
			})
		}
	}

	return events, nil
}

// ParseResendBounce parses a Resend webhook body into a BounceEvent.
// Returns (nil, nil) for event types that don't map to a bounce.
func ParseResendBounce(body []byte) (*BounceEvent, error) {
	var payload resendWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse resend webhook: %w", err)
	}

	switch payload.Type {
	case "email.bounced":
		if len(payload.Data.To) == 0 {
			return nil, fmt.Errorf("resend bounce: no recipient in payload")
		}
		bounceType := payload.Data.BounceType
		if bounceType == "" {
			bounceType = "hard"
		}
		return &BounceEvent{
			Email:     payload.Data.To[0],
			Type:      bounceType,
			Reason:    "hard_bounce",
			Timestamp: time.Now(),
		}, nil

	case "email.complained":
		if len(payload.Data.To) == 0 {
			return nil, fmt.Errorf("resend complained: no recipient in payload")
		}
		return &BounceEvent{
			Email:     payload.Data.To[0],
			Type:      "hard",
			Reason:    "complained",
			Timestamp: time.Now(),
		}, nil

	default:
		return nil, nil
	}
}
