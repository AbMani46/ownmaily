package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AbMani46/ownmaily/internal/testutil"
)

func TestGetSettings(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "GET", "/api/settings", nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if _, ok := body["smtp_credentials"]; ok {
		t.Fatal("smtp_credentials must not be present in GET /api/settings response")
	}
	if _, ok := body["smtp_provider"]; !ok {
		t.Fatal("smtp_provider field should be present")
	}
}

func TestUpdateGeneralSettings(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	update := map[string]any{
		"site_name":        "My Newsletter",
		"installation_url": "https://mail.example.com",
		"timezone":         "UTC",
		"physical_address": "123 Main St",
		"from_name":        "Newsletter Bot",
		"from_email":       "news@example.com",
		"reply_to":         "reply@example.com",
	}

	resp := testutil.MakeRequest(t, srv, "PUT", "/api/settings/general", update, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if body["site_name"] != "My Newsletter" {
		t.Fatalf("expected site_name=My Newsletter, got %v", body["site_name"])
	}
	if body["from_email"] != "news@example.com" {
		t.Fatalf("expected from_email=news@example.com, got %v", body["from_email"])
	}
}

func TestUpdateSMTPSettings_InvalidProvider(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "PUT", "/api/settings/smtp",
		map[string]any{"provider": "sendgrid", "credentials": map[string]string{"api_key": "key"}},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "invalid_provider" {
		t.Fatalf("expected error=invalid_provider, got %v", body["error"])
	}
}

func TestUpdateSMTPSettings_ValidProvider(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "PUT", "/api/settings/smtp",
		map[string]any{"provider": "resend", "credentials": map[string]string{"api_key": "re_fake_key"}},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	// Confirm GET /api/settings now shows resend as provider.
	resp2 := testutil.MakeRequest(t, srv, "GET", "/api/settings", nil, testutil.AuthHeader(token))
	defer resp2.Body.Close()

	var body map[string]any
	testutil.DecodeJSON(t, resp2, &body)
	if body["smtp_provider"] != "resend" {
		t.Fatalf("expected smtp_provider=resend, got %v", body["smtp_provider"])
	}
}

func TestRegenerateAPIKey(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/settings/api-key/regenerate", nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	key, ok := body["key"].(string)
	if !ok || !strings.HasPrefix(key, "om_") {
		t.Fatalf("expected key starting with om_, got %v", body["key"])
	}
}

func TestGetAPIKey_AfterRegenerate(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/settings/api-key/regenerate", nil, testutil.AuthHeader(token))
	resp.Body.Close()

	resp2 := testutil.MakeRequest(t, srv, "GET", "/api/settings/api-key", nil, testutil.AuthHeader(token))
	defer resp2.Body.Close()

	testutil.AssertStatus(t, resp2, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp2, &body)

	if body["key_prefix"] == nil {
		t.Fatal("expected key_prefix to be set after regenerate")
	}
}

func TestAPIKey_CanAuthenticateRequests(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	// Regenerate to get a fresh API key.
	resp := testutil.MakeRequest(t, srv, "POST", "/api/settings/api-key/regenerate", nil, testutil.AuthHeader(token))
	var regen map[string]any
	testutil.DecodeJSON(t, resp, &regen)
	apiKey := regen["key"].(string)

	// Use the raw API key as Bearer token on a protected endpoint.
	resp2 := testutil.MakeRequest(t, srv, "GET", "/api/settings", nil,
		map[string]string{"Authorization": "Bearer " + apiKey})
	defer resp2.Body.Close()

	testutil.AssertStatus(t, resp2, http.StatusOK)
}

func TestAddSuppression_Manual(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/settings/suppressions",
		map[string]string{"email": "blocked@example.com"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)
}

func TestAddSuppression_Duplicate(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	testutil.MakeRequest(t, srv, "POST", "/api/settings/suppressions",
		map[string]string{"email": "dup@example.com"},
		testutil.AuthHeader(token)).Body.Close()

	resp := testutil.MakeRequest(t, srv, "POST", "/api/settings/suppressions",
		map[string]string{"email": "dup@example.com"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)
}

func TestDeleteSuppression(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	testutil.MakeRequest(t, srv, "POST", "/api/settings/suppressions",
		map[string]string{"email": "del@example.com"},
		testutil.AuthHeader(token)).Body.Close()

	resp := testutil.MakeRequest(t, srv, "DELETE", "/api/settings/suppressions/del%40example.com",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNoContent)
}

func TestListSuppressions_Paginated(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	for i := range 3 {
		testutil.MakeRequest(t, srv, "POST", "/api/settings/suppressions",
			map[string]string{"email": "sup" + string(rune('a'+i)) + "@example.com"},
			testutil.AuthHeader(token)).Body.Close()
	}

	resp := testutil.MakeRequest(t, srv, "GET", "/api/settings/suppressions?page=1&per_page=2",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	total, ok := body["total"].(float64)
	if !ok || total != 3 {
		t.Fatalf("expected total=3, got %v", body["total"])
	}
	sups := body["suppressions"].([]any)
	if len(sups) != 2 {
		t.Fatalf("expected 2 items on page 1, got %d", len(sups))
	}
}

func TestExportSuppressions_CSV(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	testutil.MakeRequest(t, srv, "POST", "/api/settings/suppressions",
		map[string]string{"email": "export@example.com"},
		testutil.AuthHeader(token)).Body.Close()

	resp := testutil.MakeRequest(t, srv, "GET", "/api/settings/suppressions/export",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("expected Content-Type text/csv, got %s", ct)
	}
	cd := resp.Header.Get("Content-Disposition")
	if !strings.Contains(cd, "suppressions-") {
		t.Fatalf("expected Content-Disposition to contain suppressions-, got %s", cd)
	}
}
