package server

import (
	"context"

	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *OrderServiceServer) AcceptReturn(ctx context.Context, req *order_service.AcceptReturnRequest) (*emptypb.Empty, error) {
	err := s.ctrl.AcceptReturn(ctx, req.RecipientId, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to accept return: %v", err)
	}
	return &emptypb.Empty{}, nil
}
