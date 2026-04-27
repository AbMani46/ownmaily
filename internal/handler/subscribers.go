package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type SubscriberHandler struct {
	db *db.Queries
}

func NewSubscriberHandler(q *db.Queries) *SubscriberHandler {
	return &SubscriberHandler{db: q}
}

type subscriberResponse struct {
	db.Subscriber
	Tags []db.Tag `json:"tags"`
}

type subscriberDetailResponse struct {
	db.Subscriber
	Tags  []db.Tag  `json:"tags"`
	Lists []db.List `json:"lists"`
}

func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	return id, id.Scan(s)
}

func parsePage(r *http.Request) (page, perPage int32) {
	page, perPage = 1, 50
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = int32(n)
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if n, err := strconv.Atoi(pp); err == nil && n > 0 && n <= 100 {
			perPage = int32(n)
		}
	}
	return
}

func (h *SubscriberHandler) List(w http.ResponseWriter, r *http.Request) {
	page, perPage := parsePage(r)
	offset := (page - 1) * perPage
	q := r.URL.Query()
	qStr := q.Get("q")
	statusStr := q.Get("status")
	listIDStr := q.Get("list_id")
	tagIDStr := q.Get("tag_id")

	var subs []db.Subscriber
	var total int64
	var err error

	switch {
	case qStr != "":
		pattern := "%" + qStr + "%"
		subs, err = h.db.SearchSubscribers(r.Context(), db.SearchSubscribersParams{
			Email:  pattern,
			Limit:  perPage,
			Offset: offset,
		})
		total = int64(len(subs))
	case statusStr != "":
		subs, err = h.db.ListSubscribersByStatus(r.Context(), db.ListSubscribersByStatusParams{
			Status: statusStr,
			Limit:  perPage,
			Offset: offset,
		})
		total = int64(len(subs))
	case listIDStr != "":
		listID, parseErr := parseUUID(listIDStr)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid list_id")
			return
		}
		subs, err = h.db.ListSubscribersInList(r.Context(), db.ListSubscribersInListParams{
			ListID: listID,
			Limit:  perPage,
			Offset: offset,
		})
		total = int64(len(subs))
	case tagIDStr != "":
		tagID, parseErr := parseUUID(tagIDStr)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid tag_id")
			return
		}
		subs, err = h.db.ListSubscribersWithTag(r.Context(), db.ListSubscribersWithTagParams{
			TagID:  tagID,
			Limit:  perPage,
			Offset: offset,
		})
		total = int64(len(subs))
	default:
		subs, err = h.db.ListSubscribers(r.Context(), db.ListSubscribersParams{
			Limit:  perPage,
			Offset: offset,
		})
		if err == nil {
			total, err = h.db.CountSubscribers(r.Context())
		}
	}

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

func (h *SubscriberHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(body.Email))
	if email == "" || !strings.Contains(email, "@") {
		writeError(w, http.StatusBadRequest, "bad_request", "valid email is required")
		return
	}

	suppressed, err := h.db.IsSuppressed(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if suppressed {
		writeError(w, http.StatusConflict, "suppressed", "This email is on the suppression list")
		return
	}

	_, err = h.db.GetSubscriberByEmail(r.Context(), email)
	if err == nil {
		writeError(w, http.StatusConflict, "duplicate", "Subscriber already exists")
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	sub, err := h.db.CreateSubscriber(r.Context(), db.CreateSubscriberParams{
		Email:     email,
		FirstName: strings.TrimSpace(body.FirstName),
		LastName:  strings.TrimSpace(body.LastName),
		Status:    "active",
		Source:    "api",
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusCreated, subscriberResponse{Subscriber: sub, Tags: []db.Tag{}})
}

func (h *SubscriberHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	sub, err := h.db.GetSubscriberByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	tags, err := h.db.ListTagsForSubscriber(r.Context(), sub.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if tags == nil {
		tags = []db.Tag{}
	}

	lists, err := h.db.ListListsForSubscriber(r.Context(), sub.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if lists == nil {
		lists = []db.List{}
	}

	writeJSON(w, http.StatusOK, subscriberDetailResponse{
		Subscriber: sub,
		Tags:       tags,
		Lists:      lists,
	})
}

func (h *SubscriberHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	existing, err := h.db.GetSubscriberByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	var body struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Status    string `json:"status"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	newEmail := strings.TrimSpace(strings.ToLower(body.Email))
	if newEmail == "" {
		newEmail = existing.Email
	}
	newStatus := body.Status
	if newStatus == "" {
		newStatus = existing.Status
	}

	if newEmail != existing.Email {
		suppressed, suppErr := h.db.IsSuppressed(r.Context(), newEmail)
		if suppErr != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
		if suppressed {
			writeError(w, http.StatusConflict, "suppressed", "This email is on the suppression list")
			return
		}

		_, dupErr := h.db.GetSubscriberByEmail(r.Context(), newEmail)
		if dupErr == nil {
			writeError(w, http.StatusConflict, "duplicate", "Subscriber already exists")
			return
		}
		if !errors.Is(dupErr, pgx.ErrNoRows) {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
	}

	updated, err := h.db.UpdateSubscriber(r.Context(), db.UpdateSubscriberParams{
		ID:        id,
		Email:     newEmail,
		FirstName: strings.TrimSpace(body.FirstName),
		LastName:  strings.TrimSpace(body.LastName),
		Status:    newStatus,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	tags, err := h.db.ListTagsForSubscriber(r.Context(), updated.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if tags == nil {
		tags = []db.Tag{}
	}

	writeJSON(w, http.StatusOK, subscriberResponse{Subscriber: updated, Tags: tags})
}

func (h *SubscriberHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	sub, err := h.db.GetSubscriberByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.AddSuppression(r.Context(), db.AddSuppressionParams{
		Email:  sub.Email,
		Reason: "manual",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.DeleteSubscriber(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriberHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id")
		return
	}

	sub, err := h.db.GetSubscriberByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "subscriber not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.UpdateSubscriberStatus(r.Context(), db.UpdateSubscriberStatusParams{
		ID:     id,
		Status: "unsubscribed",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err := h.db.AddSuppression(r.Context(), db.AddSuppressionParams{
		Email:  sub.Email,
		Reason: "unsubscribed",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "unsubscribed"})
}

func (h *SubscriberHandler) Import(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "too_large", "file exceeds 10MB limit")
		} else {
			writeError(w, http.StatusBadRequest, "bad_request", "could not parse form")
		}
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "missing file field")
		return
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	headers, err := csvReader.Read()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_csv", "could not read CSV headers")
		return
	}

	emailIdx, firstIdx, lastIdx := -1, -1, -1
	for i, col := range headers {
		switch strings.ToLower(strings.TrimSpace(col)) {
		case "email":
			emailIdx = i
		case "first_name":
			firstIdx = i
		case "last_name":
			lastIdx = i
		}
	}

	if emailIdx == -1 {
		writeError(w, http.StatusBadRequest, "invalid_csv", "CSV must have an email column")
		return
	}

	var imported, skipped, invalid int

	for {
		row, rowErr := csvReader.Read()
		if errors.Is(rowErr, io.EOF) {
			break
		}
		if rowErr != nil {
			invalid++
			continue
		}

		if emailIdx >= len(row) {
			invalid++
			continue
		}

		email := strings.TrimSpace(strings.ToLower(row[emailIdx]))
		if email == "" || !strings.Contains(email, "@") {
			invalid++
			continue
		}

		suppressed, suppErr := h.db.IsSuppressed(r.Context(), email)
		if suppErr != nil || suppressed {
			skipped++
			continue
		}

		var firstName, lastName string
		if firstIdx >= 0 && firstIdx < len(row) {
			firstName = strings.TrimSpace(row[firstIdx])
		}
		if lastIdx >= 0 && lastIdx < len(row) {
			lastName = strings.TrimSpace(row[lastIdx])
		}

		_, createErr := h.db.CreateSubscriber(r.Context(), db.CreateSubscriberParams{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Status:    "active",
			Source:    "import",
		})
		if createErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(createErr, &pgErr) && pgErr.Code == "23505" {
				skipped++
			} else {
				skipped++
			}
			continue
		}
		imported++
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"imported": imported,
		"skipped":  skipped,
		"invalid":  invalid,
	})
}

func (h *SubscriberHandler) Export(w http.ResponseWriter, r *http.Request) {
	subs, err := h.db.ListAllSubscribers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	filename := fmt.Sprintf("subscribers-%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"email", "first_name", "last_name", "status", "created_at"})

	for _, sub := range subs {
		createdAt := ""
		if sub.CreatedAt.Valid {
			createdAt = sub.CreatedAt.Time.Format(time.RFC3339)
		}
		_ = cw.Write([]string{sub.Email, sub.FirstName, sub.LastName, sub.Status, createdAt})
	}
	cw.Flush()
}
