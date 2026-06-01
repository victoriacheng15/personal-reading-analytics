# === Go Variables ===
BIN_DIR = bin

# === Go Development Targets ===
.PHONY: go-vet go-format go-update go-test go-cov metrics-build setup-tailwind web-build

go-vet:
	go vet ./cmd/... ./internal/...

go-format:
	gofmt -w ./cmd ./internal 

go-update:
	go get -u ./... && go mod tidy 

go-test:
	go test -v ./cmd/... ./internal/... 

go-cov:
	go test -coverprofile=coverage.out ./cmd/... ./internal/... && go tool cover -func=coverage.out && rm coverage.out || exit 1

metrics-build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/metricsjson ./cmd/metrics
	./$(BIN_DIR)/metricsjson

setup-tailwind:
	@mkdir -p $(BIN_DIR)
	@if [ ! -f $(BIN_DIR)/tailwindcss ]; then \
		echo "Downloading tailwind css cli v4..."; \
		curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 -o $(BIN_DIR)/tailwindcss; \
		chmod +x $(BIN_DIR)/tailwindcss; \
	fi

web-build: setup-tailwind
	echo 'Running analytics build...'
	rm -rf dist
	mkdir -p dist
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/web-ssg ./cmd/web
	./$(BIN_DIR)/web-ssg
	mkdir -p dist/css
	./$(BIN_DIR)/tailwindcss -i ./internal/web/templates/css/input.css -o ./dist/css/styles.css --minify
