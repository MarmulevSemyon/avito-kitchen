package postgres

import (
	"context"
	"fmt"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

type MenuItemRepository struct {
	db DBTX
}

func NewMenuItemRepository(db DBTX) *MenuItemRepository {
	return &MenuItemRepository{
		db: db,
	}
}

func (r *MenuItemRepository) GetByIDsLocked(
	ctx context.Context,
	ids []int64,
) ([]domain.MenuItem, error) {
	const query = `
		SELECT
			id,
			restaurant_id,
			name,
			COALESCE(description, ''),
			price,
			is_available,
			created_at,
			updated_at
		FROM menu_items
		WHERE id = ANY($1)
		ORDER BY id
		FOR SHARE
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("query menu items by ids: %w", err)
	}
	defer rows.Close()

	menuItems := make([]domain.MenuItem, 0, len(ids))

	for rows.Next() {
		var menuItem domain.MenuItem

		if err := rows.Scan(
			&menuItem.ID,
			&menuItem.RestaurantID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.Price,
			&menuItem.IsAvailable,
			&menuItem.CreatedAt,
			&menuItem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}

		menuItems = append(menuItems, menuItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate menu items: %w", err)
	}

	return menuItems, nil
}

var _ service.MenuItemRepository = (*MenuItemRepository)(nil)
