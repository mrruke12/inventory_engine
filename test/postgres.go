package test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	conn "github.com/mrruke12/inventory_engine/internal/infrastructure/postgres"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgres() (*postgres.PostgresContainer, *pgxpool.Pool, error) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithLogger(log.New(io.Discard, "", 0)),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		return nil, nil, err
	}

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		return nil, nil, err
	}

	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, nil, fmt.Errorf("resolve test helper source path")
	}
	migrationsPath, err := filepath.Abs(filepath.Join(filepath.Dir(sourceFile), "..", "internal", "infrastructure", "migrations"))
	if err != nil {
		return nil, nil, err
	}
	migrationsURL := (&url.URL{Scheme: "file", Path: filepath.ToSlash(migrationsPath)}).String()
	err = conn.Migrate(migrationsURL, connString)

	if err != nil {
		return nil, nil, err
	}

	pool, err := conn.Connect(ctx, connString)

	if err != nil {
		return nil, nil, err
	}
	return pgContainer, pool, nil
}

func CleanupTables(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	sql := `
		DO $$ 
		DECLARE 
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE;';
			END LOOP;
		END $$;
	`

	_, err := pool.Exec(ctx, sql)

	return err
}

func CleanupPostgres(t *testing.T, pool *pgxpool.Pool) {
	t.Cleanup(func() {
		if err := CleanupTables(pool); err != nil {
			t.Errorf("cleanup failed: %v", err)
		}
	})
}
