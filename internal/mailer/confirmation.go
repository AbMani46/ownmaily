package mailer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type ConfirmationMailer struct {
	db              *db.Queries
	store           *Store
	installationURL string
	appSecret       string
}

func NewConfirmationMailer(q *db.Queries, store *Store, installationURL, appSecret string) *ConfirmationMailer {
	return &ConfirmationMailer{db: q, store: store, installationURL: installationURL, appSecret: appSecret}
}

// GenerateToken builds a self-contained signed token encoding subscriberID:listID:unixTimestamp.
// Token format: base64url(message) + "." + base64url(HMAC-SHA256(message, secret))
func (m *ConfirmationMailer) GenerateToken(subscriberID, listID pgtype.UUID) string {
	message := uuidToString(subscriberID) + ":" + uuidToString(listID) + ":" + strconv.FormatInt(time.Now().Unix(), 10)
	sig := computeHMAC(message, m.appSecret)
	return base64.RawURLEncoding.EncodeToString([]byte(message)) + "." + base64.RawURLEncoding.EncodeToString(sig)
}

// ValidateToken checks the HMAC signature, 48-hour expiry, and that sid/lid match.
func (m *ConfirmationMailer) ValidateToken(token string, subscriberID, listID pgtype.UUID) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}

	msgBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	message := string(msgBytes)
	if !hmac.Equal(computeHMAC(message, m.appSecret), sigBytes) {
		return false
	}

	// message = subscriberID:listID:unixTimestamp (UUIDs contain hyphens, not colons)
	segs := strings.Split(message, ":")
	if len(segs) != 3 {
		return false
	}

	ts, err := strconv.ParseInt(segs[2], 10, 64)
	if err != nil {
		return false
	}
	if time.Now().Unix()-ts > 48*3600 {
		return false
	}

	return segs[0] == uuidToString(subscriberID) && segs[1] == uuidToString(listID)
}

// SendConfirmation emails a double opt-in link through the configured mailer,
// using the from address and installation URL saved in settings.
func (m *ConfirmationMailer) SendConfirmation(subscriberID, listID pgtype.UUID, email string) error {
	s, err := m.db.GetSettings(context.Background())
	if err != nil {
		return fmt.Errorf("confirmation: load settings: %w", err)
	}

	installationURL := m.installationURL
	if s.InstallationUrl != "" {
		installationURL = s.InstallationUrl
	}

	from := s.FromEmail
	if from == "" {
		from = "no-reply@ownmaily.local"
	} else if s.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.FromName, s.FromEmail)
	}

	token := m.GenerateToken(subscriberID, listID)
	confirmURL := fmt.Sprintf("%s/confirm?token=%s&sid=%s&lid=%s",
		installationURL, token, uuidToString(subscriberID), uuidToString(listID))

	msg := Message{
		To:       email,
		From:     from,
		ReplyTo:  s.ReplyTo,
		Subject:  "Please confirm your subscription",
		TextBody: "Please confirm your subscription by opening this link:\n\n" + confirmURL + "\n\nIf you did not sign up, you can ignore this email.",
		HTMLBody: fmt.Sprintf(`<p>Please confirm your subscription.</p><p><a href="%s">Confirm subscription</a></p><p>If you did not sign up, you can ignore this email.</p>`,
			html.EscapeString(confirmURL)),
	}
	if err := m.store.Get().Send(msg); err != nil {
		return fmt.Errorf("confirmation: send to %s: %w", email, err)
	}
	return nil
}

func computeHMAC(message, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func uuidToString(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
