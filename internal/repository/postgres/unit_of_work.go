package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/talense-tasks/backend-trainee-assignment-autumn-2026-flow-2-marmulevsemyon-f974dedd/internal/service"
)

// UnitOfWork управляет выполнением операций внутри транзакции.
type UnitOfWork struct {
	pool *pgxpool.Pool
}

// NewUnitOfWork создаёт новый менеджер транзакций.
func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{
		pool: pool,
	}
}

// WithinTransaction выполняет функцию внутри транзакции.
func (u *UnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(repositories service.TransactionRepositories) error,
) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	repositories := service.TransactionRepositories{
		Restaurants: NewRestaurantRepository(tx),
		MenuItems:   NewMenuItemRepository(tx),
		Orders:      NewOrderRepository(tx),
	}

	if err := fn(repositories); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

var _ service.UnitOfWork = (*UnitOfWork)(nil)
