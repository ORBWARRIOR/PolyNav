.PHONY: build build-frontend build-backend run-frontend run-backend clean proto proto-go proto-java

# Variables
PROTO_DIR = proto
FRONTEND_DIR = frontend
BACKEND_DIR = backend
PROTO_FILES = $(PROTO_DIR)/polynav.proto

# Default target
build: build-frontend build-backend

# Generate Protocol Buffers
proto: proto-go proto-java

proto-go:
	@echo "Generating Go code from proto..."
	cd $(BACKEND_DIR) && protoc -I ../$(PROTO_DIR) \
		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
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
run-frontend:
	@echo "Running Frontend..."
	cd $(FRONTEND_DIR) && mvn javafx:run

run-backend:
	@echo "Running Backend..."
	cd $(BACKEND_DIR) && go run ./cmd/main.go

# Clean
clean:
	@echo "Cleaning up..."
	cd $(FRONTEND_DIR) && mvn clean
	cd $(BACKEND_DIR) && go clean && rm -rf bin/

