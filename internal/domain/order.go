package domain

import (
	"fmt"
	"time"
)

// OrderStatus представляет состояние заказа.
type OrderStatus string

const (
	// OrderStatusCreated означает, что заказ создан.
	OrderStatusCreated OrderStatus = "CREATED"

	// OrderStatusAccepted означает, что ресторан принял заказ.
	OrderStatusAccepted OrderStatus = "ACCEPTED"

	// OrderStatusPreparing означает, что заказ готовится.
	OrderStatusPreparing OrderStatus = "PREPARING"

	// OrderStatusReady означает, что заказ готов к выдаче.
	OrderStatusReady OrderStatus = "READY"

	// OrderStatusCompleted означает, что заказ завершён.
	OrderStatusCompleted OrderStatus = "COMPLETED"

	// OrderStatusRejected означает, что заказ отменён.
	OrderStatusRejected OrderStatus = "REJECTED"
)

// Order представляет заказ пользователя.
type Order struct {
	ID           int64
	UserID       int64
	RestaurantID int64
	Status       OrderStatus
	TotalPrice   int64
	Items        []OrderItem
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OrderItem представляет элемент заказа.
type OrderItem struct {
	ID         int64
	OrderID    int64
	MenuItemID int64

	// Снимок MenuItem в момент создания заказа.
	Name      string
	UnitPrice int64
	Quantity  int
}

// CanTransitionTo проверяет возможность перехода заказа в новый статус.
func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	switch s {
	case OrderStatusCreated:
		return next == OrderStatusAccepted ||
			next == OrderStatusRejected

	case OrderStatusAccepted:
		return next == OrderStatusPreparing

	case OrderStatusPreparing:
		return next == OrderStatusReady

	case OrderStatusReady:
		return next == OrderStatusCompleted

	case OrderStatusCompleted, OrderStatusRejected:
		return false

	default:
		return false
	}
}

// ChangeStatus изменяет статус заказа с проверкой допустимого перехода.
func (o *Order) ChangeStatus(next OrderStatus) error {
	if !o.Status.CanTransitionTo(next) {
		return fmt.Errorf(
			"%w: %s -> %s",
			ErrInvalidOrderStatusTransition,
			o.Status,
			next,
		)
	}

	o.Status = next

	return nil
}
