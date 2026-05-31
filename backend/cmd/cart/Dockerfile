# syntax=docker/dockerfile:1
#
# Cart service — gRPC persistent cart.
# Build from the backend module root (run `make generate-backend` first):
#
#   docker build -f cmd/cart/Dockerfile -t ecom-cart .

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/cart ./cmd/cart

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
USER app
COPY --from=builder /out/cart /usr/local/bin/cart

# gRPC. Actual bind port is CART_GRPC_PORT / GRPC_PORT (default 9001 at runtime);
# EXPOSE is documentation only.
EXPOSE 9001
ENTRYPOINT ["cart"]
