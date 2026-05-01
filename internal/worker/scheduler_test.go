package worker

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
)

func TestSchedulerWorker_QueuesOverdueCampaign(t *testing.T) {
	pool, queries, cleanup := newWorkerTestDB(t)
	t.Cleanup(cleanup)

	ctx := context.Background()

	list := wCreateList(t, queries, "sched-overdue-list")
	campaign := wCreateCampaign(t, queries, list.ID)

	// Schedule in the past — scheduler must pick it up.
	pastTime := pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Minute), Valid: true}
	if _, err := queries.ScheduleCampaign(ctx, sqlc.ScheduleCampaignParams{
		ID:          campaign.ID,
		ScheduledAt: pastTime,
	}); err != nil {
		t.Fatalf("schedule campaign: %v", err)
	}

	sched := NewSchedulerWorker(queries, pool)
	sched.tick(ctx)

	// Campaign should now be queued.
	got, err := queries.GetCampaignByID(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get campaign: %v", err)
	}
	if got.Status != "queued" {
		t.Fatalf("expected campaign status=queued, got %q", got.Status)
	}

	// A send_jobs row should exist.
	job, err := queries.GetSendJobByCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get send job: %v", err)
	}
	if job.Status != "pending" {
		t.Fatalf("expected send_job status=pending, got %q", job.Status)
	}
}

func TestSchedulerWorker_IgnoresFutureCampaign(t *testing.T) {
	pool, queries, cleanup := newWorkerTestDB(t)
	t.Cleanup(cleanup)

	ctx := context.Background()

	list := wCreateList(t, queries, "sched-future-list")
	campaign := wCreateCampaign(t, queries, list.ID)

	// Schedule in the future — scheduler must leave it alone.
	futureTime := pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}
	if _, err := queries.ScheduleCampaign(ctx, sqlc.ScheduleCampaignParams{
		ID:          campaign.ID,
		ScheduledAt: futureTime,
	}); err != nil {
		t.Fatalf("schedule campaign: %v", err)
	}

	sched := NewSchedulerWorker(queries, pool)
	sched.tick(ctx)

	got, err := queries.GetCampaignByID(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get campaign: %v", err)
	}
	if got.Status != "scheduled" {
		t.Fatalf("expected campaign status still=scheduled, got %q", got.Status)
	}

	// No send_jobs row should exist.
	_, err = queries.GetSendJobByCampaign(ctx, campaign.ID)
	if err == nil {
		t.Fatal("expected no send_jobs row for a future-scheduled campaign")
	}
}

func TestSchedulerWorker_IdempotentOnDoubleRun(t *testing.T) {
	pool, queries, cleanup := newWorkerTestDB(t)
	t.Cleanup(cleanup)

	ctx := context.Background()

	list := wCreateList(t, queries, "sched-idem-list")
	campaign := wCreateCampaign(t, queries, list.ID)

	pastTime := pgtype.Timestamptz{Time: time.Now().Add(-2 * time.Minute), Valid: true}
	if _, err := queries.ScheduleCampaign(ctx, sqlc.ScheduleCampaignParams{
		ID:          campaign.ID,
		ScheduledAt: pastTime,
	}); err != nil {
		t.Fatalf("schedule campaign: %v", err)
	}

	sched := NewSchedulerWorker(queries, pool)

	// First tick queues the campaign.
	sched.tick(ctx)

	// Second tick sees status=queued and skips; transaction re-fetch prevents re-scheduling.
	sched.tick(ctx)

	// Count send_jobs for this campaign — should be exactly 1.
	// Use GetSendJobByCampaign to fetch the latest (only one should exist).
	job, err := queries.GetSendJobByCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get send job: %v", err)
	}
	if job.CampaignID != campaign.ID {
		t.Fatalf("unexpected campaign id in job")
	}

	// Verify the campaign status is queued (not re-scheduled or doubled).
	got, err := queries.GetCampaignByID(ctx, campaign.ID)
	if err != nil {
		t.Fatalf("get campaign: %v", err)
	}
	if got.Status != "queued" {
		t.Fatalf("expected status=queued after double tick, got %q", got.Status)
	}

	// Verify only 1 send_job row exists by listing pending jobs (or checking the count).
	// The easiest check: ListPendingSendJobs returns exactly 1 job for this campaign.
	pending, err := queries.ListPendingSendJobs(ctx)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	count := 0
	for _, j := range pending {
		if j.CampaignID == campaign.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 pending job for campaign, got %d", count)
	}
}
