package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/AbMani46/ownmaily/internal/auth"
	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type SetupHandler struct {
	db          *db.Queries
	secret      string
	mailerStore *mailer.Store
}

func NewSetupHandler(queries *db.Queries, secret string, mailerStore *mailer.Store) *SetupHandler {
	return &SetupHandler{db: queries, secret: secret, mailerStore: mailerStore}
}

// guardSetupIncomplete returns true (and writes 409) if setup is already complete.
func (h *SetupHandler) guardSetupIncomplete(w http.ResponseWriter, r *http.Request) bool {
	s, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load settings")
		return true
	}
	if s.SetupComplete {
		writeError(w, http.StatusConflict, "already_setup", "setup is already complete")
		return true
	}
	return false
}

// POST /api/setup/owner — create owner account (only if setup_complete = false)
func (h *SetupHandler) CreateOwner(w http.ResponseWriter, r *http.Request) {
	if h.guardSetupIncomplete(w, r) {
		return
	}

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if body.Email == "" || !strings.Contains(body.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid_email", "valid email required")
		return
	}
	if len(body.Password) < 8 {
		writeError(w, http.StatusBadRequest, "invalid_password", "password must be at least 8 characters")
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not hash password")
		return
	}

	owner, err := h.db.CreateOwner(r.Context(), db.CreateOwnerParams{
		Email:        body.Email,
		PasswordHash: hash,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not create owner")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"email": owner.Email})
}

// PUT /api/setup/settings — update general settings (only if setup_complete = false)
func (h *SetupHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if h.guardSetupIncomplete(w, r) {
		return
	}

	var body struct {
		SiteName        string `json:"site_name"`
		InstallationURL string `json:"installation_url"`
		Timezone        string `json:"timezone"`
		PhysicalAddress string `json:"physical_address"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	current, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load settings")
		return
	}

	_, err = h.db.UpdateSettings(r.Context(), db.UpdateSettingsParams{
		SiteName:        body.SiteName,
		InstallationUrl: body.InstallationURL,
		Timezone:        body.Timezone,
		PhysicalAddress: body.PhysicalAddress,
		FromName:        current.FromName,
		FromEmail:       current.FromEmail,
		ReplyTo:         current.ReplyTo,
		SmtpProvider:    current.SmtpProvider,
		SmtpCredentials: current.SmtpCredentials,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not save settings")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "settings saved"})
}

// PUT /api/setup/smtp — save SMTP credentials + hot-swap mailer (only if setup_complete = false)
func (h *SetupHandler) UpdateSMTP(w http.ResponseWriter, r *http.Request) {
	if h.guardSetupIncomplete(w, r) {
		return
	}

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
	writeJSON(w, http.StatusOK, map[string]string{"message": "SMTP configured", "provider": body.Provider})
}

// POST /api/setup/test-smtp — send a test email to provided address (only if setup_complete = false)
func (h *SetupHandler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	if h.guardSetupIncomplete(w, r) {
		return
	}

	var body struct {
		To string `json:"to"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}
	if body.To == "" || !strings.Contains(body.To, "@") {
		writeError(w, http.StatusBadRequest, "invalid_email", "valid recipient email required")
		return
	}

	s, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load settings")
		return
	}
	if s.SmtpProvider == "" {
		writeError(w, http.StatusBadRequest, "no_smtp", "no SMTP provider configured")
		return
	}

	from := s.FromEmail
	if from == "" {
		from = "no-reply@ownmaily.local"
	}

	msg := mailer.Message{
		To:       body.To,
		From:     from,
		Subject:  "OwnMaily setup test email",
		TextBody: "Your OwnMaily SMTP configuration is working correctly.",
		HTMLBody:  "<p>Your OwnMaily SMTP configuration is working correctly.</p>",
	}
	if err := h.mailerStore.Get().Send(msg); err != nil {
		writeError(w, http.StatusInternalServerError, "send_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "test email sent to " + body.To})
}

// GET /api/setup/status — public endpoint, returns {"complete": true/false}
func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	s, err := h.db.GetSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]bool{"complete": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"complete": s.SetupComplete})
}

// POST /api/setup/complete — mark setup done + auto-login (only if setup_complete = false)
func (h *SetupHandler) Complete(w http.ResponseWriter, r *http.Request) {
	if h.guardSetupIncomplete(w, r) {
		return
	}

	if err := h.db.SetSetupComplete(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not complete setup")
		return
	}

	owner, err := h.db.GetOwner(r.Context())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "no_owner", "no owner account found; complete step 2 first")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "could not load owner")
		return
	}

	token, err := auth.GenerateToken(owner.Email, h.secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

