package mailer_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/AbMani46/ownmaily/internal/mailer"
	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

// recordingMailer records every message instead of sending it.
type recordingMailer struct{ sent []mailer.Message }

func (r *recordingMailer) Send(msg mailer.Message) error { r.sent = append(r.sent, msg); return nil }
func (r *recordingMailer) Name() string                  { return "recording" }

func TestSendConfirmation_SendsEmailFromSettings(t *testing.T) {
	_, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	ctx := context.Background()

	if _, err := db.UpdateSettings(ctx, sqlc.UpdateSettingsParams{
		SiteName:        "Test Site",
		InstallationUrl: "https://mail.example.com",
		Timezone:        "UTC",
		FromName:        "My Newsletter",
		FromEmail:       "hello@example.com",
		SmtpCredentials: []byte("{}"),
	}); err != nil {
		t.Fatalf("update settings: %v", err)
	}

	sub := testutil.CreateSubscriber(t, db, "reader@example.com")
	list := testutil.CreateList(t, db, "doi-list")

	rec := &recordingMailer{}
	cm := mailer.NewConfirmationMailer(db, mailer.NewStore(rec), "http://localhost:4400", testutil.TestAppSecret)

	if err := cm.SendConfirmation(sub.ID, list.ID, sub.Email); err != nil {
		t.Fatalf("SendConfirmation: %v", err)
	}

	if len(rec.sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(rec.sent))
	}
	msg := rec.sent[0]
	if msg.To != "reader@example.com" {
		t.Fatalf("expected To=reader@example.com, got %q", msg.To)
	}
	if msg.From != "My Newsletter <hello@example.com>" {
		t.Fatalf("expected From from settings, got %q", msg.From)
	}

	// The text body carries a working confirm link on the saved installation URL.
	var link string
	for _, line := range strings.Split(msg.TextBody, "\n") {
		if strings.HasPrefix(line, "https://mail.example.com/confirm?") {
			link = line
		}
	}
	if link == "" {
		t.Fatalf("expected confirm link in text body, got %q", msg.TextBody)
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse confirm link: %v", err)
	}
	if !cm.ValidateToken(u.Query().Get("token"), sub.ID, list.ID) {
		t.Fatal("expected confirm link token to validate")
	}
	if !strings.Contains(msg.HTMLBody, "https://mail.example.com/confirm?") {
		t.Fatalf("expected confirm link in HTML body, got %q", msg.HTMLBody)
	}
}
