.PHONY: up down logs api-logs web-logs db-logs bridge-install bridge-start bridge-stop bridge-doctor

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f --tail=200

api-logs:
	docker compose logs -f --tail=200 api

web-logs:
	docker compose logs -f --tail=200 web

db-logs:
	docker compose logs -f --tail=200 db

bridge-install:
	./scripts/bridge/install.sh

bridge-start:
	./scripts/bridge/start.sh

bridge-stop:
	./scripts/bridge/stop.sh

bridge-doctor:
	./scripts/bridge/doctor.sh
