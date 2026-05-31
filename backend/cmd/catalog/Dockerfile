# syntax=docker/dockerfile:1
#
# Catalog service — gRPC product catalog (cache-aside over Redis + Postgres).
# Build from the backend module root (run `make generate-backend` first):
#
#   docker build -f cmd/catalog/Dockerfile -t ecom-catalog .

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/catalog ./cmd/catalog

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
USER app
COPY --from=builder /out/catalog /usr/local/bin/catalog

# gRPC. Actual bind port is GRPC_PORT (default 9000); EXPOSE is documentation only.
EXPOSE 9000
ENTRYPOINT ["catalog"]
