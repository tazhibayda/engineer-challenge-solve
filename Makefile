APP_NAME  := OrbittoAuth
PROTO_DIR := api/proto
GEN_DIR   := pkg/api

DB_URL ?= postgres://postgres:1234@192.168.32.123:5432/auth_db?sslmode=disable

.PHONY: help all build run clean generate test test-unit test-full test-cover install-migrate migrate-up migrate-down up down restart log install-tools init-proto-deps generate-api

help:
	@echo "Доступные команды:"
	@echo "  make build         - Сгенерировать код и собрать бинарник"
	@echo "  make run           - Запустить приложение"
	@echo "  make clean         - Удалить бинарники, кэш и скачанные proto-файлы"
	@echo "  make test          - Запустить unit-тесты (с генерацией моков)"
	@echo "  make test-full     - Запустить все тесты (включая интеграционные)"
	@echo "  make test-cover    - Проверить покрытие кода тестами"
	@echo "  make migrate-up    - Накатить миграции БД"
	@echo "  make migrate-down  - Откатить последнюю миграцию БД"
	@echo "  make install-tools - Установить необходимые утилиты (protoc, migrate и т.д.)"

all: build

build: generate-api generate
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) cmd/main.go

run:
	@echo "Running $(APP_NAME)..."
	@go run cmd/main.go

clean:
	@echo "Cleaning up..."
	@rm -rf bin/ coverage.out vendor.protogen/ pkg/api/

generate:
	@echo "Generating mocks..."
	@go generate ./...

test: generate test-unit

test-unit:
	@echo "Running unit tests..."
	@go test -short -v -race ./...

test-full: generate
	@echo "Running all tests (Unit + Integration)..."
	@go test -v -race ./...

test-cover:
	@echo "Checking coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

install-migrate:
	@echo "Installing migrate tool..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	@echo "Running up migrations..."
	@migrate -path migrations -database "$(DB_URL)" -verbose up

migrate-down:
	@echo "Running down migrations..."
	@migrate -path migrations -database "$(DB_URL)" -verbose down 1

up:
	@echo "Starting OrbittoAuth system..."
	docker-compose --env-file .env up --build -d

down:
	@echo "Stopping system..."
	docker-compose down -v

restart:
	docker-compose restart app

log:
	docker-compose logs -f app migrate

install-tools: install-migrate
	@echo "Installing protoc plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

init-proto-deps:
	@echo "Downloading specific proto dependencies (fast)..."
	@mkdir -p vendor.protogen/google/api
	@curl -sSLk https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto -o vendor.protogen/google/api/annotations.proto
	@curl -sSLk https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto -o vendor.protogen/google/api/http.proto
	@mkdir -p vendor.protogen/protoc-gen-openapiv2/options
	@curl -sSLk https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto -o vendor.protogen/protoc-gen-openapiv2/options/annotations.proto
	@curl -sSLk https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto -o vendor.protogen/protoc-gen-openapiv2/options/openapiv2.proto
	@echo "Dependencies downloaded successfully!"

generate-api:
	@if [ ! -d "vendor.protogen" ]; then $(MAKE) init-proto-deps; fi
	@echo "Generating Go files and Swagger docs from proto..."
	@mkdir -p $(GEN_DIR)
	@protoc -I $(PROTO_DIR) \
		-I vendor.protogen \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(GEN_DIR) --openapiv2_opt=allow_merge=true,merge_file_name=swagger \
		$(PROTO_DIR)/auth/v1/auth.proto