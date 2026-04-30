package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AbMani46/ownmaily/internal/testutil"
)

func TestLogin_Success(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)

	resp := testutil.MakeRequest(t, srv, "POST", "/api/auth/login",
		map[string]string{"email": "owner@test.com", "password": "testpass123"}, nil)
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	token, ok := body["token"].(string)
	if !ok || token == "" {
		t.Fatalf("expected non-empty token in response, got %v", body)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)

	resp := testutil.MakeRequest(t, srv, "POST", "/api/auth/login",
		map[string]string{"email": "owner@test.com", "password": "wrongpassword"}, nil)
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusUnauthorized)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)

	if body["error"] != "unauthorized" {
		t.Fatalf("expected error=unauthorized, got %q", body["error"])
	}
}

func TestLogin_NoOwner(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	// No owner created — DB is empty.
	resp := testutil.MakeRequest(t, srv, "POST", "/api/auth/login",
		map[string]string{"email": "nobody@test.com", "password": "testpass123"}, nil)
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusUnauthorized)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)

	if body["error"] != "unauthorized" {
		t.Fatalf("expected error=unauthorized, got %q", body["error"])
	}
}

func TestLogin_SetsJWTCookie(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)

	resp := testutil.MakeRequest(t, srv, "POST", "/api/auth/login",
		map[string]string{"email": "owner@test.com", "password": "testpass123"}, nil)
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var jwtCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "jwt" {
			jwtCookie = c
			break
		}
	}
	if jwtCookie == nil {
		t.Fatal("expected jwt cookie in login response")
	}
	if !jwtCookie.HttpOnly {
		t.Fatal("expected jwt cookie to be HttpOnly")
	}
	if jwtCookie.Value == "" {
		t.Fatal("expected jwt cookie to have a non-empty value")
	}
}

func TestMe_WithValidToken(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "GET", "/api/auth/me", nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)

	if body["email"] != "owner@test.com" {
		t.Fatalf("expected email=owner@test.com, got %q", body["email"])
	}
}

func TestMe_WithoutToken(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	resp := testutil.MakeRequest(t, srv, "GET", "/api/auth/me", nil, nil)
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusUnauthorized)
}

func TestMe_WithAPIKey(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	// Regenerate API key using JWT auth.
	regenResp := testutil.MakeRequest(t, srv, "POST", "/api/settings/api-key/regenerate",
		nil, testutil.AuthHeader(token))
	testutil.AssertStatus(t, regenResp, http.StatusOK)

	var regenBody map[string]any
	testutil.DecodeJSON(t, regenResp, &regenBody)

	rawKey, ok := regenBody["key"].(string)
	if !ok || !strings.HasPrefix(rawKey, "om_") {
		t.Fatalf("expected raw API key with om_ prefix, got %v", regenBody["key"])
	}

	// Use the raw API key as Bearer token to authenticate.
	meResp := testutil.MakeRequest(t, srv, "GET", "/api/auth/me", nil,
		map[string]string{"Authorization": "Bearer " + rawKey})
	defer meResp.Body.Close()

	testutil.AssertStatus(t, meResp, http.StatusOK)
}

func TestLogout_ClearsCookie(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	resp := testutil.MakeRequest(t, srv, "POST", "/api/auth/logout", nil, nil)
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	// Go serializes MaxAge=-1 as Max-Age=0 in the Set-Cookie header, which
	// instructs the browser to delete the cookie immediately.
	found := false
	for _, h := range resp.Header["Set-Cookie"] {
		if strings.HasPrefix(h, "jwt=") && strings.Contains(h, "Max-Age=0") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Set-Cookie: jwt= with Max-Age=0, got: %v", resp.Header["Set-Cookie"])
	}
}
