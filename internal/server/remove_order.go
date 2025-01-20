package server

import (
	"context"

	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *OrderServiceServer) RemoveOrder(ctx context.Context, req *order_service.RemoveOrderRequest) (*emptypb.Empty, error) {
	err := s.ctrl.RemoveOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to remove order: %v", err)
	}
	return &emptypb.Empty{}, nil
}
