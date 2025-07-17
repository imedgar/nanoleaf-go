.PHONY: build test clean run lint fmt vet

# Build the application
build:
	go build -o nanoleaf-go cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f nanoleaf-go coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Full check (format, vet, test)
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o nanoleaf-go-linux-amd64 cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o nanoleaf-go-darwin-amd64 cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o nanoleaf-go-windows-amd64.exe cmd/main.go