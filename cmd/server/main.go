package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	dbembed "github.com/AbMani46/ownmaily/db"
	"github.com/AbMani46/ownmaily/internal/config"
	"github.com/AbMani46/ownmaily/internal/db"
	"github.com/AbMani46/ownmaily/internal/handler"
	"github.com/AbMani46/ownmaily/internal/mailer"
	"github.com/AbMani46/ownmaily/internal/middleware"
	db2 "github.com/AbMani46/ownmaily/internal/sqlc"
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

	authHandler := handler.NewAuthHandler(queries, cfg.AppSecret, cfg.InstallationURL)
	subscriberHandler := handler.NewSubscriberHandler(queries)
	confirmMailer := mailer.NewConfirmationMailer(queries, cfg.InstallationURL, cfg.AppSecret)
	listHandler := handler.NewListHandler(queries, confirmMailer)

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)
	r.With(middleware.RequireAuth(cfg.AppSecret, queries)).Get("/api/auth/me", authHandler.Me)

	r.Get("/confirm", listHandler.ConfirmOptIn)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.AppSecret, queries))

		r.Get("/api/subscribers", subscriberHandler.List)
		r.Post("/api/subscribers", subscriberHandler.Create)
		r.Get("/api/subscribers/export", subscriberHandler.Export)
		r.Post("/api/subscribers/import", subscriberHandler.Import)
		r.Get("/api/subscribers/{id}", subscriberHandler.Get)
		r.Put("/api/subscribers/{id}", subscriberHandler.Update)
		r.Delete("/api/subscribers/{id}", subscriberHandler.Delete)
		r.Post("/api/subscribers/{id}/unsubscribe", subscriberHandler.Unsubscribe)

		r.Get("/api/lists", listHandler.List)
		r.Post("/api/lists", listHandler.Create)
		r.Get("/api/lists/{id}", listHandler.Get)
		r.Put("/api/lists/{id}", listHandler.Update)
		r.Delete("/api/lists/{id}", listHandler.Delete)
		r.Get("/api/lists/{id}/subscribers", listHandler.ListSubscribers)
		r.Post("/api/lists/{id}/subscribers", listHandler.AddSubscriber)
		r.Delete("/api/lists/{id}/subscribers/{subscriberID}", listHandler.RemoveSubscriber)
	})

	r.Handle("/*", http.FileServer(http.Dir("frontend")))

	log.Printf("OwnMaily started on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
