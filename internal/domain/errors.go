package domain

import "errors"

var ErrInvalidOrderStatusTransition = errors.New(
	"invalid order status transition",
)
