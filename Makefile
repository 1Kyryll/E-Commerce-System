.PHONY: run-gateway run-catalog run-cart run-order frontend-dev

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
