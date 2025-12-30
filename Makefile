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