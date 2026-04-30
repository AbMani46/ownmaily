package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type ListHandler struct {
	db     *db.Queries
	mailer *mailer.ConfirmationMailer
}

func NewListHandler(q *db.Queries, m *mailer.ConfirmationMailer) *ListHandler {
	return &ListHandler{db: q, mailer: m}
}

type listResponse struct {
	db.List
	SubscriberCount int64 `json:"subscriber_count"`
}

func (h *ListHandler) List(w http.ResponseWriter, r *http.Request) {
	lists, err := h.db.ListLists(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if lists == nil {
		lists = []db.List{}
	}

	result := make([]listResponse, 0, len(lists))
	for _, l := range lists {
		count, countErr := h.db.CountSubscribersInList(r.Context(), l.ID)
		if countErr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
		result = append(result, listResponse{List: l, SubscriberCount: count})
	}

	writeJSON(w, http.StatusOK, map[string]any{"lists": result})
}

func (h *ListHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		DoubleOptIn bool   `json:"double_opt_in"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	list, err := h.db.CreateList(r.Context(), db.CreateListParams{
		Name:        name,
		Description: strings.TrimSpace(body.Description),
		DoubleOptIn: body.DoubleOptIn,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "duplicate", "List name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusCreated, listResponse{List: list, SubscriberCount: 0})
}

func (h *ListHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	list, err := h.db.GetListByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	count, err := h.db.CountSubscribersInList(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, listResponse{List: list, SubscriberCount: count})
}

func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	if _, err := h.db.GetListByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		DoubleOptIn bool   `json:"double_opt_in"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "name is required")
		return
	}

	list, err := h.db.UpdateList(r.Context(), db.UpdateListParams{
		ID:          id,
		Name:        name,
		Description: strings.TrimSpace(body.Description),
		DoubleOptIn: body.DoubleOptIn,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "duplicate", "List name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	count, err := h.db.CountSubscribersInList(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, listResponse{List: list, SubscriberCount: count})
}

func (h *ListHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	if _, err := h.db.GetListByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	count, err := h.db.CountSubscribersInList(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if count > 0 && r.URL.Query().Get("force") != "true" {
		writeError(w, http.StatusConflict, "list_not_empty",
			fmt.Sprintf("List has %d subscribers. Remove them before deleting or use force=true", count))
		return
	}

	if err := h.db.DeleteList(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	if _, err := h.db.GetListByID(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	page, perPage := parsePage(r)
	offset := (page - 1) * perPage

	subs, err := h.db.ListSubscribersInList(r.Context(), db.ListSubscribersInListParams{
		ListID: id,
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

	total, err := h.db.CountSubscribersInList(r.Context(), id)
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

func (h *ListHandler) AddSubscriber(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid list id")
		return
	}

	list, err := h.db.GetListByID(r.Context(), listID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	var body struct {
		SubscriberID string `json:"subscriber_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	subID, err := parseUUID(body.SubscriberID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid subscriber_id")
		return
	}

	sub, err := h.db.GetSubscriberByID(r.Context(), subID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if sub.Status == "unsubscribed" || sub.Status == "bounced" {
		writeError(w, http.StatusConflict, "subscriber_inactive", "Subscriber is unsubscribed or bounced")
		return
	}

	inList, err := h.db.IsSubscriberInList(r.Context(), db.IsSubscriberInListParams{
		ListID:       listID,
		SubscriberID: subID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if inList {
		writeError(w, http.StatusConflict, "already_member", "Subscriber is already in this list")
		return
	}

	if list.DoubleOptIn && sub.Status == "active" {
		if err := h.db.UpdateSubscriberStatus(r.Context(), db.UpdateSubscriberStatusParams{
			ID:     subID,
			Status: "pending",
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
		if err := h.mailer.SendConfirmation(subID, listID, sub.Email); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
	}

	if err := h.db.AddSubscriberToList(r.Context(), db.AddSubscriberToListParams{
		ListID:       listID,
		SubscriberID: subID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "added"})
}

func (h *ListHandler) RemoveSubscriber(w http.ResponseWriter, r *http.Request) {
	listID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid list id")
		return
	}

	if _, err := h.db.GetListByID(r.Context(), listID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	subID, err := parseUUID(chi.URLParam(r, "subscriberID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid subscriber id")
		return
	}

	if _, err := h.db.GetSubscriberByID(r.Context(), subID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.RemoveSubscriberFromList(r.Context(), db.RemoveSubscriberFromListParams{
		ListID:       listID,
		SubscriberID: subID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) ConfirmOptIn(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	sidStr := r.URL.Query().Get("sid")
	lidStr := r.URL.Query().Get("lid")

	if token == "" || sidStr == "" || lidStr == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "Confirmation link is invalid or expired")
		return
	}

	subID, err := parseUUID(sidStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", "Confirmation link is invalid or expired")
		return
	}

	listID, err := parseUUID(lidStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_token", "Confirmation link is invalid or expired")
		return
	}

	if !h.mailer.ValidateToken(token, subID, listID) {
		writeError(w, http.StatusBadRequest, "invalid_token", "Confirmation link is invalid or expired")
		return
	}

	if _, err := h.db.GetSubscriberByID(r.Context(), subID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	inList, err := h.db.IsSubscriberInList(r.Context(), db.IsSubscriberInListParams{
		ListID:       listID,
		SubscriberID: subID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if !inList {
		writeError(w, http.StatusBadRequest, "not_found", "Subscription not found")
		return
	}

	if err := h.db.UpdateSubscriberStatus(r.Context(), db.UpdateSubscriberStatusParams{
		ID:     subID,
		Status: "active",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Email confirmed. You are now subscribed."})
}
