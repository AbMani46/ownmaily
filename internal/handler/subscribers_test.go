package handler_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

// makeImportRequest builds and sends a multipart CSV import request.
func makeImportRequest(t *testing.T, srv *httptest.Server, csvContent string, headers map[string]string) *http.Response {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", "test.csv")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	fmt.Fprint(fw, csvContent)
	w.Close()

	req, err := http.NewRequest("POST", srv.URL+"/api/subscribers/import", &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("import request: %v", err)
	}
	return resp
}

func TestCreateSubscriber_Success(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/subscribers",
		map[string]string{"email": "alice@test.com", "first_name": "Alice"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusCreated)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if body["email"] != "alice@test.com" {
		t.Fatalf("expected email=alice@test.com, got %v", body["email"])
	}
}

func TestCreateSubscriber_DuplicateEmail(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	testutil.CreateSubscriber(t, db, "dup@test.com")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/subscribers",
		map[string]string{"email": "dup@test.com"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)

	if body["error"] != "duplicate" {
		t.Fatalf("expected error=duplicate, got %q", body["error"])
	}
}

func TestCreateSubscriber_SuppressedEmail(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	// Add email to suppression list directly.
	if err := db.AddSuppression(context.Background(), sqlc.AddSuppressionParams{
		Email:  "suppressed@test.com",
		Reason: "manual",
	}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "POST", "/api/subscribers",
		map[string]string{"email": "suppressed@test.com"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)

	if body["error"] != "suppressed" {
		t.Fatalf("expected error=suppressed, got %q", body["error"])
	}
}

func TestCreateSubscriber_InvalidEmail(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/subscribers",
		map[string]string{"email": "notanemail"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)
}

func TestListSubscribers_Pagination(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	testutil.CreateSubscriber(t, db, "page1@test.com")
	testutil.CreateSubscriber(t, db, "page2@test.com")
	testutil.CreateSubscriber(t, db, "page3@test.com")

	// Page 1: 2 of 3.
	resp1 := testutil.MakeRequest(t, srv, "GET", "/api/subscribers?per_page=2&page=1",
		nil, testutil.AuthHeader(token))
	defer resp1.Body.Close()
	testutil.AssertStatus(t, resp1, http.StatusOK)

	var page1 map[string]any
	testutil.DecodeJSON(t, resp1, &page1)

	subs1 := page1["subscribers"].([]any)
	if len(subs1) != 2 {
		t.Fatalf("page 1: expected 2 subscribers, got %d", len(subs1))
	}
	if int(page1["total"].(float64)) != 3 {
		t.Fatalf("expected total=3, got %v", page1["total"])
	}

	// Page 2: 1 of 3.
	resp2 := testutil.MakeRequest(t, srv, "GET", "/api/subscribers?per_page=2&page=2",
		nil, testutil.AuthHeader(token))
	defer resp2.Body.Close()
	testutil.AssertStatus(t, resp2, http.StatusOK)

	var page2 map[string]any
	testutil.DecodeJSON(t, resp2, &page2)

	subs2 := page2["subscribers"].([]any)
	if len(subs2) != 1 {
		t.Fatalf("page 2: expected 1 subscriber, got %d", len(subs2))
	}
}

func TestListSubscribers_SearchFilter(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	testutil.CreateSubscriber(t, db, "alice@search.com")
	testutil.CreateSubscriber(t, db, "bob@search.com")
	testutil.CreateSubscriber(t, db, "charlie@search.com")

	resp := testutil.MakeRequest(t, srv, "GET", "/api/subscribers?q=alice",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()
	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	subs := body["subscribers"].([]any)
	if len(subs) != 1 {
		t.Fatalf("expected 1 result for q=alice, got %d", len(subs))
	}
	got := subs[0].(map[string]any)["email"]
	if got != "alice@search.com" {
		t.Fatalf("expected alice@search.com, got %v", got)
	}
}

func TestListSubscribers_StatusFilter(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	active1 := testutil.CreateSubscriber(t, db, "active1@test.com")
	testutil.CreateSubscriber(t, db, "active2@test.com")

	// Mark one subscriber as unsubscribed.
	if err := db.UpdateSubscriberStatus(context.Background(), sqlc.UpdateSubscriberStatusParams{
		ID:     active1.ID,
		Status: "unsubscribed",
	}); err != nil {
		t.Fatalf("update subscriber status: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "GET", "/api/subscribers?status=active",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()
	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	subs := body["subscribers"].([]any)
	if len(subs) != 1 {
		t.Fatalf("expected 1 active subscriber, got %d", len(subs))
	}
	got := subs[0].(map[string]any)["email"]
	if got != "active2@test.com" {
		t.Fatalf("expected active2@test.com, got %v", got)
	}
}

func TestGetSubscriber_NotFound(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "GET",
		"/api/subscribers/00000000-0000-0000-0000-000000000000",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNotFound)
}

func TestUpdateSubscriber(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "update@test.com")

	resp := testutil.MakeRequest(t, srv, "PUT",
		"/api/subscribers/"+testutil.UUIDString(sub.ID),
		map[string]string{"email": "update@test.com", "first_name": "Updated"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if body["first_name"] != "Updated" {
		t.Fatalf("expected first_name=Updated, got %v", body["first_name"])
	}
}

func TestDeleteSubscriber_AddsSuppression(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "delete@test.com")

	resp := testutil.MakeRequest(t, srv, "DELETE",
		"/api/subscribers/"+testutil.UUIDString(sub.ID),
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNoContent)

	// Verify email is now in the suppression list.
	suppressed, err := db.IsSuppressed(context.Background(), "delete@test.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if !suppressed {
		t.Fatal("expected delete@test.com to be in the suppression list after delete")
	}
}

func TestUnsubscribeSubscriber(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "unsub@test.com")

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/subscribers/"+testutil.UUIDString(sub.ID)+"/unsubscribe",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	// Status in DB should be "unsubscribed".
	updated, err := db.GetSubscriberByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if updated.Status != "unsubscribed" {
		t.Fatalf("expected status=unsubscribed, got %q", updated.Status)
	}

	// Suppression row should exist.
	suppressed, err := db.IsSuppressed(context.Background(), "unsub@test.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if !suppressed {
		t.Fatal("expected unsub@test.com in suppression list after unsubscribe")
	}
}

func TestImportCSV_ValidAndInvalid(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	// Pre-suppress one email.
	if err := db.AddSuppression(context.Background(), sqlc.AddSuppressionParams{
		Email:  "suppressed@import.com",
		Reason: "manual",
	}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}

	// CSV: 2 valid, 1 duplicate, 1 invalid, 1 suppressed.
	csvContent := strings.Join([]string{
		"email,first_name,last_name",
		"alice@import.com,Alice,Smith",      // imported
		"bob@import.com,Bob,Jones",          // imported
		"alice@import.com,Alice,Dup",        // skipped (duplicate)
		"notanemail,,",                      // invalid
		"suppressed@import.com,,",           // skipped (suppressed)
	}, "\n") + "\n"

	resp := makeImportRequest(t, srv, csvContent, testutil.AuthHeader(token))
	defer resp.Body.Close()
	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if int(body["imported"].(float64)) != 2 {
		t.Fatalf("expected imported=2, got %v", body["imported"])
	}
	if int(body["skipped"].(float64)) != 2 {
		t.Fatalf("expected skipped=2, got %v", body["skipped"])
	}
	if int(body["invalid"].(float64)) != 1 {
		t.Fatalf("expected invalid=1, got %v", body["invalid"])
	}
}

func TestExportCSV(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	testutil.CreateSubscriber(t, db, "export1@test.com")
	testutil.CreateSubscriber(t, db, "export2@test.com")

	resp := testutil.MakeRequest(t, srv, "GET", "/api/subscribers/export",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/csv") {
		t.Fatalf("expected text/csv content type, got %q", ct)
	}

	// Parse and verify CSV content.
	r := csv.NewReader(resp.Body)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}

	// First row is header; remaining rows are subscribers.
	if len(rows) < 3 { // header + 2 subscribers
		t.Fatalf("expected at least 3 rows (header + 2 subs), got %d", len(rows))
	}

	emails := map[string]bool{}
	for _, row := range rows[1:] {
		emails[row[0]] = true
	}
	if !emails["export1@test.com"] {
		t.Fatal("expected export1@test.com in CSV export")
	}
	if !emails["export2@test.com"] {
		t.Fatal("expected export2@test.com in CSV export")
	}

	// Verify Content-Disposition includes a filename.
	cd := resp.Header.Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") {
		t.Fatalf("expected Content-Disposition: attachment, got %q", cd)
	}

	// Drain body to allow connection reuse.
	_, _ = io.Copy(io.Discard, resp.Body)
}
