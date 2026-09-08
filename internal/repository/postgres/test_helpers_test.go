package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const defaultTestDatabaseURL = "postgres://avito:avito@localhost:5433/postgres?sslmode=disable"

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	adminURL := os.Getenv("TEST_DATABASE_URL")
	explicitURL := adminURL != ""

	if adminURL == "" {
		adminURL = defaultTestDatabaseURL
	}

	adminPool, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		t.Fatalf("create admin pool: %v", err)
	}

	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()

		if explicitURL {
			t.Fatalf("connect to test postgres: %v", err)
		}

		t.Skipf("test postgres is unavailable: %v", err)
	}

	databaseName := fmt.Sprintf(
		"avito_kitchen_test_%d",
		time.Now().UnixNano(),
	)

	_, err = adminPool.Exec(
		ctx,
		"CREATE DATABASE "+databaseName,
	)
	if err != nil {
		adminPool.Close()
		t.Fatalf("create test database: %v", err)
	}

	t.Cleanup(func() {
		_, _ = adminPool.Exec(
			context.Background(),
			"DROP DATABASE IF EXISTS "+databaseName+" WITH (FORCE)",
		)

		adminPool.Close()
	})

	databaseURL := replaceDatabaseName(
		adminURL,
		databaseName,
	)

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test database pool: %v", err)
	}

	t.Cleanup(pool.Close)

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	if err := applyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return pool
}

func applyMigrations(
	ctx context.Context,
	pool *pgxpool.Pool,
) error {
	migrationsDir, err := migrationsDirectory()
	if err != nil {
		return err
	}

	migrationFiles, err := filepath.Glob(
		filepath.Join(
			migrationsDir,
			"*.up.sql",
		),
	)
	if err != nil {
		return fmt.Errorf(
			"find migration files: %w",
			err,
		)
	}

	sort.Strings(migrationFiles)

	for _, migrationFile := range migrationFiles {
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			return fmt.Errorf(
				"read migration %s: %w",
				filepath.Base(migrationFile),
				err,
			)
		}

		if _, err := pool.Exec(
			ctx,
			string(content),
		); err != nil {
			return fmt.Errorf(
				"execute migration %s: %w",
				filepath.Base(migrationFile),
				err,
			)
		}
	}

	return nil
}

func migrationsDirectory() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf(
			"determine current test file path",
		)
	}

	projectRoot := filepath.Clean(
		filepath.Join(
			filepath.Dir(filename),
			"..",
			"..",
			"..",
		),
	)

	return filepath.Join(
		projectRoot,
		"migrations",
	), nil
}

func replaceDatabaseName(
	databaseURL string,
	databaseName string,
) string {
	const separator = "/"

	queryIndex := strings.Index(
		databaseURL,
		"?",
	)

	var base string
	var query string

	if queryIndex >= 0 {
		base = databaseURL[:queryIndex]
		query = databaseURL[queryIndex:]
	} else {
		base = databaseURL
	}

	lastSlash := strings.LastIndex(
		base,
		separator,
	)

	if lastSlash < 0 {
		return databaseURL
	}

	return base[:lastSlash+1] +
		databaseName +
		query
}

func insertTestRestaurant(
	t *testing.T,
	pool *pgxpool.Pool,
	name string,
	description any,
	isActive bool,
) int64 {
	t.Helper()

	const query = `
		INSERT INTO restaurants (
			name,
			description,
			is_active
		)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64

	err := pool.QueryRow(
		context.Background(),
		query,
		name,
		description,
		isActive,
	).Scan(&id)
	require.NoError(t, err)

	return id
}

func insertTestMenuItem(
	t *testing.T,
	pool *pgxpool.Pool,
	restaurantID int64,
	name string,
	description any,
	price int64,
	isAvailable bool,
) int64 {
	t.Helper()

	const query = `
		INSERT INTO menu_items (
			restaurant_id,
			name,
			description,
			price,
			is_available
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64

	err := pool.QueryRow(
		context.Background(),
		query,
		restaurantID,
		name,
		description,
		price,
		isAvailable,
	).Scan(&id)
	require.NoError(t, err)

	return id
}
