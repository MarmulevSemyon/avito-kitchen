package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

// OrderRepository реализует работу с заказами в PostgreSQL.
type OrderRepository struct {
	db DBTX
}

// NewOrderRepository создаёт новый репозиторий заказов.
func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// Create создаёт новый заказ.
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

// GetByID получает заказ по идентификатору.
func (r *OrderRepository) GetByID(
	ctx context.Context,
	id int64,
) (domain.Order, error) {
	order, err := r.getOrderHeader(ctx, id, false)
	if err != nil {
		return domain.Order{}, err
	}

	items, err := r.getOrderItems(
		ctx,
		[]int64{id},
	)
	if err != nil {
		return domain.Order{}, err
	}

	order.Items = items[id]

	return order, nil
}

// GetByIDLocked получает заказ с блокировкой.
func (r *OrderRepository) GetByIDLocked(
	ctx context.Context,
	id int64,
) (domain.Order, error) {
	return r.getOrderHeader(ctx, id, true)
}

// UpdateStatus изменяет статус заказа.
func (r *OrderRepository) UpdateStatus(
	ctx context.Context,
	order domain.Order,
) (domain.Order, error) {
	const query = `
		UPDATE orders AS o
		SET
			status_id = s.id,
			updated_at = NOW()
		FROM order_statuses AS s
		WHERE
			o.id = $1
			AND s.code = $2
		RETURNING o.updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		order.ID,
		order.Status,
	).Scan(
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}

		return domain.Order{}, fmt.Errorf(
			"update order status: %w",
			err,
		)
	}

	return order, nil
}

// ListByRestaurant возвращает список заказов указанного ресторана с пагинацией.
func (r *OrderRepository) ListByRestaurant(
	ctx context.Context,
	restaurantID int64,
	status *domain.OrderStatus,
	limit int,
	cursor *service.OrderCursor,
) ([]domain.Order, error) {
	var query strings.Builder

	query.WriteString(`
		SELECT
			o.id,
			o.user_id,
			o.restaurant_id,
			s.code,
			o.total_price,
			o.created_at,
			o.updated_at
		FROM orders AS o
		JOIN order_statuses AS s
			ON s.id = o.status_id
		WHERE o.restaurant_id = $1
	`)

	args := []any{restaurantID}
	nextArg := 2

	if status != nil {
		fmt.Fprintf(&query, `
        AND o.status_id = (
            SELECT id
            FROM order_statuses
            WHERE code = $%d
        )
        `, nextArg)

		args = append(args, *status)
		nextArg++
	}

	if cursor != nil {
		fmt.Fprintf(&query, `
        AND (o.created_at, o.id) < ($%d, $%d)
        `, nextArg, nextArg+1)

		args = append(
			args,
			cursor.CreatedAt,
			cursor.ID,
		)

		nextArg += 2
	}

	fmt.Fprintf(&query, `
    ORDER BY o.created_at DESC, o.id DESC
    LIMIT $%d
    `, nextArg)

	args = append(args, limit)

	rows, err := r.db.Query(
		ctx,
		query.String(),
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query restaurant orders: %w",
			err,
		)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0, limit)
	orderIDs := make([]int64, 0, limit)

	for rows.Next() {
		var order domain.Order
		var statusCode string

		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.RestaurantID,
			&statusCode,
			&order.TotalPrice,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan order: %w",
				err,
			)
		}

		order.Status = domain.OrderStatus(statusCode)

		orders = append(orders, order)
		orderIDs = append(orderIDs, order.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate orders: %w",
			err,
		)
	}

	if len(orderIDs) == 0 {
		return orders, nil
	}

	itemsByOrderID, err := r.getOrderItems(
		ctx,
		orderIDs,
	)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		orders[i].Items = itemsByOrderID[orders[i].ID]
	}

	return orders, nil
}

func (r *OrderRepository) getOrderHeader(
	ctx context.Context,
	id int64,
	forUpdate bool,
) (domain.Order, error) {
	query := `
		SELECT
			o.id,
			o.user_id,
			o.restaurant_id,
			s.code,
			o.total_price,
			o.created_at,
			o.updated_at
		FROM orders AS o
		JOIN order_statuses AS s
			ON s.id = o.status_id
		WHERE o.id = $1
	`

	if forUpdate {
		query += ` FOR UPDATE OF o`
	}

	var order domain.Order
	var statusCode string

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.RestaurantID,
		&statusCode,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}

		return domain.Order{}, fmt.Errorf(
			"query order by id: %w",
			err,
		)
	}

	order.Status = domain.OrderStatus(statusCode)

	return order, nil
}

func (r *OrderRepository) getOrderItems(
	ctx context.Context,
	orderIDs []int64,
) (map[int64][]domain.OrderItem, error) {
	const query = `
		SELECT
			id,
			order_id,
			menu_item_id,
			name,
			unit_price,
			quantity
		FROM order_items
		WHERE order_id = ANY($1)
		ORDER BY order_id, id
	`

	rows, err := r.db.Query(
		ctx,
		query,
		orderIDs,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"query order items: %w",
			err,
		)
	}
	defer rows.Close()

	result := make(
		map[int64][]domain.OrderItem,
		len(orderIDs),
	)

	for rows.Next() {
		var item domain.OrderItem

		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.Name,
			&item.UnitPrice,
			&item.Quantity,
		); err != nil {
			return nil, fmt.Errorf(
				"scan order item: %w",
				err,
			)
		}

		result[item.OrderID] = append(
			result[item.OrderID],
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate order items: %w",
			err,
		)
	}

	return result, nil
}

var _ service.OrderRepository = (*OrderRepository)(nil)
var _ service.OrderQueryRepository = (*OrderRepository)(nil)
