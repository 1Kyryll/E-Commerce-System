# Contributing

Thanks for contributing. This project follows a few simple conventions — please
keep changes consistent witoh them.

## Source of truth

The `docs/` directory is authritative for every design decision. Read the
relevant doc before changing behavior, and update the doc **first** if a
decision needs to change (add an ADR under `docs/adr/` for significant
changes). The code reflects the docs, not the other way around.

## Branches

- One branch per major feature. The **frontend** is a single branch; **each
  backend service** has its own branch; **everything else** goes to its own
  branch.
- Branch names are short and bare — the feature or service name, no prefixes.
  Examples: `cart`, `catalog`, `auth`, `gateway`, `order`, `frontend`,
  `observability`, `docker`, `docs`, `load-testing`.
- Don't create too many branches. Keep related work together.

## Commits

- Commit frequently, with short, explicit, simple messages.
- Prefix every commit title. Use a scope in parentheses where it adds clarity:
  - `feat(...): ...` — new functionality (e.g. `feat(cart): optimistic CartProvider`)
  - `fix: ...` / `fix(...): ...` — bug fixes (e.g. `fix(ci): backend job regenerates only backend code`)
  - `docs: ...` / `docs(...): ...` — documentation (e.g. `docs(adr): frontend feature slicing`)
  - `chore: ...` / `chore(...): ...` — maintenance, tooling, housekeeping
  - `refactor: ...` — code change with no behavior change
  - `test: ...` — tests
  - `ci: ...` — CI/CD pipeline changes
- Keep the subject in the imperative and to the point. 

## Pull requests

- Open PRs against `main`.
- One PR per feature branch; keep it focused on a single concern.
- Reference the issue it closes (e.g. `Closes #12`).
- Make sure CI is green (typecheck, lint, build, generated-code checks) before
  requesting a merge.
- Merge commits read `Merge PR from <owner>/<branch>`.

## Before you push

- Backend: regenerate code (sqlc / proto) when you touch queries or `.proto`
  files, and run the relevant service tests.
- Frontend: typecheck, lint, and build.
