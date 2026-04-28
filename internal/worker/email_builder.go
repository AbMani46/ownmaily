package worker

import (
	"fmt"
	"strings"

	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

func BuildMessage(campaign db.Campaign, subscriber db.Subscriber, installationURL, appSecret string) mailer.Message {
	// TODO Session 14: replace PLACEHOLDER with real signed unsubscribe token
	unsubURL := installationURL + "/unsubscribe?token=PLACEHOLDER"

	replaceVars := func(s string) string {
		s = strings.ReplaceAll(s, "{{unsubscribe_url}}", unsubURL)
		s = strings.ReplaceAll(s, "{{first_name}}", subscriber.FirstName)
		return s
	}

	return mailer.Message{
		To:       subscriber.Email,
		From:     fmt.Sprintf("%s <%s>", campaign.FromName, campaign.FromEmail),
		Subject:  campaign.Subject,
		HTMLBody: replaceVars(campaign.HtmlBody),
		TextBody: replaceVars(campaign.TextBody),
		ReplyTo:  campaign.ReplyTo,
		Headers: map[string]string{
			"List-Unsubscribe": "<" + unsubURL + ">",
		},
	}
}
