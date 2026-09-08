package domain

import "errors"

var (
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")

	ErrInvalidUserID       = errors.New("user id must be positive")
	ErrInvalidRestaurantID = errors.New("restaurant id must be positive")
	ErrInvalidMenuItemID   = errors.New("menu item id must be positive")

	ErrRestaurantNotFound    = errors.New("restaurant not found")
	ErrRestaurantUnavailable = errors.New("restaurant unavailable")

	ErrMenuItemNotFound           = errors.New("menu item not found")
	ErrMenuItemUnavailable        = errors.New("menu item unavailable")
	ErrMenuItemRestaurantMismatch = errors.New("menu item belongs to another restaurant")
	ErrInvalidMenuItemName        = errors.New("menu item name must not be empty")
	ErrInvalidMenuItemPrice       = errors.New("menu item price must not be negative")
	ErrEmptyMenuItemUpdate        = errors.New("menu item update must contain at least one field")

	ErrEmptyOrder               = errors.New("order must contain at least one item")
	ErrInvalidOrderItemQuantity = errors.New("order item quantity must be positive")
	ErrDuplicateMenuItem        = errors.New("duplicate menu item")
)
