package testutil

import (
	"net"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AbMani46/ownmaily/internal/handler"
	"github.com/AbMani46/ownmaily/internal/mailer"
	"github.com/AbMani46/ownmaily/internal/middleware"
	sqlcdb "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/worker"
)

const TestAppSecret = "test-secret-do-not-use"

// NewTestServer wires up the full chi router with real handlers connected to
// the provided DB. Returns an httptest.Server and cleanup func.
func NewTestServer(t *testing.T, db *sqlcdb.Queries, pool *pgxpool.Pool) (*httptest.Server, func()) {
	t.Helper()

	// Pre-allocate a listener so we know the address before creating handlers.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	installationURL := "http://" + ln.Addr().String()

	mailerStore := mailer.NewStore(&mailer.LogMailer{})
	sendWorker := worker.NewSendWorker(db, mailerStore, installationURL, TestAppSecret)
	schedulerWorker := worker.NewSchedulerWorker(db, pool)
	_ = schedulerWorker // not started in tests

	webhookHandler := handler.NewWebhookHandler(db)
	authHandler := handler.NewAuthHandler(db, TestAppSecret, installationURL)
	subscriberHandler := handler.NewSubscriberHandler(db)
	confirmMailer := mailer.NewConfirmationMailer(db, installationURL, TestAppSecret)
	listHandler := handler.NewListHandler(db, confirmMailer)
	tagHandler := handler.NewTagHandler(db)
	campaignHandler := handler.NewCampaignHandler(db, installationURL, TestAppSecret, sendWorker)
	trackingHandler := handler.NewTrackingHandler(db, TestAppSecret)
	analyticsHandler := handler.NewAnalyticsHandler(db)
	settingsHandler := handler.NewSettingsHandler(db, mailerStore)

	r := chi.NewRouter()

	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)
	r.With(middleware.RequireAuth(TestAppSecret, db)).Get("/api/auth/me", authHandler.Me)

	r.Get("/confirm", listHandler.ConfirmOptIn)
	r.Get("/unsubscribe", trackingHandler.Unsubscribe)
	r.Get("/track/open/{token}", trackingHandler.Open)
	r.Get("/track/click/{token}", trackingHandler.Click)
	r.Post("/webhooks/resend", webhookHandler.Resend)
	r.Post("/webhooks/mailgun", webhookHandler.Mailgun)
	r.Post("/webhooks/ses", webhookHandler.SES)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(TestAppSecret, db))

		r.Get("/api/subscribers", subscriberHandler.List)
		r.Post("/api/subscribers", subscriberHandler.Create)
		r.Get("/api/subscribers/export", subscriberHandler.Export)
		r.Post("/api/subscribers/import", subscriberHandler.Import)
		r.Post("/api/subscribers/bulk-tag", subscriberHandler.BulkTag)
		r.Get("/api/subscribers/{id}", subscriberHandler.Get)
		r.Put("/api/subscribers/{id}", subscriberHandler.Update)
		r.Delete("/api/subscribers/{id}", subscriberHandler.Delete)
		r.Post("/api/subscribers/{id}/unsubscribe", subscriberHandler.Unsubscribe)
		r.Get("/api/subscribers/{id}/tags", subscriberHandler.ListTags)
		r.Post("/api/subscribers/{id}/tags", subscriberHandler.AddTag)
		r.Delete("/api/subscribers/{id}/tags/{tagID}", subscriberHandler.RemoveTag)

		r.Get("/api/lists", listHandler.List)
		r.Post("/api/lists", listHandler.Create)
		r.Get("/api/lists/{id}", listHandler.Get)
		r.Put("/api/lists/{id}", listHandler.Update)
		r.Delete("/api/lists/{id}", listHandler.Delete)
		r.Get("/api/lists/{id}/subscribers", listHandler.ListSubscribers)
		r.Post("/api/lists/{id}/subscribers", listHandler.AddSubscriber)
		r.Delete("/api/lists/{id}/subscribers/{subscriberID}", listHandler.RemoveSubscriber)

		r.Get("/api/tags", tagHandler.List)
		r.Post("/api/tags", tagHandler.Create)
		r.Get("/api/tags/{id}", tagHandler.Get)
		r.Put("/api/tags/{id}", tagHandler.Update)
		r.Delete("/api/tags/{id}", tagHandler.Delete)
		r.Get("/api/tags/{id}/subscribers", tagHandler.ListSubscribers)

		r.Get("/api/campaigns", campaignHandler.List)
		r.Post("/api/campaigns", campaignHandler.Create)
		r.Get("/api/campaigns/{id}", campaignHandler.Get)
		r.Put("/api/campaigns/{id}", campaignHandler.Update)
		r.Delete("/api/campaigns/{id}", campaignHandler.Delete)
		r.Get("/api/campaigns/{id}/preview", campaignHandler.Preview)
		r.Get("/api/campaigns/{id}/stats", campaignHandler.Stats)
		r.Post("/api/campaigns/{id}/send", campaignHandler.Send)
		r.Post("/api/campaigns/{id}/schedule", campaignHandler.Schedule)
		r.Post("/api/campaigns/{id}/cancel", campaignHandler.Cancel)
		r.Post("/api/campaigns/{id}/duplicate", campaignHandler.Duplicate)

		r.Get("/api/analytics/overview", analyticsHandler.Overview)
		r.Get("/api/subscribers/{id}/stats", subscriberHandler.Stats)

		r.Get("/api/settings", settingsHandler.Get)
		r.Put("/api/settings/general", settingsHandler.UpdateGeneral)
		r.Put("/api/settings/smtp", settingsHandler.UpdateSMTP)
		r.Post("/api/settings/smtp/test", settingsHandler.TestSMTP)
		r.Get("/api/settings/api-key", settingsHandler.GetAPIKey)
		r.Post("/api/settings/api-key/regenerate", settingsHandler.RegenerateAPIKey)
		r.Get("/api/settings/suppressions/export", settingsHandler.ExportSuppressions)
		r.Get("/api/settings/suppressions", settingsHandler.ListSuppressions)
		r.Post("/api/settings/suppressions", settingsHandler.AddSuppression)
		r.Delete("/api/settings/suppressions/{email}", settingsHandler.DeleteSuppression)
	})

	srv := httptest.NewUnstartedServer(r)
	srv.Listener = ln
	srv.Start()

	return srv, srv.Close
}
