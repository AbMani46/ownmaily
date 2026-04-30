package handler_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

// campaignBody is the minimum valid campaign request body.
func campaignBody(listID string) map[string]any {
	return map[string]any{
		"name":         "Test Campaign",
		"subject":      "Hello World",
		"from_name":    "Sender",
		"from_email":   "sender@example.com",
		"html_body":    "<p>Hello</p>",
		"text_body":    "Hello",
		"send_to_type": "list",
		"send_to_id":   listID,
	}
}

func TestCreateCampaign_Success(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Campaign list")

	resp := testutil.MakeRequest(t, srv, "POST", "/api/campaigns",
		campaignBody(testutil.UUIDString(list.ID)),
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusCreated)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if body["name"] != "Test Campaign" {
		t.Fatalf("expected name=Test Campaign, got %v", body["name"])
	}
	if body["status"] != "draft" {
		t.Fatalf("expected status=draft, got %v", body["status"])
	}
}

func TestCreateCampaign_MissingRequiredFields(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	cases := []map[string]any{
		{"subject": "S", "from_name": "N", "from_email": "e@x.com"}, // missing name
		{"name": "N", "from_name": "N", "from_email": "e@x.com"},    // missing subject
	}

	for _, body := range cases {
		resp := testutil.MakeRequest(t, srv, "POST", "/api/campaigns", body, testutil.AuthHeader(token))
		testutil.AssertStatus(t, resp, http.StatusBadRequest)
		resp.Body.Close()
	}
}

func TestUpdateCampaign_Locked_WhenSent(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Sent campaign list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	// Mark campaign as sent directly in DB.
	if err := db.UpdateCampaignStatus(context.Background(), sqlc.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "sent",
	}); err != nil {
		t.Fatalf("update status: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "PUT",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID),
		campaignBody(testutil.UUIDString(list.ID)),
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "campaign_locked" {
		t.Fatalf("expected error=campaign_locked, got %q", body["error"])
	}
}

func TestScheduleCampaign_Success(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Schedule list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	futureTime := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/schedule",
		map[string]string{"scheduled_at": futureTime},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if body["status"] != "scheduled" {
		t.Fatalf("expected status=scheduled, got %v", body["status"])
	}
	if body["scheduled_at"] == nil {
		t.Fatal("expected scheduled_at to be set")
	}
}

func TestScheduleCampaign_PastTime(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Past schedule list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	pastTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/schedule",
		map[string]string{"scheduled_at": pastTime},
		testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "invalid_time" {
		t.Fatalf("expected error=invalid_time, got %q", body["error"])
	}
}

func TestCancelCampaign(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Cancel list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	// Schedule it first.
	futureTime := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	schedResp := testutil.MakeRequest(t, srv, "POST",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/schedule",
		map[string]string{"scheduled_at": futureTime},
		testutil.AuthHeader(token))
	testutil.AssertStatus(t, schedResp, http.StatusOK)
	schedResp.Body.Close()

	// Cancel it.
	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/cancel",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	if body["status"] != "draft" {
		t.Fatalf("expected status=draft after cancel, got %v", body["status"])
	}
	if body["scheduled_at"] != nil {
		t.Fatalf("expected scheduled_at=null after cancel, got %v", body["scheduled_at"])
	}
}

func TestDuplicateCampaign(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Dup campaign list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/duplicate",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusCreated)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)
	expectedName := "Copy of " + campaign.Name
	if body["name"] != expectedName {
		t.Fatalf("expected name=%q, got %q", expectedName, body["name"])
	}
	if body["status"] != "draft" {
		t.Fatalf("expected status=draft for duplicate, got %v", body["status"])
	}
}

func TestDeleteCampaign_WhenSending_Returns409(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Sending delete list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	if err := db.UpdateCampaignStatus(context.Background(), sqlc.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "sending",
	}); err != nil {
		t.Fatalf("update status: %v", err)
	}

	resp := testutil.MakeRequest(t, srv, "DELETE",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID),
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusConflict)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "campaign_locked" {
		t.Fatalf("expected error=campaign_locked, got %q", body["error"])
	}
}

func TestCampaignPreview(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Preview list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	resp := testutil.MakeRequest(t, srv, "GET",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/preview",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("expected text/html Content-Type, got %q", ct)
	}
}

func TestCampaignStats_ZerosForNewCampaign(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")
	list := testutil.CreateList(t, db, "Stats list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	resp := testutil.MakeRequest(t, srv, "GET",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/stats",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	for _, field := range []string{"sent", "failed", "opens", "clicks"} {
		if int(body[field].(float64)) != 0 {
			t.Fatalf("expected %s=0 for new campaign, got %v", field, body[field])
		}
	}

	links, ok := body["links"].([]any)
	if !ok {
		t.Fatalf("expected links to be an array, got %T", body["links"])
	}
	if len(links) != 0 {
		t.Fatalf("expected empty links for new campaign, got %d", len(links))
	}
}

func TestSendCampaign_NoRecipients(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	// Empty list — no active subscribers.
	list := testutil.CreateList(t, db, "Empty send list")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	resp := testutil.MakeRequest(t, srv, "POST",
		"/api/campaigns/"+testutil.UUIDString(campaign.ID)+"/send",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)

	var body map[string]string
	testutil.DecodeJSON(t, resp, &body)
	if body["error"] != "no_recipients" {
		t.Fatalf("expected error=no_recipients, got %q", body["error"])
	}
}
