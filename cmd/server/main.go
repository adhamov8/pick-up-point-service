package main

import (
	"context"
	"log"
	"time"

	"gitlab.ozon.dev/ashadkhamov/homework/internal/cache"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/controller"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/domain"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/kafka"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/metrics"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/repository/postgres"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/server"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/tracer"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/usecase"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

func main() {
	initConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tracer.SetupTracer(ctx, "order_service")

	metricsInstance := metrics.GetMetrics()

	connectionString := viper.GetString("database.dsn")
	if connectionString == "" {
		log.Fatal("Database connection string is not set in the config")
	}

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	cacheConfig := cache.CacheConfig{
		Strategy:        cache.LRUStrategy,
		MaxEntries:      1000,
		DefaultTTL:      5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
	orderCache, err := cache.NewCache[string, *domain.Order](cacheConfig)
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.CloseCache(orderCache)

	orderRepo := postgres.NewOrderRepository(db, orderCache)
	returnRepo := postgres.NewReturnRepository(db)

	txManager := postgres.NewTxManager(db)

	producer, err := kafka.NewProducer(viper.GetStringSlice("kafka.brokers"))
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	orderUseCase := usecase.NewOrderUseCase(orderRepo, returnRepo, txManager, metricsInstance)

	orderController := controller.NewOrderController(orderUseCase, producer, viper.GetString("kafka.topic"))

	grpcServer := server.NewOrderServiceServer(orderController)

	go func() {
		grpcAddress := viper.GetString("server.grpc_port")
		if err := server.RunGRPCServer(ctx, grpcAddress, grpcServer, metricsInstance); err != nil {
			log.Fatalf("Failed to run gRPC server: %v", err)
		}
	}()

	<-ctx.Done()
}
