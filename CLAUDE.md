# System Design, Code Architecture and System Requirements
 
- High-level overview of application's System Design is documented explicitly in the docs/system-design.md file
   - System Design decisions and tradeoffs are listed in the docs/adr/ directory
   - Database Models are listed in the docs/data-layer.md files, also every other decision related to how data is being handled can be found in the docs/adr/ directory
- Code architecture and reductive folder structure are listed in the docs/code-architecture.md file
- MVP System Requirements and personal goals for the project can be found in the docs/system-requirements.md file
# Observability, Load Testing, CI/CD pipeline
 
- Observability and monitoring in the docs/observability.md
- Load Testing in the docs/load-testing.md
- CI/CD pipeline in the docs/pipeline.md
# Git conventions
- Commit frequently
- Write short, explicit and simple messages
- Create a branch for every new major feature(frontend is a one branch, each of the backend services has its own branch, and eveything else goes to separate branches). Do not create too many branches. 
- For every commit title write prefixes such as: 
    - docs: ... 
    - fix: ...
    - chore: ...
    - feat(...): ... 
    - refactor: ... 
    - test: ...

# Tech Stack
 
- **Backend**: Go — gRPC internal, REST gateway, `pgx` + `sqlc` for Postgres
- **Frontend**: Next.js (App Router) + TypeScript + Tailwind v4
- **Data**: Postgres, Redis (cache only)
- **Observability**: OpenTelemetry → Grafana LGTM (Loki, Tempo, Prometheus)
- **Load testing**: k6
- **CI/CD**: GitHub Actions
- **Deployment**: Docker Compose on Hetzner Cloud
# Source of Truth
 
The `docs/` directory is authoritative for every design decision in this project. Consult it before answering questions about the system's design, rationale, or trade-offs — do not infer from code alone. If a decision needs to change, update the relevant file in `docs/` first (and add an ADR under `docs/adr/` if the change is significant). The code reflects the docs, not the other way around.