APP_NAME := OrbittoAuth

generate:
	@echo "Generating mocks..."
	@go generate ./...

test: generate
	@echo "Running tests..."
	@go test -v -race ./...