APP_NAME = gitlab.ozon.dev/ashadkhamov/homework
MAIN_FILE = cmd/server/main.go
PROTO_DIR = api
OUT_DIR = pkg
GOOGLEAPIS_DIR = googleapis
DB_DSN = postgres://postgres:postgres@localhost:5433/postgres?sslmode=disable

.PHONY: all deps generate build run clean up down migrate test

all: deps generate build

deps:
	go mod tidy

generate:
	rm -f api/order_service/v1/*.pb.go
	rm -f pkg/order_service/v1/*.pb.go
	protoc \
		--proto_path=. \
		--proto_path=api \
		--proto_path=$(GOOGLEAPIS_DIR) \
		--go_out=paths=source_relative:./ \
		--go-grpc_out=paths=source_relative:./ \
		api/order_service/v1/order_service.proto
	mkdir -p pkg/order_service/v1
	mv api/order_service/v1/*.pb.go pkg/order_service/v1/

build: generate
	go build -o cmd/server/server ./cmd/server
	go build -o cmd/notifier/notifier ./cmd/notifier
	go build -o cmd/client/client ./cmd/client

run:
	go run cmd/server/main.go

clean:
	go clean
	rm -f cmd/server/server cmd/notifier/notifier cmd/client/client

up:
	docker-compose up -d

down:
	docker-compose down

migrate:
	goose -dir ./migrations postgres "$(DB_DSN)" up

test:
	go test ./...

