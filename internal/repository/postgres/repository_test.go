package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

func TestRestaurantRepository_GetByIDLocked(t *testing.T) {
	pool := newTestPool(t)

	restaurantID := insertTestRestaurant(
		t,
		pool,
		"Test Restaurant",
		nil,
		false,
	)

	repository := NewRestaurantRepository(pool)

	restaurant, err := repository.GetByIDLocked(
		context.Background(),
		restaurantID,
	)
	require.NoError(t, err)

	require.Equal(t, restaurantID, restaurant.ID)
	require.Equal(t, "Test Restaurant", restaurant.Name)

	// В БД description = NULL,
	// repository должен преобразовать его в пустую строку.
	require.Empty(t, restaurant.Description)

	require.False(t, restaurant.IsActive)
	require.False(t, restaurant.CreatedAt.IsZero())
	require.False(t, restaurant.UpdatedAt.IsZero())
}

func TestRestaurantRepository_GetByIDLocked_NotFound(
	t *testing.T,
) {
	pool := newTestPool(t)

	repository := NewRestaurantRepository(pool)

	_, err := repository.GetByIDLocked(
		context.Background(),
		999999,
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrRestaurantNotFound,
	)
}

func TestRestaurantRepository_GetByIDLocked_BlocksUpdate(
	t *testing.T,
) {
	pool := newTestPool(t)

	restaurantID := insertTestRestaurant(
		t,
		pool,
		"Locked Restaurant",
		nil,
		true,
	)

	ctx := context.Background()

	firstTx, err := pool.Begin(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = firstTx.Rollback(context.Background())
	})

	repository := NewRestaurantRepository(firstTx)

	_, err = repository.GetByIDLocked(
		ctx,
		restaurantID,
	)
	require.NoError(t, err)

	secondTx, err := pool.Begin(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = secondTx.Rollback(context.Background())
	})

	// Не ждём блокировку бесконечно.
	_, err = secondTx.Exec(
		ctx,
		`SET LOCAL lock_timeout = '100ms'`,
	)
	require.NoError(t, err)

	_, err = secondTx.Exec(
		ctx,
		`
			UPDATE restaurants
			SET is_active = FALSE
			WHERE id = $1
		`,
		restaurantID,
	)

	require.Error(t, err)

	var pgErr *pgconn.PgError

	require.True(
		t,
		errors.As(err, &pgErr),
	)

	// lock_not_available.
	require.Equal(t, "55P03", pgErr.Code)
}

func TestMenuItemRepository_GetByIDsLocked(t *testing.T) {
	pool := newTestPool(t)

	restaurantID := insertTestRestaurant(
		t,
		pool,
		"Test Restaurant",
		nil,
		true,
	)

	firstID := insertTestMenuItem(
		t,
		pool,
		restaurantID,
		"Pizza",
		nil,
		700,
		true,
	)

	secondID := insertTestMenuItem(
		t,
		pool,
		restaurantID,
		"Pasta",
		"Carbonara",
		600,
		false,
	)

	repository := NewMenuItemRepository(pool)

	menuItems, err := repository.GetByIDsLocked(
		context.Background(),
		[]int64{
			secondID,
			firstID,
			999999,
		},
	)
	require.NoError(t, err)

	// Несуществующий id просто не возвращается.
	// Проверку полноты выполняет OrderService.
	require.Len(t, menuItems, 2)

	menuItemsByID := make(
		map[int64]domain.MenuItem,
		len(menuItems),
	)

	for _, menuItem := range menuItems {
		menuItemsByID[menuItem.ID] = menuItem
	}

	first := menuItemsByID[firstID]

	require.Equal(t, restaurantID, first.RestaurantID)
	require.Equal(t, "Pizza", first.Name)
	require.Empty(t, first.Description)
	require.Equal(t, int64(700), first.Price)
	require.True(t, first.IsAvailable)

	second := menuItemsByID[secondID]

	require.Equal(t, restaurantID, second.RestaurantID)
	require.Equal(t, "Pasta", second.Name)
	require.Equal(t, "Carbonara", second.Description)
	require.Equal(t, int64(600), second.Price)
	require.False(t, second.IsAvailable)
}

func TestOrderRepository_Create(t *testing.T) {
	pool := newTestPool(t)

	restaurantID := insertTestRestaurant(
		t,
		pool,
		"Test Restaurant",
		nil,
		true,
	)

	firstMenuItemID := insertTestMenuItem(
		t,
		pool,
		restaurantID,
		"Pizza",
		nil,
		700,
		true,
	)

	secondMenuItemID := insertTestMenuItem(
		t,
		pool,
		restaurantID,
		"Pasta",
		nil,
		600,
		true,
	)

	repository := NewOrderRepository(pool)

	order := domain.Order{
		RestaurantID: restaurantID,
		Status:       domain.OrderStatusCreated,
		TotalPrice:   2000,
		Items: []domain.OrderItem{
			{
				MenuItemID: firstMenuItemID,
				Name:       "Pizza snapshot",
				UnitPrice:  700,
				Quantity:   2,
			},
			{
				MenuItemID: secondMenuItemID,
				Name:       "Pasta snapshot",
				UnitPrice:  600,
				Quantity:   1,
			},
		},
	}

	createdOrder, err := repository.Create(
		context.Background(),
		order,
	)
	require.NoError(t, err)

	require.Positive(t, createdOrder.ID)
	require.False(t, createdOrder.CreatedAt.IsZero())
	require.False(t, createdOrder.UpdatedAt.IsZero())
	require.Len(t, createdOrder.Items, 2)

	for _, item := range createdOrder.Items {
		require.Positive(t, item.ID)
		require.Equal(t, createdOrder.ID, item.OrderID)
	}

	var (
		dbRestaurantID int64
		dbStatus       string
		dbTotalPrice   int64
	)

	err = pool.QueryRow(
		context.Background(),
		`
			SELECT
				restaurant_id,
				status,
				total_price
			FROM orders
			WHERE id = $1
		`,
		createdOrder.ID,
	).Scan(
		&dbRestaurantID,
		&dbStatus,
		&dbTotalPrice,
	)
	require.NoError(t, err)

	require.Equal(t, restaurantID, dbRestaurantID)
	require.Equal(
		t,
		string(domain.OrderStatusCreated),
		dbStatus,
	)
	require.Equal(t, int64(2000), dbTotalPrice)

	rows, err := pool.Query(
		context.Background(),
		`
			SELECT
				menu_item_id,
				name,
				unit_price,
				quantity
			FROM order_items
			WHERE order_id = $1
			ORDER BY id
		`,
		createdOrder.ID,
	)
	require.NoError(t, err)
	defer rows.Close()

	type persistedOrderItem struct {
		MenuItemID int64
		Name       string
		UnitPrice  int64
		Quantity   int
	}

	persistedItems := make(
		[]persistedOrderItem,
		0,
		2,
	)

	for rows.Next() {
		var item persistedOrderItem

		err := rows.Scan(
			&item.MenuItemID,
			&item.Name,
			&item.UnitPrice,
			&item.Quantity,
		)
		require.NoError(t, err)

		persistedItems = append(
			persistedItems,
			item,
		)
	}

	require.NoError(t, rows.Err())

	require.Equal(
		t,
		[]persistedOrderItem{
			{
				MenuItemID: firstMenuItemID,
				Name:       "Pizza snapshot",
				UnitPrice:  700,
				Quantity:   2,
			},
			{
				MenuItemID: secondMenuItemID,
				Name:       "Pasta snapshot",
				UnitPrice:  600,
				Quantity:   1,
			},
		},
		persistedItems,
	)
}
