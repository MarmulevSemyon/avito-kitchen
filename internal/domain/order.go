package domain

import (
	"fmt"
	"time"
)

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "CREATED"
	OrderStatusAccepted  OrderStatus = "ACCEPTED"
	OrderStatusPreparing OrderStatus = "PREPARING"
	OrderStatusReady     OrderStatus = "READY"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

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

type OrderItem struct {
	ID         int64
	OrderID    int64
	MenuItemID int64

	// Снимок MenuItem в момент создания заказа.
	Name      string
	UnitPrice int64
	Quantity  int
}

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
