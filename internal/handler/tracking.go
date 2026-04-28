package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/tracking"
)

const unsubscribeErrorPage = `<html><body><h1>Invalid unsubscribe link</h1><p>This link is invalid or has already been used.</p></body></html>`
const unsubscribeSuccessPage = `<html><body><h1>You have been unsubscribed</h1><p>You will no longer receive emails from this sender.</p></body></html>`

type TrackingHandler struct {
	db        *db.Queries
	appSecret string
}

func NewTrackingHandler(q *db.Queries, appSecret string) *TrackingHandler {
	return &TrackingHandler{db: q, appSecret: appSecret}
}

// transparentGIF is a 1x1 transparent GIF — hardcoded, never changes.
var transparentGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
	0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff,
	0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

func (h *TrackingHandler) Open(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	subscriberID, campaignID, err := tracking.ParseOpenToken(token, h.appSecret)
	if err != nil {
		log.Printf("tracking/open: invalid token: %v", err)
		servePixel(w)
		return
	}

	if err := h.db.RecordOpen(r.Context(), db.RecordOpenParams{
		CampaignID:   campaignID,
		SubscriberID: subscriberID,
	}); err != nil {
		log.Printf("tracking/open: db error: %v", err)
	}

	servePixel(w)
}

func (h *TrackingHandler) Click(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	destination := r.URL.Query().Get("url")
	if destination == "" {
		http.Error(w, "missing url param", http.StatusBadRequest)
		return
	}

	// Decode in case it was double-encoded, but QueryEscape on the sender side
	// already handles standard URL encoding; just use destination as-is.
	decoded, err := url.QueryUnescape(destination)
	if err != nil {
		decoded = destination
	}

	subscriberID, campaignID, linkIndex, err := tracking.ParseClickToken(token, h.appSecret)
	if err != nil {
		log.Printf("tracking/click: invalid token: %v", err)
		http.Redirect(w, r, decoded, http.StatusFound)
		return
	}

	if _, err := h.db.RecordClick(r.Context(), db.RecordClickParams{
		CampaignID:   campaignID,
		SubscriberID: subscriberID,
		LinkIndex:    int32(linkIndex),
		LinkUrl:      decoded,
	}); err != nil {
		log.Printf("tracking/click: db error: %v", err)
	}

	http.Redirect(w, r, decoded, http.StatusFound)
}

func (h *TrackingHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	subscriberID, _, err := tracking.ParseUnsubscribeToken(token, h.appSecret)
	if err != nil {
		log.Printf("unsubscribe: invalid token: %v", err)
		serveHTML(w, http.StatusBadRequest, unsubscribeErrorPage)
		return
	}

	sub, err := h.db.GetSubscriberByID(r.Context(), subscriberID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			serveHTML(w, http.StatusBadRequest, unsubscribeErrorPage)
			return
		}
		log.Printf("unsubscribe: db get subscriber: %v", err)
		serveHTML(w, http.StatusBadRequest, unsubscribeErrorPage)
		return
	}

	if sub.Status == "unsubscribed" || sub.Status == "bounced" {
		serveHTML(w, http.StatusOK, unsubscribeSuccessPage)
		return
	}

	if err := h.db.UpdateSubscriberStatus(r.Context(), db.UpdateSubscriberStatusParams{
		ID:     subscriberID,
		Status: "unsubscribed",
	}); err != nil {
		log.Printf("unsubscribe: update status: %v", err)
	}

	if err := h.db.AddSuppression(r.Context(), db.AddSuppressionParams{
		Email:  sub.Email,
		Reason: "unsubscribed",
	}); err != nil {
		log.Printf("unsubscribe: add suppression: %v", err)
	}

	serveHTML(w, http.StatusOK, unsubscribeSuccessPage)
}

func servePixel(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(transparentGIF)
}

func serveHTML(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(body))
}
