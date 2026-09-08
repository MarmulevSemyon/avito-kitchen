// Package main запускает HTTP API сервиса.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/httpapi"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/repository/postgres"
	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

const (
	defaultHTTPAddress = ":8080"

	defaultDatabaseURL = "postgres://avito:avito@localhost:5432/avito_kitchen?sslmode=disable"
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

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error(
			"failed to create database pool",
			"error",
			err,
		)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error(
			"failed to connect to database",
			"error",
			err,
		)
		os.Exit(1)
	}

	restaurantRepository := postgres.NewRestaurantRepository(pool)
	menuItemRepository := postgres.NewMenuItemRepository(pool)

	restaurantService := service.NewRestaurantService(
		restaurantRepository,
		menuItemRepository,
	)

	orderRepository := postgres.NewOrderRepository(pool)

	uow := postgres.NewUnitOfWork(pool)

	orderService := service.NewOrderService(
		uow,
		orderRepository,
	)

	router := httpapi.NewRouter(
		restaurantService,
		orderService,
	)

	httpAddress := os.Getenv("HTTP_ADDRESS")
	if httpAddress == "" {
		httpAddress = defaultHTTPAddress
	}

	server := &http.Server{
		Addr:              httpAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"api server started",
			"address",
			httpAddress,
		)

		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")

	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error(
				"http server failed",
				"error",
				err,
			)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"failed to shutdown http server",
			"error",
			err,
		)
	}

	logger.Info("api server stopped")
}
