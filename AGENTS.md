# AGENTS.md

Single source of rules for every coding agent (Claude Code and Cursor). Project decisions live in
`.specs/STATE.md` `## Decisions` (AD-NNN) — read them before changing anything; they are constraints.

## Stack

- Backend `app/`: Go, `net/http` + Huma v2, PostgreSQL 17 (`pgx/v5`, `sqlc`, `goose`), `log/slog`.
- Frontend `web/`: React 19, Vite, TypeScript strict, TanStack Router/Query, shadcn/ui, Tailwind v4, Zod.

## Architecture rules (Vertical Slice)

1. One use case = one folder: `app/internal/features/<feature>/<slice>/` holding its endpoint, `queries.sql`, generated code and tests.
2. A slice imports `app/internal/platform/...` and the standard library/approved deps. **Never another feature.** `app/archtest` fails otherwise.
3. Shared code goes to `platform/` only when a second feature needs it — not in anticipation.
4. Every operation declares its permission (`rbac`) at registration. Every mutating operation records an audit event in the same transaction.
5. Errors are RFC 9457 Problem Details. Never invent another error shape.
6. Frontend mirrors the backend: `web/src/features/<feature>/`. Call the API only through the generated client.
7. Generated files (`sqlc`, OpenAPI, TS client) are never edited by hand — change the source and regenerate.
8. No comments in code. Names, types and small functions carry the meaning; a comment that explains code means the code needs a better name or shape. Only machine-read annotations are allowed: `//go:` directives, `//nolint:<linter>`, `-- +goose`, sqlc `-- name:`, generator markers (`// slices:imports`, `// features:register`) and `biome-ignore`.

## Test policy

Classify code by its shape, not by its layer name. *Decision* = anything that changes an outcome (dispatch,
validation, guard, state transition, mapping with more than one row). *Instrumentation* = forwards or reshapes
with no conditional.

| Code | Required proofs | Coverage expectation |
| --- | --- | --- |
| Decides, reached across a boundary | one at the boundary **and** one at its own layer | the contract at the boundary; one asserted case per row of the decision table at its own layer |
| Decides, not reached across a boundary | one at its own layer | one asserted case per row of the decision table |
| Entry point that decides nothing | one at the boundary | accepted input, each rejected input, each error path |
| Instrumentation, pass-throughs | none of its own | covered by its consumer's proof |

Tests hit a real Postgres (testcontainers-go), never a mocked database. Write tests from the requirement, never
by reading the implementation.

## Workflow

Commands (all from the repository root):

| Command | What it does |
| --- | --- |
| `task dev` | Postgres in Docker + migrations, API on :8080, Vite on :5173 |
| `task new:slice FEATURE=<f> NAME=<n>` | scaffold a slice (endpoint answering 501, `queries.sql`, test) and regenerate code |
| `task gen` | regenerate sqlc code, `app/openapi.json` and `web/src/api/schema.d.ts` |
| `task check` | the gate: format, lint, generated-code drift, archtest, Go tests, web typecheck/lint/tests |
| `task e2e` | Playwright against the built binary |

A slice declares its own permission in `endpoint.go` (`const permission op.Permission = "<feature>:<action>"`);
there is no central permission list.

- Start a new use case with `task new:slice FEATURE=<feature> NAME=<slice>`.
- Before declaring work done: `task check` must pass. Never weaken, skip or delete a test to make it pass.
- `git push`, deploys and production data changes need explicit approval.

## tlc-spec-lean

profile: standard
budget: 150k
