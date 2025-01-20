package controller

import (
	"context"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/dto"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/usecase"
)

type Task interface {
	Execute() error
}

type AddOrderTask struct {
	orderUseCase *usecase.OrderUseCase
	req          *dto.AddOrderDTO
}

func (t *AddOrderTask) Execute() error {
	return t.orderUseCase.AddOrder(context.Background(), t.req)
}

type RemoveOrderTask struct {
	orderUseCase *usecase.OrderUseCase
	orderID      string
}

func (t *RemoveOrderTask) Execute() error {
	return t.orderUseCase.RemoveOrder(context.Background(), t.orderID)
}

type DeliverOrdersTask struct {
	orderUseCase *usecase.OrderUseCase
	recipientID  string
	orderIDs     []string
}

func (t *DeliverOrdersTask) Execute() error {
	return t.orderUseCase.DeliverOrders(context.Background(), t.recipientID, t.orderIDs)
}

type AcceptReturnTask struct {
	orderUseCase *usecase.OrderUseCase
	recipientID  string
	orderID      string
}

func (t *AcceptReturnTask) Execute() error {
	return t.orderUseCase.AcceptReturn(context.Background(), t.recipientID, t.orderID)
}
