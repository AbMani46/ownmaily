package worker

import (
	"fmt"
	"strings"

	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/tracking"
)

func BuildMessage(campaign db.Campaign, subscriber db.Subscriber, installationURL, appSecret string) mailer.Message {
	unsubToken := tracking.GenerateUnsubscribeToken(subscriber.ID, campaign.ID, appSecret)
	unsubURL := fmt.Sprintf("%s/unsubscribe?token=%s", installationURL, unsubToken)

	replaceVars := func(s string) string {
		s = strings.ReplaceAll(s, "{{unsubscribe_url}}", unsubURL)
		s = strings.ReplaceAll(s, "{{first_name}}", subscriber.FirstName)
		return s
	}

	htmlBody := replaceVars(campaign.HtmlBody)

	rewritten, _ := tracking.RewriteLinks(htmlBody, installationURL, appSecret, subscriber.ID, campaign.ID)
	htmlBody = rewritten

	openToken := tracking.GenerateOpenToken(subscriber.ID, campaign.ID, appSecret)
	pixelURL := fmt.Sprintf("%s/track/open/%s", installationURL, openToken)
	pixelHTML := fmt.Sprintf(`<img src="%s" width="1" height="1" style="display:none" alt="" />`, pixelURL)

	if strings.Contains(htmlBody, "</body>") {
		htmlBody = strings.Replace(htmlBody, "</body>", pixelHTML+"</body>", 1)
	} else {
		htmlBody = htmlBody + pixelHTML
	}

	return mailer.Message{
		To:       subscriber.Email,
		From:     fmt.Sprintf("%s <%s>", campaign.FromName, campaign.FromEmail),
		Subject:  campaign.Subject,
		HTMLBody: htmlBody,
		TextBody: replaceVars(campaign.TextBody),
		ReplyTo:  campaign.ReplyTo,
		Headers: map[string]string{
			"List-Unsubscribe":      fmt.Sprintf("<%s>", unsubURL),
			"List-Unsubscribe-Post": "List-Unsubscribe=One-Click",
		},
	}
}
