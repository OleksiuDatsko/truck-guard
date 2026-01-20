.PHONY: env-up
env-up:
	cp .env.example .env
	${EDITOR} .env

.PHONY: up-build
up-build:
	docker compose up -d --build

.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down -v

.PHONY: logs
logs:
	docker compose logs -f

.PHONY: rebuild
rebuild: down up-build

.PHONY: dev-up
dev-up:
	docker compose -f docker-compose.dev.yaml up -d

.PHONY: dev-down
dev-down:
	docker compose -f docker-compose.dev.yaml down -v

.PHONY: dev-rebuild
dev-rebuild: dev-down dev-up
