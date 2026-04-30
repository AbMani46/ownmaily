package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AbMani46/ownmaily/internal/mailer"
	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/time/rate"
)

type SendWorker struct {
	db              *db.Queries
	mailerStore     *mailer.Store
	installationURL string
	appSecret       string
	limiter         *rate.Limiter
}

func NewSendWorker(q *db.Queries, s *mailer.Store, installationURL, appSecret string) *SendWorker {
	return &SendWorker{
		db:              q,
		mailerStore:     s,
		installationURL: installationURL,
		appSecret:       appSecret,
		limiter:         rate.NewLimiter(rate.Limit(2), 1),
	}
}

func (w *SendWorker) Start(ctx context.Context) {
	log.Printf("send worker started")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processPendingJobs(ctx)
		}
	}
}

func (w *SendWorker) processPendingJobs(ctx context.Context) {
	jobs, err := w.db.ListPendingSendJobs(ctx)
	if err != nil {
		log.Printf("send worker: list pending jobs: %v", err)
		return
	}
	for _, job := range jobs {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("send worker: job %s failed: %v", job.ID.String(), err)
		}
	}
}

func (w *SendWorker) processJob(ctx context.Context, job db.SendJob) error {
	if err := w.db.UpdateSendJobStatus(ctx, db.UpdateSendJobStatusParams{
		ID:     job.ID,
		Status: "running",
	}); err != nil {
		return err
	}

	campaign, err := w.db.GetCampaignByID(ctx, job.CampaignID)
	if err != nil {
		return w.failJob(ctx, job.ID, campaign.ID, "get campaign: "+err.Error())
	}

	if err := w.db.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
		ID:     campaign.ID,
		Status: "sending",
	}); err != nil {
		return w.failJob(ctx, job.ID, campaign.ID, "update campaign sending: "+err.Error())
	}

	recipients, err := ExpandRecipients(ctx, w.db, campaign)
	if err != nil {
		return w.failJob(ctx, job.ID, campaign.ID, "expand recipients: "+err.Error())
	}

	if err := w.db.UpdateSendJobCounts(ctx, db.UpdateSendJobCountsParams{
		ID:         job.ID,
		TotalCount: int32(len(recipients)),
	}); err != nil {
		return w.failJob(ctx, job.ID, campaign.ID, "update job counts: "+err.Error())
	}

	if len(recipients) > 0 {
		rows := bulkRecipientParams(campaign.ID, recipients)
		if _, err := w.db.BulkCreateCampaignRecipients(ctx, rows); err != nil {
			return w.failJob(ctx, job.ID, campaign.ID, "bulk create recipients: "+err.Error())
		}
	}

	const batchSize = 100
	offset := int32(0)
	for {
		batch, err := w.db.ListPendingRecipients(ctx, db.ListPendingRecipientsParams{
			CampaignID: campaign.ID,
			Limit:      batchSize,
			Offset:     offset,
		})
		if err != nil {
			return w.failJob(ctx, job.ID, campaign.ID, "list pending recipients: "+err.Error())
		}
		if len(batch) == 0 {
			break
		}

		for _, recipient := range batch {
			sub, err := w.db.GetSubscriberByID(ctx, recipient.SubscriberID)
			if err != nil {
				log.Printf("send worker: get subscriber %s: %v", recipient.SubscriberID.String(), err)
				_ = w.db.UpdateRecipientStatus(ctx, db.UpdateRecipientStatusParams{
					CampaignID:   campaign.ID,
					SubscriberID: recipient.SubscriberID,
					Status:       "failed",
					SentAt:       pgtype.Timestamptz{},
				})
				_ = w.db.IncrementFailedCount(ctx, job.ID)
				if err := w.limiter.Wait(ctx); err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					return fmt.Errorf("rate limiter: %w", err)
				}
				continue
			}

			msg := BuildMessage(campaign, sub, w.installationURL, w.appSecret)
			if err := w.mailerStore.Get().Send(msg); err != nil {
				log.Printf("send worker: send to %s: %v", sub.Email, err)
				_ = w.db.UpdateRecipientStatus(ctx, db.UpdateRecipientStatusParams{
					CampaignID:   campaign.ID,
					SubscriberID: sub.ID,
					Status:       "failed",
					SentAt:       pgtype.Timestamptz{},
				})
				_ = w.db.IncrementFailedCount(ctx, job.ID)
				if err := w.limiter.Wait(ctx); err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					return fmt.Errorf("rate limiter: %w", err)
				}
				continue
			}

			_ = w.db.UpdateRecipientStatus(ctx, db.UpdateRecipientStatusParams{
				CampaignID:   campaign.ID,
				SubscriberID: sub.ID,
				Status:       "sent",
				SentAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			})
			_ = w.db.IncrementSentCount(ctx, job.ID)
			if err := w.limiter.Wait(ctx); err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return fmt.Errorf("rate limiter: %w", err)
			}
		}

		if len(batch) < batchSize {
			break
		}
		offset += batchSize
	}

	if err := w.db.UpdateSendJobStatus(ctx, db.UpdateSendJobStatusParams{
		ID:     job.ID,
		Status: "complete",
	}); err != nil {
		log.Printf("send worker: mark job complete: %v", err)
	}
	if err := w.db.MarkCampaignSent(ctx, campaign.ID); err != nil {
		log.Printf("send worker: mark campaign sent: %v", err)
	}

	return nil
}

func (w *SendWorker) failJob(ctx context.Context, jobID, campaignID pgtype.UUID, msg string) error {
	log.Printf("send worker: failing job %s: %s", jobID.String(), msg)
	_ = w.db.UpdateSendJobError(ctx, db.UpdateSendJobErrorParams{
		ID:           jobID,
		Status:       "failed",
		ErrorMessage: msg,
	})
	_ = w.db.UpdateCampaignStatus(ctx, db.UpdateCampaignStatusParams{
		ID:     campaignID,
		Status: "failed",
	})
	return nil
}
