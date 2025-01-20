package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

func (m *TxManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error, opts *sql.TxOptions) (err error) {
	tx, err := m.db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("panic occurred during transaction: %v", p)
			log.Println(err)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx)
	return err
}

func GetTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	if !ok || tx == nil {
		return nil, fmt.Errorf("transaction not found in context")
	}
	return tx, nil
}
