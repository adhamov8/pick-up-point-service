package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/controller"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/middleware"
	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc"
)

type OrderServiceServer struct {
	order_service.UnimplementedOrderServiceServer
	ctrl *controller.OrderController
}

func NewOrderServiceServer(ctrl *controller.OrderController) *OrderServiceServer {
	return &OrderServiceServer{ctrl: ctrl}
}

func RunGRPCServer(ctx context.Context, address string, server *OrderServiceServer, metrics interfaces.Metrics) error {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggingInterceptor(server.ctrl.Producer, "order_service")), // Используем Producer
	)
	order_service.RegisterOrderServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	return grpcServer.Serve(lis)
}
