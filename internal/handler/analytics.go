package handler

import (
	"net/http"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type AnalyticsHandler struct {
	db *db.Queries
}

func NewAnalyticsHandler(q *db.Queries) *AnalyticsHandler {
	return &AnalyticsHandler{db: q}
}

func (h *AnalyticsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	total, err := h.db.CountSubscribers(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	active, err := h.db.CountSubscribersByStatus(ctx, "active")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	unsubscribed, err := h.db.CountSubscribersByStatus(ctx, "unsubscribed")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	bounced, err := h.db.CountSubscribersByStatus(ctx, "bounced")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	campaignsSent, err := h.db.CountCampaignsByStatus(ctx, "sent")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	emailsSent, err := h.db.CountTotalSent(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	opens, err := h.db.CountTotalOpens(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	clicks, err := h.db.CountTotalClicks(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	var openRate, clickRate float64
	if emailsSent > 0 {
		openRate = float64(opens) / float64(emailsSent)
		clickRate = float64(clicks) / float64(emailsSent)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"total_subscribers":    total,
		"active_subscribers":   active,
		"unsubscribed":         unsubscribed,
		"bounced":              bounced,
		"total_campaigns_sent": campaignsSent,
		"total_emails_sent":    emailsSent,
		"total_opens":          opens,
		"total_clicks":         clicks,
		"overall_open_rate":    openRate,
		"overall_click_rate":   clickRate,
	})
}
