# LESSONS - auto-maintained by scripts/lessons.py

> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.py` write.
> Canonical state lives in `.specs/lessons.json`. Edit lessons only via the script.
> promote_threshold=2 distinct features · window_days=45 · quarantine_threshold=2

## Confirmed (load these at Plan/Checks)

Corroborated across multiple features. Safe to apply as guidance.

_none_

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

### L-008 - When the plan applies a screen state to a second screen, prove that state on the second screen too
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: users
- evidence: web/src/features/users/UserDetail.tsx:94 (web)
- last seen: 2026-10-08T16:13:38Z

### L-009 - When a criterion names two forms such as creation or edition, prove each form, not only the first
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: users
- evidence: web/src/features/users/UserDetail.tsx:70 (web)
- last seen: 2026-10-08T16:13:39Z

### L-010 - Classify branching web components under the Test policy and assert every branch, including fallback messages and redirect guards
- signal: `ac_gap` · recurrence: 1 feature(s) · scope: `web` · harmful: 0
- features: users
- evidence: web/src/features/users/LoginPage.tsx:21 (web)
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

## Quarantined (failed when applied - ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
