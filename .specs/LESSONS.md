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

## Quarantined (failed when applied - ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
