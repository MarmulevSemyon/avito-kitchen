package postgres

import (
	"context"
	"fmt"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

type OrderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) Create(
	ctx context.Context,
	order domain.Order,
) (domain.Order, error) {
	const createOrderQuery = `
		INSERT INTO orders (
			user_id,
			restaurant_id,
			status_id,
			total_price
		)
		SELECT
			$1,
			$2,
			s.id,
			$4
		FROM order_statuses AS s
		WHERE s.code = $3
		RETURNING
			id,
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		createOrderQuery,
		order.UserID,
		order.RestaurantID,
		order.Status,
		order.TotalPrice,
	).Scan(
		&order.ID,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return domain.Order{}, fmt.Errorf("insert order: %w", err)
	}

	const createOrderItemQuery = `
		INSERT INTO order_items (
			order_id,
			menu_item_id,
			name,
			unit_price,
			quantity
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	for i := range order.Items {
		order.Items[i].OrderID = order.ID

		err := r.db.QueryRow(
			ctx,
			createOrderItemQuery,
			order.Items[i].OrderID,
			order.Items[i].MenuItemID,
			order.Items[i].Name,
			order.Items[i].UnitPrice,
			order.Items[i].Quantity,
		).Scan(&order.Items[i].ID)
		if err != nil {
			return domain.Order{}, fmt.Errorf(
				"insert order item: menu_item_id=%d: %w",
				order.Items[i].MenuItemID,
				err,
			)
		}
	}

	return order, nil
}

var _ service.OrderRepository = (*OrderRepository)(nil)
