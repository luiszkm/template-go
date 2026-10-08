# Project state

## Decisions

| ID | Decision | Rationale | Status | Date |
| --- | --- | --- | --- | --- |
| AD-001 | Vertical Slice Architecture: one folder per use case under `app/internal/features/<feature>/<slice>/`; slices import only `app/internal/platform/...`, never another feature; enforced by `app/archtest` | an LLM reads and changes one folder; coupling is caught by a failing test, not by review | active | 2026-10-07 |
| AD-002 | PostgreSQL 17 + `pgx/v5` + `sqlc` (hand-written SQL, `queries.sql` inside each slice) + `goose` SQL migrations embedded in the binary | explicit SQL is what LLMs get right; rejected GORM/ent (hidden queries) and SQL Server (no mature sqlc support) | active | 2026-10-07 |
| AD-003 | HTTP via `net/http` mux + Huma v2, code-first; OpenAPI 3.1 exported and committed at `app/openapi.json`, diffed by the gate | one source of truth for the contract; rejected spec-first YAML (two places to edit) and Gin/Echo (no native OpenAPI) | active | 2026-10-07 |
| AD-004 | Every HTTP error is RFC 9457 Problem Details (`application/problem+json`) | the error shape is decided once, before the first handler copies an accident | active | 2026-10-07 |
| AD-005 | Authentication is local only: password hashed with argon2id; opaque session token in an httpOnly, SameSite=Lax cookie, stored server-side and revocable | role removal or user deactivation must take effect on the next request; rejected JWT (not revocable) and OIDC (out of template scope) | active | 2026-10-07 |
| AD-006 | RBAC is global (no tenant): user ↔ roles ↔ permissions; permissions are Go constants (`resource:action`), roles and assignments live in the DB; every operation declares its permission at registration; seeded role `admin` holds all permissions | readable in the slice diff; rejected Casbin (extra DSL) and multi-tenant (out of scope) | active | 2026-10-07 |
| AD-007 | Audit is an append-only `audit_events` table written in the same DB transaction as the mutation via `platform/audit`; carries actor, action, resource, before/after JSON, ip, request_id; the `audit` feature only reads | an event can never be lost when the mutation commits; rejected async logging | active | 2026-10-07 |
| AD-008 | Frontend in `web/`: React 19 + Vite + strict TypeScript + TanStack Router (file-based) + TanStack Query + shadcn/ui + Tailwind v4 + Zod + react-hook-form; API client generated from `app/openapi.json` (`openapi-typescript` + `openapi-fetch`); Vitest + Testing Library + Playwright; Biome | highest LLM accuracy; contract break becomes a compile error; rejected Angular (verbosity) and Next.js (Node server duplicates the Go backend) | active | 2026-10-07 |
| AD-009 | `task check` (Taskfile) is the single gate, run locally and in CI: format, golangci-lint v2, sqlc diff, archtest, go test (testcontainers Postgres), OpenAPI export diff, web typecheck/lint/test | one command means an agent cannot pick a weaker subset | active | 2026-10-07 |
| AD-010 | Spec workflow is `tlc-spec-lean` with `profile: standard` | catches unproven set members and tests that pass under a wrong implementation | active | 2026-10-07 |

## Handoff

**Feature**: foundation - done
**Where**: C1-C75 built and green; independent Verifier round 4 PASS at `b5b76c1` (`task check` exit 0, `task e2e` pass, 4/4 faults killed, `validate_verification.py` exit 0). Rounds 1-3 FAIL closed by S7 (C56-C68), S8 (C69-C73), S9 (C74-C75); 4th round authorized by the user on 2026-10-08
**In progress**: `users` - `.specs/features/users/plan.md` written (65 AC, 16 doors), `validate_plan.py` exit 0, awaiting human review
**Next step**: on approval, derive `.specs/features/users/checks.md`
**Blockers**: none
**Uncommitted**: `.claude/skills/auth-security/`, `.cursor/skills/auth-security/` (untracked, user-added)
**Branch**: `main` (local only, no remote)
