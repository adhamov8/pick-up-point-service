package usecase_test

import (
	"context"
	"database/sql"
	"log"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/domain"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/repository/postgres"
)

func TestOrderRepository(t *testing.T) {
	connStr := "postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM orders")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	txManager := postgres.NewTxManager(db)
	orderRepo := postgres.NewOrderRepository()

	order := &domain.Order{
		OrderID:       "order1",
		RecipientID:   "recipient1",
		ExpiryDate:    "2024-09-30",
		Status:        "new",
		Weight:        10.0,
		Cost:          100.0,
		PackagingType: "box",
	}

	err = txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		err := orderRepo.AddOrder(ctx, order)
		assert.NoError(t, err)

		gotOrder, err := orderRepo.GetOrder(ctx, "order1")
		assert.NoError(t, err)
		assert.Equal(t, order, gotOrder)

		order.Status = "delivered"
		err = orderRepo.UpdateOrder(ctx, order)
		assert.NoError(t, err)

		gotOrder, err = orderRepo.GetOrder(ctx, "order1")
		assert.NoError(t, err)
		assert.Equal(t, "delivered", gotOrder.Status)

		err = orderRepo.DeleteOrder(ctx, "order1")
		assert.NoError(t, err)

		_, err = orderRepo.GetOrder(ctx, "order1")
		assert.Error(t, err)

		return nil
	}, nil)
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}
}
