package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AbMani46/ownmaily/internal/auth"
	sqlcdb "github.com/AbMani46/ownmaily/internal/sqlc"
)

// CreateOwner hashes a known password and inserts an owner row.
func CreateOwner(t *testing.T, db *sqlcdb.Queries) sqlcdb.Owner {
	t.Helper()
	hash, err := auth.HashPassword("testpass123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	owner, err := db.CreateOwner(context.Background(), sqlcdb.CreateOwnerParams{
		Email:        "owner@test.com",
		PasswordHash: hash,
	})
	if err != nil {
		t.Fatalf("create owner: %v", err)
	}
	return owner
}

// LoginAndGetToken posts to /api/auth/login and returns the JWT token string.
func LoginAndGetToken(t *testing.T, server *httptest.Server, email, password string) string {
	t.Helper()
	resp := MakeRequest(t, server, "POST", "/api/auth/login",
		map[string]string{"email": email, "password": password}, nil)
	defer resp.Body.Close()
	var result map[string]string
	DecodeJSON(t, resp, &result)
	token, ok := result["token"]
	if !ok {
		t.Fatalf("login response missing token field: %v", result)
	}
	return token
}

// CreateSubscriber inserts an active subscriber directly via DB.
func CreateSubscriber(t *testing.T, db *sqlcdb.Queries, email string) sqlcdb.Subscriber {
	t.Helper()
	sub, err := db.CreateSubscriber(context.Background(), sqlcdb.CreateSubscriberParams{
		Email:     email,
		FirstName: "",
		LastName:  "",
		Status:    "active",
		Source:    "api",
	})
	if err != nil {
		t.Fatalf("create subscriber %s: %v", email, err)
	}
	return sub
}

// CreateList inserts a list directly via DB.
func CreateList(t *testing.T, db *sqlcdb.Queries, name string) sqlcdb.List {
	t.Helper()
	list, err := db.CreateList(context.Background(), sqlcdb.CreateListParams{
		Name:        name,
		Description: "",
		DoubleOptIn: false,
	})
	if err != nil {
		t.Fatalf("create list %s: %v", name, err)
	}
	return list
}

// CreateTag inserts a tag directly via DB.
func CreateTag(t *testing.T, db *sqlcdb.Queries, name string) sqlcdb.Tag {
	t.Helper()
	tag, err := db.CreateTag(context.Background(), name)
	if err != nil {
		t.Fatalf("create tag %s: %v", name, err)
	}
	return tag
}

// CreateCampaign inserts a draft campaign directly via DB.
func CreateCampaign(t *testing.T, db *sqlcdb.Queries, listID pgtype.UUID) sqlcdb.Campaign {
	t.Helper()
	c, err := db.CreateCampaign(context.Background(), sqlcdb.CreateCampaignParams{
		Name:        "Test Campaign",
		Subject:     "Test Subject",
		PreviewText: "",
		FromName:    "Tester",
		FromEmail:   "tester@example.com",
		ReplyTo:     "",
		HtmlBody:    "<p>Hello</p>",
		TextBody:    "Hello",
		SendToType:  "list",
		SendToID:    listID,
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	return c
}

// AuthHeader returns a header map with a JWT Bearer token.
func AuthHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// MakeRequest sends an HTTP request to the test server and returns the response.
// body is marshalled to JSON if non-nil. headers are added to the request.
func MakeRequest(t *testing.T, server *httptest.Server, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// DecodeJSON decodes the response body into v. Closes the body.
func DecodeJSON(t *testing.T, r *http.Response, v any) {
	t.Helper()
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("decode JSON response: %v", err)
	}
}

// AssertStatus fails the test if r.StatusCode != expected, printing the body.
func AssertStatus(t *testing.T, r *http.Response, expected int) {
	t.Helper()
	if r.StatusCode == expected {
		return
	}
	body, _ := io.ReadAll(r.Body)
	r.Body.Close()
	t.Fatalf("expected status %d, got %d\nbody: %s", expected, r.StatusCode, body)
}

// UUIDString converts a pgtype.UUID to its string representation.
func UUIDString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		id.Bytes[0:4], id.Bytes[4:6], id.Bytes[6:8], id.Bytes[8:10], id.Bytes[10:16])
}
