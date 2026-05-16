# CI/CD Pipeline
 
This document describes how every commit makes it from a `git push` to a running deployment. Kept deliberately simple — one workflow file, one environment, one Hetzner VM — to be expanded as the project's needs grow. For the deployment target and the broader system this pipeline serves, see [system-design.md](./system-design.md). For the load test that runs as a deploy gate, see `load-tests` directory.

## Pipeline shape
 
Every push to `main` flows through four sequential stages, each gating the next:
 
```mermaid
flowchart LR
    push([git push origin main])
    test[1. Test & lint]
    build[2. Build & push images]
    deploy[3. Deploy to Hetzner]
    smoke[4. Smoke load test]
 
    push --> test --> build --> deploy --> smoke
 
    classDef plain fill:none,stroke:#666
    class push,test,build,deploy,smoke plain
```
 
A failure at any stage stops the pipeline and surfaces in the GitHub UI. Failures *after* deploy mean a bad version is already live — automatic rollback isn't wired up yet, that's covered under "What's deferred."
 
## The workflow
 
The whole pipeline lives in a single file. Pull requests run only the test job; everything past that runs on pushes to `main`.
 
```yaml
# .github/workflows/ci-cd.yml
name: CI/CD
 
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
 
env:
  REGISTRY: ghcr.io
  IMAGE_PREFIX: ${{ github.repository }}
 
jobs:
  # ---------- 1. Test & lint ----------
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 5s
          --health-retries 10
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      - name: Run tests
        working-directory: backend
        run: go test ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/test?sslmode=disable
      - uses: golangci/golangci-lint-action@v6
        with:
          working-directory: backend
 
  # ---------- 2. Build & push images ----------
  build:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    strategy:
      matrix:
        service: [gateway, catalog, cart, order, outbox-publisher, cleanup-worker]
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./backend
          target: ${{ matrix.service }}
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-${{ matrix.service }}:${{ github.sha }}
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-${{ matrix.service }}:latest
 
  build-frontend:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-frontend:${{ github.sha }}
            ${{ env.REGISTRY }}/${{ env.IMAGE_PREFIX }}-frontend:latest
 
  # ---------- 3. Deploy to Hetzner ----------
  deploy:
    needs: [build, build-frontend]
    runs-on: ubuntu-latest
    steps:
      - uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key:  ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /opt/ecommerce-demo
            export IMAGE_TAG=${{ github.sha }}
            docker compose pull
            docker compose up -d --wait
            curl -fsS http://localhost:8080/healthz
 
  # ---------- 4. Smoke load test ----------
  smoke:
    needs: deploy
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: grafana/setup-k6-action@v1
      - run: k6 run load-tests/smoke.js
        env:
          BASE_URL: ${{ secrets.PRODUCTION_URL }}
```
 
A few choices baked into this:
 
**Images are tagged with the commit SHA, plus a moving `latest`.** The SHA tag makes deployments deterministic and lets any historical version be redeployed by changing `IMAGE_TAG` on the server; `latest` is the convenience tag for local development.
 
**Backend services share one matrix job.** Same Dockerfile, different `--target`. The Dockerfile is structured as a multi-stage build with one named stage per service binary. Adding a new service is one line in the matrix.
 
**Deployment is `docker compose up -d --wait`.** The server-side `docker-compose.yml` references images via `${IMAGE_TAG}`, so exporting that variable before `docker compose pull` switches the running set to the new SHA. `--wait` blocks until containers report healthy via their compose-defined healthchecks. The follow-up `curl` catches the case where containers are healthy individually but the gateway isn't actually serving.
 
**The smoke test runs against the live production URL.** Low concurrency, short duration — enough to catch obvious breakage, not enough to stress-test. If it fails the workflow turns red, but the deploy has already happened.

## What you need to configure
 
GitHub repository secrets the workflow expects:
 
- `SERVER_HOST` — Hetzner box IP or DNS name
- `SERVER_USER` — SSH user (typically `deploy`)
- `SSH_PRIVATE_KEY` — private key for that user
- `PRODUCTION_URL` — public URL the smoke test hits
Server-side, before the first deploy: install Docker and Docker Compose, place a `docker-compose.yml` at `/opt/ecommerce-demo/` that references the GHCR images via `${IMAGE_TAG}`, set up Caddy for TLS termination, and ensure the `deploy` user can reach Docker without sudo. This is one-time setup, scripted in `deploy/scripts/bootstrap.sh`.

## What's deferred
 
A few obvious improvements consciously not in this iteration:
 
- **Automatic rollback on smoke failure.** Currently a failed smoke test produces a red build but leaves the bad version running. A rollback step would SSH in and run `IMAGE_TAG=<previous-sha> docker compose up -d`. Worth adding once the project has weathered its first bad deploy.
- **Separate staging environment.** A second Hetzner box receiving every push, with production deploying only from tags or manual approval. Out of scope while one VM is sufficient.
- **Integration tests in CI.** Currently only unit tests run. Tagged `//go:build integration` Go tests booted against a real Postgres container would catch contract regressions; tying that into the test job is the natural next step.
- **Dependency vulnerability scanning.** Trivy or grype on the built images, gated as part of the pipeline.
- **Smarter Docker build caching.** The `cache: true` on setup-go covers the Go cache; layered caching via buildx would speed up image builds significantly.
- **Build only what changed.** Currently every push rebuilds every service. Path filters or `dorny/paths-filter` would skip rebuilds when the affected service's directory is untouched.
None of these block the pipeline from doing its job, each is additive when the project earns it.
 