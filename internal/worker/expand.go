package worker

import (
	"context"
	"fmt"

	db "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

func ExpandRecipients(ctx context.Context, q *db.Queries, campaign db.Campaign) ([]db.Subscriber, error) {
	const pageSize = 1000
	var all []db.Subscriber

	switch campaign.SendToType {
	case "list":
		offset := int32(0)
		for {
			page, err := q.ListSubscribersInList(ctx, db.ListSubscribersInListParams{
				ListID: campaign.SendToID,
				Limit:  pageSize,
				Offset: offset,
			})
			if err != nil {
				return nil, err
			}
			all = append(all, page...)
			if len(page) < pageSize {
				break
			}
			offset += pageSize
		}
	case "tag":
		offset := int32(0)
		for {
			page, err := q.ListSubscribersWithTag(ctx, db.ListSubscribersWithTagParams{
				TagID:  campaign.SendToID,
				Limit:  pageSize,
				Offset: offset,
			})
			if err != nil {
				return nil, err
			}
			all = append(all, page...)
			if len(page) < pageSize {
				break
			}
			offset += pageSize
		}
	default:
		return nil, fmt.Errorf("unknown send_to_type: %s", campaign.SendToType)
	}

	var filtered []db.Subscriber
	for _, sub := range all {
		if sub.Status != "active" {
			continue
		}
		suppressed, err := q.IsSuppressed(ctx, sub.Email)
		if err != nil {
			return nil, err
		}
		if suppressed {
			continue
		}
		filtered = append(filtered, sub)
	}

	return filtered, nil
}

func bulkRecipientParams(campaignID pgtype.UUID, subs []db.Subscriber) []db.BulkCreateCampaignRecipientsParams {
	params := make([]db.BulkCreateCampaignRecipientsParams, len(subs))
	for i, s := range subs {
		params[i] = db.BulkCreateCampaignRecipientsParams{
			CampaignID:   campaignID,
			SubscriberID: s.ID,
		}
	}
	return params
}
