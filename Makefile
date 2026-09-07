POSTGRES_USER ?= avito
POSTGRES_PASSWORD ?= avito
POSTGRES_DB ?= avito_kitchen

DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: up down logs migrate-up migrate-down db-shell db-reset

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate \
		-path /migrations \
		-database "$(DATABASE_URL)" \
		down 1

db-shell:
	docker compose exec postgres \
		psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

db-reset:
	docker compose down -v
	docker compose up -d