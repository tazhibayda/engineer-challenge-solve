APP_NAME := OrbittoAuth

test:
	@echo "Running tests..."
	@go test -v -race ./...