.PHONY: build test test-coverage clean run lint fmt vet install uninstall check

# Build the application
build:
	go build -o nanoleaf-go cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Install the application
install: build
	@echo "Installing nanoleaf-go to /usr/local/bin"
	@mv nanoleaf-go /usr/local/bin/nanoleaf-go

# Uninstall the application
uninstall:
	@echo "Removing nanoleaf-go from /usr/local/bin"
	@rm -f /usr/local/bin/nanoleaf-go

# Install dependencies
deps:
	go mod tidy
	go mod download


# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o nanoleaf-go-linux-amd64 cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o nanoleaf-go-darwin-amd64 cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o nanoleaf-go-windows-amd64.exe cmd/main.go

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all quality checks
check: fmt vet test

# Clean build artifacts and test files
clean:
	rm -f nanoleaf-go nanoleaf-go-* coverage.out coverage.html
