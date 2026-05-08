package main

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

	mailerStore := mailer.NewStore(loadMailer(context.Background(), queries))
	sendWorker := worker.NewSendWorker(queries, mailerStore, cfg.InstallationURL, cfg.AppSecret)
	schedulerWorker := worker.NewSchedulerWorker(queries, pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sendWorker.Start(ctx)
	go schedulerWorker.Start(ctx)

	setupHandler := handler.NewSetupHandler(queries, cfg.AppSecret, mailerStore)
	webhookHandler := handler.NewWebhookHandler(queries)
	authHandler := handler.NewAuthHandler(queries, cfg.AppSecret, cfg.InstallationURL)
	subscriberHandler := handler.NewSubscriberHandler(queries)
	confirmMailer := mailer.NewConfirmationMailer(queries, cfg.InstallationURL, cfg.AppSecret)
	listHandler := handler.NewListHandler(queries, confirmMailer)
	embedHandler := handler.NewEmbedHandler(queries, cfg.InstallationURL, confirmMailer)
	tagHandler := handler.NewTagHandler(queries)
	campaignHandler := handler.NewCampaignHandler(queries, cfg.InstallationURL, cfg.AppSecret, sendWorker)
	trackingHandler := handler.NewTrackingHandler(queries, cfg.AppSecret)
	analyticsHandler := handler.NewAnalyticsHandler(queries)
	settingsHandler := handler.NewSettingsHandler(queries, mailerStore)
	uploadStorage := handler.NewLocalStorage(cfg.UploadDir + "/images")
	uploadHandler := handler.NewUploadHandler(uploadStorage)

	r := chi.NewRouter()
	r.Use(setupGuard(queries))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Setup API — no auth required, guarded internally by setup_complete flag
	r.Get("/api/setup/status", setupHandler.Status)
	r.Post("/api/setup/owner", setupHandler.CreateOwner)
	r.Put("/api/setup/settings", setupHandler.UpdateSettings)
	r.Put("/api/setup/smtp", setupHandler.UpdateSMTP)
	r.Post("/api/setup/test-smtp", setupHandler.TestSMTP)
	r.Post("/api/setup/complete", setupHandler.Complete)

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

	r.Get("/embed/{listID}.js", embedHandler.ServeJS)
	r.Post("/api/public/subscribe", embedHandler.Subscribe)

	// Uploaded images are publicly accessible so email clients can fetch them.
	uploadDir := cfg.UploadDir
	r.Get("/uploads/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))).ServeHTTP(w, r)
	})

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
		r.Post("/api/lists/{id}/subscribers/bulk", listHandler.BulkAddSubscribers)
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

		r.Post("/api/uploads/images", uploadHandler.UploadImage)

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

	// Serve Vue build with SPA fallback — must be mounted after all /api/* routes.
	distFS := os.DirFS("frontend/dist")
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := fs.Stat(distFS, path); err == nil {
			http.FileServer(http.Dir("frontend/dist")).ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, "frontend/dist/index.html")
	})

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

// setupGuard redirects to /setup if setup is not complete.
// Skips API routes, setup routes, public tracking/webhook routes, and health.
func setupGuard(queries *db2.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if strings.HasPrefix(path, "/api/") ||
				strings.HasPrefix(path, "/setup") ||
				strings.HasPrefix(path, "/track/") ||
				strings.HasPrefix(path, "/unsubscribe") ||
				strings.HasPrefix(path, "/confirm") ||
				strings.HasPrefix(path, "/webhooks/") ||
				strings.HasPrefix(path, "/embed/") ||
				strings.HasPrefix(path, "/uploads/") ||
				strings.HasPrefix(path, "/assets/") ||
				path == "/favicon.ico" ||
				path == "/favicon.svg" ||
				path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			settings, err := queries.GetSettings(r.Context())
			if err != nil || !settings.SetupComplete {
				http.Redirect(w, r, "/setup", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// loadMailer reads smtp_provider + smtp_credentials from settings at startup.
// Falls back to LogMailer on any error so startup never fails.
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
