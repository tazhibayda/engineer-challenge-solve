APP_NAME := OrbittoAuth
PROTO_DIR := api/proto
GEN_DIR := pkg/api


generate:
	@echo "Generating mocks..."
	@go generate ./...

test: generate
	@echo "Running tests..."
	@go test -v -race ./...


DB_URL := postgres://postgres:1234@192.168.32.123:5432/auth_db?sslmode=disable

install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	@echo "Running up migrations..."
	migrate -path migrations -database "$(DB_URL)" -verbose up

migrate-down:
	@echo "Running down migrations..."
	migrate -path migrations -database "$(DB_URL)" -verbose down 1


# Запуск только быстрых unit-тестов
test-unit:
	@echo "Running unit tests..."
	@go test -short -v ./...

# Запуск всех тестов, включая интеграционные (требует поднятых БД)
test-full:
	@echo "Running all tests (Unit + Integration)..."
	@go test -v ./...

# Хелпер для проверки покрытия
test-cover:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out



.PHONY: install-tools generate-api


init-proto-deps:
	@echo "Downloading proto dependencies..."
	@rm -rf vendor.protogen
	@mkdir -p vendor.protogen
	@git clone --depth 1 https://github.com/googleapis/googleapis vendor.protogen/googleapis
	@git clone --depth 1 https://github.com/grpc-ecosystem/grpc-gateway vendor.protogen/grpc-gateway
	@cp -r vendor.protogen/googleapis/google vendor.protogen/
	@cp -r vendor.protogen/grpc-gateway/protoc-gen-openapiv2 vendor.protogen/
	@rm -rf vendor.protogen/googleapis vendor.protogen/grpc-gateway
	@echo "Dependencies downloaded to vendor.protogen/"

install-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

generate-api:
	@echo "Generating Go files and Swagger docs from proto..."
	@mkdir -p $(GEN_DIR)
	protoc -I $(PROTO_DIR) \
		-I vendor.protogen \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GEN_DIR) --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(GEN_DIR) --openapiv2_opt=allow_merge=true,merge_file_name=swagger \
		$(PROTO_DIR)/auth/v1/auth.proto