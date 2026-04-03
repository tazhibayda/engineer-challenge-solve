APP_NAME := OrbittoAuth
PROTO_DIR := api/proto
GEN_DIR := pkg/api

.PHONY: install-tools generate-api

generate:
	@echo "Generating mocks..."
	@go generate ./...

test: generate
	@echo "Running tests..."
	@go test -v -race ./...


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