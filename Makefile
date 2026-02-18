.PHONY: backend-run backend-tidy backend-build frontend-install frontend-dev frontend-build up down

COMPOSE_CMD := $(shell if command -v docker >/dev/null 2>&1; then echo "docker compose"; elif command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo ""; fi)

backend-run:
	cd backend && go run ./cmd/api

backend-tidy:
	cd backend && go mod tidy

backend-build:
	cd backend && go build ./cmd/api

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

up:
	@if [ -z "$(COMPOSE_CMD)" ]; then \
		echo "Error: docker/docker-compose bulunamadi. Docker Desktop veya OrbStack kurup tekrar deneyin."; \
		exit 1; \
	fi
	@$(COMPOSE_CMD) up --build

down:
	@if [ -z "$(COMPOSE_CMD)" ]; then \
		echo "Error: docker/docker-compose bulunamadi."; \
		exit 1; \
	fi
	@$(COMPOSE_CMD) down
