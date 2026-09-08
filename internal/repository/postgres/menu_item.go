package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

// MenuItemRepository реализует работу с меню ресторана в PostgreSQL.
type MenuItemRepository struct {
	db DBTX
}

// NewMenuItemRepository создаёт новый репозиторий меню.
func NewMenuItemRepository(db DBTX) *MenuItemRepository {
	return &MenuItemRepository{
		db: db,
	}
}

// ListByRestaurantID возвращает меню указанного ресторана.
func (r *MenuItemRepository) ListByRestaurantID(
	ctx context.Context,
	restaurantID int64,
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
		WHERE restaurant_id = $1
		ORDER BY id
	`

	rows, err := r.db.Query(
		ctx,
		query,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query menu items by restaurant id: %w",
			err,
		)
	}
	defer rows.Close()

	menuItems := make([]domain.MenuItem, 0)

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

// GetByID возвращает элемент меню по идентификатору.
func (r *MenuItemRepository) GetByID(
	ctx context.Context,
	id int64,
) (domain.MenuItem, error) {
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
		WHERE id = $1
	`

	var menuItem domain.MenuItem

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&menuItem.ID,
		&menuItem.RestaurantID,
		&menuItem.Name,
		&menuItem.Description,
		&menuItem.Price,
		&menuItem.IsAvailable,
		&menuItem.CreatedAt,
		&menuItem.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.MenuItem{}, domain.ErrMenuItemNotFound
		}

		return domain.MenuItem{}, fmt.Errorf(
			"query menu item by id: %w",
			err,
		)
	}

	return menuItem, nil
}

// Create создаёт новый элемент меню.
func (r *MenuItemRepository) Create(
	ctx context.Context,
	menuItem domain.MenuItem,
) (domain.MenuItem, error) {
	const query = `
		INSERT INTO menu_items (
			restaurant_id,
			name,
			description,
			price,
			is_available
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		menuItem.RestaurantID,
		menuItem.Name,
		menuItem.Description,
		menuItem.Price,
		menuItem.IsAvailable,
	).Scan(
		&menuItem.ID,
		&menuItem.CreatedAt,
		&menuItem.UpdatedAt,
	)
	if err != nil {
		return domain.MenuItem{}, fmt.Errorf(
			"insert menu item: %w",
			err,
		)
	}

	return menuItem, nil
}

// Update обновляет элемент меню.
func (r *MenuItemRepository) Update(
	ctx context.Context,
	menuItem domain.MenuItem,
) (domain.MenuItem, error) {
	const query = `
		UPDATE menu_items
		SET
			name = $2,
			description = $3,
			price = $4,
			is_available = $5,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		menuItem.ID,
		menuItem.Name,
		menuItem.Description,
		menuItem.Price,
		menuItem.IsAvailable,
	).Scan(
		&menuItem.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.MenuItem{}, domain.ErrMenuItemNotFound
		}

		return domain.MenuItem{}, fmt.Errorf(
			"update menu item: %w",
			err,
		)
	}

	return menuItem, nil
}

// GetByIDsLocked получает элементы меню с блокировкой.
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
var _ service.RestaurantMenuRepository = (*MenuItemRepository)(nil)
