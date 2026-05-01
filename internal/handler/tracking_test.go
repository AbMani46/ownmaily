package handler_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/testutil"
	"github.com/AbMani46/ownmaily/internal/tracking"
)

func TestOpenPixel_ValidToken_RecordsOpen(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "track@example.com")
	list := testutil.CreateList(t, db, "Track List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	token := tracking.GenerateOpenToken(sub.ID, campaign.ID, testutil.TestAppSecret)

	resp, err := http.Get(srv.URL + "/track/open/" + token)
	if err != nil {
		t.Fatalf("GET open pixel: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	opens, err := db.CountOpensByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("count opens: %v", err)
	}
	if opens != 1 {
		t.Fatalf("expected 1 open row, got %d", opens)
	}
}

func TestOpenPixel_InvalidToken_StillReturnsGIF(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	resp, err := http.Get(srv.URL + "/track/open/invalid.token")
	if err != nil {
		t.Fatalf("GET open pixel: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)
	if resp.Header.Get("Content-Type") != "image/gif" {
		t.Fatalf("expected image/gif, got %s", resp.Header.Get("Content-Type"))
	}
}

func TestOpenPixel_Idempotent(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "track-idem@example.com")
	list := testutil.CreateList(t, db, "Track Idem List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	token := tracking.GenerateOpenToken(sub.ID, campaign.ID, testutil.TestAppSecret)

	for i := 0; i < 2; i++ {
		resp, err := http.Get(srv.URL + "/track/open/" + token)
		if err != nil {
			t.Fatalf("hit %d: GET open pixel: %v", i, err)
		}
		resp.Body.Close()
	}

	opens, err := db.CountOpensByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("count opens: %v", err)
	}
	if opens != 1 {
		t.Fatalf("expected exactly 1 open row after 2 hits, got %d", opens)
	}
}

func TestOpenPixel_ReturnsGIF(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "track-gif@example.com")
	list := testutil.CreateList(t, db, "GIF List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	token := tracking.GenerateOpenToken(sub.ID, campaign.ID, testutil.TestAppSecret)

	resp, err := http.Get(srv.URL + "/track/open/" + token)
	if err != nil {
		t.Fatalf("GET open pixel: %v", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "image/gif" {
		t.Fatalf("expected Content-Type image/gif, got %s", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) != 43 {
		t.Fatalf("expected 43 byte GIF, got %d bytes", len(body))
	}
}

func TestClickRedirect_ValidToken_RecordsClick(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "click@example.com")
	list := testutil.CreateList(t, db, "Click List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	dest := "https://example.com/landing"
	token := tracking.GenerateClickToken(sub.ID, campaign.ID, 0, testutil.TestAppSecret)

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(srv.URL + "/track/click/" + token + "?url=" + dest)
	if err != nil {
		t.Fatalf("GET click: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusFound)
	if loc := resp.Header.Get("Location"); loc != dest {
		t.Fatalf("expected Location %s, got %s", dest, loc)
	}

	clicks, err := db.CountClicksByCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("count clicks: %v", err)
	}
	if clicks != 1 {
		t.Fatalf("expected 1 click row, got %d", clicks)
	}
}

func TestClickRedirect_InvalidToken_StillRedirects(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	dest := "https://example.com/landing"

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	resp, err := client.Get(srv.URL + "/track/click/invalid.token?url=" + dest)
	if err != nil {
		t.Fatalf("GET click: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusFound)
	if loc := resp.Header.Get("Location"); loc != dest {
		t.Fatalf("expected redirect to %s, got %s", dest, loc)
	}
}

func TestClickRedirect_MissingURL_Returns400(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "click-nourl@example.com")
	list := testutil.CreateList(t, db, "No URL List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	token := tracking.GenerateClickToken(sub.ID, campaign.ID, 0, testutil.TestAppSecret)

	resp, err := http.Get(srv.URL + "/track/click/" + token)
	if err != nil {
		t.Fatalf("GET click: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)
}

func TestUnsubscribe_ValidToken(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "unsub@example.com")
	list := testutil.CreateList(t, db, "Unsub List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	token := tracking.GenerateUnsubscribeToken(sub.ID, campaign.ID, testutil.TestAppSecret)

	resp, err := http.Get(srv.URL + "/unsubscribe?token=" + token)
	if err != nil {
		t.Fatalf("GET unsubscribe: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)
	if ct := resp.Header.Get("Content-Type"); ct == "" || ct[:9] != "text/html" {
		t.Fatalf("expected text/html, got %s", ct)
	}

	updated, err := db.GetSubscriberByID(context.Background(), sub.ID)
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if updated.Status != "unsubscribed" {
		t.Fatalf("expected status unsubscribed, got %s", updated.Status)
	}

	suppressed, err := db.IsSuppressed(context.Background(), sub.Email)
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if !suppressed {
		t.Fatal("expected suppression row to exist")
	}
}

func TestUnsubscribe_InvalidToken_Returns400HTML(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	resp, err := http.Get(srv.URL + "/unsubscribe?token=bad.token")
	if err != nil {
		t.Fatalf("GET unsubscribe: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusBadRequest)
	ct := resp.Header.Get("Content-Type")
	if ct == "" || ct[:9] != "text/html" {
		t.Fatalf("expected text/html, got %s", ct)
	}
}

func TestUnsubscribe_Idempotent(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "unsub-idem@example.com")
	list := testutil.CreateList(t, db, "Unsub Idem List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	token := tracking.GenerateUnsubscribeToken(sub.ID, campaign.ID, testutil.TestAppSecret)

	for i := 0; i < 2; i++ {
		resp, err := http.Get(srv.URL + "/unsubscribe?token=" + token)
		if err != nil {
			t.Fatalf("hit %d GET unsubscribe: %v", i, err)
		}
		resp.Body.Close()
		testutil.AssertStatus(t, resp, http.StatusOK)
	}

	count, err := db.CountSuppressions(context.Background())
	if err != nil {
		t.Fatalf("count suppressions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 suppression row, got %d", count)
	}
}

func TestUnsubscribe_AlreadyUnsubscribed_Returns200(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateOwner(t, db)
	sub := testutil.CreateSubscriber(t, db, "already-unsub@example.com")
	list := testutil.CreateList(t, db, "Already Unsub List")
	campaign := testutil.CreateCampaign(t, db, list.ID)

	// Pre-set status to unsubscribed
	if err := db.UpdateSubscriberStatus(context.Background(), sqlc.UpdateSubscriberStatusParams{
		ID:     sub.ID,
		Status: "unsubscribed",
	}); err != nil {
		t.Fatalf("update subscriber status: %v", err)
	}

	token := tracking.GenerateUnsubscribeToken(sub.ID, campaign.ID, testutil.TestAppSecret)

	resp, err := http.Get(srv.URL + "/unsubscribe?token=" + token)
	if err != nil {
		t.Fatalf("GET unsubscribe: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)
}
