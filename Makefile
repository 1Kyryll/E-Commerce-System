.PHONY: generate generate-proto generate-db \
        migrate-up migrate-down \
        run-gateway run-catalog run-cart run-order \
        frontend-dev

DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/ecommerce?sslmode=disable
MIGRATE_DATABASE_URL ?= pgx5://postgres:postgres@localhost:5433/ecommerce?sslmode=disable

generate-proto:
	cd proto && buf generate

generate-db:
	cd backend && sqlc generate

generate: generate-proto generate-db

migrate-up:
	migrate -path backend/migrations -database "$(MIGRATE_DATABASE_URL)" up

migrate-down:
	migrate -path backend/migrations -database "$(MIGRATE_DATABASE_URL)" down -all

run-gateway:
	cd backend && go run ./cmd/gateway

run-catalog:
	cd backend && go run ./cmd/catalog

run-cart:
	cd backend && go run ./cmd/cart

run-order:
	cd backend && go run ./cmd/order

frontend-dev:
	cd frontend && npm run dev
