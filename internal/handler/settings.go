package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/AbMani46/ownmaily/internal/auth"
	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type SettingsHandler struct {
	db          *db.Queries
	mailerStore *mailer.Store
}

func NewSettingsHandler(queries *db.Queries, mailerStore *mailer.Store) *SettingsHandler {
	return &SettingsHandler{db: queries, mailerStore: mailerStore}
}

type settingsResponse struct {
	SiteName        string `json:"site_name"`
	InstallationURL string `json:"installation_url"`
	Timezone        string `json:"timezone"`
	PhysicalAddress string `json:"physical_address"`
	FromName        string `json:"from_name"`
	FromEmail       string `json:"from_email"`
	ReplyTo         string `json:"reply_to"`
	SmtpProvider    string `json:"smtp_provider"`
	SetupComplete   bool   `json:"setup_complete"`
}

func toSettingsResponse(s db.Setting) settingsResponse {
	return settingsResponse{
		SiteName:        s.SiteName,
		InstallationURL: s.InstallationUrl,
		Timezone:        s.Timezone,
		PhysicalAddress: s.PhysicalAddress,
		FromName:        s.FromName,
		FromEmail:       s.FromEmail,
		ReplyTo:         s.ReplyTo,
		SmtpProvider:    s.SmtpProvider,
		SetupComplete:   s.SetupComplete,
	}
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	s, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load settings")
		return
	}
	writeJSON(w, http.StatusOK, toSettingsResponse(s))
}

func (h *SettingsHandler) UpdateGeneral(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SiteName        string `json:"site_name"`
		InstallationURL string `json:"installation_url"`
		Timezone        string `json:"timezone"`
		PhysicalAddress string `json:"physical_address"`
		FromName        string `json:"from_name"`
		FromEmail       string `json:"from_email"`
		ReplyTo         string `json:"reply_to"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if body.FromEmail != "" && !strings.Contains(body.FromEmail, "@") {
		writeError(w, http.StatusBadRequest, "invalid_email", "from_email must contain @")
		return
	}

	current, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load settings")
		return
	}

	updated, err := h.db.UpdateSettings(r.Context(), db.UpdateSettingsParams{
		SiteName:        body.SiteName,
		InstallationUrl: body.InstallationURL,
		Timezone:        body.Timezone,
		PhysicalAddress: body.PhysicalAddress,
		FromName:        body.FromName,
		FromEmail:       body.FromEmail,
		ReplyTo:         body.ReplyTo,
		SmtpProvider:    current.SmtpProvider,
		SmtpCredentials: current.SmtpCredentials,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not update settings")
		return
	}
	writeJSON(w, http.StatusOK, toSettingsResponse(updated))
}

func (h *SettingsHandler) UpdateSMTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider    string          `json:"provider"`
		Credentials json.RawMessage `json:"credentials"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	validProviders := map[string]bool{"resend": true, "mailgun": true, "ses": true}
	if !validProviders[body.Provider] {
		writeError(w, http.StatusBadRequest, "invalid_provider", "provider must be one of: resend, mailgun, ses")
		return
	}

	credJSON := string(body.Credentials)
	m, err := mailer.NewMailer(body.Provider, credJSON)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_credentials", err.Error())
		return
	}

	if err := h.db.UpdateSMTPSettings(r.Context(), db.UpdateSMTPSettingsParams{
		SmtpProvider:    body.Provider,
		SmtpCredentials: []byte(credJSON),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not save SMTP settings")
		return
	}

	h.mailerStore.Set(m)

	writeJSON(w, http.StatusOK, map[string]string{
		"message":  "SMTP settings updated",
		"provider": body.Provider,
	})
}

func (h *SettingsHandler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	s, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load settings")
		return
	}
	if s.SmtpProvider == "" {
		writeError(w, http.StatusBadRequest, "no_smtp", "No SMTP provider configured")
		return
	}

	owner, err := h.db.GetOwner(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load owner")
		return
	}

	msg := mailer.Message{
		To:       owner.Email,
		From:     s.FromEmail,
		Subject:  "OwnMaily test email",
		TextBody: "This is a test email from OwnMaily. Your SMTP settings are working correctly.",
		HTMLBody:  "<p>This is a test email from OwnMaily. Your SMTP settings are working correctly.</p>",
	}
	if err := h.mailerStore.Get().Send(msg); err != nil {
		writeError(w, http.StatusInternalServerError, "send_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Test email sent to " + owner.Email})
}

func (h *SettingsHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	key, err := h.db.GetLatestAPIKey(r.Context())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"key_prefix": nil,
				"created_at": nil,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load API key")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key_prefix": key.KeyPrefix,
		"created_at": key.CreatedAt.Time,
	})
}

func (h *SettingsHandler) RegenerateAPIKey(w http.ResponseWriter, r *http.Request) {
	if err := h.db.DeleteAllAPIKeys(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not delete old keys")
		return
	}

	rawKey := auth.GenerateAPIKey()
	hash, err := auth.HashKey(rawKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not hash key")
		return
	}

	created, err := h.db.CreateAPIKey(r.Context(), db.CreateAPIKeyParams{
		KeyHash:   hash,
		KeyPrefix: auth.KeyPrefix(rawKey),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create key")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"key":        rawKey,
		"key_prefix": created.KeyPrefix,
		"created_at": created.CreatedAt.Time,
		"message":    "Save this key — it will not be shown again",
	})
}

func (h *SettingsHandler) ListSuppressions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	q := r.URL.Query().Get("q")
	offset := int32((page - 1) * perPage)

	total, err := h.db.CountSuppressions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not count suppressions")
		return
	}

	var suppressions []db.SuppressedEmail
	if q != "" {
		suppressions, err = h.db.SearchSuppressions(r.Context(), db.SearchSuppressionsParams{
			Email:  "%" + q + "%",
			Limit:  int32(perPage),
			Offset: offset,
		})
	} else {
		suppressions, err = h.db.ListSuppressions(r.Context(), db.ListSuppressionsParams{
			Limit:  int32(perPage),
			Offset: offset,
		})
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not list suppressions")
		return
	}
	if suppressions == nil {
		suppressions = []db.SuppressedEmail{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"suppressions": suppressions,
		"total":        total,
		"page":         page,
		"per_page":     perPage,
	})
}

func (h *SettingsHandler) AddSuppression(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	email := strings.ToLower(strings.TrimSpace(body.Email))
	if !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "invalid_email", "email must contain @")
		return
	}

	existing, err := h.db.IsSuppressed(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not check suppression")
		return
	}
	if existing {
		writeError(w, http.StatusConflict, "already_suppressed", "email is already suppressed")
		return
	}

	if err := h.db.AddSuppression(r.Context(), db.AddSuppressionParams{
		Email:  email,
		Reason: "manual",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not add suppression")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "suppressed"})
}

func (h *SettingsHandler) DeleteSuppression(w http.ResponseWriter, r *http.Request) {
	emailParam := chi.URLParam(r, "email")
	email, err := url.QueryUnescape(emailParam)
	if err != nil {
		email = emailParam
	}

	if err := h.db.DeleteSuppression(r.Context(), email); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not delete suppression")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SettingsHandler) ExportSuppressions(w http.ResponseWriter, r *http.Request) {
	suppressions, err := h.db.ListAllSuppressions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load suppressions")
		return
	}

	date := time.Now().Format("2006-01-02")
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="suppressions-%s.csv"`, date))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"email", "reason", "created_at"})
	for _, s := range suppressions {
		_ = cw.Write([]string{s.Email, s.Reason, s.CreatedAt.Time.Format(time.RFC3339)})
	}
	cw.Flush()
}
