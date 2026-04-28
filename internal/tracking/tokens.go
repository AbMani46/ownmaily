package tracking

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// GenerateOpenToken encodes subscriberID:campaignID into an HMAC-signed token.
// Token format: base64url(payload) + "." + base64url(HMAC-SHA256(message, appSecret))
func GenerateOpenToken(subscriberID, campaignID pgtype.UUID, appSecret string) string {
	payload := uuidStr(subscriberID) + ":" + uuidStr(campaignID)
	message := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(computeHMAC(message, appSecret))
	return message + "." + sig
}

// ParseOpenToken verifies the signature and returns the encoded subscriber and campaign IDs.
func ParseOpenToken(token, appSecret string) (subscriberID, campaignID pgtype.UUID, err error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return subscriberID, campaignID, errors.New("invalid token format")
	}
	message, sigPart := parts[0], parts[1]

	sigBytes, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return subscriberID, campaignID, errors.New("invalid signature encoding")
	}
	if !hmac.Equal(computeHMAC(message, appSecret), sigBytes) {
		return subscriberID, campaignID, errors.New("signature mismatch")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(message)
	if err != nil {
		return subscriberID, campaignID, errors.New("invalid message encoding")
	}

	segs := strings.SplitN(string(payloadBytes), ":", 2)
	if len(segs) != 2 {
		return subscriberID, campaignID, errors.New("invalid payload format")
	}

	if err := subscriberID.Scan(segs[0]); err != nil {
		return subscriberID, campaignID, fmt.Errorf("invalid subscriber id: %w", err)
	}
	if err := campaignID.Scan(segs[1]); err != nil {
		return subscriberID, campaignID, fmt.Errorf("invalid campaign id: %w", err)
	}

	return subscriberID, campaignID, nil
}

// GenerateClickToken encodes subscriberID:campaignID:linkIndex into an HMAC-signed token.
func GenerateClickToken(subscriberID, campaignID pgtype.UUID, linkIndex int, appSecret string) string {
	payload := uuidStr(subscriberID) + ":" + uuidStr(campaignID) + ":" + strconv.Itoa(linkIndex)
	message := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(computeHMAC(message, appSecret))
	return message + "." + sig
}

// ParseClickToken verifies the signature and returns the encoded subscriber ID, campaign ID, and link index.
func ParseClickToken(token, appSecret string) (subscriberID, campaignID pgtype.UUID, linkIndex int, err error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return subscriberID, campaignID, 0, errors.New("invalid token format")
	}
	message, sigPart := parts[0], parts[1]

	sigBytes, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return subscriberID, campaignID, 0, errors.New("invalid signature encoding")
	}
	if !hmac.Equal(computeHMAC(message, appSecret), sigBytes) {
		return subscriberID, campaignID, 0, errors.New("signature mismatch")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(message)
	if err != nil {
		return subscriberID, campaignID, 0, errors.New("invalid message encoding")
	}

	segs := strings.SplitN(string(payloadBytes), ":", 3)
	if len(segs) != 3 {
		return subscriberID, campaignID, 0, errors.New("invalid payload format")
	}

	if err := subscriberID.Scan(segs[0]); err != nil {
		return subscriberID, campaignID, 0, fmt.Errorf("invalid subscriber id: %w", err)
	}
	if err := campaignID.Scan(segs[1]); err != nil {
		return subscriberID, campaignID, 0, fmt.Errorf("invalid campaign id: %w", err)
	}
	linkIndex, err = strconv.Atoi(segs[2])
	if err != nil {
		return subscriberID, campaignID, 0, fmt.Errorf("invalid link index: %w", err)
	}

	return subscriberID, campaignID, linkIndex, nil
}

// GenerateUnsubscribeToken encodes subscriberID:campaignID into an HMAC-signed token.
// No expiry — must work forever since emails sit in inboxes for years.
func GenerateUnsubscribeToken(subscriberID, campaignID pgtype.UUID, appSecret string) string {
	payload := uuidStr(subscriberID) + ":" + uuidStr(campaignID)
	message := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(computeHMAC(message, appSecret))
	return message + "." + sig
}

// ParseUnsubscribeToken verifies the signature and returns the encoded subscriber and campaign IDs.
func ParseUnsubscribeToken(token, appSecret string) (subscriberID, campaignID pgtype.UUID, err error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return subscriberID, campaignID, errors.New("invalid token format")
	}
	message, sigPart := parts[0], parts[1]

	sigBytes, err := base64.RawURLEncoding.DecodeString(sigPart)
	if err != nil {
		return subscriberID, campaignID, errors.New("invalid signature encoding")
	}
	if !hmac.Equal(computeHMAC(message, appSecret), sigBytes) {
		return subscriberID, campaignID, errors.New("signature mismatch")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(message)
	if err != nil {
		return subscriberID, campaignID, errors.New("invalid message encoding")
	}

	segs := strings.SplitN(string(payloadBytes), ":", 2)
	if len(segs) != 2 {
		return subscriberID, campaignID, errors.New("invalid payload format")
	}

	if err := subscriberID.Scan(segs[0]); err != nil {
		return subscriberID, campaignID, fmt.Errorf("invalid subscriber id: %w", err)
	}
	if err := campaignID.Scan(segs[1]); err != nil {
		return subscriberID, campaignID, fmt.Errorf("invalid campaign id: %w", err)
	}

	return subscriberID, campaignID, nil
}

func computeHMAC(message, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func uuidStr(id pgtype.UUID) string {
	b := id.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
