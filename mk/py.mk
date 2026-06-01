# === Python Variables ===
IMAGE_NAME = extraction

# === Python Development Targets ===
.PHONY: run install freeze update py-check py-format py-test py-cov py-run

run:
	$(DOCKER) build -t $(IMAGE_NAME) .
	$(DOCKER) run --rm $(IMAGE_NAME)
	$(DOCKER) image rm $(IMAGE_NAME)

install:
	if [ ! -d .venv ]; then python3 -m venv .venv; fi && \
	.venv/bin/pip install --upgrade pip && \
	.venv/bin/pip install -r requirements.txt

freeze:
	.venv/bin/pip freeze > requirements.txt

update:
	.venv/bin/pip install --upgrade -r requirements.txt

py-check:
	.venv/bin/python -m ruff check script/ --diff

py-format:
	.venv/bin/python -m ruff format script/

py-test:
	.venv/bin/python -m pytest script/

py-cov:
	.venv/bin/python -m pytest --cov=script --cov-report=term-missing

py-run:
	.venv/bin/python script/main.py
