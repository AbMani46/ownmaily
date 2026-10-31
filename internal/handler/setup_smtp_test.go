package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbMani46/ownmaily/internal/handler"
	"github.com/AbMani46/ownmaily/internal/mailer"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

// captureMailer records every message instead of sending it.
type captureMailer struct{ sent []mailer.Message }

func (c *captureMailer) Send(msg mailer.Message) error { c.sent = append(c.sent, msg); return nil }
func (c *captureMailer) Name() string                  { return "capture" }

func callSetupSMTP(t *testing.T, fn http.HandlerFunc, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	rec := httptest.NewRecorder()
	fn(rec, httptest.NewRequest(method, path, bytes.NewReader(b)))
	return rec
}

func TestSetupSMTP_SavesFromAddress(t *testing.T) {
	_, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)

	store := mailer.NewStore(&mailer.LogMailer{})
	h := handler.NewSetupHandler(db, testutil.TestAppSecret, store)

	rec := callSetupSMTP(t, h.UpdateSMTP, "PUT", "/api/setup/smtp", map[string]any{
		"provider":    "resend",
		"from_email":  "hello@example.com",
		"from_name":   "My Newsletter",
		"credentials": map[string]string{"api_key": "re_test"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	s, err := db.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if s.FromEmail != "hello@example.com" || s.FromName != "My Newsletter" {
		t.Fatalf("expected from saved, got email=%q name=%q", s.FromEmail, s.FromName)
	}
	if s.SmtpProvider != "resend" {
		t.Fatalf("expected smtp_provider=resend, got %q", s.SmtpProvider)
	}

	// Swap in a capturing mailer so the test email does not hit the network.
	capture := &captureMailer{}
	store.Set(capture)

	rec = callSetupSMTP(t, h.TestSMTP, "POST", "/api/setup/test-smtp", map[string]string{"to": "owner@example.com"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(capture.sent) != 1 {
		t.Fatalf("expected 1 test email, got %d", len(capture.sent))
	}
	if got := capture.sent[0].From; got != "My Newsletter <hello@example.com>" {
		t.Fatalf("expected test email from the saved address, got %q", got)
	}
}
