# === Base Variables ===
LINT_IMAGE = ghcr.io/igorshubovych/markdownlint-cli:v0.44.0
DOCKER ?= podman

# Include language-specific Makefiles
include mk/go.mk
include mk/py.mk

.PHONY: help lint clean

# === Help ===
help:
	@echo "Available commands:"
	@echo "  make help             - Show this help message"
	@echo ""
	@echo "  make run              - [Python] Build and run extraction via Docker"
	@echo "  make py-run           - [Python] Run extraction via local venv"
	@echo "  make install          - [Python] Create .venv and install dependencies"
	@echo "  make py-check         - [Python] Run ruff check (lint)"
	@echo "  make py-format        - [Python] Format files with ruff"
	@echo "  make py-test          - [Python] Run tests"
	@echo ""
	@echo "  make go-vet           - [Go] Verify Go code quality via vet"
	@echo "  make go-format        - [Go] Format files with gofmt"
	@echo "  make go-test          - [Go] Run tests"
	@echo "  make go-cov           - [Go] Run tests with coverage summary"
	@echo "  make metrics-build    - [Go] Build metrics json"
	@echo "  make web-build        - [Go] Build web site"
	@echo ""
	@echo "  make lint             - [Quality] Run markdownlint via Docker"
	@echo "  make clean            - [Utils] Remove build artifacts and caches"

# === Quality & Linting ===
lint:
	$(DOCKER) run --rm -v "$(PWD):/data:Z" -w /data $(LINT_IMAGE) --fix "**/*.md"

# === Utilities ===
clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null
	rm -rf $(BIN_DIR) dist
	rm -f coverage.out coverage.html *.exe
