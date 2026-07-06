# PolyNav — Wails3 Makefile

WAILS := wails3
BIN   := bin

.PHONY: dev build build-server test test-engine test-integration \
        bindings tidy clean help frontend-install frontend-build vet

dev: ## Dev mode with hot-reload
	$(WAILS) dev

build: ## Production build
	$(WAILS) build

build-server: ## Headless HTTP server build
	$(WAILS) task build:server

bindings: ## Regenerate TS bindings from Go services
	$(WAILS) generate bindings -clean=true -ts

test: ## Run unit tests (engine + root)
	go test ./... -count=1
	go vet ./...

test-engine: ## Engine package tests only
	go test ./engine/... -v -count=1

test-integration: ## Integration tests (build-tagged)
	go test -tags=integration ./tests/... -v -count=1

vet: ## Run go vet
	go vet ./...

tidy: ## Go mod tidy
	go mod tidy

clean: ## Remove build artifacts
	rm -rf $(BIN) polynav polynav-server frontend/dist

frontend-install: ## npm install frontend deps
	cd frontend && npm install

frontend-build: ## Vite build frontend
	cd frontend && npm run build

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
