package middleware

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/kafka"

	"google.golang.org/grpc"
)

func LoggingInterceptor(producer kafka.Producer, topic string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			event := map[string]interface{}{
				"method":    info.FullMethod,
				"timestamp": start.Format(time.RFC3339),
				"duration":  duration.String(),
				"error":     err.Error(),
			}

			eventBytes, _ := json.Marshal(event)
			producer.SendMessage(topic, info.FullMethod, eventBytes)
		}

		return resp, err
	}
}
