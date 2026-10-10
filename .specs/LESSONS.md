# LESSONS - auto-maintained by scripts/lessons.py

> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.py` write.
> Canonical state lives in `.specs/lessons.json`. Edit lessons only via the script.
> promote_threshold=2 distinct features · window_days=45 · quarantine_threshold=2

## Confirmed (load these at Plan/Checks)

Corroborated across multiple features. Safe to apply as guidance.

### L-008 - When the plan applies a screen state to a second screen, prove that state on the second screen too
- signal: `ac_gap` · recurrence: 3 feature(s) · scope: `web` · harmful: 0
- features: users, audit, audit-links
- evidence: web/src/features/users/UserDetail.tsx:94 (web) (+2 more)
- last seen: 2026-10-09T12:39:25Z

### L-010 - Classify branching web components under the Test policy and assert every branch, including fallback messages and redirect guards
- signal: `ac_gap` · recurrence: 3 feature(s) · scope: `web` · harmful: 0
- features: users, rbac, audit
- evidence: web/src/features/users/LoginPage.tsx:21 (web) (+2 more)
- last seen: 2026-10-08T23:47:10Z

## Candidates (under observation - do NOT load as guidance yet)

Seen once or not yet corroborated. Tracked, not trusted.

### L-001 - Enumerate every subcommand of a CLI entry point when listing its exit codes, including the error exit of auxiliary subcommands
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `cmd` · harmful: 0
- features: foundation
- evidence: app/cmd/api/main.go:122-125 (cmd)
- last seen: 2026-10-08T13:23:07Z

### L-002 - When a fix adds a branch to a component tested at its own layer, add the own-layer case too, not only an assembled-server case
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `platform` · harmful: 0
- features: foundation
- evidence: app/internal/platform/webui/webui.go:31 (platform)
- last seen: 2026-10-08T13:23:07Z

### L-003 - Give process dispatch exits (missing argument, unknown command) a proof wherever a binary's main switches on its arguments
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `cmd` · harmful: 0
- features: foundation
- evidence: app/cmd/agenthooks/main.go:23-36 (cmd)
- last seen: 2026-10-08T13:23:07Z

### L-004 - Do not rely on the default 1s findBy timeout in tests that run inside the gate; set an explicit timeout so load cannot turn the gate red
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: foundation
- evidence: web/src/features/status/ApiStatus.test.tsx:21 (web)
- last seen: 2026-10-08T13:23:08Z

### L-005 - Prove every success arm of a binary's argument dispatch by running the built binary, not only its failure exits
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `cmd` · harmful: 0
- features: foundation
- evidence: verification.md round 3 gap 1 - app/cmd/agenthooks/main.go:29-32 (cmd)
- last seen: 2026-10-08T13:48:49Z

### L-006 - When a requirement names the literal command a hook runs, assert that literal, not only an injected substitute
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `cmd` · harmful: 0
- features: foundation
- evidence: C51 - app/cmd/agenthooks/main.go:32 (cmd)
- last seen: 2026-10-08T13:48:50Z

### L-007 - Enumerate an auth middleware's decision rows from its code, including the branch taken when no database is configured, and assert each one at its own layer
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `platform` · harmful: 0
- features: users
- evidence: app/internal/platform/auth/auth.go:133 (platform)
- last seen: 2026-10-08T16:13:37Z

### L-009 - When a criterion names two forms such as creation or edition, prove each form, not only the first
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: users
- evidence: web/src/features/users/UserDetail.tsx:70 (web)
- last seen: 2026-10-08T16:13:39Z

### L-011 - Make a check's example input satisfy the precondition it states; a case variant must normalize to the existing value
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: users
- evidence: C2 (checks)
- last seen: 2026-10-08T16:13:40Z

### L-012 - List every test that carries a clause of a check in its Proof lines, not only the main test
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: users
- evidence: C50 (checks)
- last seen: 2026-10-08T16:13:41Z

### L-013 - Assert the exact count of audit rows a mutation writes, not only that one can be read
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `audit` · harmful: 0
- features: users
- evidence: app/internal/features/users/change_password/change_password_test.go:72 (audit)
- last seen: 2026-10-08T16:13:41Z

### L-014 - Do not assert wall-clock latency ratios from a few samples in tests that run inside the gate; prove timing equivalence by counting the work done instead
- signal: `gate_fail` · recurrence: 1 feature(s) · scope: `login` · harmful: 0
- features: users
- evidence: C13 - app/internal/features/users/login/login_test.go:224 failed in task check at fa85de1 (login)
- last seen: 2026-10-08T17:59:48Z

### L-015 - Prove a lock-based race guard by forcing both transactions to be in flight before either commits (hold a blocking lock and wait for waiters in pg_stat_activity), never by repeating unsynchronised requests
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `concurrency` · harmful: 0
- features: rbac
- evidence: F1 - app/internal/features/rbac/assign_roles/queries.sql:11 (concurrency)
- last seen: 2026-10-08T20:57:57Z

### L-016 - When a check says exactly, assert set equality, and list framework-added statuses such as 500 in the Surface
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `contract` · harmful: 0
- features: rbac
- evidence: C25 - app/internal/app/rbac_test.go:126 (contract)
- last seen: 2026-10-08T20:57:58Z

### L-017 - When a criterion lists several screens for one state, prove the state on every screen named, including the API-403 variant
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: rbac
- evidence: AC 27 - web/src/features/rbac/RoleForm.tsx:37 (web)
- last seen: 2026-10-08T20:57:58Z

### L-018 - When a claim says either of two requests, prove the failure of each request separately
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: rbac
- evidence: C46 - web/src/features/rbac/UserRoles.tsx:73 (web)
- last seen: 2026-10-08T20:57:58Z

### L-019 - Assert where an error message renders (its id and the aria-describedby link), not only that the text exists
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: rbac
- evidence: C34 - web/src/features/rbac/RoleForm.test.tsx:60 (web)
- last seen: 2026-10-08T20:57:58Z

### L-020 - When a criterion names a and b or both, give the both case its own coverage member and proof
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: rbac
- evidence: AC 9 - app/internal/features/rbac/update_role (checks)
- last seen: 2026-10-08T20:57:58Z

### L-021 - Choose a LOCK TABLE mode for a concurrency gate that lets the auth and middleware reads through, and prove the waiters are inside their transactions (backend_xid set), not only that two are waiting
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `concurrency` · harmful: 0
- features: rbac
- evidence: C21 - app/internal/features/rbac/assign_roles/assign_roles_test.go:202 (concurrency)
- last seen: 2026-10-08T21:17:21Z

### L-022 - A killed mutant does not prove the mechanism a check describes; when the check claims a forced interleaving, verify where the requests wait
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: rbac
- evidence: C21 round 2 (checks)
- last seen: 2026-10-08T21:17:22Z

### L-023 - To confirm a concurrency claim, list the ungranted locks (pg_stat_activity joined to pg_locks) at the moment the gate releases
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `concurrency` · harmful: 0
- features: rbac
- evidence: C21 round 3 - assign_roles_test.go:210 (concurrency)
- last seen: 2026-10-08T21:36:44Z

### L-024 - When a shared mapper branches on a nullable column, give each arm a coverage member and an asserted case, and classify the mapper in the Test policy even when it is not a slice
- signal: `surviving_mutant` · recurrence: 1 feature(s) · scope: `audit` · harmful: 0
- features: audit
- evidence: app/internal/features/audit/event/event.go:58 (audit)
- last seen: 2026-10-08T23:47:09Z

### L-025 - When a Landing door fixes an index's columns, assert the index definition, not only that its name exists
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `migrations` · harmful: 0
- features: audit
- evidence: C16 - app/migrations/schema_test.go:82 (migrations)
- last seen: 2026-10-08T23:47:09Z

### L-026 - A test that inserts a fixture with a null field but asserts only the status does not prove the null arm; assert the value
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: audit
- evidence: audit round 2 - app/internal/features/audit/get_event/get_event_test.go:68 (checks)
- last seen: 2026-10-09T00:11:19Z

### L-027 - When a fix adds a Test policy classification, add the proofs that row requires at the code's own layer in the same fix
- signal: `spec_deviation` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: audit
- evidence: audit round 2 - checks.md Test policy event.From (checks)
- last seen: 2026-10-09T00:11:20Z

### L-028 - Removing a key from a TanStack Router validateSearch schema is an equivalent mutant because raw search params pass through; inject URL faults in the component instead
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: audit
- evidence: audit round 2 - web/src/routes/_authed/audit/index.tsx (web)
- last seen: 2026-10-09T00:11:20Z

### L-029 - In a table-driven own-layer test, make each case vary only the decision it names
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `checks` · harmful: 0
- features: audit
- evidence: audit round 3 - app/internal/features/audit/event/event_test.go:25 (checks)
- last seen: 2026-10-09T00:30:50Z

### L-030 - Before running a proof against an injected fault, confirm the mutation diff is non-empty and is the intended change
- signal: `gate_fail` · recurrence: 1 feature(s) · scope: `verify` · harmful: 0
- features: rename
- evidence: .specs/features/rename/verification.md (round 2, faults F10-F15) (verify)
- last seen: 2026-10-09T15:09:09Z

### L-031 - Write control characters in specs as escaped text, never as the literal byte
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `specs` · harmful: 0
- features: rename
- evidence: .specs/features/rename/checks.md:43 (C13) (specs)
- last seen: 2026-10-09T15:09:09Z

### L-032 - Never put a pipe character, even escaped, inside a report table cell; the validator splits columns on it
- signal: `gate_fail` · recurrence: 1 feature(s) · scope: `verify` · harmful: 0
- features: rename
- evidence: .specs/features/rename/verification.md (round 2, validate_verification first run) (verify)
- last seen: 2026-10-09T15:09:09Z

### L-033 - When a coverage row names a status for a sampled property, the cited proof must issue a request that returns that status
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `http` · harmful: 0
- features: hardening
- evidence: C21 / app/internal/app/hardening_test.go:123 (http)
- last seen: 2026-10-10T17:31:45Z

### L-034 - When a mapping rewrites some cases and preserves others, assert the preserved value too, not only the rewritten one
- signal: `spec_precision_gap` · recurrence: 1 feature(s) · scope: `httpx` · harmful: 0
- features: hardening
- evidence: C2, C4 / app/internal/platform/httpx/problem.go:45 (httpx)
- last seen: 2026-10-10T17:31:45Z

### L-035 - A branch added during the build to keep an older check true needs its own check for each side of the branch
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `httpx` · harmful: 0
- features: hardening
- evidence: app/internal/platform/httpx/middleware.go:108,126 (httpx)
- last seen: 2026-10-10T17:31:46Z

## Quarantined (failed when applied - ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
