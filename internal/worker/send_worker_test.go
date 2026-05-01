package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/AbMani46/ownmaily/internal/mailer"
	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
)

// failMailer implements mailer.Mailer and always returns an error.
type failMailer struct{}

func (f *failMailer) Send(msg mailer.Message) error { return fmt.Errorf("send failed") }
func (f *failMailer) Name() string                  { return "fail" }

func TestSendWorker_ProcessesJob_EndToEnd(t *testing.T) {
	pool, queries, cleanup := newWorkerTestDB(t)
	t.Cleanup(cleanup)
	_ = pool

	ctx := context.Background()

	list := wCreateList(t, queries, "e2e-list")

	for i := 0; i < 3; i++ {
		sub := wCreateSubscriber(t, queries, fmt.Sprintf("e2e%d@test.com", i))
		if err := queries.AddSubscriberToList(ctx, sqlc.AddSubscriberToListParams{
			ListID:       list.ID,
			SubscriberID: sub.ID,
		}); err != nil {
			t.Fatalf("add subscriber %d to list: %v", i, err)
		}
	}

	campaign := wCreateCampaign(t, queries, list.ID)

	job, err := queries.CreateSendJob(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("create send job: %v", err)
	}

	if err := queries.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "queued",
	}); err != nil {
		t.Fatalf("update campaign to queued: %v", err)
	}

	mailerStore := mailer.NewStore(&mailer.LogMailer{})
	w := NewSendWorker(queries, mailerStore, "http://localhost:4400", workerTestSecret)

	if err := w.processJob(ctx, job); err != nil {
		t.Fatalf("processJob: %v", err)
	}

	// Assert send_job is complete with correct counts.
	gotJob, err := queries.GetSendJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("get send job: %v", err)
	}
	if gotJob.Status != "complete" {
		t.Fatalf("expected job status=complete, got %q", gotJob.Status)
	}
	if gotJob.SentCount != 3 {
		t.Fatalf("expected sent_count=3, got %d", gotJob.SentCount)
	}
	if gotJob.FailedCount != 0 {
		t.Fatalf("expected failed_count=0, got %d", gotJob.FailedCount)
	}

	// Assert campaign is marked sent with sent_at non-null.
	gotCampaign, err := queries.GetCampaignByID(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get campaign: %v", err)
	}
	if gotCampaign.Status != "sent" {
		t.Fatalf("expected campaign status=sent, got %q", gotCampaign.Status)
	}
	if !gotCampaign.SentAt.Valid {
		t.Fatal("expected campaign sent_at to be non-null")
	}

	// Assert all 3 campaign_recipients have status=sent.
	sentCount, err := queries.CountRecipientsByStatus(ctx, sqlc.CountRecipientsByStatusParams{
		CampaignID: campaign.ID,
		Status:     "sent",
	})
	if err != nil {
		t.Fatalf("count recipients sent: %v", err)
	}
	if sentCount != 3 {
		t.Fatalf("expected 3 sent recipients, got %d", sentCount)
	}
}

func TestSendWorker_SkipsSuppressedRecipients(t *testing.T) {
	pool, queries, cleanup := newWorkerTestDB(t)
	t.Cleanup(cleanup)
	_ = pool

	ctx := context.Background()

	list := wCreateList(t, queries, "suppress-list")

	var subs []sqlc.Subscriber
	for i := 0; i < 3; i++ {
		sub := wCreateSubscriber(t, queries, fmt.Sprintf("supsub%d@test.com", i))
		subs = append(subs, sub)
		if err := queries.AddSubscriberToList(ctx, sqlc.AddSubscriberToListParams{
			ListID:       list.ID,
			SubscriberID: sub.ID,
		}); err != nil {
			t.Fatalf("add subscriber %d to list: %v", i, err)
		}
	}

	// Suppress the first subscriber.
	if err := queries.AddSuppression(ctx, sqlc.AddSuppressionParams{
		Email:  subs[0].Email,
		Reason: "manual",
	}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}

	campaign := wCreateCampaign(t, queries, list.ID)

	job, err := queries.CreateSendJob(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("create send job: %v", err)
	}

	if err := queries.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "queued",
	}); err != nil {
		t.Fatalf("update campaign to queued: %v", err)
	}

	mailerStore := mailer.NewStore(&mailer.LogMailer{})
	w := NewSendWorker(queries, mailerStore, "http://localhost:4400", workerTestSecret)

	if err := w.processJob(ctx, job); err != nil {
		t.Fatalf("processJob: %v", err)
	}

	gotJob, err := queries.GetSendJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("get send job: %v", err)
	}
	if gotJob.SentCount != 2 {
		t.Fatalf("expected sent_count=2 (suppressed 1), got %d", gotJob.SentCount)
	}

	// The suppressed subscriber should have no campaign_recipients row.
	_, err = queries.GetRecipientStatus(ctx, sqlc.GetRecipientStatusParams{
		CampaignID:   campaign.ID,
		SubscriberID: subs[0].ID,
	})
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected suppressed subscriber to have no campaign_recipients row, got err=%v", err)
	}
}

func TestSendWorker_HandlesWorkerFailure(t *testing.T) {
	pool, queries, cleanup := newWorkerTestDB(t)
	t.Cleanup(cleanup)
	_ = pool

	ctx := context.Background()

	list := wCreateList(t, queries, "fail-list")

	for i := 0; i < 3; i++ {
		sub := wCreateSubscriber(t, queries, fmt.Sprintf("failsub%d@test.com", i))
		if err := queries.AddSubscriberToList(ctx, sqlc.AddSubscriberToListParams{
			ListID:       list.ID,
			SubscriberID: sub.ID,
		}); err != nil {
			t.Fatalf("add subscriber %d to list: %v", i, err)
		}
	}

	campaign := wCreateCampaign(t, queries, list.ID)

	job, err := queries.CreateSendJob(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("create send job: %v", err)
	}

	if err := queries.UpdateCampaignStatus(ctx, sqlc.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "queued",
	}); err != nil {
		t.Fatalf("update campaign to queued: %v", err)
	}

	mailerStore := mailer.NewStore(&failMailer{})
	w := NewSendWorker(queries, mailerStore, "http://localhost:4400", workerTestSecret)

	if err := w.processJob(ctx, job); err != nil {
		t.Fatalf("processJob: %v", err)
	}

	// Job completes even though all sends failed.
	gotJob, err := queries.GetSendJob(ctx, job.ID)
	if err != nil {
		t.Fatalf("get send job: %v", err)
	}
	if gotJob.Status != "complete" {
		t.Fatalf("expected job status=complete even on send failure, got %q", gotJob.Status)
	}
	if gotJob.FailedCount != 3 {
		t.Fatalf("expected failed_count=3, got %d", gotJob.FailedCount)
	}
	if gotJob.SentCount != 0 {
		t.Fatalf("expected sent_count=0 when all fail, got %d", gotJob.SentCount)
	}

	// Campaign is still marked sent (the job completed).
	gotCampaign, err := queries.GetCampaignByID(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get campaign: %v", err)
	}
	if gotCampaign.Status != "sent" {
		t.Fatalf("expected campaign status=sent after job completes, got %q", gotCampaign.Status)
	}
}
