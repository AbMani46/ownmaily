package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AbMani46/ownmaily/internal/auth"
	"github.com/AbMani46/ownmaily/internal/middleware"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/jackc/pgx/v5"
)

type AuthHandler struct {
	queries *db.Queries
	secret  string
	baseURL string
}

func NewAuthHandler(queries *db.Queries, secret, baseURL string) *AuthHandler {
	return &AuthHandler{queries: queries, secret: secret, baseURL: baseURL}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	owner, err := h.queries.GetOwner(r.Context())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if !auth.CheckPassword(body.Password, owner.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(owner.Email, h.secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "could not generate token")
		return
	}

	secure := strings.HasPrefix(h.baseURL, "https://")
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		Path:     "/",
	})

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Path:     "/",
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*auth.Claims)
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing claims")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"email": claims.Subject})
}
