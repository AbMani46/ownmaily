package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/AbMani46/ownmaily/internal/testutil"
)

func TestResendWebhook_HardBounce(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "bounce@example.com")

	payload := `{"type":"email.bounced","data":{"email_id":"abc","from":"from@x.com","to":["bounce@example.com"],"bounce_type":"hard"}}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/resend", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	suppressed, err := db.IsSuppressed(context.Background(), "bounce@example.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if !suppressed {
		t.Fatal("expected bounce@example.com to be suppressed")
	}

	sub, err := db.GetSubscriberByEmail(context.Background(), "bounce@example.com")
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if sub.Status != "bounced" {
		t.Fatalf("expected status bounced, got %s", sub.Status)
	}
}

func TestResendWebhook_SoftBounce_NoSuppression(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "soft@example.com")

	payload := `{"type":"email.bounced","data":{"email_id":"abc","from":"from@x.com","to":["soft@example.com"],"bounce_type":"soft"}}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/resend", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	suppressed, err := db.IsSuppressed(context.Background(), "soft@example.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if suppressed {
		t.Fatal("soft bounce should NOT add to suppression list")
	}
}

func TestResendWebhook_Complaint(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "complaint@example.com")

	payload := `{"type":"email.complained","data":{"email_id":"abc","from":"from@x.com","to":["complaint@example.com"]}}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/resend", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	sup, err := db.GetSuppression(context.Background(), "complaint@example.com")
	if err != nil {
		t.Fatalf("get suppression: %v", err)
	}
	if sup.Reason != "complained" {
		t.Fatalf("expected reason complained, got %s", sup.Reason)
	}
}

func TestResendWebhook_UnknownEvent_Returns200(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	payload := `{"type":"email.opened","data":{"email_id":"abc"}}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/resend", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	count, err := db.CountSuppressions(context.Background())
	if err != nil {
		t.Fatalf("count suppressions: %v", err)
	}
	if count != 0 {
		t.Fatalf("unknown event should not create suppressions, got %d", count)
	}
}

func TestMailgunWebhook_HardBounce(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "mg-bounce@example.com")

	body := "event=bounced&recipient=mg-bounce@example.com&severity=hard"
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/mailgun", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	suppressed, err := db.IsSuppressed(context.Background(), "mg-bounce@example.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if !suppressed {
		t.Fatal("expected mg-bounce@example.com to be suppressed")
	}

	sub, err := db.GetSubscriberByEmail(context.Background(), "mg-bounce@example.com")
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if sub.Status != "bounced" {
		t.Fatalf("expected status bounced, got %s", sub.Status)
	}
}

func TestMailgunWebhook_Complaint(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "mg-complaint@example.com")

	body := "event=complained&recipient=mg-complaint@example.com"
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/mailgun", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	sup, err := db.GetSuppression(context.Background(), "mg-complaint@example.com")
	if err != nil {
		t.Fatalf("get suppression: %v", err)
	}
	if sup.Reason != "complained" {
		t.Fatalf("expected reason complained, got %s", sup.Reason)
	}
}

func TestMailgunWebhook_SoftBounce_NoSuppression(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "mg-soft@example.com")

	body := "event=bounced&recipient=mg-soft@example.com&severity=soft"
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/mailgun", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	suppressed, err := db.IsSuppressed(context.Background(), "mg-soft@example.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if suppressed {
		t.Fatal("mailgun soft bounce should NOT add to suppression list")
	}
}

func TestSESWebhook_HardBounce(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "ses-bounce@example.com")

	payload := `{"Type":"Notification","Message":"{\"notificationType\":\"Bounce\",\"bounce\":{\"bounceType\":\"Permanent\",\"bouncedRecipients\":[{\"emailAddress\":\"ses-bounce@example.com\"}]}}"}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/ses", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	suppressed, err := db.IsSuppressed(context.Background(), "ses-bounce@example.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if !suppressed {
		t.Fatal("expected ses-bounce@example.com to be suppressed")
	}

	sub, err := db.GetSubscriberByEmail(context.Background(), "ses-bounce@example.com")
	if err != nil {
		t.Fatalf("get subscriber: %v", err)
	}
	if sub.Status != "bounced" {
		t.Fatalf("expected status bounced, got %s", sub.Status)
	}
}

func TestSESWebhook_Complaint(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "ses-complaint@example.com")

	payload := `{"Type":"Notification","Message":"{\"notificationType\":\"Complaint\",\"complaint\":{\"complainedRecipients\":[{\"emailAddress\":\"ses-complaint@example.com\"}]}}"}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/ses", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	sup, err := db.GetSuppression(context.Background(), "ses-complaint@example.com")
	if err != nil {
		t.Fatalf("get suppression: %v", err)
	}
	if sup.Reason != "complained" {
		t.Fatalf("expected reason complained, got %s", sup.Reason)
	}
}

func TestSESWebhook_SoftBounce_NoSuppression(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	testutil.CreateSubscriber(t, db, "ses-soft@example.com")

	payload := `{"Type":"Notification","Message":"{\"notificationType\":\"Bounce\",\"bounce\":{\"bounceType\":\"Transient\",\"bouncedRecipients\":[{\"emailAddress\":\"ses-soft@example.com\"}]}}"}`
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/ses", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)

	suppressed, err := db.IsSuppressed(context.Background(), "ses-soft@example.com")
	if err != nil {
		t.Fatalf("check suppression: %v", err)
	}
	if suppressed {
		t.Fatal("SES soft (Transient) bounce should NOT add to suppression list")
	}
}

func TestSESWebhook_SubscriptionConfirmation(t *testing.T) {
	pool, db, cleanup := testutil.NewTestDB(t)
	t.Cleanup(cleanup)
	srv, srvCleanup := testutil.NewTestServer(t, db, pool)
	t.Cleanup(srvCleanup)

	// SubscribeURL points to /health on our own test server, which returns 200.
	// The test server has no /health route, so we add one inline using the fact
	// that the handler just GETs the URL and ignores the response body.
	// Actually the test server doesn't have /health — use the srv.URL itself.
	// The SES handler does http.Get(SubscribeURL) and ignores errors.
	// We'll point to a path that returns any response.
	subscribeURL := fmt.Sprintf("%s/api/auth/login", srv.URL)

	payload := fmt.Sprintf(`{"Type":"SubscriptionConfirmation","SubscribeURL":"%s"}`, subscribeURL)
	req, _ := http.NewRequest("POST", srv.URL+"/webhooks/ses", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST webhook: %v", err)
	}
	defer resp.Body.Close()

	testutil.AssertStatus(t, resp, http.StatusOK)
}
