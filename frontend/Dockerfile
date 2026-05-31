# syntax=docker/dockerfile:1
#
# Frontend — Next.js (App Router) storefront, built as a standalone server.
# Build context is frontend/:
#
#   docker build -t ecom-frontend ./frontend
#
# Runtime env:
#   PORT          listen port (default 3000)
#   GATEWAY_URL   gateway base URL used by server-side fetches in
#                 lib/api/server.ts (e.g. http://gateway:8080)
#
# Note: the `/api/*` -> gateway rewrite in next.config.ts is resolved at
# BUILD time. Pass --build-arg GATEWAY_URL=... to bake a non-default
# destination; otherwise it falls back to http://localhost:8080. In a
# reverse-proxy deployment (Caddy) the proxy routes /api to the gateway and
# this rewrite is unused.

FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS builder
WORKDIR /app
ARG GATEWAY_URL
ENV GATEWAY_URL=${GATEWAY_URL}
ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

# Self-hosted fonts/css live in public; standalone copies the rest.
COPY --from=builder /app/public ./public
COPY --from=builder --chown=node:node /app/.next/standalone ./
COPY --from=builder --chown=node:node /app/.next/static ./.next/static

USER node
EXPOSE 3000
CMD ["node", "server.js"]
