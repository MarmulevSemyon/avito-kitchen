package main

import (
	"context"
	"log/slog"
	"time"
)

func processCreatedOrders(
	ctx context.Context,
	logger *slog.Logger,
	client *apiClient,
	statusDelay time.Duration,
) {
	orders, err := client.listCreatedOrders(ctx)
	if err != nil {
		logger.Warn(
			"failed to get created orders",
			"error",
			err,
		)
		return
	}

	if len(orders) == 0 {
		return
	}

	logger.Info(
		"created orders received",
		"count",
		len(orders),
	)

	for _, currentOrder := range orders {
		if ctx.Err() != nil {
			return
		}

		logger.Info(
			"processing order",
			"order_id",
			currentOrder.ID,
			"items",
			len(currentOrder.Items),
			"total_price",
			currentOrder.TotalPrice,
		)

		if !processOrder(
			ctx,
			logger,
			client,
			currentOrder,
			statusDelay,
		) {
			return
		}
	}
}

func processOrder(
	ctx context.Context,
	logger *slog.Logger,
	client *apiClient,
	currentOrder order,
	statusDelay time.Duration,
) bool {
	statuses := []string{
		"ACCEPTED",
		"PREPARING",
		"READY",
		"COMPLETED",
	}

	for index, status := range statuses {
		if err := client.updateStatus(
			ctx,
			currentOrder.ID,
			status,
		); err != nil {
			logger.Warn(
				"failed to update order status",
				"order_id",
				currentOrder.ID,
				"status",
				status,
				"error",
				err,
			)

			return true
		}

		logger.Info(
			"order status updated",
			"order_id",
			currentOrder.ID,
			"status",
			status,
		)

		if index == len(statuses)-1 {
			continue
		}

		if !wait(ctx, statusDelay) {
			return false
		}
	}

	return true
}

func wait(
	ctx context.Context,
	duration time.Duration,
) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false

	case <-timer.C:
		return true
	}
}

func newPollTicker(
	interval time.Duration,
) *time.Ticker {
	return time.NewTicker(interval)
}
