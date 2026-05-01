package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/mailer"
)

type EmbedHandler struct {
	db              *db.Queries
	installationURL string
	mailer          *mailer.ConfirmationMailer

	rateMu  sync.Mutex
	rateMap map[string][]time.Time
}

func NewEmbedHandler(q *db.Queries, installationURL string, m *mailer.ConfirmationMailer) *EmbedHandler {
	return &EmbedHandler{
		db:              q,
		installationURL: installationURL,
		mailer:          m,
		rateMap:         make(map[string][]time.Time),
	}
}

func (h *EmbedHandler) ServeJS(w http.ResponseWriter, r *http.Request) {
	listIDStr := chi.URLParam(r, "listID")
	// Strip .js suffix (chi pattern captures it as part of the param)
	listIDStr = strings.TrimSuffix(listIDStr, ".js")

	listID, err := parseUUID(listIDStr)
	if err != nil {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprint(w, "// OwnMaily: list not found\n")
		return
	}

	if _, err := h.db.GetListByID(r.Context(), listID); err != nil {
		w.Header().Set("Content-Type", "application/javascript")
		fmt.Fprint(w, "// OwnMaily: list not found\n")
		return
	}

	js := buildEmbedJS(listIDStr, h.installationURL)
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=300")
	fmt.Fprint(w, js)
}

func (h *EmbedHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	}
	if !h.checkRate(ip) {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
		return
	}

	var body struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		ListID    string `json:"list_id"`
	}
	if err := readJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body")
		return
	}

	if !strings.Contains(body.Email, "@") {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid email")
		return
	}

	listID, err := parseUUID(body.ListID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid list_id")
		return
	}

	list, err := h.db.GetListByID(r.Context(), listID)
	if err != nil {
		if isNoRows(err) {
			writeError(w, http.StatusBadRequest, "bad_request", "list not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	suppressed, err := h.db.IsSuppressed(r.Context(), body.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}
	if suppressed {
		writeJSON(w, http.StatusOK, map[string]string{"message": "already_subscribed"})
		return
	}

	existing, err := h.db.GetSubscriberByEmail(r.Context(), body.Email)
	if err != nil && !isNoRows(err) {
		writeError(w, http.StatusInternalServerError, "internal_error", "server error")
		return
	}

	if err == nil && existing.Status == "active" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "already_subscribed"})
		return
	}

	var sub db.Subscriber
	if isNoRows(err) {
		status := "active"
		if list.DoubleOptIn {
			status = "pending"
		}
		sub, err = h.db.CreateSubscriber(r.Context(), db.CreateSubscriberParams{
			Email:     body.Email,
			FirstName: body.FirstName,
			Status:    status,
			Source:    "form",
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
	} else {
		sub = existing
	}

	if list.DoubleOptIn {
		if err := h.mailer.SendConfirmation(sub.ID, listID, sub.Email); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "server error")
			return
		}
	}

	if err := h.db.AddSubscriberToList(r.Context(), db.AddSubscriberToListParams{
		ListID:       listID,
		SubscriberID: sub.ID,
	}); err != nil {
		// Ignore duplicate — subscriber already in list
		_ = err
	}

	if list.DoubleOptIn {
		writeJSON(w, http.StatusOK, map[string]string{"message": "check_email"})
	} else {
		writeJSON(w, http.StatusOK, map[string]string{"message": "subscribed"})
	}
}

// checkRate allows max 5 requests per IP per minute.
func (h *EmbedHandler) checkRate(ip string) bool {
	now := time.Now()
	cutoff := now.Add(-time.Minute)

	h.rateMu.Lock()
	defer h.rateMu.Unlock()

	timestamps := h.rateMap[ip]
	filtered := timestamps[:0]
	for _, t := range timestamps {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= 5 {
		h.rateMap[ip] = filtered
		return false
	}
	h.rateMap[ip] = append(filtered, now)
	return true
}

func isNoRows(err error) bool {
	return err == pgx.ErrNoRows
}

func buildEmbedJS(listID, baseURL string) string {
	return fmt.Sprintf(`(function() {
  var listId = '%s';
  var baseUrl = '%s';

  var scripts = document.getElementsByTagName('script');
  var currentScript = scripts[scripts.length - 1];

  var form = document.createElement('form');
  form.style.cssText = 'margin:0;padding:0;';

  form.innerHTML = '<div style="display:flex;flex-direction:column;gap:8px;max-width:400px;">' +
    '<input type="text" name="first_name" placeholder="Your name" ' +
    'style="padding:8px 12px;border:1px solid #ddd;border-radius:6px;font-size:14px;" />' +
    '<input type="email" name="email" placeholder="Your email" required ' +
    'style="padding:8px 12px;border:1px solid #ddd;border-radius:6px;font-size:14px;" />' +
    '<button type="submit" ' +
    'style="padding:8px 16px;background:#10b981;color:#fff;border:none;border-radius:6px;font-size:14px;cursor:pointer;">' +
    'Subscribe</button>' +
    '<div class="om-message" style="font-size:13px;display:none;"></div>' +
    '</div>';

  form.addEventListener('submit', function(e) {
    e.preventDefault();
    var msg = form.querySelector('.om-message');
    var btn = form.querySelector('button');
    btn.disabled = true;
    btn.textContent = 'Subscribing...';

    fetch(baseUrl + '/api/public/subscribe', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        email: form.querySelector('[name=email]').value,
        first_name: form.querySelector('[name=first_name]').value,
        list_id: listId
      })
    })
    .then(function(r) { return r.json(); })
    .then(function(data) {
      msg.style.display = 'block';
      if (data.message === 'subscribed' || data.message === 'check_email' || data.message === 'already_subscribed') {
        msg.style.color = '#059669';
        msg.textContent = data.message === 'check_email' ? 'Check your email to confirm!' : 'Thanks for subscribing!';
        form.querySelector('[name=email]').value = '';
        form.querySelector('[name=first_name]').value = '';
      } else {
        msg.style.color = '#dc2626';
        msg.textContent = 'Something went wrong.';
      }
      btn.disabled = false;
      btn.textContent = 'Subscribe';
    })
    .catch(function() {
      msg.style.display = 'block';
      msg.style.color = '#dc2626';
      msg.textContent = 'Something went wrong. Please try again.';
      btn.disabled = false;
      btn.textContent = 'Subscribe';
    });
  });

  currentScript.parentNode.insertBefore(form, currentScript.nextSibling);
})();
`, listID, baseURL)
}
