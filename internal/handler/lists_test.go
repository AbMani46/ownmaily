package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/AbMani46/ownmaily/internal/mailer"
	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

func TestCreateList_Success(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/lists",
		map[string]any{"name": "My Newsletter", "description": "Weekly news"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusCreated)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if body["name"] != "My Newsletter" {
		t.Fatalf("expected name=My Newsletter, got %v", body["name"])
	}
	if int(body["subscriber_count"].(float64)) != 0 {
		t.Fatalf("expected subscriber_count=0, got %v", body["subscriber_count"])
	}
}

func TestCreateList_DuplicateName(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	testutil.CreateList(t, db, "Duplicate Name")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/lists",
		map[string]any{"name": "Duplicate Name"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "duplicate" {
		t.Fatalf("expected error=duplicate, got %q", body["error"])
	}
}

func TestCreateList_NameRequired(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/lists",
		map[string]any{"name": ""},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)
}

func TestGetList_NotFound(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "GET",
		"/api/lists/00000000-0000-0000-0000-000000000000",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNotFound)
}

func TestUpdateList(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Original Name")

	resp := testutil.MakeRequest(t, srv, "PUT",
		"/api/lists/"+testutil.UUIDString(list.ID),
		map[string]any{"name": "Updated Name", "description": "New desc", "double_opt_in": false},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if body["name"] != "Updated Name" {
		t.Fatalf("expected name=Updated Name, got %v", body["name"])
	}
}

func TestDeleteList_NonEmpty_Returns409(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Non-empty list")
	sub := testutil.CreateSubscriber(t, db, "member@test.com")

	if err := db.AddSubscriberToList(context.Background(), sqlc.AddSubscriberToListParams{
		ListID:       list.ID,
		SubscriberID: sub.ID,
	}); err != nil {
		t.Fatalf("add subscriber to list: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "DELETE",
		"/api/lists/"+testutil.UUIDString(list.ID),
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "list_not_empty" {
		t.Fatalf("expected error=list_not_empty, got %q", body["error"])
	}
}

func TestDeleteList_ForceDelete(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Force-delete list")
	sub := testutil.CreateSubscriber(t, db, "member2@test.com")

	if err := db.AddSubscriberToList(context.Background(), sqlc.AddSubscriberToListParams{
		ListID:       list.ID,
		SubscriberID: sub.ID,
	}); err != nil {
		t.Fatalf("add subscriber to list: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "DELETE",
		"/api/lists/"+testutil.UUIDString(list.ID)+"?force=true",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNoContent)
}

func TestAddSubscriberToList(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Add-sub list")
	sub := testutil.CreateSubscriber(t, db, "addsub@test.com")

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/lists/"+testutil.UUIDString(list.ID)+"/subscribers",
		map[string]string{"subscriber_id": testutil.UUIDString(sub.ID)},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	count, err := db.CountSubscribersInList(context.Background(), list.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 subscriber in list, got %d", count)
	}
}

func TestAddSubscriberToList_AlreadyMember(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Already-member list")
	sub := testutil.CreateSubscriber(t, db, "already@test.com")

	if err := db.AddSubscriberToList(context.Background(), sqlc.AddSubscriberToListParams{
		ListID:       list.ID,
		SubscriberID: sub.ID,
	}); err != nil {
		t.Fatalf("add subscriber to list: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/lists/"+testutil.UUIDString(list.ID)+"/subscribers",
		map[string]string{"subscriber_id": testutil.UUIDString(sub.ID)},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "already_member" {
		t.Fatalf("expected error=already_member, got %q", body["error"])
	}
}

func TestAddSubscriberToList_InactiveSubscriber(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Inactive-sub list")
	sub := testutil.CreateSubscriber(t, db, "inactive@test.com")

	if err := db.UpdateSubscriberStatus(context.Background(), sqlc.UpdateSubscriberStatusParams{
		ID:     sub.ID,
		Status: "unsubscribed",
	}); err != nil {
		t.Fatalf("update status: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/lists/"+testutil.UUIDString(list.ID)+"/subscribers",
		map[string]string{"subscriber_id": testutil.UUIDString(sub.ID)},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "subscriber_inactive" {
		t.Fatalf("expected error=subscriber_inactive, got %q", body["error"])
	}
}

func TestRemoveSubscriberFromList(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Remove-sub list")
	sub := testutil.CreateSubscriber(t, db, "removesub@test.com")

	if err := db.AddSubscriberToList(context.Background(), sqlc.AddSubscriberToListParams{
		ListID:       list.ID,
		SubscriberID: sub.ID,
	}); err != nil {
		t.Fatalf("add subscriber to list: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "DELETE",
		fmt.Sprintf("/api/lists/%s/subscribers/%s", testutil.UUIDString(list.ID), testutil.UUIDString(sub.ID)),
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNoContent)

	count, err := db.CountSubscribersInList(context.Background(), list.ID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 subscribers after remove, got %d", count)
	}
}

func TestListSubscribersInList_Paginated(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Paged list")

	for i := 0; i < 5; i++ {
		sub := testutil.CreateSubscriber(t, db, fmt.Sprintf("paged%d@test.com", i))
		if err := db.AddSubscriberToList(context.Background(), sqlc.AddSubscriberToListParams{
			ListID:       list.ID,
			SubscriberID: sub.ID,
		}); err != nil {
			t.Fatalf("add subscriber %d: %v", i, err)
		}
	}

	resp := testutil.MakeRequest(t, srv, "GET",
		"/api/lists/"+testutil.UUIDString(list.ID)+"/subscribers?per_page=2&page=1",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	subs := body["subscribers"].([]any)
	if len(subs) != 2 {
		t.Fatalf("expected 2 subscribers on page 1, got %d", len(subs))
	}
	if int(body["total"].(float64)) != 5 {
		t.Fatalf("expected total=5, got %v", body["total"])
	}
}

func TestDoubleOptIn_ConfirmationFlow(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	sub := testutil.CreateSubscriber(t, db, "doi@test.com")

	list, err := db.CreateList(context.Background(), sqlc.CreateListParams{
		Name:        "DOI List",
		Description: "",
		DoubleOptIn: true,
	})
	if err != nil {
		t.Fatalf("create doi list: %v", err)
	}

	// Adding to a double opt-in list sets subscriber status to pending.
	addResp := testutil.MakeRequest(t, srv, "POST",
		"/api/lists/"+testutil.UUIDString(list.ID)+"/subscribers",
		map[string]string{"subscriber_id": testutil.UUIDString(sub.ID)},
		testutil.AuthHeader(token))
	defer addResp.Body.Close()

	testutil.AssertStatus(t, addResp, http.StatusOK)

	updated, err := db.GetSubscriberByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if updated.Status != "pending" {
		t.Fatalf("expected status=pending after DOI add, got %q", updated.Status)
	}

	// Generate a valid confirmation token using the same secret the server uses.
	cm := mailer.NewConfirmationMailer(db, srv.URL, testutil.TestAppSecret)
	tok := cm.GenerateToken(sub.ID, list.ID)

	confirmURL := fmt.Sprintf("/confirm?token=%s&sid=%s&lid=%s",
		tok, testutil.UUIDString(sub.ID), testutil.UUIDString(list.ID))
	confirmResp := testutil.MakeRequest(t, srv, "GET", confirmURL, nil, nil)
	defer confirmResp.Body.Close()

	testutil.AssertStatus(t, confirmResp, http.StatusOK)

	confirmed, err := db.GetSubscriberByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("get subscriber after confirm: %v", err)
	}
	if confirmed.Status != "active" {
		t.Fatalf("expected status=active after confirmation, got %q", confirmed.Status)
	}
}
