POSTGRES_USER ?= avito
POSTGRES_PASSWORD ?= avito
POSTGRES_DB ?= avito_kitchen
POSTGRES_PORT ?= 5433

DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)?sslmode=disable
TEST_DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/postgres?sslmode=disable

.PHONY: up down restart logs status migrate-up migrate-down db-shell db-reset test test-integration

up:
	docker compose up -d --build

down:
	docker compose down

restart:
	docker compose restart
	
status:
	docker compose ps

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

test:
	go test ./...

test-integration:
	docker compose up -d --wait postgres
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" \
		go test ./internal/repository/postgres -v -count=1
