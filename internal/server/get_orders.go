package server

import (
	"context"

	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OrderServiceServer) GetOrders(ctx context.Context, req *order_service.GetOrdersRequest) (*order_service.GetOrdersResponse, error) {
	orders, err := s.ctrl.GetOrders(ctx, req.RecipientId, int(req.LastN))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get orders: %v", err)
	}

	var orderProtos []*order_service.Order
	for _, order := range orders {
		orderProtos = append(orderProtos, &order_service.Order{
			OrderId:      order.OrderID,
			Status:       order.Status,
			DeliveryDate: order.DeliveryDate,
		})
	}

	return &order_service.GetOrdersResponse{Orders: orderProtos}, nil
}
