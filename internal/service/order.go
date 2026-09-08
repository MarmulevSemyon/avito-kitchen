package service

import (
	"context"
	"fmt"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

// RestaurantRepository describes restaurant operations required
// during order creation.
//
// GetByIDLocked must prevent concurrent modification of the returned
// restaurant until the current transaction is completed.
type RestaurantRepository interface {
	GetByIDLocked(
		ctx context.Context,
		id int64,
	) (domain.Restaurant, error)
}

// MenuItemRepository describes menu operations required
// during order creation.
//
// GetByIDsLocked must prevent concurrent modification of the returned
// menu items until the current transaction is completed.
type MenuItemRepository interface {
	GetByIDsLocked(
		ctx context.Context,
		ids []int64,
	) ([]domain.MenuItem, error)
}

// OrderRepository describes order persistence operations.
type OrderRepository interface {
	Create(
		ctx context.Context,
		order domain.Order,
	) (domain.Order, error)
}

// TransactionRepositories contains repositories bound to the same
// database transaction.
type TransactionRepositories struct {
	Restaurants RestaurantRepository
	MenuItems   MenuItemRepository
	Orders      OrderRepository
}

// UnitOfWork defines the transaction boundary for an application use case.
//
// The implementation must:
//   - start a transaction before calling fn;
//   - commit if fn returns nil;
//   - rollback if fn returns an error.
type UnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(repositories TransactionRepositories) error,
	) error
}

type OrderService struct {
	uow UnitOfWork
}

func NewOrderService(uow UnitOfWork) *OrderService {
	return &OrderService{
		uow: uow,
	}
}

type CreateOrderInput struct {
	UserID       int64
	RestaurantID int64
	Items        []CreateOrderItemInput
}

type CreateOrderItemInput struct {
	MenuItemID int64
	Quantity   int
}

func (s *OrderService) CreateOrder(
	ctx context.Context,
	input CreateOrderInput,
) (domain.Order, error) {
	if input.UserID <= 0 {
		return domain.Order{}, domain.ErrInvalidUserID
	}

	if len(input.Items) == 0 {
		return domain.Order{}, domain.ErrEmptyOrder
	}

	if err := validateOrderItems(input.Items); err != nil {
		return domain.Order{}, err
	}

	menuItemIDs := make([]int64, 0, len(input.Items))

	for _, item := range input.Items {
		menuItemIDs = append(menuItemIDs, item.MenuItemID)
	}

	var createdOrder domain.Order

	err := s.uow.WithinTransaction(
		ctx,
		func(repositories TransactionRepositories) error {
			restaurant, err := repositories.Restaurants.GetByIDLocked(
				ctx,
				input.RestaurantID,
			)
			if err != nil {
				return fmt.Errorf("get restaurant: %w", err)
			}

			if !restaurant.IsActive {
				return domain.ErrRestaurantUnavailable
			}

			menuItems, err := repositories.MenuItems.GetByIDsLocked(
				ctx,
				menuItemIDs,
			)
			if err != nil {
				return fmt.Errorf("get menu items: %w", err)
			}

			menuItemsByID := make(
				map[int64]domain.MenuItem,
				len(menuItems),
			)

			for _, menuItem := range menuItems {
				menuItemsByID[menuItem.ID] = menuItem
			}

			order := domain.Order{
				UserID:       input.UserID,
				RestaurantID: input.RestaurantID,
				Status:       domain.OrderStatusCreated,
				Items: make(
					[]domain.OrderItem,
					0,
					len(input.Items),
				),
			}

			for _, requestedItem := range input.Items {
				menuItem, exists := menuItemsByID[requestedItem.MenuItemID]
				if !exists {
					return fmt.Errorf(
						"%w: id=%d",
						domain.ErrMenuItemNotFound,
						requestedItem.MenuItemID,
					)
				}

				if menuItem.RestaurantID != input.RestaurantID {
					return fmt.Errorf(
						"%w: menu_item_id=%d",
						domain.ErrMenuItemRestaurantMismatch,
						menuItem.ID,
					)
				}

				if !menuItem.IsAvailable {
					return fmt.Errorf(
						"%w: menu_item_id=%d",
						domain.ErrMenuItemUnavailable,
						menuItem.ID,
					)
				}

				orderItem := domain.OrderItem{
					MenuItemID: menuItem.ID,
					Name:       menuItem.Name,
					UnitPrice:  menuItem.Price,
					Quantity:   requestedItem.Quantity,
				}

				order.Items = append(order.Items, orderItem)

				order.TotalPrice += menuItem.Price *
					int64(requestedItem.Quantity)
			}

			createdOrder, err = repositories.Orders.Create(
				ctx,
				order,
			)
			if err != nil {
				return fmt.Errorf("create order: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return domain.Order{}, err
	}

	return createdOrder, nil
}

func validateOrderItems(items []CreateOrderItemInput) error {
	seen := make(map[int64]struct{}, len(items))

	for _, item := range items {
		if item.Quantity <= 0 {
			return fmt.Errorf(
				"%w: menu_item_id=%d",
				domain.ErrInvalidOrderItemQuantity,
				item.MenuItemID,
			)
		}

		if _, exists := seen[item.MenuItemID]; exists {
			return fmt.Errorf(
				"%w: menu_item_id=%d",
				domain.ErrDuplicateMenuItem,
				item.MenuItemID,
			)
		}

		seen[item.MenuItemID] = struct{}{}
	}

	return nil
}
