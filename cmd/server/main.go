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

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)
	r.With(middleware.RequireAuth(cfg.AppSecret, queries)).Get("/api/auth/me", authHandler.Me)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(cfg.AppSecret, queries))
		// sessions 4+ mount here
	})

	r.Handle("/*", http.FileServer(http.Dir("frontend")))

	log.Printf("OwnMaily started on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
