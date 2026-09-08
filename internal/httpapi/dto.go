package httpapi

import (
	"time"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

type restaurantResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type menuItemResponse struct {
	ID           int64     `json:"id"`
	RestaurantID int64     `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        int64     `json:"price"`
	IsAvailable  bool      `json:"is_available"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type createOrderRequest struct {
	UserID       int64                    `json:"user_id"`
	RestaurantID int64                    `json:"restaurant_id"`
	Items        []createOrderItemRequest `json:"items"`
}

type updateOrderStatusRequest struct {
	Status domain.OrderStatus `json:"status"`
}

func toMenuItemResponse(
	menuItem domain.MenuItem,
) menuItemResponse {
	return menuItemResponse{
		ID:           menuItem.ID,
		RestaurantID: menuItem.RestaurantID,
		Name:         menuItem.Name,
		Description:  menuItem.Description,
		Price:        menuItem.Price,
		IsAvailable:  menuItem.IsAvailable,
		CreatedAt:    menuItem.CreatedAt,
		UpdatedAt:    menuItem.UpdatedAt,
	}
}

func toOrderResponse(
	order domain.Order,
) orderResponse {
	items := make(
		[]orderItemResponse,
		0,
		len(order.Items),
	)

	for _, item := range order.Items {
		items = append(
			items,
			orderItemResponse{
				MenuItemID: item.MenuItemID,
				Name:       item.Name,
				UnitPrice:  item.UnitPrice,
				Quantity:   item.Quantity,
			},
		)
	}

	return orderResponse{
		ID:           order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		Status:       order.Status,
		TotalPrice:   order.TotalPrice,
		Items:        items,
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
	}
}

type orderItemResponse struct {
	MenuItemID int64  `json:"menu_item_id"`
	Name       string `json:"name"`
	UnitPrice  int64  `json:"unit_price"`
	Quantity   int    `json:"quantity"`
}

type orderResponse struct {
	ID           int64               `json:"id"`
	UserID       int64               `json:"user_id"`
	RestaurantID int64               `json:"restaurant_id"`
	Status       domain.OrderStatus  `json:"status"`
	TotalPrice   int64               `json:"total_price"`
	Items        []orderItemResponse `json:"items"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

type listOrdersResponse struct {
	Orders     []orderResponse `json:"orders"`
	NextCursor *string         `json:"next_cursor"`
}

type createMenuItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
}

type updateMenuItemRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Price       *int64  `json:"price"`
	IsAvailable *bool   `json:"is_available"`
}

type createOrderItemRequest struct {
	MenuItemID int64 `json:"menu_item_id"`
	Quantity   int   `json:"quantity"`
}
