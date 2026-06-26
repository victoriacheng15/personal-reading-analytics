# === Go Variables ===
BIN_DIR = bin

.PHONY: lint-go fmt-go test-go cov-go go-check go-update metrics-build setup-tailwind web-build

# ==============================================================================
# GO DEVELOPMENT TARGETS
# ==============================================================================

lint-go: ## Verify Go code quality via vet
	go vet ./cmd/... ./internal/...

fmt-go: ## Format Go files with gofmt
	gofmt -w ./cmd ./internal

test-go: ## Run all Go unit tests
	go test -v ./cmd/... ./internal/...

cov-go: ## Run Go tests with coverage summary
	go test -coverprofile=coverage.out ./cmd/... ./internal/... && go tool cover -func=coverage.out && rm -f coverage.out || exit 1

go-check: ## Check Go code formatting without modifying files
	@if [ -n "$$(gofmt -l ./cmd ./internal)" ]; then \
		echo "Go files not formatted! Run 'make fmt-go'"; \
		gofmt -d ./cmd ./internal; \
		exit 1; \
	fi

go-update: ## Update Go dependencies and run go mod tidy
	go get -u ./... && go mod tidy

metrics-build: ## Build and run the metrics calculator
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/metricsjson ./cmd/metrics
	./$(BIN_DIR)/metricsjson

setup-tailwind: ## Set up Tailwind CSS CLI
	@mkdir -p $(BIN_DIR)
	@if [ ! -f $(BIN_DIR)/tailwindcss ]; then \
		echo "Downloading tailwind css cli v4..."; \
		curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o $(BIN_DIR)/tailwindcss; \
		chmod +x $(BIN_DIR)/tailwindcss; \
	fi

web-build: setup-tailwind ## Build and run the dashboard generator
	echo 'Running analytics build...'
	rm -rf dist
	mkdir -p dist
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/web-ssg ./cmd/web
	./$(BIN_DIR)/web-ssg
	mkdir -p dist/css
	./$(BIN_DIR)/tailwindcss -i ./internal/web/templates/css/input.css -o ./dist/css/styles.css --minify
	rm -rf $(BIN_DIR)
