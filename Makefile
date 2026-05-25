# Use bash for recipes. On Windows GNU make defaults to cmd.exe, which can't
# parse the `VAR=value command` syntax the load-* targets use. Bash ships
# with Git for Windows / WSL / Docker Desktop.
SHELL := bash

.PHONY: generate generate-proto generate-db \
        migrate-up migrate-down \
        run-gateway run-catalog run-cart run-order \
        frontend-dev seed

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

run-binaries: 
	cd backend && go run ./cmd/gateway ./cmd/catalog ./cmd/cart ./cmd/order

frontend-dev:
	cd frontend && npm run dev

seed:
	docker exec -i ecom-postgres psql -U postgres -d ecommerce < backend/seeds/products.sql

# -----------------------------------------------------------------------------
# Load testing. Each target runs one k6 script against the local stack and
# pushes metrics to Mimir. Scale by overriding K6_VUS / K6_DURATION:
#
#   make load-browse K6_VUS=1000 K6_DURATION=5m
#   make load-checkout K6_VUS=10000 K6_DURATION=10m
#
# Prerequisites: app stack (make run-gateway, etc.) + observability stack
# (docker compose -f observability/docker-compose.yml up -d) running.
# -----------------------------------------------------------------------------
K6_VUS ?= 100
K6_DURATION ?= 3m
K6_BASE_URL ?= http://localhost:8080
K6_RW_URL ?= http://localhost:9009/api/v1/push
K6_RUN := BASE_URL=$(K6_BASE_URL) K6_VUS=$(K6_VUS) K6_DURATION=$(K6_DURATION) \
          K6_PROMETHEUS_RW_SERVER_URL=$(K6_RW_URL) \
          K6_PROMETHEUS_RW_PUSH_INTERVAL=5s \
          k6 run --out experimental-prometheus-rw

load-smoke:
	$(K6_RUN) load-tests/smoke.js

load-browse:
	$(K6_RUN) load-tests/browse.js

load-cart:
	$(K6_RUN) load-tests/cart.js

load-checkout:
	$(K6_RUN) load-tests/checkout.js

.PHONY: load-smoke load-browse load-cart load-checkout
