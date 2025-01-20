package server

import (
	"context"

	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *OrderServiceServer) DeliverOrders(ctx context.Context, req *order_service.DeliverOrdersRequest) (*emptypb.Empty, error) {
	err := s.ctrl.DeliverOrders(ctx, req.RecipientId, req.OrderIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to deliver orders: %v", err)
	}
	return &emptypb.Empty{}, nil
}
