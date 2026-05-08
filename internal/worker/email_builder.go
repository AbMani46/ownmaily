package worker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/tracking"
)

// relativeImgSrc matches src attributes on img tags that point to /uploads/ paths.
var relativeImgSrc = regexp.MustCompile(`(<img\s[^>]*src=")(/uploads/[^"]+)(")`)

func BuildMessage(campaign db.Campaign, subscriber db.Subscriber, installationURL, appSecret string) mailer.Message {
	unsubToken := tracking.GenerateUnsubscribeToken(subscriber.ID, campaign.ID, appSecret)
	unsubURL := fmt.Sprintf("%s/unsubscribe?token=%s", installationURL, unsubToken)

	replaceVars := func(s string) string {
		s = strings.ReplaceAll(s, "{{unsubscribe_url}}", unsubURL)
		s = strings.ReplaceAll(s, "{{first_name}}", subscriber.FirstName)
		return s
	}

	htmlBody := replaceVars(campaign.HtmlBody)

	// Absolutize root-relative image src paths (/uploads/images/...) using the current
	// installationURL from settings. Stored paths are relative so changing the domain
	// doesn't break old campaigns.
	base := strings.TrimRight(installationURL, "/")
	htmlBody = relativeImgSrc.ReplaceAllString(htmlBody, "${1}"+base+"${2}${3}")

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
