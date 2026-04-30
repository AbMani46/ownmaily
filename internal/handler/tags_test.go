package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

func TestCreateTag_Success(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/tags",
		map[string]string{"name": "vip"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusCreated)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if body["name"] != "vip" {
		t.Fatalf("expected name=vip, got %v", body["name"])
	}
}

func TestCreateTag_Duplicate(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	testutil.CreateTag(t, db, "existing-tag")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/tags",
		map[string]string{"name": "existing-tag"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "duplicate" {
		t.Fatalf("expected error=duplicate, got %q", body["error"])
	}
}

func TestUpdateTag_NameConflict(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	tagA := testutil.CreateTag(t, db, "tag-a")
	testutil.CreateTag(t, db, "tag-b")

	// Try renaming tag-a to tag-b.
	resp := testutil.MakeRequest(t, srv, "PUT",
		"/api/tags/"+testutil.UUIDString(tagA.ID),
		map[string]string{"name": "tag-b"},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "duplicate" {
		t.Fatalf("expected error=duplicate, got %q", body["error"])
	}
}

func TestDeleteTag_CascadesSubscriberTags(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "cascade@test.com")
	tag := testutil.CreateTag(t, db, "cascade-tag")

	if err := db.AddTagToSubscriber(context.Background(), sqlc.AddTagToSubscriberParams{
		SubscriberID: sub.ID,
		TagID:        tag.ID,
	}); err != nil {
		t.Fatalf("add tag to subscriber: %v", err)
	}

	// Delete the tag — subscriber_tags cascade.
	delResp := testutil.MakeRequest(t, srv, "DELETE",
		"/api/tags/"+testutil.UUIDString(tag.ID),
		nil, testutil.AuthHeader(token))
	defer delResp.Body.Close()

	testutil.AssertStatus(t, delResp, http.StatusNoContent)

	// Subscriber still exists.
	got, err := db.GetSubscriberByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("subscriber should still exist after tag delete: %v", err)
	}
	if got.Email != "cascade@test.com" {
		t.Fatalf("expected cascade@test.com, got %v", got.Email)
	}

	// Subscriber's tag list is now empty.
	tags, err := db.ListTagsForSubscriber(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after cascade delete, got %d", len(tags))
	}
}

func TestAddTagToSubscriber(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "tagged@test.com")
	tag := testutil.CreateTag(t, db, "new-tag")

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/subscribers/"+testutil.UUIDString(sub.ID)+"/tags",
		map[string]string{"tag_id": testutil.UUIDString(tag.ID)},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	tags, err := db.ListTagsForSubscriber(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}
	if tags[0].Name != "new-tag" {
		t.Fatalf("expected new-tag, got %q", tags[0].Name)
	}
}

func TestAddTagToSubscriber_AlreadyTagged(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "alreadytagged@test.com")
	tag := testutil.CreateTag(t, db, "dup-tag")

	if err := db.AddTagToSubscriber(context.Background(), sqlc.AddTagToSubscriberParams{
		SubscriberID: sub.ID,
		TagID:        tag.ID,
	}); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/subscribers/"+testutil.UUIDString(sub.ID)+"/tags",
		map[string]string{"tag_id": testutil.UUIDString(tag.ID)},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "already_tagged" {
		t.Fatalf("expected error=already_tagged, got %q", body["error"])
	}
}

func TestRemoveTagFromSubscriber(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	sub := testutil.CreateSubscriber(t, db, "removetag@test.com")
	tag := testutil.CreateTag(t, db, "remove-me")

	if err := db.AddTagToSubscriber(context.Background(), sqlc.AddTagToSubscriberParams{
		SubscriberID: sub.ID,
		TagID:        tag.ID,
	}); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "DELETE",
		fmt.Sprintf("/api/subscribers/%s/tags/%s", testutil.UUIDString(sub.ID), testutil.UUIDString(tag.ID)),
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusNoContent)

	tags, err := db.ListTagsForSubscriber(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after remove, got %d", len(tags))
	}
}

func TestBulkTag_Add(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	tag := testutil.CreateTag(t, db, "bulk-add-tag")

	subIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		sub := testutil.CreateSubscriber(t, db, fmt.Sprintf("bulkadd%d@test.com", i))
		subIDs[i] = testutil.UUIDString(sub.ID)
	}

	resp := testutil.MakeRequest(t, srv, "POST", "/api/subscribers/bulk-tag",
		map[string]any{
			"subscriber_ids": subIDs,
			"tag_id":         testutil.UUIDString(tag.ID),
			"action":         "add",
		},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if int(body["processed"].(float64)) != 3 {
		t.Fatalf("expected processed=3, got %v", body["processed"])
	}
}

func TestBulkTag_Remove(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	tag := testutil.CreateTag(t, db, "bulk-remove-tag")

	subIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		sub := testutil.CreateSubscriber(t, db, fmt.Sprintf("bulkrem%d@test.com", i))
		subIDs[i] = testutil.UUIDString(sub.ID)
		if err := db.AddTagToSubscriber(context.Background(), sqlc.AddTagToSubscriberParams{
			SubscriberID: sub.ID,
			TagID:        tag.ID,
		}); err != nil {
			t.Fatalf("pre-tag subscriber %d: %v", i, err)
		}
	}

	resp := testutil.MakeRequest(t, srv, "POST", "/api/subscribers/bulk-tag",
		map[string]any{
			"subscriber_ids": subIDs,
			"tag_id":         testutil.UUIDString(tag.ID),
			"action":         "remove",
		},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if int(body["processed"].(float64)) != 3 {
		t.Fatalf("expected processed=3, got %v", body["processed"])
	}

	// Filter by tag should return 0 subscribers.
	filterResp := testutil.MakeRequest(t, srv, "GET",
		"/api/subscribers?tag_id="+testutil.UUIDString(tag.ID),
		nil, testutil.AuthHeader(token))
	defer filterResp.Body.Close()

	testutil.AssertStatus(t, filterResp, http.StatusOK)

	var filterBody map[string]any
	testutil.DecodeJSON(t, filterResp, &filterBody)
	subs := filterBody["subscribers"].([]any)
	if len(subs) != 0 {
		t.Fatalf("expected 0 subscribers after bulk remove, got %d", len(subs))
	}
}

func TestFilterSubscribersByTag(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	tag := testutil.CreateTag(t, db, "filter-tag")

	// Create 5 subscribers, tag only 3.
	for i := 0; i < 5; i++ {
		sub := testutil.CreateSubscriber(t, db, fmt.Sprintf("filter%d@test.com", i))
		if i < 3 {
			if err := db.AddTagToSubscriber(context.Background(), sqlc.AddTagToSubscriberParams{
				SubscriberID: sub.ID,
				TagID:        tag.ID,
			}); err != nil {
				t.Fatalf("tag subscriber %d: %v", i, err)
			}
		}
	}

	resp := testutil.MakeRequest(t, srv, "GET",
		"/api/subscribers?tag_id="+testutil.UUIDString(tag.ID),
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	subs := body["subscribers"].([]any)
	if len(subs) != 3 {
		t.Fatalf("expected 3 tagged subscribers, got %d", len(subs))
	}
}
