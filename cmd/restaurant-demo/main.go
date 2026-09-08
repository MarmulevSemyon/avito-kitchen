// Package main запускает демонстрационный клиент ресторана.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error(
			"failed to load config",
			"error",
			err,
		)
		os.Exit(1)
	}

	client := newAPIClient(
		cfg.APIBaseURL,
		cfg.RestaurantID,
	)

	logger.Info(
		"demo restaurant started",
		"restaurant_id",
		cfg.RestaurantID,
		"api_base_url",
		cfg.APIBaseURL,
		"poll_interval",
		cfg.PollInterval.String(),
	)

	processCreatedOrders(
		ctx,
		logger,
		client,
		cfg.StatusDelay,
	)

	ticker := newPollTicker(cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("demo restaurant stopped")
			return

		case <-ticker.C:
			processCreatedOrders(
				ctx,
				logger,
				client,
				cfg.StatusDelay,
			)
		}
	}
}
