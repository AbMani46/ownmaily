package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SESMailer struct {
	accessKeyID     string
	secretAccessKey string
	region          string
	client          *http.Client
}

func NewSESMailer(accessKeyID, secretAccessKey, region string) *SESMailer {
	return &SESMailer{
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		client:          &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SESMailer) Name() string { return "ses" }

func (s *SESMailer) Send(msg Message) error {
	type charset struct {
		Data    string `json:"Data"`
		Charset string `json:"Charset"`
	}
	type header struct {
		Name  string `json:"Name"`
		Value string `json:"Value"`
	}

	simple := map[string]any{
		"Subject": charset{Data: msg.Subject, Charset: "UTF-8"},
		"Body": map[string]any{
			"Html": charset{Data: msg.HTMLBody, Charset: "UTF-8"},
			"Text": charset{Data: msg.TextBody, Charset: "UTF-8"},
		},
	}
	if len(msg.Headers) > 0 {
		headers := make([]header, 0, len(msg.Headers))
		for k, v := range msg.Headers {
			headers = append(headers, header{Name: k, Value: v})
		}
		simple["Headers"] = headers
	}

	payload := map[string]any{
		"FromEmailAddress": msg.From,
		"Destination":      map[string]any{"ToAddresses": []string{msg.To}},
		"Content":          map[string]any{"Simple": simple},
	}

	if msg.ReplyTo != "" {
		payload["ReplyToAddresses"] = []string{msg.ReplyTo}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://email.%s.amazonaws.com/v2/email/outbound-emails", s.region)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if err := signRequest(req, s.accessKeyID, s.secretAccessKey, s.region, body); err != nil {
		return fmt.Errorf("ses: sign request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ses: %d %s", resp.StatusCode, string(respBody))
	}
	return nil
}
