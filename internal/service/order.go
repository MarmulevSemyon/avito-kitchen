// Package service содержит бизнес-логику приложения.
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

const (
	defaultOrdersLimit = 20
	maxOrdersLimit     = 100
)

// RestaurantRepository содержит операции с ресторанами,
// необходимые внутри транзакционных сценариев оформления заказа.
type RestaurantRepository interface {
	GetByIDLocked(
		ctx context.Context,
		id int64,
	) (domain.Restaurant, error)
}

// MenuItemRepository содержит операции с пунктами меню,
// необходимые внутри транзакционных сценариев оформления заказа.
type MenuItemRepository interface {
	GetByIDsLocked(
		ctx context.Context,
		ids []int64,
	) ([]domain.MenuItem, error)
}

// OrderRepository содержит операции с заказами,
// которые должны выполняться в рамках текущей транзакции.
type OrderRepository interface {
	Create(
		ctx context.Context,
		order domain.Order,
	) (domain.Order, error)

	GetByIDLocked(
		ctx context.Context,
		id int64,
	) (domain.Order, error)

	UpdateStatus(
		ctx context.Context,
		order domain.Order,
	) (domain.Order, error)
}

// OrderQueryRepository содержит нетранзакционные операции чтения заказов.
type OrderQueryRepository interface {
	GetByID(
		ctx context.Context,
		id int64,
	) (domain.Order, error)

	ListByRestaurant(
		ctx context.Context,
		restaurantID int64,
		status *domain.OrderStatus,
		limit int,
		cursor *OrderCursor,
	) ([]domain.Order, error)
}

// TransactionRepositories содержит репозитории, работающие внутри транзакции.
type TransactionRepositories struct {
	Restaurants RestaurantRepository
	MenuItems   MenuItemRepository
	Orders      OrderRepository
}

// UnitOfWork управляет выполнением операций внутри транзакции.
type UnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(repositories TransactionRepositories) error,
	) error
}

// OrderService реализует бизнес-логику работы с заказами.
type OrderService struct {
	uow     UnitOfWork
	queries OrderQueryRepository
}

// NewOrderService создаёт сервис заказов.
func NewOrderService(
	uow UnitOfWork,
	queries OrderQueryRepository,
) *OrderService {
	return &OrderService{
		uow:     uow,
		queries: queries,
	}
}

// CreateOrderInput содержит данные для создания заказа.
type CreateOrderInput struct {
	UserID       int64
	RestaurantID int64
	Items        []CreateOrderItemInput
}

// CreateOrderItemInput содержит данные позиции заказа.
type CreateOrderItemInput struct {
	MenuItemID int64
	Quantity   int
}

// OrderCursor содержит курсор пагинации заказов.
type OrderCursor struct {
	CreatedAt time.Time
	ID        int64
}

// ListRestaurantOrdersInput содержит параметры списка заказов ресторана.
type ListRestaurantOrdersInput struct {
	RestaurantID int64
	Status       *domain.OrderStatus
	Limit        int
	Cursor       *OrderCursor
}

// ListRestaurantOrdersResult содержит результат пагинации заказов.
type ListRestaurantOrdersResult struct {
	Orders     []domain.Order
	NextCursor *OrderCursor
}

// UpdateOrderStatusInput содержит данные для изменения статуса заказа.
type UpdateOrderStatusInput struct {
	RestaurantID int64
	OrderID      int64
	Status       domain.OrderStatus
}

// CreateOrder создаёт новый заказ пользователя.
func (s *OrderService) CreateOrder(
	ctx context.Context,
	input CreateOrderInput,
) (domain.Order, error) {
	if input.UserID <= 0 {
		return domain.Order{}, domain.ErrInvalidUserID
	}

	if input.RestaurantID <= 0 {
		return domain.Order{}, domain.ErrInvalidRestaurantID
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

// GetOrder возвращает заказ по идентификатору.
func (s *OrderService) GetOrder(
	ctx context.Context,
	orderID int64,
) (domain.Order, error) {
	if orderID <= 0 {
		return domain.Order{}, domain.ErrInvalidOrderID
	}

	order, err := s.queries.GetByID(ctx, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("get order: %w", err)
	}

	return order, nil
}

// ListRestaurantOrders возвращает список заказов ресторана с пагинацией.
func (s *OrderService) ListRestaurantOrders(
	ctx context.Context,
	input ListRestaurantOrdersInput,
) (ListRestaurantOrdersResult, error) {
	if input.RestaurantID <= 0 {
		return ListRestaurantOrdersResult{},
			domain.ErrInvalidRestaurantID
	}

	if input.Status != nil && !isValidOrderStatus(*input.Status) {
		return ListRestaurantOrdersResult{},
			domain.ErrInvalidOrderStatus
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultOrdersLimit
	}

	if limit < 1 || limit > maxOrdersLimit {
		return ListRestaurantOrdersResult{},
			domain.ErrInvalidLimit
	}

	// Запрашиваем на одну запись больше, чтобы определить,
	// существует ли следующая страница.
	orders, err := s.queries.ListByRestaurant(
		ctx,
		input.RestaurantID,
		input.Status,
		limit+1,
		input.Cursor,
	)
	if err != nil {
		return ListRestaurantOrdersResult{},
			fmt.Errorf("list restaurant orders: %w", err)
	}

	result := ListRestaurantOrdersResult{
		Orders: orders,
	}

	if len(orders) > limit {
		result.Orders = orders[:limit]

		last := result.Orders[len(result.Orders)-1]

		result.NextCursor = &OrderCursor{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		}
	}

	return result, nil
}

// UpdateStatus изменяет статус заказа.
func (s *OrderService) UpdateStatus(
	ctx context.Context,
	input UpdateOrderStatusInput,
) (domain.Order, error) {
	if input.RestaurantID <= 0 {
		return domain.Order{}, domain.ErrInvalidRestaurantID
	}

	if input.OrderID <= 0 {
		return domain.Order{}, domain.ErrInvalidOrderID
	}

	if !isValidOrderStatus(input.Status) {
		return domain.Order{}, domain.ErrInvalidOrderStatus
	}

	var updatedOrder domain.Order

	err := s.uow.WithinTransaction(
		ctx,
		func(repositories TransactionRepositories) error {
			order, err := repositories.Orders.GetByIDLocked(
				ctx,
				input.OrderID,
			)
			if err != nil {
				return fmt.Errorf("get order: %w", err)
			}

			if order.RestaurantID != input.RestaurantID {
				return domain.ErrOrderRestaurantMismatch
			}

			if err := order.ChangeStatus(input.Status); err != nil {
				return err
			}

			updatedOrder, err = repositories.Orders.UpdateStatus(
				ctx,
				order,
			)
			if err != nil {
				return fmt.Errorf("update order status: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return domain.Order{}, err
	}

	return updatedOrder, nil
}

func validateOrderItems(items []CreateOrderItemInput) error {
	seen := make(map[int64]struct{}, len(items))

	for _, item := range items {
		if item.MenuItemID <= 0 {
			return domain.ErrInvalidMenuItemID
		}

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

func isValidOrderStatus(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderStatusCreated,
		domain.OrderStatusAccepted,
		domain.OrderStatusPreparing,
		domain.OrderStatusReady,
		domain.OrderStatusCompleted,
		domain.OrderStatusRejected:
		return true

	default:
		return false
	}
}
