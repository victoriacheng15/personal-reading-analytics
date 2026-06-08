# === Python Variables ===
IMAGE_NAME = extraction

.PHONY: lint-py fmt-py test-py cov-py run install freeze update py-run

# ==============================================================================
# PYTHON DEVELOPMENT TARGETS
# ==============================================================================

lint-py: ## Run Python linter (ruff)
	.venv/bin/python -m ruff check script/ --diff

fmt-py: ## Format Python files with ruff
	.venv/bin/python -m ruff format script/

test-py: ## Run Python unit tests via pytest
	.venv/bin/python -m pytest script/

cov-py: ## Run Python test coverage report
	.venv/bin/python -m pytest --cov=script --cov-report=term-missing

run: ## Build and run extraction via Docker
	$(DOCKER) build -t $(IMAGE_NAME) .
	$(DOCKER) run --rm $(IMAGE_NAME)
	$(DOCKER) image rm $(IMAGE_NAME)

install: ## Create .venv and install Python dependencies
	if [ ! -d .venv ]; then python3 -m venv .venv; fi && \
	.venv/bin/pip install --upgrade pip && \
	.venv/bin/pip install -r requirements.txt

freeze: ## Freeze current Python dependencies to requirements.txt
	.venv/bin/pip freeze > requirements.txt

update: ## Update Python dependencies in .venv from requirements.txt
	.venv/bin/pip install --upgrade -r requirements.txt

py-run: ## Run extraction via local venv
	.venv/bin/python script/main.py
