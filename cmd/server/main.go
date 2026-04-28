package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	dbembed "github.com/AbMani46/ownmaily/db"
	"github.com/AbMani46/ownmaily/internal/config"
	"github.com/AbMani46/ownmaily/internal/db"
	"github.com/AbMani46/ownmaily/internal/handler"
	"github.com/AbMani46/ownmaily/internal/mailer"
	"github.com/AbMani46/ownmaily/internal/middleware"
	db2 "github.com/AbMani46/ownmaily/internal/sqlc"
	"github.com/AbMani46/ownmaily/internal/worker"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := db.RunMigrations(cfg.DBUrl, dbembed.Migrations); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pool, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	queries := db2.New(pool)

	activeMailer := loadMailer(context.Background(), queries)
	sendWorker := worker.NewSendWorker(queries, activeMailer, cfg.InstallationURL, cfg.AppSecret)
	schedulerWorker := worker.NewSchedulerWorker(queries, pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sendWorker.Start(ctx)
	go schedulerWorker.Start(ctx)

	webhookHandler := handler.NewWebhookHandler(queries)
	authHandler := handler.NewAuthHandler(queries, cfg.AppSecret, cfg.InstallationURL)
	subscriberHandler := handler.NewSubscriberHandler(queries)
	confirmMailer := mailer.NewConfirmationMailer(queries, cfg.InstallationURL, cfg.AppSecret)
	listHandler := handler.NewListHandler(queries, confirmMailer)
	tagHandler := handler.NewTagHandler(queries)
	campaignHandler := handler.NewCampaignHandler(queries, cfg.InstallationURL, cfg.AppSecret, sendWorker)
	trackingHandler := handler.NewTrackingHandler(queries, cfg.AppSecret)
	analyticsHandler := handler.NewAnalyticsHandler(queries)

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)
	r.With(middleware.RequireAuth(cfg.AppSecret, queries)).Get("/api/auth/me", authHandler.Me)

	r.Get("/confirm", listHandler.ConfirmOptIn)
	r.Get("/unsubscribe", trackingHandler.Unsubscribe)
	r.Get("/track/open/{token}", trackingHandler.Open)
	r.Get("/track/click/{token}", trackingHandler.Click)
	r.Post("/webhooks/resend", webhookHandler.Resend)
	r.Post("/webhooks/mailgun", webhookHandler.Mailgun)
	r.Post("/webhooks/ses", webhookHandler.SES)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.AppSecret, queries))

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
	})

	r.Handle("/*", http.FileServer(http.Dir("frontend")))

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Printf("shutting down...")
		cancel()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		_ = srv.Shutdown(shutCtx)
	}()

	log.Printf("OwnMaily started on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}

// loadMailer reads smtp_provider + smtp_credentials from settings at startup.
// Falls back to LogMailer on any error so startup never fails.
// TODO Session 17: reload mailer when SMTP settings change via the settings API.
func loadMailer(ctx context.Context, queries *db2.Queries) mailer.Mailer {
	settings, err := queries.GetSettings(ctx)
	if err != nil || settings.SmtpProvider == "" {
		return &mailer.LogMailer{}
	}
	m, err := mailer.NewMailer(settings.SmtpProvider, string(settings.SmtpCredentials))
	if err != nil {
		log.Printf("warn: could not load mailer (%v), falling back to LogMailer", err)
		return &mailer.LogMailer{}
	}
	log.Printf("mailer: %s loaded", m.Name())
	return m
}
