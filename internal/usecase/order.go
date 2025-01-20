package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/domain"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/dto"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"
)

type OrderUseCase struct {
	orderRepo  interfaces.OrderRepository
	returnRepo interfaces.ReturnRepository
	txManager  interfaces.TxManager
	metrics    interfaces.Metrics
}

func NewOrderUseCase(orderRepo interfaces.OrderRepository, returnRepo interfaces.ReturnRepository, txManager interfaces.TxManager, metrics interfaces.Metrics) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:  orderRepo,
		returnRepo: returnRepo,
		txManager:  txManager,
		metrics:    metrics,
	}
}

func (uc *OrderUseCase) AddOrder(ctx context.Context, req *dto.AddOrderDTO) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validate add order request: %w", err)
	}

	expiryDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return domain.ErrInvalidInput
	}

	order := &domain.Order{
		OrderID:       req.OrderID,
		RecipientID:   req.RecipientID,
		ExpiryDate:    expiryDate,
		Status:        "new",
		Weight:        req.Weight,
		Cost:          calculateCost(req.Weight, req.PackagingType),
		PackagingType: req.PackagingType,
	}

	err = uc.orderRepo.AddOrder(ctx, order)
	if err != nil {
		return err
	}

	uc.metrics.IncOrdersServed()

	return nil
}

func (uc *OrderUseCase) RemoveOrder(ctx context.Context, orderID string) error {
	return uc.orderRepo.DeleteOrder(ctx, orderID)
}

func (uc *OrderUseCase) DeliverOrders(ctx context.Context, recipientID string, orderIDs []string) error {
	for _, orderID := range orderIDs {
		order, err := uc.orderRepo.GetOrder(ctx, orderID)
		if err != nil {
			return err
		}
		if order.RecipientID != recipientID {
			return domain.ErrPermissionDenied
		}
		order.Status = "delivered"
		order.DeliveryDate = sql.NullTime{Time: time.Now(), Valid: true}
		if err := uc.orderRepo.UpdateOrder(ctx, order); err != nil {
			return err
		}
	}
	return nil
}

func (uc *OrderUseCase) AcceptReturn(ctx context.Context, recipientID, orderID string) error {
	return uc.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		order, err := uc.orderRepo.GetOrder(ctx, orderID)
		if err != nil {
			return err
		}
		if order.RecipientID != recipientID {
			return domain.ErrPermissionDenied
		}
		order.Status = "returned"
		order.ReturnDate = sql.NullTime{Time: time.Now(), Valid: true}
		if err := uc.orderRepo.UpdateOrder(ctx, order); err != nil {
			return err
		}
		ret := &domain.Return{
			OrderID:     orderID,
			RecipientID: recipientID,
			ReturnDate:  time.Now(),
		}
		return uc.returnRepo.AddReturn(ctx, ret)
	}, nil)
}

func (uc *OrderUseCase) GetOrders(ctx context.Context, recipientID string, lastN int) ([]*dto.OrderDTO, error) {
	orders, err := uc.orderRepo.ListOrdersByRecipient(ctx, recipientID, lastN)
	if err != nil {
		return nil, err
	}

	var orderDTOs []*dto.OrderDTO
	for _, order := range orders {
		dtoOrder := &dto.OrderDTO{
			OrderID:       order.OrderID,
			RecipientID:   order.RecipientID,
			ExpiryDate:    order.ExpiryDate.Format("2006-01-02"),
			Status:        order.Status,
			Weight:        order.Weight,
			Cost:          order.Cost,
			PackagingType: order.PackagingType,
		}
		if order.DeliveryDate.Valid {
			dtoOrder.DeliveryDate = order.DeliveryDate.Time.Format("2006-01-02")
		}
		if order.ReturnDate.Valid {
			dtoOrder.ReturnDate = order.ReturnDate.Time.Format("2006-01-02")
		}
		orderDTOs = append(orderDTOs, dtoOrder)
	}
	return orderDTOs, nil
}

func (uc *OrderUseCase) GetReturns(ctx context.Context, page int) ([]*dto.ReturnDTO, error) {
	pageSize := 10
	offset := (page - 1) * pageSize
	returns, err := uc.returnRepo.ListReturns(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	var returnDTOs []*dto.ReturnDTO
	for _, ret := range returns {
		returnDTOs = append(returnDTOs, &dto.ReturnDTO{
			OrderID:     ret.OrderID,
			RecipientID: ret.RecipientID,
			ReturnDate:  ret.ReturnDate.Format("2006-01-02"),
		})
	}
	return returnDTOs, nil
}

func calculateCost(weight float32, packagingType string) float32 {
	baseCost := weight * 10
	switch packagingType {
	case "bag":
		return baseCost + 5
	case "box":
		return baseCost + 10
	case "film":
		return baseCost + 2
	default:
		return baseCost
	}
}
