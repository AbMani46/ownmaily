package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/worker"
)

type CampaignHandler struct {
	db              *db.Queries
	installationURL string
	appSecret       string
	worker          *worker.SendWorker
}

func NewCampaignHandler(q *db.Queries, installationURL, appSecret string, w *worker.SendWorker) *CampaignHandler {
	return &CampaignHandler{db: q, installationURL: installationURL, appSecret: appSecret, worker: w}
}

type campaignRequest struct {
	Name        string `json:"name"`
	Subject     string `json:"subject"`
	PreviewText string `json:"preview_text"`
	FromName    string `json:"from_name"`
	FromEmail   string `json:"from_email"`
	ReplyTo     string `json:"reply_to"`
	HtmlBody    string `json:"html_body"`
	TextBody    string `json:"text_body"`
	SendToType  string `json:"send_to_type"`
	SendToID    string `json:"send_to_id"`
}

func parseSendToID(raw string) (pgtype.UUID, error) {
	var id pgtype.UUID
	if raw == "" {
		return id, nil
	}
	return id, id.Scan(raw)
}

func (h *CampaignHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePage(r)
	offset := (page - 1) * perPage
	statusFilter := r.URL.Query().Get("status")

	var campaigns []db.Campaign
	var total int64
	var err error

	if statusFilter != "" {
		campaigns, err = h.db.ListCampaignsByStatus(r.Context(), db.ListCampaignsByStatusParams{
			Status: statusFilter,
			Limit:  perPage,
			Offset: offset,
		})
		if err == nil {
			total, err = h.db.CountCampaignsByStatus(r.Context(), statusFilter)
		}
	} else {
		campaigns, err = h.db.ListCampaigns(r.Context(), db.ListCampaignsParams{
			Limit:  perPage,
			Offset: offset,
		})
		if err == nil {
			total, err = h.db.CountCampaigns(r.Context())
		}
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if campaigns == nil {
		campaigns = []db.Campaign{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"campaigns": campaigns,
		"total":     total,
		"page":      page,
		"per_page":  perPage,
	})
}

func (h *CampaignHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body campaignRequest
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}
	if strings.TrimSpace(body.Subject) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "subject is required")
		return
	}
	if strings.TrimSpace(body.FromName) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "from_name is required")
		return
	}
	if !strings.Contains(body.FromEmail, "@") {
		writeError(w, http.StatusBadRequest, "bad_request", "from_email must contain @")
		return
	}
	if body.SendToType == "" {
		body.SendToType = "list"
	}
	if body.SendToType != "list" && body.SendToType != "tag" {
		writeError(w, http.StatusBadRequest, "bad_request", `send_to_type must be "list" or "tag"`)
		return
	}

	sendToID, err := parseSendToID(body.SendToID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid send_to_id")
		return
	}

	campaign, err := h.db.CreateCampaign(r.Context(), db.CreateCampaignParams{
		Name:        strings.TrimSpace(body.Name),
		Subject:     strings.TrimSpace(body.Subject),
		PreviewText: strings.TrimSpace(body.PreviewText),
		FromName:    strings.TrimSpace(body.FromName),
		FromEmail:   strings.TrimSpace(body.FromEmail),
		ReplyTo:     strings.TrimSpace(body.ReplyTo),
		HtmlBody:    body.HtmlBody,
		TextBody:    body.TextBody,
		SendToType:  body.SendToType,
		SendToID:    sendToID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusCreated, campaign)
}

func (h *CampaignHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	campaign, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	existing, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if existing.Status == "sending" || existing.Status == "sent" {
		writeError(w, http.StatusConflict, "campaign_locked", "Cannot edit a campaign that is sending or has been sent")
		return
	}

	var body campaignRequest
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}
	if strings.TrimSpace(body.Subject) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "subject is required")
		return
	}
	if strings.TrimSpace(body.FromName) == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "from_name is required")
		return
	}
	if !strings.Contains(body.FromEmail, "@") {
		writeError(w, http.StatusBadRequest, "bad_request", "from_email must contain @")
		return
	}
	if body.SendToType == "" {
		body.SendToType = "list"
	}
	if body.SendToType != "list" && body.SendToType != "tag" {
		writeError(w, http.StatusBadRequest, "bad_request", `send_to_type must be "list" or "tag"`)
		return
	}

	sendToID, err := parseSendToID(body.SendToID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid send_to_id")
		return
	}

	updated, err := h.db.UpdateCampaign(r.Context(), db.UpdateCampaignParams{
		ID:          id,
		Name:        strings.TrimSpace(body.Name),
		Subject:     strings.TrimSpace(body.Subject),
		PreviewText: strings.TrimSpace(body.PreviewText),
		FromName:    strings.TrimSpace(body.FromName),
		FromEmail:   strings.TrimSpace(body.FromEmail),
		ReplyTo:     strings.TrimSpace(body.ReplyTo),
		HtmlBody:    body.HtmlBody,
		TextBody:    body.TextBody,
		SendToType:  body.SendToType,
		SendToID:    sendToID,
		ScheduledAt: existing.ScheduledAt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *CampaignHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	existing, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if existing.Status == "sending" {
		writeError(w, http.StatusConflict, "campaign_locked", "Cannot delete a campaign that is currently sending")
		return
	}

	if err := h.db.DeleteCampaign(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CampaignHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	original, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	dupe, err := h.db.CreateCampaign(r.Context(), db.CreateCampaignParams{
		Name:        "Copy of " + original.Name,
		Subject:     original.Subject,
		PreviewText: original.PreviewText,
		FromName:    original.FromName,
		FromEmail:   original.FromEmail,
		ReplyTo:     original.ReplyTo,
		HtmlBody:    original.HtmlBody,
		TextBody:    original.TextBody,
		SendToType:  original.SendToType,
		SendToID:    original.SendToID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusCreated, dupe)
}

func (h *CampaignHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	existing, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if existing.Status != "draft" {
		writeError(w, http.StatusConflict, "invalid_status", "Only draft campaigns can be scheduled")
		return
	}

	var body struct {
		ScheduledAt string `json:"scheduled_at"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	t, err := time.Parse(time.RFC3339, body.ScheduledAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "scheduled_at must be in RFC3339 format")
		return
	}

	if !t.After(time.Now()) {
		writeError(w, http.StatusBadRequest, "invalid_time", "Scheduled time must be in the future")
		return
	}

	campaign, err := h.db.ScheduleCampaign(r.Context(), db.ScheduleCampaignParams{
		ID:          id,
		ScheduledAt: pgtype.Timestamptz{Time: t, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	existing, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if existing.Status != "scheduled" {
		writeError(w, http.StatusConflict, "invalid_status", "Only scheduled campaigns can be cancelled")
		return
	}

	campaign, err := h.db.CancelCampaign(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, campaign)
}

func (h *CampaignHandler) Send(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	campaign, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if campaign.Status != "draft" && campaign.Status != "scheduled" {
		writeError(w, http.StatusConflict, "invalid_status", "Campaign cannot be sent in its current status")
		return
	}

	if !campaign.SendToID.Valid {
		writeError(w, http.StatusBadRequest, "missing_target", "Campaign has no send target. Set a list or tag before sending.")
		return
	}

	if strings.TrimSpace(campaign.Subject) == "" || strings.TrimSpace(campaign.FromEmail) == "" {
		writeError(w, http.StatusBadRequest, "missing_fields", "Campaign subject and from_email are required before sending")
		return
	}

	recipients, err := worker.ExpandRecipients(r.Context(), h.db, campaign)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if len(recipients) == 0 {
		writeError(w, http.StatusBadRequest, "no_recipients", "No active subscribers found for this campaign's target")
		return
	}

	if _, err := h.db.CreateSendJob(r.Context(), campaign.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.UpdateCampaignStatus(r.Context(), db.UpdateCampaignStatusParams{
		ID:     id,
		Status: "queued",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	log.Printf("campaign %s queued for sending (%d recipients)", campaign.ID.String(), len(recipients))

	writeJSON(w, http.StatusAccepted, map[string]any{
		"message":     "Campaign queued for sending",
		"campaign_id": campaign.ID,
	})
}

func (h *CampaignHandler) Preview(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	campaign, err := h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(campaign.HtmlBody))
}

func (h *CampaignHandler) Stats(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	_, err = h.db.GetCampaignByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "campaign not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	sent, err := h.db.CountRecipientsByStatus(r.Context(), db.CountRecipientsByStatusParams{
		CampaignID: id,
		Status:     "sent",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	failed, err := h.db.CountRecipientsByStatus(r.Context(), db.CountRecipientsByStatusParams{
		CampaignID: id,
		Status:     "failed",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	opens, err := h.db.CountOpensByCampaign(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	clicks, err := h.db.CountClicksByCampaign(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	links, err := h.db.CountClicksByLink(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if links == nil {
		links = []db.CountClicksByLinkRow{}
	}

	var openRate, clickRate float64
	if sent > 0 {
		openRate = float64(opens) / float64(sent)
		clickRate = float64(clicks) / float64(sent)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"campaign_id": id,
		"sent":        sent,
		"failed":      failed,
		"opens":       opens,
		"clicks":      clicks,
		"open_rate":   openRate,
		"click_rate":  clickRate,
		"links":       links,
	})
}
