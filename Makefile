.PHONY: install update format run up logs down

install:
	python -m pip install -r requirements.txt

update:
	pur -r requirements.txt

format:
	ruff format src/main.py src/utils

run:
	cd src && python main.py

up:
	docker compose up --build

logs:
	docker logs extractor > logs.txt

down:
	docker compose down
