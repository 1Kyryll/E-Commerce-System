# syntax=docker/dockerfile:1
#
# Cleanup worker — background job that releases expired reservations.
# No listening port. Build from the backend module root (run
# `make generate-backend` first):
#
#   docker build -f cmd/cleanup-worker/Dockerfile -t ecom-cleanup-worker .

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/cleanup-worker ./cmd/cleanup-worker

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
USER app
COPY --from=builder /out/cleanup-worker /usr/local/bin/cleanup-worker

ENTRYPOINT ["cleanup-worker"]
