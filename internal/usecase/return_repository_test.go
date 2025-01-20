package usecase_test

import (
	"context"
	"database/sql"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"homework-1/internal/domain"
	"homework-1/internal/repository/postgres"
)

func TestReturnRepository(t *testing.T) {
	connStr := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM returns")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	txManager := postgres.NewTxManager(db)
	returnRepo := postgres.NewReturnRepository()

	ret := &domain.Return{
		OrderID:     "order1",
		RecipientID: "recipient1",
		ReturnDate:  "2024-09-30",
	}

	err = txManager.RunInTransaction(ctx, func(ctx context.Context) error {

		err := returnRepo.AddReturn(ctx, ret)
		assert.NoError(t, err)

		returns, err := returnRepo.ListReturns(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(returns))
		assert.Equal(t, ret, returns[0])

		return nil
	}, nil)
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}
