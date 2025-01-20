package server

import (
	"context"

	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *OrderServiceServer) GetReturns(ctx context.Context, req *order_service.GetReturnsRequest) (*order_service.GetReturnsResponse, error) {
	returns, err := s.ctrl.GetReturns(ctx, int(req.Page))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get returns: %v", err)
	}

	var returnProtos []*order_service.Return
	for _, ret := range returns {
		returnProtos = append(returnProtos, &order_service.Return{
			OrderId:    ret.OrderID,
			ReturnDate: ret.ReturnDate,
		})
	}

	return &order_service.GetReturnsResponse{Returns: returnProtos}, nil
}
