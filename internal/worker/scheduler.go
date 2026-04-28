package worker

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
)

type SchedulerWorker struct {
	db   *db.Queries
	pool *pgxpool.Pool
}

func NewSchedulerWorker(q *db.Queries, pool *pgxpool.Pool) *SchedulerWorker {
	return &SchedulerWorker{db: q, pool: pool}
}

func (w *SchedulerWorker) Start(ctx context.Context) {
	log.Printf("scheduler worker started")
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *SchedulerWorker) tick(ctx context.Context) {
	campaigns, err := w.db.ListScheduledCampaignsDue(ctx)
	if err != nil {
		log.Printf("scheduler: list scheduled campaigns: %v", err)
		return
	}

	for _, campaign := range campaigns {
		tx, err := w.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			log.Printf("scheduler: begin tx for campaign %s: %v", campaign.ID.String(), err)
			continue
		}

		qtx := w.db.WithTx(tx)

		current, err := qtx.GetCampaignByID(ctx, campaign.ID)
		if err != nil {
			log.Printf("scheduler: re-fetch campaign %s: %v", campaign.ID.String(), err)
			_ = tx.Rollback(ctx)
			continue
		}
		if current.Status != "scheduled" {
			_ = tx.Rollback(ctx)
			continue
		}

		if _, err := qtx.CreateSendJob(ctx, campaign.ID); err != nil {
			log.Printf("scheduler: create send job for campaign %s: %v", campaign.ID.String(), err)
			_ = tx.Rollback(ctx)
			continue
		}

		if err := qtx.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
			ID:     campaign.ID,
			Status: "queued",
		}); err != nil {
			log.Printf("scheduler: update campaign status for %s: %v", campaign.ID.String(), err)
			_ = tx.Rollback(ctx)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			log.Printf("scheduler: commit tx for campaign %s: %v", campaign.ID.String(), err)
			continue
		}

		log.Printf("scheduler: queued campaign %s", campaign.ID.String())
	}
}
