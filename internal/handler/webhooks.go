package handler

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type WebhookHandler struct {
	db *db.Queries
}

func NewWebhookHandler(q *db.Queries) *WebhookHandler {
	return &WebhookHandler{db: q}
}

// Resend handles POST /webhooks/resend.
// Always returns 200 to prevent Resend from retrying.
//
// TODO: verify Resend webhook signature using svix header
// Header: webhook-signature, webhook-id, webhook-timestamp
// See https://docs.resend.com/webhooks for verification docs
func (h *WebhookHandler) Resend(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("webhook/resend: read body: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	event, err := mailer.ParseResendBounce(body)
	if err != nil {
		log.Printf("webhook/resend: parse: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}
	if event == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := context.Background()

	switch event.Type {
	case "hard":
		h.handleHardBounce(ctx, event)
	case "soft":
		log.Printf("webhook/resend: soft bounce for %s — logged, no suppression", event.Email)
	}

	w.WriteHeader(http.StatusOK)
}

// Mailgun handles POST /webhooks/mailgun.
// Always returns 200 to prevent Mailgun from retrying.
//
// TODO: verify Mailgun webhook signature
// Fields: timestamp, token, signature (in form body)
// HMAC-SHA256(timestamp+token, mailgun_webhook_signing_key)
func (h *WebhookHandler) Mailgun(w http.ResponseWriter, r *http.Request) {
	event, err := mailer.ParseMailgunBounce(r)
	if err != nil {
		log.Printf("webhook/mailgun: parse: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}
	if event == nil {
		log.Printf("webhook/mailgun: unhandled event type — ignored")
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := context.Background()

	switch event.Type {
	case "hard":
		h.handleHardBounce(ctx, event)
	case "soft":
		log.Printf("webhook/mailgun: soft bounce for %s — logged, no suppression", event.Email)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebhookHandler) handleHardBounce(ctx context.Context, event *mailer.BounceEvent) {
	if err := h.db.AddSuppression(ctx, db.AddSuppressionParams{
		Email:  event.Email,
		Reason: event.Reason,
	}); err != nil {
		log.Printf("webhook/resend: add suppression for %s: %v", event.Email, err)
	}

	sub, err := h.db.GetSubscriberByEmail(ctx, event.Email)
	if err != nil {
		log.Printf("webhook/resend: subscriber %s not found, skipping status update", event.Email)
		return
	}

	if err := h.db.UpdateSubscriberStatus(ctx, db.UpdateSubscriberStatusParams{
		ID:     sub.ID,
		Status: "bounced",
	}); err != nil {
		log.Printf("webhook/resend: update subscriber status for %s: %v", event.Email, err)
	}
}
