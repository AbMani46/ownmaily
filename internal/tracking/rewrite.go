package tracking

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/net/html"
)

// RewriteLinks walks the HTML, replaces qualifying <a href> values with click
// tracking URLs, and returns the rewritten HTML plus the original URLs in
// link-index order.
func RewriteLinks(htmlBody, installationURL, appSecret string, subscriberID, campaignID pgtype.UUID) (string, []string) {
	isFullDoc := strings.Contains(strings.ToLower(htmlBody), "<html")

	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return htmlBody, nil
	}

	var links []string
	idx := 0

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for i, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					if !skipLink(href) {
						token := GenerateClickToken(subscriberID, campaignID, idx, appSecret)
						n.Attr[i].Val = fmt.Sprintf("%s/track/click/%s?url=%s",
							installationURL, token, url.QueryEscape(href))
						links = append(links, href)
						idx++
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if isFullDoc {
		var buf strings.Builder
		if err := html.Render(&buf, doc); err != nil {
			return htmlBody, nil
		}
		return buf.String(), links
	}

	// Fragment: reconstruct from body children only to avoid injected html/head tags.
	body := findBodyNode(doc)
	if body == nil {
		return htmlBody, nil
	}
	var buf strings.Builder
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			return htmlBody, nil
		}
	}
	return buf.String(), links
}

func skipLink(href string) bool {
	return strings.HasPrefix(href, "#") ||
		strings.HasPrefix(href, "mailto:") ||
		strings.HasPrefix(href, "tel:") ||
		strings.Contains(href, "/unsubscribe")
}

func findBodyNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findBodyNode(c); found != nil {
			return found
		}
	}
	return nil
}
