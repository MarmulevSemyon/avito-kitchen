package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/domain"
)

type userValidationUnitOfWorkStub struct {
	calls int
}

func (u *userValidationUnitOfWorkStub) WithinTransaction(
	_ context.Context,
	_ func(repositories TransactionRepositories) error,
) error {
	u.calls++

	return nil
}

func TestOrderService_CreateOrder_InvalidUserID(t *testing.T) {
	tests := []struct {
		name   string
		userID int64
	}{
		{
			name:   "zero user id",
			userID: 0,
		},
		{
			name:   "negative user id",
			userID: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := &userValidationUnitOfWorkStub{}

			orderService := NewOrderService(uow)

			_, err := orderService.CreateOrder(
				context.Background(),
				CreateOrderInput{
					UserID:       tt.userID,
					RestaurantID: 1,
					Items: []CreateOrderItemInput{
						{
							MenuItemID: 10,
							Quantity:   1,
						},
					},
				},
			)

			require.ErrorIs(
				t,
				err,
				domain.ErrInvalidUserID,
			)

			require.Equal(t, 0, uow.calls)
		})
	}
}
