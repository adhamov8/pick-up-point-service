package controller

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/dto"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/events"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/kafka"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/usecase"
)

type OrderController struct {
	orderUseCase *usecase.OrderUseCase
	Producer     kafka.Producer
	topic        string
}

func NewOrderController(orderUseCase *usecase.OrderUseCase, producer kafka.Producer, topic string) *OrderController {
	return &OrderController{
		orderUseCase: orderUseCase,
		Producer:     producer,
		topic:        topic,
	}
}

func (c *OrderController) Close() error {
	if c.Producer != nil {
		return c.Producer.Close()
	}
	return nil
}

func (c *OrderController) AddOrder(ctx context.Context, req *dto.AddOrderDTO) error {
	err := c.orderUseCase.AddOrder(ctx, req)
	if err != nil {
		log.Printf("Failed to add order: %v", err)
		return err
	}

	event := events.OrderEvent{
		OrderID:   req.OrderID,
		Event:     "AddOrder",
		Timestamp: time.Now().Format(time.RFC3339),
		Details:   req,
	}
	c.sendEvent(req.OrderID, event)

	return nil
}

func (c *OrderController) RemoveOrder(ctx context.Context, orderID string) error {
	err := c.orderUseCase.RemoveOrder(ctx, orderID)
	if err != nil {
		log.Printf("Failed to remove order: %v", err)
		return err
	}

	event := events.OrderEvent{
		OrderID:   orderID,
		Event:     "RemoveOrder",
		Timestamp: time.Now().Format(time.RFC3339),
	}
	c.sendEvent(orderID, event)

	return nil
}

func (c *OrderController) DeliverOrders(ctx context.Context, recipientID string, orderIDs []string) error {
	err := c.orderUseCase.DeliverOrders(ctx, recipientID, orderIDs)
	if err != nil {
		log.Printf("Failed to deliver orders: %v", err)
		return err
	}

	for _, orderID := range orderIDs {
		event := events.OrderEvent{
			OrderID:     orderID,
			RecipientID: recipientID,
			Event:       "DeliverOrder",
			Timestamp:   time.Now().Format(time.RFC3339),
		}
		c.sendEvent(orderID, event)
	}

	return nil
}

func (c *OrderController) AcceptReturn(ctx context.Context, recipientID, orderID string) error {
	err := c.orderUseCase.AcceptReturn(ctx, recipientID, orderID)
	if err != nil {
		log.Printf("Failed to accept return: %v", err)
		return err
	}

	event := events.OrderEvent{
		OrderID:     orderID,
		RecipientID: recipientID,
		Event:       "AcceptReturn",
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	c.sendEvent(orderID, event)

	return nil
}

func (c *OrderController) GetOrders(ctx context.Context, recipientID string, lastN int) ([]*dto.OrderDTO, error) {
	orders, err := c.orderUseCase.GetOrders(ctx, recipientID, lastN)
	if err != nil {
		log.Printf("Failed to get orders: %v", err)
		return nil, err
	}
	return orders, nil
}

func (c *OrderController) GetReturns(ctx context.Context, page int) ([]*dto.ReturnDTO, error) {
	returns, err := c.orderUseCase.GetReturns(ctx, page)
	if err != nil {
		log.Printf("Failed to get returns: %v", err)
		return nil, err
	}
	return returns, nil
}

func (c *OrderController) sendEvent(key string, event events.OrderEvent) {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	err = c.Producer.SendMessage(c.topic, key, eventBytes)
	if err != nil {
		log.Printf("Failed to send event to Kafka: %v", err)
	}
}
