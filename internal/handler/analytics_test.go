package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
)

func TestAnalyticsOverview_EmptyDB(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	resp := testutil.MakeRequest(t, srv, "GET", "/api/analytics/overview", nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	for _, field := range []string{"total_subscribers", "active_subscribers", "unsubscribed", "bounced",
		"total_campaigns_sent", "total_emails_sent", "total_opens", "total_clicks"} {
		v, ok := body[field].(float64)
		if !ok || v != 0 {
			t.Fatalf("expected %s=0, got %v", field, body[field])
		}
	}
	for _, field := range []string{"overall_open_rate", "overall_click_rate"} {
		v, ok := body[field].(float64)
		if !ok || v != 0.0 {
			t.Fatalf("expected %s=0.0, got %v", field, body[field])
		}
	}
}

func TestAnalyticsOverview_WithData(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	// 3 active, 1 unsubscribed, 1 bounced
	sub1 := testutil.CreateSubscriber(t, db, "a1@example.com")
	sub2 := testutil.CreateSubscriber(t, db, "a2@example.com")
	sub3 := testutil.CreateSubscriber(t, db, "a3@example.com")
	subU := testutil.CreateSubscriber(t, db, "unsub@example.com")
	subB := testutil.CreateSubscriber(t, db, "bounced@example.com")

	ctx := context.Background()
	db.UpdateSubscriberStatus(ctx, sqlc.UpdateSubscriberStatusParams{ID: subU.ID, Status: "unsubscribed"})
	db.UpdateSubscriberStatus(ctx, sqlc.UpdateSubscriberStatusParams{ID: subB.ID, Status: "bounced"})

	// Create 2 campaigns, mark them sent.
	list := testutil.CreateList(t, db, "Overview List")
	c1 := testutil.CreateCampaign(t, db, list.ID)
	c2 := testutil.CreateCampaign(t, db, list.ID)

	db.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{ID: c1.ID, Status: "sent"})
	db.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{ID: c2.ID, Status: "sent"})

	// Seed 3 campaign recipients as sent for each campaign.
	sentAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	for _, sub := range []sqlc.Subscriber{sub1, sub2, sub3} {
		for _, camp := range []sqlc.Campaign{c1, c2} {
			r, err := db.CreateCampaignRecipient(ctx, sqlc.CreateCampaignRecipientParams{
				CampaignID:   camp.ID,
				SubscriberID: sub.ID,
			})
			if err != nil {
				t.Fatalf("create recipient: %v", err)
			}
			db.UpdateRecipientStatus(ctx, sqlc.UpdateRecipientStatusParams{
				CampaignID:   r.CampaignID,
				SubscriberID: r.SubscriberID,
				Status:       "sent",
				SentAt:       sentAt,
			})
		}
	}

	resp := testutil.MakeRequest(t, srv, "GET", "/api/analytics/overview", nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if v := int(body["active_subscribers"].(float64)); v != 3 {
		t.Fatalf("expected active_subscribers=3, got %d", v)
	}
	if v := int(body["unsubscribed"].(float64)); v != 1 {
		t.Fatalf("expected unsubscribed=1, got %d", v)
	}
	if v := int(body["bounced"].(float64)); v != 1 {
		t.Fatalf("expected bounced=1, got %d", v)
	}
	if v := int(body["total_campaigns_sent"].(float64)); v != 2 {
		t.Fatalf("expected total_campaigns_sent=2, got %d", v)
	}
	if v := int(body["total_emails_sent"].(float64)); v != 6 {
		t.Fatalf("expected total_emails_sent=6 (3 subs × 2 campaigns), got %d", v)
	}
}

func TestAnalyticsOverview_RatesCalculation(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	ctx := context.Background()
	list := testutil.CreateList(t, db, "Rates List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	sentAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}

	// Seed 10 sent recipients, 4 opens, 2 clicks.
	var subs [10]sqlc.Subscriber
	for i := range 10 {
		email := "rate" + string(rune('a'+i)) + "@example.com"
		subs[i] = testutil.CreateSubscriber(t, db, email)
		r, err := db.CreateCampaignRecipient(ctx, sqlc.CreateCampaignRecipientParams{
			CampaignID:   campaign.ID,
			SubscriberID: subs[i].ID,
		})
		if err != nil {
			t.Fatalf("create recipient: %v", err)
		}
		db.UpdateRecipientStatus(ctx, sqlc.UpdateRecipientStatusParams{
			CampaignID:   r.CampaignID,
			SubscriberID: r.SubscriberID,
			Status:       "sent",
			SentAt:       sentAt,
		})
	}

	// 4 opens
	for i := range 4 {
		db.RecordOpen(ctx, sqlc.RecordOpenParams{
			CampaignID:   campaign.ID,
			SubscriberID: subs[i].ID,
		})
	}

	// 2 clicks
	for i := range 2 {
		db.RecordClick(ctx, sqlc.RecordClickParams{
			CampaignID:   campaign.ID,
			SubscriberID: subs[i].ID,
			LinkIndex:    0,
			LinkUrl:      "https://example.com",
		})
	}

	resp := testutil.MakeRequest(t, srv, "GET", "/api/analytics/overview", nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	openRate := body["overall_open_rate"].(float64)
	if openRate != 0.4 {
		t.Fatalf("expected overall_open_rate=0.4, got %f", openRate)
	}
	clickRate := body["overall_click_rate"].(float64)
	if clickRate != 0.2 {
		t.Fatalf("expected overall_click_rate=0.2, got %f", clickRate)
	}
}

func TestSubscriberStats_NoActivity(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	sub := testutil.CreateSubscriber(t, db, "noact@example.com")

	resp := testutil.MakeRequest(t, srv, "GET", "/api/subscribers/"+testutil.UUIDString(sub.ID)+"/stats",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if v := int(body["campaigns_received"].(float64)); v != 0 {
		t.Fatalf("expected campaigns_received=0, got %d", v)
	}
	if v := int(body["total_opens"].(float64)); v != 0 {
		t.Fatalf("expected total_opens=0, got %d", v)
	}
	if v := int(body["total_clicks"].(float64)); v != 0 {
		t.Fatalf("expected total_clicks=0, got %d", v)
	}
	if body["last_active"] != nil {
		t.Fatalf("expected last_active=nil, got %v", body["last_active"])
	}
}

func TestSubscriberStats_WithActivity(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	token := testutil.LoginAndGetToken(t, srv, "owner@test.com", "testpass123")

	sub := testutil.CreateSubscriber(t, db, "active@example.com")
	list := testutil.CreateList(t, db, "Active List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	ctx := context.Background()

	// Create a sent campaign_recipient so it shows in campaigns_received.
	sentAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	r, err := db.CreateCampaignRecipient(ctx, sqlc.CreateCampaignRecipientParams{
		CampaignID:   campaign.ID,
		SubscriberID: sub.ID,
	})
	if err != nil {
		t.Fatalf("create recipient: %v", err)
	}
	db.UpdateRecipientStatus(ctx, sqlc.UpdateRecipientStatusParams{
		CampaignID:   r.CampaignID,
		SubscriberID: r.SubscriberID,
		Status:       "sent",
		SentAt:       sentAt,
	})

	// Seed 1 open and 1 click.
	db.RecordOpen(ctx, sqlc.RecordOpenParams{
		CampaignID:   campaign.ID,
		SubscriberID: sub.ID,
	})
	db.RecordClick(ctx, sqlc.RecordClickParams{
		CampaignID:   campaign.ID,
		SubscriberID: sub.ID,
		LinkIndex:    0,
		LinkUrl:      "https://example.com/link",
	})

	resp := testutil.MakeRequest(t, srv, "GET", "/api/subscribers/"+testutil.UUIDString(sub.ID)+"/stats",
		nil, testutil.AuthHeader(token))
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	var body map[string]any
	testutil.DecodeJSON(t, resp, &body)

	if v := int(body["campaigns_received"].(float64)); v != 1 {
		t.Fatalf("expected campaigns_received=1, got %d", v)
	}
	if v := int(body["total_opens"].(float64)); v != 1 {
		t.Fatalf("expected total_opens=1, got %d", v)
	}
	if v := int(body["total_clicks"].(float64)); v != 1 {
		t.Fatalf("expected total_clicks=1, got %d", v)
	}
	if body["last_active"] == nil {
		t.Fatal("expected last_active to be non-null")
	}
}
