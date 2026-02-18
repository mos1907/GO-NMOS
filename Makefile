.PHONY: backend-run backend-tidy backend-build backend-test backend-fmt frontend-install frontend-dev frontend-build frontend-doctor test fmt up down

COMPOSE_CMD := $(shell if command -v docker >/dev/null 2>&1; then echo "docker compose"; elif command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo ""; fi)

backend-run:
	cd backend && go run ./cmd/api

backend-tidy:
	cd backend && go mod tidy

backend-build:
	cd backend && go build ./cmd/api

backend-test:
	cd backend && go test ./...

backend-fmt:
	cd backend && gofmt -w .

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-doctor:
	cd frontend && npm run doctor

test: backend-test frontend-doctor frontend-build

fmt: backend-fmt

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
