package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const defaultTestDatabaseURL = "postgres://avito:avito@localhost:5432/postgres?sslmode=disable"

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	usingDefaultURL := databaseURL == ""

	if usingDefaultURL {
		databaseURL = defaultTestDatabaseURL
	}

	adminConfig, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)

	// Для CREATE DATABASE подключаемся к системной БД,
	// а не к avito_kitchen.
	adminConfig.ConnConfig.Database = "postgres"

	adminPool, err := pgxpool.NewWithConfig(ctx, adminConfig)
	require.NoError(t, err)

	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()

		if usingDefaultURL {
			t.Skipf(
				"PostgreSQL is not available at default address: %v",
				err,
			)
		}

		require.NoError(t, err)
	}

	t.Cleanup(adminPool.Close)

	databaseName := fmt.Sprintf(
		"avito_kitchen_test_%d",
		time.Now().UnixNano(),
	)

	databaseIdentifier := pgx.Identifier{databaseName}.Sanitize()

	_, err = adminPool.Exec(
		ctx,
		"CREATE DATABASE "+databaseIdentifier,
	)
	require.NoError(t, err)

	var testPool *pgxpool.Pool

	t.Cleanup(func() {
		if testPool != nil {
			testPool.Close()
		}

		_, _ = adminPool.Exec(
			context.Background(),
			"DROP DATABASE IF EXISTS "+
				databaseIdentifier+
				" WITH (FORCE)",
		)
	})

	testConfig, err := pgxpool.ParseConfig(databaseURL)
	require.NoError(t, err)

	testConfig.ConnConfig.Database = databaseName

	testPool, err = pgxpool.NewWithConfig(ctx, testConfig)
	require.NoError(t, err)

	require.NoError(t, testPool.Ping(ctx))

	applyTestMigration(t, testPool)

	return testPool
}

func applyTestMigration(
	t *testing.T,
	pool *pgxpool.Pool,
) {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	migrationPath := filepath.Join(
		filepath.Dir(currentFile),
		"..",
		"..",
		"..",
		"migrations",
		"000001_init.up.sql",
	)

	content, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	statements := strings.Split(string(content), ";")

	ctx := context.Background()

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)

		if statement == "" {
			continue
		}

		_, err := pool.Exec(ctx, statement)
		require.NoErrorf(
			t,
			err,
			"failed migration statement:\n%s",
			statement,
		)
	}
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
