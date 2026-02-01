.PHONY: build build-frontend build-backend run run-frontend run-backend clean proto proto-go proto-java test-delaunay test-delaunay-benchmark

# Variables
PROTO_DIR = proto
FRONTEND_DIR = frontend/polynav
BACKEND_DIR = backend
PROTO_FILES = $(PROTO_DIR)/polynav.proto

# Default target
build: build-frontend build-backend

# Generate Protocol Buffers
proto: proto-go proto-java

proto-go:
	@echo "Generating Go code from proto..."
	cd $(BACKEND_DIR) && protoc -I ../$(PROTO_DIR) \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		../$(PROTO_FILES)

proto-java:
	@echo "Generating Java code from proto..."
	cd $(FRONTEND_DIR) && mvn protobuf:compile protobuf:compile-custom

# Build
build-frontend:
	@echo "Building Frontend..."
	cd $(FRONTEND_DIR) && mvn clean install

build-backend:
	@echo "Building Backend..."
	cd $(BACKEND_DIR) && go mod tidy && go build -o ./build/service ./cmd/main.go

# Run
run:
	@echo "Starting both Backend and Frontend..."
	make run-frontend &
	make run-backend

run-frontend:
	@echo "Running Frontend..."
	cd $(FRONTEND_DIR) && mvn javafx:run

run-backend:
	@echo "Running Backend Service..."
	cd $(BACKEND_DIR) && go run ./cmd/main.go

# Test
test-delaunay:
	@echo "Running all tests..."
	go test -v ./backend/internal/algo/

test-delaunay-benchmark:
	@echo "Running benchmarks only..."
	cd backend/internal/algo/ && go test -bench=. -benchmem -run ^$

# Clean
clean:
	@echo "Cleaning up..."
	cd $(FRONTEND_DIR) && mvn clean
	cd $(BACKEND_DIR) && go clean && rm -rf build/
