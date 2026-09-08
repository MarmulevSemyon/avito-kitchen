package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

func TestUnitOfWork_Commit(t *testing.T) {
	pool := newTestPool(t)

	restaurantID := insertTestRestaurant(
		t,
		pool,
		"Test Restaurant",
		nil,
		true,
	)

	menuItemID := insertTestMenuItem(
		t,
		pool,
		restaurantID,
		"Pizza",
		nil,
		700,
		true,
	)

	uow := NewUnitOfWork(pool)

	var createdOrderID int64

	err := uow.WithinTransaction(
		context.Background(),
		func(
			repositories service.TransactionRepositories,
		) error {
			order := domain.Order{
				UserID:       12345,
				RestaurantID: restaurantID,
				Status:       domain.OrderStatusCreated,
				TotalPrice:   1400,
				Items: []domain.OrderItem{
					{
						MenuItemID: menuItemID,
						Name:       "Pizza",
						UnitPrice:  700,
						Quantity:   2,
					},
				},
			}

			createdOrder, err := repositories.Orders.Create(
				context.Background(),
				order,
			)
			if err != nil {
				return err
			}

			createdOrderID = createdOrder.ID

			return nil
		},
	)
	require.NoError(t, err)
	require.Positive(t, createdOrderID)

	var (
		orderCount int
		dbUserID   int64
		dbStatus   string
	)

	err = pool.QueryRow(
		context.Background(),
		`
			SELECT
				COUNT(*),
				MAX(o.user_id),
				MAX(s.code)
			FROM orders AS o
			JOIN order_statuses AS s
				ON s.id = o.status_id
			WHERE o.id = $1
		`,
		createdOrderID,
	).Scan(
		&orderCount,
		&dbUserID,
		&dbStatus,
	)
	require.NoError(t, err)

	require.Equal(t, 1, orderCount)
	require.Equal(t, int64(12345), dbUserID)
	require.Equal(
		t,
		string(domain.OrderStatusCreated),
		dbStatus,
	)

	var itemCount int

	err = pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM order_items
			WHERE order_id = $1
		`,
		createdOrderID,
	).Scan(&itemCount)
	require.NoError(t, err)

	require.Equal(t, 1, itemCount)
}

func TestUnitOfWork_Rollback(t *testing.T) {
	pool := newTestPool(t)

	restaurantID := insertTestRestaurant(
		t,
		pool,
		"Test Restaurant",
		nil,
		true,
	)

	menuItemID := insertTestMenuItem(
		t,
		pool,
		restaurantID,
		"Pizza",
		nil,
		700,
		true,
	)

	uow := NewUnitOfWork(pool)

	expectedErr := errors.New("force rollback")

	var createdOrderID int64

	err := uow.WithinTransaction(
		context.Background(),
		func(
			repositories service.TransactionRepositories,
		) error {
			order := domain.Order{
				UserID:       12345,
				RestaurantID: restaurantID,
				Status:       domain.OrderStatusCreated,
				TotalPrice:   1400,
				Items: []domain.OrderItem{
					{
						MenuItemID: menuItemID,
						Name:       "Pizza",
						UnitPrice:  700,
						Quantity:   2,
					},
				},
			}

			createdOrder, err := repositories.Orders.Create(
				context.Background(),
				order,
			)
			if err != nil {
				return err
			}

			createdOrderID = createdOrder.ID

			return expectedErr
		},
	)

	require.ErrorIs(t, err, expectedErr)
	require.Positive(t, createdOrderID)

	var orderCount int

	err = pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM orders
			WHERE id = $1
		`,
		createdOrderID,
	).Scan(&orderCount)
	require.NoError(t, err)

	require.Zero(t, orderCount)

	var itemCount int

	err = pool.QueryRow(
		context.Background(),
		`
			SELECT COUNT(*)
			FROM order_items
			WHERE order_id = $1
		`,
		createdOrderID,
	).Scan(&itemCount)
	require.NoError(t, err)

	require.Zero(t, itemCount)
}
