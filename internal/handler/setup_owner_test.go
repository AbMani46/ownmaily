package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbMani46/ownmaily/internal/auth"
	"github.com/AbMani46/ownmaily/internal/handler"
	"github.com/AbMani46/ownmaily/internal/mailer"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

func callSetupOwner(t *testing.T, fn http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	rec := httptest.NewRecorder()
	fn(rec, httptest.NewRequest("POST", path, bytes.NewReader(b)))
	return rec
}

func TestSetupCreateOwner_SecondCallReplacesOwner(t *testing.T) {
	_, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)

	h := handler.NewSetupHandler(db, testutil.TestAppSecret, mailer.NewStore(&mailer.LogMailer{}))

	rec := callSetupOwner(t, h.CreateOwner, "/api/setup/owner",
		map[string]string{"email": "first@example.com", "password": "password1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("first call: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = callSetupOwner(t, h.CreateOwner, "/api/setup/owner",
		map[string]string{"email": "second@example.com", "password": "password2"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("second call: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	owner, err := db.GetOwner(context.Background())
	if err != nil {
		t.Fatalf("get owner: %v", err)
	}
	if owner.Email != "second@example.com" {
		t.Fatalf("expected owner email to be replaced, got %q", owner.Email)
	}
	if !auth.CheckPassword("password2", owner.PasswordHash) {
		t.Fatal("expected owner password to be replaced")
	}
}

func TestSetupComplete_WithoutOwner_DoesNotFinishSetup(t *testing.T) {
	_, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)

	h := handler.NewSetupHandler(db, testutil.TestAppSecret, mailer.NewStore(&mailer.LogMailer{}))

	rec := callSetupOwner(t, h.Complete, "/api/setup/complete", map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	s, err := db.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if s.SetupComplete {
		t.Fatal("expected setup to stay incomplete without an owner")
	}
}
