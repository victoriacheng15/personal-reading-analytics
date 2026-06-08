# Global Makefile configurations and flags
MAKEFLAGS += --no-print-directory

# === Base Variables ===
DOCKER ?= podman

# Include modular makefiles
include mk/go.mk
include mk/python.mk

.PHONY: all lint test fmt cov lint-md fmt-md clean help

all: lint test fmt

# ==============================================================================
# MARKDOWN TARGETS (LINTING, FORMATTING)
# ==============================================================================

lint-md: ## Lint Markdown files
	@echo "==> Linting Markdown files..."
	npx markdownlint-cli '**/*.md' --ignore .venv

fmt-md: ## Format Markdown files using markdownlint-cli
	@echo "==> Formatting Markdown files..."
	npx markdownlint-cli '**/*.md' --ignore .venv --fix

# ==============================================================================
# COMPOSITE & AUTOMATION TARGETS
# ==============================================================================

lint: ## Run all linters
	@$(MAKE) lint-go
	@$(MAKE) lint-py
	@$(MAKE) lint-md

test: ## Run all tests
	@$(MAKE) test-go
	@$(MAKE) test-py

fmt: ## Format all code
	@$(MAKE) fmt-go
	@$(MAKE) fmt-py
	@$(MAKE) fmt-md

cov: ## Run all test coverages
	@$(MAKE) cov-go
	@$(MAKE) cov-py

# ==============================================================================
# UTILITIES
# ==============================================================================

clean: ## Remove build artifacts and caches
	@echo "==> Cleaning build artifacts and caches..."
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null
	find . -type f -name "*.py[co]" -delete 2>/dev/null
	rm -rf bin dist
	rm -f coverage.out coverage.html *.exe

# ==============================================================================
# DOCUMENTATION
# ==============================================================================

help: ## Show this help menu
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
