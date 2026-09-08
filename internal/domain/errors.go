// Package domain содержит основные бизнес-сущности приложения.
package domain

import "errors"

var (
	// ErrInvalidOrderStatusTransition возникает при попытке выполнить недопустимый переход статуса заказа.
	ErrInvalidOrderStatusTransition = errors.New("invalid order status transition")

	// ErrInvalidOrderStatus возникает при попытке установить неизвестный статус заказа.
	ErrInvalidOrderStatus = errors.New("invalid order status")

	// ErrInvalidUserID возникает при передаче некорректного идентификатора пользователя.
	ErrInvalidUserID = errors.New("user id must be positive")

	// ErrInvalidRestaurantID возникает при передаче некорректного идентификатора ресторана.
	ErrInvalidRestaurantID = errors.New("restaurant id must be positive")

	// ErrInvalidMenuItemID возникает при передаче некорректного идентификатора элемента меню.
	ErrInvalidMenuItemID = errors.New("menu item id must be positive")

	// ErrInvalidOrderID возникает при передаче некорректного идентификатора заказа.
	ErrInvalidOrderID = errors.New("order id must be positive")

	// ErrInvalidLimit возникает при передаче некорректного ограничения количества элементов.
	ErrInvalidLimit = errors.New("limit must be between 1 and 100")

	// ErrRestaurantNotFound возникает, если ресторан не найден.
	ErrRestaurantNotFound = errors.New("restaurant not found")

	// ErrRestaurantUnavailable возникает, если ресторан недоступен.
	ErrRestaurantUnavailable = errors.New("restaurant unavailable")

	// ErrMenuItemNotFound возникает, если элемент меню не найден.
	ErrMenuItemNotFound = errors.New("menu item not found")

	// ErrMenuItemUnavailable возникает, если элемент меню недоступен.
	ErrMenuItemUnavailable = errors.New("menu item unavailable")

	// ErrMenuItemRestaurantMismatch возникает, если элемент меню принадлежит другому ресторану.
	ErrMenuItemRestaurantMismatch = errors.New("menu item belongs to another restaurant")

	// ErrInvalidMenuItemName возникает при пустом названии элемента меню.
	ErrInvalidMenuItemName = errors.New("menu item name must not be empty")

	// ErrInvalidMenuItemPrice возникает при отрицательной цене элемента меню.
	ErrInvalidMenuItemPrice = errors.New("menu item price must not be negative")

	// ErrEmptyMenuItemUpdate возникает, если при обновлении меню не передано ни одного поля.
	ErrEmptyMenuItemUpdate = errors.New("menu item update must contain at least one field")

	// ErrOrderNotFound возникает, если заказ не найден.
	ErrOrderNotFound = errors.New("order not found")

	// ErrOrderRestaurantMismatch возникает, если заказ принадлежит другому ресторану.
	ErrOrderRestaurantMismatch = errors.New("order belongs to another restaurant")

	// ErrEmptyOrder возникает при попытке создать заказ без товаров.
	ErrEmptyOrder = errors.New("order must contain at least one item")

	// ErrInvalidOrderItemQuantity возникает при некорректном количестве товара в заказе.
	ErrInvalidOrderItemQuantity = errors.New("order item quantity must be positive")

	// ErrDuplicateMenuItem возникает при попытке добавить дублирующий элемент меню.
	ErrDuplicateMenuItem = errors.New("duplicate menu item")
)
