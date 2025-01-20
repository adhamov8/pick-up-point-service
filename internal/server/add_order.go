package server

import (
	"context"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/dto"
	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *OrderServiceServer) AddOrder(ctx context.Context, req *order_service.AddOrderRequest) (*emptypb.Empty, error) {
	dtoReq := &dto.AddOrderDTO{
		OrderID:       req.OrderId,
		RecipientID:   req.RecipientId,
		ExpiryDate:    req.ExpiryDate,
		Weight:        req.Weight,
		PackagingType: req.PackagingType,
	}
	err := s.ctrl.AddOrder(ctx, dtoReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to add order: %v", err)
	}
	return &emptypb.Empty{}, nil
}
