package domain

import "errors"

var (
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")
	ErrInvalidOrderStatus           = errors.New("invalid order status")

	ErrInvalidUserID       = errors.New("user id must be positive")
	ErrInvalidRestaurantID = errors.New("restaurant id must be positive")
	ErrInvalidMenuItemID   = errors.New("menu item id must be positive")
	ErrInvalidOrderID      = errors.New("order id must be positive")
	ErrInvalidLimit        = errors.New("limit must be between 1 and 100")

	ErrRestaurantNotFound    = errors.New("restaurant not found")
	ErrRestaurantUnavailable = errors.New("restaurant unavailable")

	ErrMenuItemNotFound           = errors.New("menu item not found")
	ErrMenuItemUnavailable        = errors.New("menu item unavailable")
	ErrMenuItemRestaurantMismatch = errors.New("menu item belongs to another restaurant")
	ErrInvalidMenuItemName        = errors.New("menu item name must not be empty")
	ErrInvalidMenuItemPrice       = errors.New("menu item price must not be negative")
	ErrEmptyMenuItemUpdate        = errors.New("menu item update must contain at least one field")

	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderRestaurantMismatch = errors.New("order belongs to another restaurant")

	ErrEmptyOrder               = errors.New("order must contain at least one item")
	ErrInvalidOrderItemQuantity = errors.New("order item quantity must be positive")
	ErrDuplicateMenuItem        = errors.New("duplicate menu item")
)
