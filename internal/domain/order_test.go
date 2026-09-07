package domain

import (
	"errors"
	"testing"
)

func TestOrderChangeStatus(t *testing.T) {
	tests := []struct {
		name    string
		from    OrderStatus
		to      OrderStatus
		wantErr bool
	}{
		{
			name: "created to accepted",
			from: OrderStatusCreated,
			to:   OrderStatusAccepted,
		},
		{
			name: "created to rejected",
			from: OrderStatusCreated,
			to:   OrderStatusRejected,
		},
		{
			name: "accepted to preparing",
			from: OrderStatusAccepted,
			to:   OrderStatusPreparing,
		},
		{
			name: "preparing to ready",
			from: OrderStatusPreparing,
			to:   OrderStatusReady,
		},
		{
			name: "ready to completed",
			from: OrderStatusReady,
			to:   OrderStatusCompleted,
		},

		{
			name:    "created to ready is forbidden",
			from:    OrderStatusCreated,
			to:      OrderStatusReady,
			wantErr: true,
		},
		{
			name:    "completed to created is forbidden",
			from:    OrderStatusCompleted,
			to:      OrderStatusCreated,
			wantErr: true,
		},
		{
			name:    "rejected to accepted is forbidden",
			from:    OrderStatusRejected,
			to:      OrderStatusAccepted,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := Order{
				Status: tt.from,
			}

			err := order.ChangeStatus(tt.to)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidOrderStatusTransition) {
					t.Fatalf(
						"expected ErrInvalidOrderStatusTransition, got %v",
						err,
					)
				}

				if order.Status != tt.from {
					t.Fatalf(
						"status changed after invalid transition: got %s, want %s",
						order.Status,
						tt.from,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if order.Status != tt.to {
				t.Fatalf(
					"unexpected status: got %s, want %s",
					order.Status,
					tt.to,
				)
			}
		})
	}
}
