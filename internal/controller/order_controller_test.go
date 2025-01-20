package controller

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/dto"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/repository/postgres"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/usecase"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(topic string, key string, value []byte) error {
	args := m.Called(topic, key, value)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestOrderController_AddOrder(t *testing.T) {
	db := sqlx.NewDb(nil, "postgres")
	orderRepo := postgres.NewOrderRepository(db)
	returnRepo := postgres.NewReturnRepository(db)
	txManager := postgres.NewTxManager(db)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, returnRepo, txManager)

	mockProducer := new(MockProducer)
	topic := "test-topic"

	controller := &OrderController{
		orderUseCase: orderUseCase,
		producer:     mockProducer,
		topic:        topic,
	}

	ctx := context.Background()
	orderID := "order123"
	recipientID := "recipient123"
	expiryDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	weight := float32(5.0)
	packagingType := "bag"

	req := &dto.AddOrderDTO{
		OrderID:       orderID,
		RecipientID:   recipientID,
		ExpiryDate:    expiryDate,
		Weight:        weight,
		PackagingType: packagingType,
	}

	mockProducer.On("SendMessage", topic, orderID, mock.Anything).Return(nil)

	err := controller.AddOrder(ctx, req)
	assert.NoError(t, err)

	mockProducer.AssertCalled(t, "SendMessage", topic, orderID, mock.MatchedBy(func(value []byte) bool {
		var sentEvent map[string]interface{}
		json.Unmarshal(value, &sentEvent)
		return sentEvent["order_id"] == orderID && sentEvent["event"] == "AddOrder"
	}))
}

func TestOrderController_DeliverOrders(t *testing.T) {
	db := sqlx.NewDb(nil, "postgres")
	orderRepo := postgres.NewOrderRepository(db)
	returnRepo := postgres.NewReturnRepository(db)
	txManager := postgres.NewTxManager(db)
	orderUseCase := usecase.NewOrderUseCase(orderRepo, returnRepo, txManager)

	mockProducer := new(MockProducer)
	topic := "test-topic"

	controller := &OrderController{
		orderUseCase: orderUseCase,
		producer:     mockProducer,
		topic:        topic,
	}

	ctx := context.Background()
	recipientID := "recipient123"
	orderIDs := []string{"order1", "order2"}

	for _, orderID := range orderIDs {
		mockProducer.On("SendMessage", topic, orderID, mock.Anything).Return(nil)
	}

	err := controller.DeliverOrders(ctx, recipientID, orderIDs)
	assert.NoError(t, err)

	for _, orderID := range orderIDs {
		mockProducer.AssertCalled(t, "SendMessage", topic, orderID, mock.MatchedBy(func(value []byte) bool {
			var sentEvent map[string]interface{}
			json.Unmarshal(value, &sentEvent)
			return sentEvent["order_id"] == orderID && sentEvent["event"] == "DeliverOrder"
		}))
	}
}
