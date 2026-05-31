# syntax=docker/dockerfile:1
#
# Gateway service — REST surface for the frontend.
# Build from the backend module root so the shared module + generated code
# are in context (run `make generate-backend` first):
#
#   docker build -f cmd/gateway/Dockerfile -t ecom-gateway .
#
# (or, from the repo root: docker build -f backend/cmd/gateway/Dockerfile -t ecom-gateway backend)

FROM golang:1.25-alpine AS builder
WORKDIR /src

# Download modules first so this layer is cached unless go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Application source, including the .gitignored backend/gen produced by
# `make generate-backend`.
COPY . .

# Static binary (CGO off) so it runs on a minimal runtime image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./cmd/gateway

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates wget \
 && adduser -D -u 10001 app
USER app
COPY --from=builder /out/gateway /usr/local/bin/gateway

# HTTP API. The actual bind port is HTTP_PORT (set to 8080 at runtime);
# EXPOSE is documentation only.
EXPOSE 8080
ENTRYPOINT ["gateway"]
