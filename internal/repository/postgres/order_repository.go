// internal/repository/postgres/order_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/domain"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"

	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db    *sqlx.DB
	cache interfaces.Cache[string, *domain.Order]
}

func NewOrderRepository(db *sqlx.DB, cache interfaces.Cache[string, *domain.Order]) *OrderRepository {
	return &OrderRepository{
		db:    db,
		cache: cache,
	}
}

func (r *OrderRepository) AddOrder(ctx context.Context, order *domain.Order) error {
	query := `
        INSERT INTO orders (
            order_id, recipient_id, expiry_date, status, weight, cost, packaging_type
        ) VALUES (
            :order_id, :recipient_id, :expiry_date, :status, :weight, :cost, :packaging_type
        )
    `

	_, err := r.db.NamedExecContext(ctx, query, order)
	if err != nil {
		return fmt.Errorf("failed to add order: %w", err)
	}

	r.cache.Set(ctx, order.OrderID, order)
	return nil
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "OrderRepository.GetOrder")
	defer span.Finish()

	order, found := r.cache.Get(ctx, orderID)
	if found {
		return order, nil
	}

	query := `
        SELECT order_id, recipient_id, expiry_date, status, delivery_date, return_date, weight, cost, packaging_type
        FROM orders WHERE order_id = $1
    `

	var fetchedOrder domain.Order
	err := r.db.GetContext(ctx, &fetchedOrder, query, orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	r.cache.Set(ctx, fetchedOrder.OrderID, &fetchedOrder)
	return &fetchedOrder, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, order *domain.Order) error {
	query := `
        UPDATE orders SET
            status = :status,
            delivery_date = :delivery_date,
            return_date = :return_date
        WHERE order_id = :order_id
    `

	_, err := r.db.NamedExecContext(ctx, query, order)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	r.cache.Set(ctx, order.OrderID, order)
	return nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, orderID string) error {
	query := `DELETE FROM orders WHERE order_id = $1`
	result, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	r.cache.Delete(ctx, orderID)
	return nil
}

func (r *OrderRepository) ListOrdersByRecipient(ctx context.Context, recipientID string, limit int) ([]*domain.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "OrderRepository.ListOrdersByRecipient")
	defer span.Finish()

	query := `
        SELECT order_id, recipient_id, expiry_date, status, delivery_date, return_date, weight, cost, packaging_type
        FROM orders
        WHERE recipient_id = $1
        ORDER BY order_id DESC
        LIMIT $2
    `

	var orders []*domain.Order
	err := r.db.SelectContext(ctx, &orders, query, recipientID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	for _, order := range orders {
		r.cache.Set(ctx, order.OrderID, order)
	}

	return orders, nil
}
