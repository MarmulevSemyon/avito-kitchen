package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

// RestaurantRepository реализует работу с ресторанами в PostgreSQL.
type RestaurantRepository struct {
	db DBTX
}

// NewRestaurantRepository создаёт новый репозиторий ресторанов.
func NewRestaurantRepository(db DBTX) *RestaurantRepository {
	return &RestaurantRepository{
		db: db,
	}
}

// ListActive возвращает список активных ресторанов.
func (r *RestaurantRepository) ListActive(
	ctx context.Context,
) ([]domain.Restaurant, error) {
	const query = `
		SELECT
			id,
			name,
			COALESCE(description, ''),
			is_active,
			created_at,
			updated_at
		FROM restaurants
		WHERE is_active = TRUE
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query active restaurants: %w", err)
	}
	defer rows.Close()

	restaurants := make([]domain.Restaurant, 0)

	for rows.Next() {
		var restaurant domain.Restaurant

		if err := rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Description,
			&restaurant.IsActive,
			&restaurant.CreatedAt,
			&restaurant.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan restaurant: %w", err)
		}

		restaurants = append(restaurants, restaurant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate restaurants: %w", err)
	}

	return restaurants, nil
}

// GetByID возвращает ресторан по идентификатору.
func (r *RestaurantRepository) GetByID(
	ctx context.Context,
	id int64,
) (domain.Restaurant, error) {
	const query = `
		SELECT
			id,
			name,
			COALESCE(description, ''),
			is_active,
			created_at,
			updated_at
		FROM restaurants
		WHERE id = $1
	`

	var restaurant domain.Restaurant

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&restaurant.ID,
		&restaurant.Name,
		&restaurant.Description,
		&restaurant.IsActive,
		&restaurant.CreatedAt,
		&restaurant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Restaurant{}, domain.ErrRestaurantNotFound
		}

		return domain.Restaurant{}, fmt.Errorf(
			"query restaurant by id: %w",
			err,
		)
	}

	return restaurant, nil
}

// GetByIDLocked возвращает ресторан по идентификатору с блокировкой строки.
func (r *RestaurantRepository) GetByIDLocked(
	ctx context.Context,
	id int64,
) (domain.Restaurant, error) {
	const query = `
		SELECT
			id,
			name,
			COALESCE(description, ''),
			is_active,
			created_at,
			updated_at
		FROM restaurants
		WHERE id = $1
		FOR SHARE
	`

	var restaurant domain.Restaurant

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&restaurant.ID,
		&restaurant.Name,
		&restaurant.Description,
		&restaurant.IsActive,
		&restaurant.CreatedAt,
		&restaurant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Restaurant{}, domain.ErrRestaurantNotFound
		}

		return domain.Restaurant{}, fmt.Errorf(
			"query restaurant by id: %w",
			err,
		)
	}

	return restaurant, nil
}

var _ service.RestaurantRepository = (*RestaurantRepository)(nil)
var _ service.RestaurantCatalogRepository = (*RestaurantRepository)(nil)
