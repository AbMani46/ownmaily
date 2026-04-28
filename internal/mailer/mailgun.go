package mailer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type MailgunMailer struct {
	apiKey string
	domain string
	client *http.Client
}

func NewMailgunMailer(apiKey, domain string) *MailgunMailer {
	return &MailgunMailer{
		apiKey: apiKey,
		domain: domain,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *MailgunMailer) Name() string { return "mailgun" }

func (m *MailgunMailer) Send(msg Message) error {
	form := url.Values{}
	form.Set("from", msg.From)
	form.Set("to", msg.To)
	form.Set("subject", msg.Subject)
	form.Set("html", msg.HTMLBody)
	form.Set("text", msg.TextBody)
	if msg.ReplyTo != "" {
		form.Set("h:Reply-To", msg.ReplyTo)
	}
	for k, v := range msg.Headers {
		form.Set("h:"+k, v)
	}

	endpoint := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", m.domain)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", m.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailgun: %d %s", resp.StatusCode, string(body))
	}
	return nil
}
