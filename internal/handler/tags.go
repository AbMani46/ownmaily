package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type TagHandler struct {
	db *db.Queries
}

func NewTagHandler(q *db.Queries) *TagHandler {
	return &TagHandler{db: q}
}

type tagResponse struct {
	db.Tag
	SubscriberCount int64 `json:"subscriber_count"`
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.db.ListTags(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if tags == nil {
		tags = []db.Tag{}
	}

	result := make([]tagResponse, 0, len(tags))
	for _, t := range tags {
		count, countErr := h.db.CountSubscribersWithTag(r.Context(), t.ID)
		if countErr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
		result = append(result, tagResponse{Tag: t, SubscriberCount: count})
	}

	writeJSON(w, http.StatusOK, map[string]any{"tags": result})
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	name := strings.TrimSpace(strings.ToLower(body.Name))
	if name == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	tag, err := h.db.CreateTag(r.Context(), name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "duplicate", "Tag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusCreated, tagResponse{Tag: tag, SubscriberCount: 0})
}

func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	tag, err := h.db.GetTagByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	count, err := h.db.CountSubscribersWithTag(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, tagResponse{Tag: tag, SubscriberCount: count})
}

func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	_, err = h.db.GetTagByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	name := strings.TrimSpace(strings.ToLower(body.Name))
	if name == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	updated, err := h.db.UpdateTag(r.Context(), db.UpdateTagParams{ID: id, Name: name})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "duplicate", "Tag name already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	count, err := h.db.CountSubscribersWithTag(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, tagResponse{Tag: updated, SubscriberCount: count})
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	_, err = h.db.GetTagByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.DeleteTag(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	_, err = h.db.GetTagByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	page, perPage := parsePage(r)
	offset := (page - 1) * perPage

	subs, err := h.db.ListSubscribersWithTag(r.Context(), db.ListSubscribersWithTagParams{
		TagID:  id,
		Limit:  perPage,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if subs == nil {
		subs = []db.Subscriber{}
	}

	total, err := h.db.CountSubscribersWithTag(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	result := make([]subscriberResponse, 0, len(subs))
	for _, sub := range subs {
		tags, tagErr := h.db.ListTagsForSubscriber(r.Context(), sub.ID)
		if tagErr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
		if tags == nil {
			tags = []db.Tag{}
		}
		result = append(result, subscriberResponse{Subscriber: sub, Tags: tags})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"subscribers": result,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
	})
}
