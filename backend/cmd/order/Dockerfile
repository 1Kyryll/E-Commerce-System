# syntax=docker/dockerfile:1
#
# Order service — gRPC order placement (reservations, payment, outbox).
# Build from the backend module root (run `make generate-backend` first):
#
#   docker build -f cmd/order/Dockerfile -t ecom-order .

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/order ./cmd/order

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
USER app
COPY --from=builder /out/order /usr/local/bin/order

# gRPC. Actual bind port is ORDER_GRPC_PORT / GRPC_PORT (default 9002 at runtime);
# EXPOSE is documentation only.
EXPOSE 9002
ENTRYPOINT ["order"]
