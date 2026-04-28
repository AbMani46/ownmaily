package mailer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type ConfirmationMailer struct {
	db              *db.Queries
	installationURL string
	appSecret       string
}

func NewConfirmationMailer(q *db.Queries, installationURL, appSecret string) *ConfirmationMailer {
	return &ConfirmationMailer{db: q, installationURL: installationURL, appSecret: appSecret}
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

func (m *ConfirmationMailer) SendConfirmation(subscriberID, listID pgtype.UUID, email string) error {
	token := m.GenerateToken(subscriberID, listID)
	confirmURL := fmt.Sprintf("%s/confirm?token=%s&sid=%s&lid=%s",
		m.installationURL, token, uuidToString(subscriberID), uuidToString(listID))

	// TODO Session 9: replace with real SMTP send
	log.Printf("CONFIRMATION EMAIL to %s: %s", email, confirmURL)
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
