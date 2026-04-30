package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	dbembed "github.com/AbMani46/ownmaily/db"
	idb "github.com/AbMani46/ownmaily/internal/db"
	sqlcdb "github.com/AbMani46/ownmaily/internal/sqlc"
)

const defaultAdminURL = "postgres://ownmaily:ownmaily@localhost:5433/postgres?sslmode=disable"

// NewTestDB creates a fresh ownmaily_test database, runs all migrations,
// and returns a pool + cleanup func. Call t.Cleanup(cleanup) in each test.
func NewTestDB(t *testing.T) (*pgxpool.Pool, *sqlcdb.Queries, func()) {
	t.Helper()

	adminURL := os.Getenv("TEST_DB_URL")
	if adminURL == "" {
		adminURL = defaultAdminURL
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

	// Terminate any existing connections to ownmaily_test before dropping.
	_, _ = adminPool.Exec(ctx,
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'ownmaily_test' AND pid <> pg_backend_pid()")

	if _, err := adminPool.Exec(ctx, "DROP DATABASE IF EXISTS ownmaily_test"); err != nil {
		adminPool.Close()
		t.Fatalf("drop test db: %v", err)
	}

	if _, err := adminPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE ownmaily_test OWNER %s", "ownmaily")); err != nil {
		adminPool.Close()
		t.Fatalf("create test db: %v", err)
	}

	adminPool.Close()

	// Derive test DB URL from the admin URL by replacing the database name.
	testURL := strings.Replace(adminURL, "/postgres?", "/ownmaily_test?", 1)

	if err := idb.RunMigrations(testURL, dbembed.Migrations); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := idb.Connect(testURL)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	queries := sqlcdb.New(pool)

	cleanup := func() {
		pool.Close()
	}

	return pool, queries, cleanup
}
