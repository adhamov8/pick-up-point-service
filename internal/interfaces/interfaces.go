package interfaces

import (
	"context"
	"database/sql"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/domain"
)

type OrderRepository interface {
	AddOrder(ctx context.Context, order *domain.Order) error
	GetOrder(ctx context.Context, orderID string) (*domain.Order, error)
	UpdateOrder(ctx context.Context, order *domain.Order) error
	DeleteOrder(ctx context.Context, orderID string) error
	ListOrdersByRecipient(ctx context.Context, recipientID string, limit int) ([]*domain.Order, error)
}

type ReturnRepository interface {
	AddReturn(ctx context.Context, ret *domain.Return) error
	ListReturns(ctx context.Context, offset, limit int) ([]*domain.Return, error)
}

type TxManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error, opts *sql.TxOptions) error
}

type Metrics interface {
	IncOrdersServed()
}

type Cache[K comparable, V any] interface {
	Set(ctx context.Context, key K, value V, ttl ...time.Duration)
	Get(ctx context.Context, key K) (V, bool)
	Delete(ctx context.Context, key K)
	Flush(ctx context.Context)
}
