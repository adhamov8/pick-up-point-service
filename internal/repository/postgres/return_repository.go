package postgres

import (
	"context"
	"fmt"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/domain"

	"github.com/jmoiron/sqlx"
)

type ReturnRepository struct {
	db *sqlx.DB
}

func NewReturnRepository(db *sqlx.DB) *ReturnRepository {
	return &ReturnRepository{db: db}
}

func (r *ReturnRepository) AddReturn(ctx context.Context, ret *domain.Return) error {
	query :=
		`INSERT INTO returns (order_id, recipient_id, return_date)
	VALUES (:order_id, :recipient_id, :return_date)`

	_, err := r.db.NamedExecContext(ctx, query, ret)
	if err != nil {
		return fmt.Errorf("failed to add return: %w", err)
	}
	return nil
}

func (r *ReturnRepository) ListReturns(ctx context.Context, offset, limit int) ([]*domain.Return, error) {
	query :=
		`SELECT id, order_id, recipient_id, return_date FROM returns
	ORDER BY return_date DESC
	OFFSET $1 LIMIT $2`

	var returns []*domain.Return
	err := r.db.SelectContext(ctx, &returns, query, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list returns: %w", err)
	}
	return returns, nil
}
