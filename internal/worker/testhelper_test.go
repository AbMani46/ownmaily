package worker

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	dbembed "github.com/AbMani46/ownmaily/db"
	idb "github.com/AbMani46/ownmaily/internal/db"
	sqlc "github.com/AbMani46/ownmaily/internal/sqlc"
)

const workerTestSecret = "test-secret-do-not-use"
const workerDefaultAdminURL = "postgres://ownmaily:ownmaily@localhost:5433/postgres?sslmode=disable"

// newWorkerTestDB creates a fresh ownmaily_test_worker database, runs migrations,
// and returns a pool + cleanup func.
func newWorkerTestDB(t *testing.T) (*pgxpool.Pool, *sqlc.Queries, func()) {
	t.Helper()

	adminURL := os.Getenv("TEST_DB_URL")
	if adminURL == "" {
		adminURL = workerDefaultAdminURL
	}

	ctx := context.Background()

	adminPool, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
		return nil, nil, func() {}
	}

	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("postgres unavailable: %v", err)
		return nil, nil, func() {}
	}

	const testDBName = "ownmaily_test_worker"

	_, _ = adminPool.Exec(ctx,
		fmt.Sprintf("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()", testDBName))

	if _, err := adminPool.Exec(ctx, "DROP DATABASE IF EXISTS "+testDBName); err != nil {
		adminPool.Close()
		t.Fatalf("drop test db: %v", err)
	}

	if _, err := adminPool.Exec(ctx, "CREATE DATABASE "+testDBName+" OWNER ownmaily"); err != nil {
		adminPool.Close()
		t.Fatalf("create test db: %v", err)
	}

	adminPool.Close()

	testURL := strings.Replace(adminURL, "/postgres?", "/"+testDBName+"?", 1)

	if err := idb.RunMigrations(testURL, dbembed.Migrations); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := idb.Connect(testURL)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	queries := sqlc.New(pool)

	return pool, queries, func() { pool.Close() }
}

func wCreateSubscriber(t *testing.T, queries *sqlc.Queries, email string) sqlc.Subscriber {
	t.Helper()
	sub, err := queries.CreateSubscriber(context.Background(), sqlc.CreateSubscriberParams{
		Email:  email,
		Status: "active",
		Source: "api",
	})
	if err != nil {
		t.Fatalf("create subscriber %s: %v", email, err)
	}
	return sub
}

func wCreateList(t *testing.T, queries *sqlc.Queries, name string) sqlc.List {
	t.Helper()
	list, err := queries.CreateList(context.Background(), sqlc.CreateListParams{
		Name:        name,
		Description: "",
		DoubleOptIn: false,
	})
	if err != nil {
		t.Fatalf("create list %s: %v", name, err)
	}
	return list
}

func wCreateCampaign(t *testing.T, queries *sqlc.Queries, listID pgtype.UUID) sqlc.Campaign {
	t.Helper()
	c, err := queries.CreateCampaign(context.Background(), sqlc.CreateCampaignParams{
		Name:        "Worker Test Campaign",
		Subject:     "Test Subject",
		PreviewText: "",
		FromName:    "Tester",
		FromEmail:   "tester@example.com",
		ReplyTo:     "",
		HtmlBody:    "<p>Hello <a href=\"https://example.com\">click</a></p><body>",
		TextBody:    "Hello",
		SendToType:  "list",
		SendToID:    listID,
	})
	if err != nil {
		t.Fatalf("create campaign: %v", err)
	}
	return c
}
