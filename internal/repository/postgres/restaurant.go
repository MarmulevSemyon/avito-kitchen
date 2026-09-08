package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

type RestaurantRepository struct {
	db DBTX
}

func NewRestaurantRepository(db DBTX) *RestaurantRepository {
	return &RestaurantRepository{
		db: db,
	}
}

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
