package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

type restaurantRepositoryStub struct {
	restaurant domain.Restaurant
	err        error
	calls      int
}

func (r *restaurantRepositoryStub) GetByIDLocked(
	_ context.Context,
	_ int64,
) (domain.Restaurant, error) {
	r.calls++

	if r.err != nil {
		return domain.Restaurant{}, r.err
	}

	return r.restaurant, nil
}

type menuItemRepositoryStub struct {
	items []domain.MenuItem
	err   error
	calls int
}

func (r *menuItemRepositoryStub) GetByIDsLocked(
	_ context.Context,
	_ []int64,
) ([]domain.MenuItem, error) {
	r.calls++

	if r.err != nil {
		return nil, r.err
	}

	return r.items, nil
}

type orderRepositoryStub struct {
	err      error
	calls    int
	received domain.Order
}

func (r *orderRepositoryStub) Create(
	_ context.Context,
	order domain.Order,
) (domain.Order, error) {
	r.calls++
	r.received = order

	if r.err != nil {
		return domain.Order{}, r.err
	}

	order.ID = 1

	return order, nil
}

type unitOfWorkStub struct {
	repositories TransactionRepositories

	calls      int
	committed  bool
	rolledBack bool
}

func (u *unitOfWorkStub) WithinTransaction(
	ctx context.Context,
	fn func(repositories TransactionRepositories) error,
) error {
	u.calls++

	err := fn(u.repositories)
	if err != nil {
		u.rolledBack = true
		return err
	}

	u.committed = true

	return nil
}

func TestOrderService_CreateOrder(t *testing.T) {
	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: true,
		},
	}

	menuRepo := &menuItemRepositoryStub{
		items: []domain.MenuItem{
			{
				ID:           10,
				RestaurantID: 1,
				Name:         "Burger",
				Price:        50000,
				IsAvailable:  true,
			},
			{
				ID:           20,
				RestaurantID: 1,
				Name:         "Cola",
				Price:        15000,
				IsAvailable:  true,
			},
		},
	}

	orderRepo := &orderRepositoryStub{}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      orderRepo,
		},
	}

	orderService := NewOrderService(uow)

	order, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   2,
				},
				{
					MenuItemID: 20,
					Quantity:   1,
				},
			},
		},
	)

	require.NoError(t, err)

	require.Equal(t, 1, uow.calls)
	require.True(t, uow.committed)
	require.False(t, uow.rolledBack)

	require.Equal(t, 1, restaurantRepo.calls)
	require.Equal(t, 1, menuRepo.calls)
	require.Equal(t, 1, orderRepo.calls)

	require.Equal(t, int64(1), order.ID)
	require.Equal(t, int64(12345), order.UserID)
	require.Equal(t, int64(1), order.RestaurantID)
	require.Equal(t, domain.OrderStatusCreated, order.Status)
	require.Equal(t, int64(115000), order.TotalPrice)

	require.Equal(t, int64(12345), orderRepo.received.UserID)

	require.Len(t, order.Items, 2)

	require.Equal(t, int64(10), order.Items[0].MenuItemID)
	require.Equal(t, "Burger", order.Items[0].Name)
	require.Equal(t, int64(50000), order.Items[0].UnitPrice)
	require.Equal(t, 2, order.Items[0].Quantity)

	require.Equal(t, int64(20), order.Items[1].MenuItemID)
	require.Equal(t, "Cola", order.Items[1].Name)
	require.Equal(t, int64(15000), order.Items[1].UnitPrice)
	require.Equal(t, 1, order.Items[1].Quantity)
}

func TestOrderService_CreateOrder_EmptyOrder(t *testing.T) {
	uow := &unitOfWorkStub{}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
		},
	)

	require.ErrorIs(t, err, domain.ErrEmptyOrder)

	// Некорректный input обнаруживается до открытия транзакции.
	require.Equal(t, 0, uow.calls)
}

func TestOrderService_CreateOrder_InvalidQuantity(t *testing.T) {
	uow := &unitOfWorkStub{}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   0,
				},
			},
		},
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrInvalidOrderItemQuantity,
	)

	// Для проверки quantity база данных не нужна.
	require.Equal(t, 0, uow.calls)
}

func TestOrderService_CreateOrder_DuplicateMenuItem(t *testing.T) {
	uow := &unitOfWorkStub{}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
				{
					MenuItemID: 10,
					Quantity:   2,
				},
			},
		},
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrDuplicateMenuItem,
	)

	require.Equal(t, 0, uow.calls)
}

func TestOrderService_CreateOrder_RestaurantUnavailable(t *testing.T) {
	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: false,
		},
	}

	menuRepo := &menuItemRepositoryStub{}
	orderRepo := &orderRepositoryStub{}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      orderRepo,
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrRestaurantUnavailable,
	)

	require.Equal(t, 1, uow.calls)
	require.False(t, uow.committed)
	require.True(t, uow.rolledBack)
	require.Equal(t, 1, restaurantRepo.calls)

	// Если ресторан недоступен, меню уже читать незачем.
	require.Equal(t, 0, menuRepo.calls)

	// Заказ не должен попасть в repository.
	require.Equal(t, 0, orderRepo.calls)
}

func TestOrderService_CreateOrder_MenuItemNotFound(t *testing.T) {
	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: true,
		},
	}

	// Пользователь запросил item=10,
	// но repository его не вернул.
	menuRepo := &menuItemRepositoryStub{
		items: nil,
	}

	orderRepo := &orderRepositoryStub{}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      orderRepo,
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(t, err, domain.ErrMenuItemNotFound)

	require.True(t, uow.rolledBack)
	require.False(t, uow.committed)

	require.Equal(t, 0, orderRepo.calls)
}

func TestOrderService_CreateOrder_MenuItemUnavailable(t *testing.T) {
	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: true,
		},
	}

	menuRepo := &menuItemRepositoryStub{
		items: []domain.MenuItem{
			{
				ID:           10,
				RestaurantID: 1,
				Name:         "Burger",
				Price:        50000,
				IsAvailable:  false,
			},
		},
	}

	orderRepo := &orderRepositoryStub{}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      orderRepo,
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrMenuItemUnavailable,
	)

	require.True(t, uow.rolledBack)
	require.False(t, uow.committed)

	require.Equal(t, 0, orderRepo.calls)
}

func TestOrderService_CreateOrder_MenuItemRestaurantMismatch(t *testing.T) {
	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: true,
		},
	}

	menuRepo := &menuItemRepositoryStub{
		items: []domain.MenuItem{
			{
				ID:           10,
				RestaurantID: 2,
				Name:         "Burger",
				Price:        50000,
				IsAvailable:  true,
			},
		},
	}

	orderRepo := &orderRepositoryStub{}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      orderRepo,
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(
		t,
		err,
		domain.ErrMenuItemRestaurantMismatch,
	)

	require.True(t, uow.rolledBack)
	require.False(t, uow.committed)

	require.Equal(t, 0, orderRepo.calls)
}

func TestOrderService_CreateOrder_RestaurantRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	restaurantRepo := &restaurantRepositoryStub{
		err: repositoryErr,
	}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   &menuItemRepositoryStub{},
			Orders:      &orderRepositoryStub{},
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(t, err, repositoryErr)

	require.True(t, uow.rolledBack)
	require.False(t, uow.committed)
}

func TestOrderService_CreateOrder_MenuRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: true,
		},
	}

	menuRepo := &menuItemRepositoryStub{
		err: repositoryErr,
	}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      &orderRepositoryStub{},
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(t, err, repositoryErr)

	require.True(t, uow.rolledBack)
	require.False(t, uow.committed)
}

func TestOrderService_CreateOrder_OrderRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database error")

	restaurantRepo := &restaurantRepositoryStub{
		restaurant: domain.Restaurant{
			ID:       1,
			IsActive: true,
		},
	}

	menuRepo := &menuItemRepositoryStub{
		items: []domain.MenuItem{
			{
				ID:           10,
				RestaurantID: 1,
				Name:         "Burger",
				Price:        50000,
				IsAvailable:  true,
			},
		},
	}

	orderRepo := &orderRepositoryStub{
		err: repositoryErr,
	}

	uow := &unitOfWorkStub{
		repositories: TransactionRepositories{
			Restaurants: restaurantRepo,
			MenuItems:   menuRepo,
			Orders:      orderRepo,
		},
	}

	orderService := NewOrderService(uow)

	_, err := orderService.CreateOrder(
		context.Background(),
		CreateOrderInput{
			UserID:       12345,
			RestaurantID: 1,
			Items: []CreateOrderItemInput{
				{
					MenuItemID: 10,
					Quantity:   1,
				},
			},
		},
	)

	require.ErrorIs(t, err, repositoryErr)
	require.Equal(t, 1, orderRepo.calls)

	require.Equal(t, int64(12345), orderRepo.received.UserID)

	require.True(t, uow.rolledBack)
	require.False(t, uow.committed)
}
