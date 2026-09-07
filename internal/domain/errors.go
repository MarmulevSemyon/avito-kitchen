package domain

import "errors"

var (
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")

	ErrRestaurantNotFound    = errors.New("restaurant not found")
	ErrRestaurantUnavailable = errors.New("restaurant unavailable")

	ErrMenuItemNotFound           = errors.New("menu item not found")
	ErrMenuItemUnavailable        = errors.New("menu item unavailable")
	ErrMenuItemRestaurantMismatch = errors.New("menu item belongs to another restaurant")

	ErrEmptyOrder               = errors.New("order must contain at least one item")
	ErrInvalidOrderItemQuantity = errors.New("order item quantity must be positive")
	ErrDuplicateMenuItem        = errors.New("duplicate menu item")
)
