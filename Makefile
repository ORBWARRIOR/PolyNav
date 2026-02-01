.PHONY: build build-frontend build-backend run run-frontend run-backend clean proto proto-go proto-java test test-delaunay test-delaunay-benchmark test-java kill

# Variables
PROTO_DIR = proto
FRONTEND_DIR = frontend/polynav
BACKEND_DIR = backend
PROTO_FILES = $(PROTO_DIR)/polynav.proto

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

# Default target
build: build-frontend build-backend

build-frontend:
	@echo "Building Frontend..."
	cd $(FRONTEND_DIR) && mvn clean install

build-backend:
	@echo "Building Backend..."
	cd $(BACKEND_DIR) && go mod tidy && go build -o ./build/service ./cmd/main.go

# Run
run: kill
	@echo "Starting both Backend and Frontend..."
	@# Run backend in background, then run frontend
	@(cd $(BACKEND_DIR) && go run ./cmd/main.go) & \
	(cd $(FRONTEND_DIR) && mvn javafx:run)

run-frontend:
	@echo "Running Frontend..."
	cd $(FRONTEND_DIR) && mvn javafx:run

run-backend:
	@echo "Running Backend Service..."
	cd $(BACKEND_DIR) && go run ./cmd/main.go

# Kill
kill:
	@echo "Killing backend service..."
	@# Kill process holding port 50051
	@fuser -k 50051/tcp 2>/dev/null || true
	@# Kill go run process
	@ps aux | grep "go run ./cmd/main.go" | grep -v grep | awk '{print $$2}' | xargs -r kill 2>/dev/null || true
	@# Kill built binary
	@ps aux | grep "backend/build/service" | grep -v grep | awk '{print $$2}' | xargs -r kill 2>/dev/null || true
	@echo "Killing frontend application..."
	@# Kill Maven wrapper
	@ps aux | grep "mvn javafx:run" | grep -v grep | awk '{print $$2}' | xargs -r kill 2>/dev/null || true
	@# Kill the actual Java process
	@ps aux | grep "io.github.orbwarrior.App" | grep -v grep | awk '{print $$2}' | xargs -r kill 2>/dev/null || true

# Test
test: test-delaunay test-ui

test-delaunay:
	@echo "Running delaunay tests..."
	go test -v ./backend/internal/algo/

test-delaunay-benchmark:
	@echo "Running delaunay benchmarks..."
	cd backend/internal/algo/ && go test -bench=. -benchmem -run ^$

test-ui:
	@echo "Running UI tests..."
	cd $(FRONTEND_DIR) && mvn test

# Clean
clean:
	@echo "Cleaning up..."
	cd $(FRONTEND_DIR) && mvn clean
	cd $(BACKEND_DIR) && go clean && rm -rf build/