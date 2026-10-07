# Foundation checks

Profile: standard
Plan: `.specs/features/foundation/plan.md`

## Intent

55 checks in 6 slices · 11 one-way doors · 0 open

Repositório greenfield: nenhum comando existe ainda. Todo `Proof:` abaixo usa comandos que esta feature cria
(`go -C app test`, `npm --prefix web run test`, `task ...`); o build cria o comando antes de rodar a prova.
Testes que precisam de Postgres usam testcontainers-go e exigem Docker.

## Checks

### S1 - O backend sobe, responde saúde e migra · ~12 files · ~40 KB · ~10k

**C1** - `GET /healthz` returns `200` with body `{"status":"ok"}` when the database pool is closed (FND-01, AC 1) `[done]`
Proof: `go -C app test ./internal/platform/health -run '^TestHealthz_OKWithoutDatabase$'`

**C2** - `GET /readyz` returns `200` with body `{"status":"ready"}` against a live Postgres (FND-01, AC 2) `[done]`
Proof: `go -C app test ./internal/platform/health -run '^TestReadyz_ReadyWhenDatabaseAnswers$'`

**C3** - `GET /readyz` returns `503` with content type `application/problem+json` when the database is stopped, and answers within 3 seconds (FND-01, AC 3) `[done]`
Proof: `go -C app test ./internal/platform/health -run '^TestReadyz_503WhenDatabaseDown$'`

**C4** - Running the built `api serve` binary with `DATABASE_URL` unset exits with code `1` and stderr contains `DATABASE_URL` (FND-01, AC 4) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestServe_MissingDatabaseURLExits1$'`

**C5** - Cancelling the serve context while a request is in flight lets that request finish with `200`, refuses a new connection, and returns `nil` (exit `0`) (FND-01, AC 5) `[done]`
Proof: `go -C app test ./internal/app -run '^TestRun_ShutdownDrainsInFlight$'`

**C6** - The shutdown drain is bounded: the default timeout is exactly `10s`, and a handler blocking longer than the configured timeout does not keep `Run` from returning (FND-01, AC 5) `[done]`
Proof: `go -C app test ./internal/platform/config -run '^TestDefaults_ShutdownTimeoutIs10s$'`
Proof: `go -C app test ./internal/app -run '^TestRun_ShutdownTimeoutBoundsDrain$'`

**C7** - `api migrate up` on an empty database applies every file in `app/migrations` (goose version equals the newest file's version) and exits `0` (FND-01, AC 6) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestMigrateUp_AppliesPending$'`

**C8** - A second `api migrate up` applies nothing (goose version unchanged) and exits `0` (FND-01, AC 7) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestMigrateUp_RerunIsNoop$'`

**C9** - `api serve` against an empty database creates no `goose_db_version` table (door 8) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestServe_DoesNotMigrate$'`

**C10** - A migration file not matching `^\d{14}_[a-z0-9_]+\.sql$` makes archtest fail naming the file; the real `app/migrations` passes (door 8) `[done]`
Proof: `go -C app test ./archtest -run '^TestMigrationNames_RejectsNonTimestamp$'`
Proof: `go -C app test ./archtest -run '^TestMigrationNames_RepositoryIsClean$'`

### S2 - Contrato HTTP compartilhado · ~10 files · ~35 KB · ~9k

**C11** - Error responses for `404`, `500` and `503` each carry content type `application/problem+json` and the fields `type`, `title`, `status`, `detail`, `request_id`, with `request_id` equal to the `X-Request-ID` response header (FND-02, AC 8) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestProblem_HasRequiredFields$'`

**C12** - `GET /api/does-not-exist` returns `404` with an `application/problem+json` body (FND-02, AC 9) `[done]`
Proof: `go -C app test ./internal/app -run '^TestRouting_UnknownAPIPathIs404Problem$'`

**C13** - A panicking handler yields `500` `application/problem+json` whose `detail` contains neither `goroutine` nor `.go:` (FND-02, AC 10) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestRecover_500WithoutStack$'`

**C14** - A panicking handler produces exactly one log entry at level `ERROR` whose `request_id` equals the response `X-Request-ID` (FND-02, AC 11) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestRecover_LogsErrorWithRequestID$'`

**C15** - A request with `X-Request-ID: abc-123` gets `X-Request-ID: abc-123` back (FND-02, AC 12) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestRequestID_EchoesIncoming$'`

**C16** - A request with no `X-Request-ID` gets a response header that parses as a UUID, distinct across two requests (FND-02, AC 13) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestRequestID_GeneratesUUID$'`

**C17** - Each request writes one JSON log line containing all 8 keys `time`, `level`, `msg`, `request_id`, `method`, `path`, `status`, `duration_ms` (FND-02, AC 14) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestAccessLog_HasAllKeys$'`

**C18** - `GET /api/openapi.json` returns `200` with a body byte-identical to `app/openapi.json` and `openapi` field starting with `3.1` (FND-02, AC 15) `[done]`
Proof: `go -C app test ./internal/app -run '^TestOpenAPI_ServedMatchesCommitted$'`

**C19** - Registering a spec with empty `Permission` and `Public: false` returns an error whose message contains the operation ID (FND-02, AC 16) `[done]`
Proof: `go -C app test ./internal/platform/op -run '^TestRegister_RejectsMissingPermission$'`

**C20** - Registering a `POST`, `PUT`, `PATCH` or `DELETE` spec with empty `AuditAction` returns an error containing the operation ID, for each of the 4 methods (FND-02, AC 17) `[done]`
Proof: `go -C app test ./internal/platform/op -run '^TestRegister_RejectsMutationWithoutAudit$'`

**C21** - Valid specs register without error: a `GET` with `Public: true` and no audit action, a `GET` with a permission, and a `POST` with a permission and an audit action (FND-02, AC 16, AC 17) `[done]`
Proof: `go -C app test ./internal/platform/op -run '^TestRegister_AcceptsValidSpecs$'`

**C22** - `app.New` fails, naming the operation ID, when a registered feature declares an operation without permission - the same function `cmd/api` uses (FND-02, AC 16, door 11) `[done]`
Proof: `go -C app test ./internal/app -run '^TestNew_FailsOnInvalidOperation$'`
Proof: `go -C app test ./archtest -run '^TestCompositionRoot_CmdAPIUsesAppNew$'`

### S3 - Regras de arquitetura executáveis · ~10 files · ~30 KB · ~8k

**C23** - Against a fixture module where `features/a` imports `features/b`, archtest reports a violation whose text contains both import paths (FND-03, AC 18) `[done]`
Proof: `go -C app test ./archtest -run '^TestImports_RejectsCrossFeature$'`

**C24** - Against a fixture module where `platform/x` imports `features/a`, archtest reports a violation containing both import paths (FND-03, AC 19) `[done]`
Proof: `go -C app test ./archtest -run '^TestImports_RejectsPlatformToFeature$'`

**C25** - Against a fixture where `features/a/slice` imports `features/a` and `internal/app` imports `features/a`, archtest reports no violation (FND-03, AC 18, door 11) `[done]`
Proof: `go -C app test ./archtest -run '^TestImports_AllowsSameFeatureAndCompositionRoot$'`

**C26** - archtest reports no violation on the real `app` module (FND-03, AC 18, AC 19) `[done]`
Proof: `go -C app test ./archtest -run '^TestImports_RepositoryIsClean$'`

**C27** - With one sqlc-generated file edited in a temp copy, `task gen:sqlc:check` exits non-zero; on the committed tree it exits `0` (FND-03, AC 20) `[done]`
Proof: `go -C app test ./archtest -run '^TestGenCheck_SqlcDriftFails$'`
Proof: `task gen:sqlc:check`

**C28** - With `app/openapi.json` edited in a temp copy, `task gen:openapi:check` exits non-zero; on the committed tree it exits `0` (FND-03, AC 21) `[done]`
Proof: `go -C app test ./archtest -run '^TestGenCheck_OpenAPIDriftFails$'`
Proof: `task gen:openapi:check`

**C29** - With `web/src/api/schema.d.ts` edited in a temp copy, `npm run gen:check` exits non-zero; on the committed tree it exits `0` (FND-03, AC 22) `[done]`
Proof: `npm --prefix web run test -- src/api/genCheck.test.ts -t "drift fails"`
Proof: `npm --prefix web run gen:check`

**C30** - The `check` task in `Taskfile.yml` invokes all 8 steps - format check, golangci-lint, generated-code diffs, archtest, `go test`, web typecheck, web lint, web unit tests - as sequential `cmds` (Task stops at the first non-zero) (FND-03, AC 23) `[done]`
Proof: `go -C app test ./archtest -run '^TestTaskfile_CheckRunsAllSteps$'`

**C31** - `app/go.mod` declares module `github.com/luiszkm/template-go`; `app/tools/go.mod` declares `tool` directives for `sqlc`, `goose` and `golangci-lint`; there is no `go.mod` at the repository root (door 1, door 13 - amended with user approval 2026-10-07) `[done]`
Proof: `go -C app test ./archtest -run '^TestModuleLayout$'`

**C32** - `app/go.mod` requires `huma/v2`, `pgx/v5`, `goose/v3`, `caarlos0/env`, `testify`, `testcontainers-go`, `otel`; `web/package.json` lists `react` 19, `vite`, `@tanstack/react-router`, `@tanstack/react-query`, `tailwindcss` 4, `zod`, `react-hook-form`, `openapi-typescript`, `openapi-fetch`, `vitest`, `@testing-library/react`, `@playwright/test`, `@biomejs/biome` (door 10) `[done]`
Proof: `go -C app test ./archtest -run '^TestDependencies_Declared$'`

### S4 - Gerador de slice · ~8 files · ~30 KB · ~8k

**C33** - `newslice --feature users --name create_user` on an existing feature creates `features/users/create_user/{endpoint.go,queries.sql,create_user_test.go}`, adds the slice to `features/users/register.go`, and adds one `sqlc.yaml` entry (FND-04, AC 24, door 6) `[done]`
Proof: `go -C app test ./cmd/newslice -run '^TestGenerate_CreatesSliceAndRegisters$'`

**C34** - `newslice` for a feature with no folder creates `features/<f>/register.go` and adds the feature to `features/registry.go` (FND-04, AC 27) `[done]`
Proof: `go -C app test ./cmd/newslice -run '^TestGenerate_NewFeatureRegistersFeature$'`

**C35** - In a temp copy of the repo, after `task new:slice FEATURE=demo NAME=get_thing`, `task check` exits `0` (FND-04, AC 25) `[done]`
Proof: `go -C app test ./cmd/newslice -run '^TestGenerated_TaskCheckPasses$'`

**C36** - The generated endpoint is registered at a path starting `/api/v1/demo/`, answers `501` `application/problem+json`, and the generated test file contains an assertion on `501` that passes (FND-04, AC 26, door 2) `[done]`
Proof: `go -C app test ./cmd/newslice -run '^TestGenerated_SliceAnswers501$'`

**C37** - When the slice folder already exists, `newslice` exits non-zero and the SHA-256 of every file under `app/` is unchanged (FND-04, AC 28) `[done]`
Proof: `go -C app test ./cmd/newslice -run '^TestGenerate_RefusesExistingSlice$'`

**C38** - Each invalid input - `FEATURE=Users`, `FEATURE=1x`, `NAME=a-b`, `NAME=` (empty) - makes `newslice` exit non-zero and create no file (FND-04, AC 29) `[done]`
Proof: `go -C app test ./cmd/newslice -run '^TestGenerate_RejectsInvalidNames$'`

### S5 - Web shell · ~20 files · ~45 KB · ~12k

**C39** - When `GET /readyz` resolves `200`, `/` shows the text `API: online` (FND-05, AC 30) `[done]`
Proof: `npm --prefix web run test -- src/features/status/ApiStatus.test.tsx -t "online"`

**C40** - When `GET /readyz` resolves `503`, and separately when it rejects with a network error, `/` shows `API: offline` and a button named `Tentar novamente` (FND-05, AC 31) `[done]`
Proof: `npm --prefix web run test -- src/features/status/ApiStatus.test.tsx -t "offline"`

**C41** - Clicking `Tentar novamente` issues a second `GET /readyz` (request count goes from 1 to 2) (FND-05, AC 32) `[done]`
Proof: `npm --prefix web run test -- src/features/status/ApiStatus.test.tsx -t "retry"`

**C42** - While `GET /readyz` is unresolved, an element with `role="status"` is present, and it is gone after resolution (FND-05, AC 33) `[done]`
Proof: `npm --prefix web run test -- src/features/status/ApiStatus.test.tsx -t "pending"`

**C43** - Navigating to `/nao-existe` shows `Página não encontrada` and a link whose `href` is `/` (FND-05, AC 34) `[done]`
Proof: `npm --prefix web run test -- src/routes/notFound.test.tsx -t "not found"`

**C44** - The `webui` handler returns `200` with the bytes of `index.html` for `GET /users/42` when no such file exists (FND-05, AC 35, door 9) `[done]`
Proof: `go -C app test ./internal/platform/webui -run '^TestHandler_FallsBackToIndex$'`

**C45** - The `webui` handler returns an existing static file (`/assets/app.js`) with its own bytes, not `index.html` (FND-05, AC 35, door 9) `[done]`
Proof: `go -C app test ./internal/platform/webui -run '^TestHandler_ServesStaticFile$'`

**C46** - In the assembled server, `GET /api/x`, `GET /healthz` and `GET /readyz` are never answered with `index.html` (FND-05, AC 35, AC 9) `[done]`
Proof: `go -C app test ./internal/app -run '^TestRouting_APIAndHealthNotSwallowedBySPA$'`

**C47** - No file under `web/src` outside `web/src/api/` calls `fetch(`, imports `axios`, or constructs `XMLHttpRequest` (FND-05, AC 36) `[done]`
Proof: `npm --prefix web run test -- src/api/boundary.test.ts -t "only generated client"`

**C48** - The Vite dev server proxies `/api`, `/healthz` and `/readyz` to `http://localhost:8080` (door 9) `[done]`
Proof: `npm --prefix web run test -- vite.config.test.ts -t "proxies backend paths"`

**C49** - Against the running stack (`task e2e` starts it), Playwright opens `/` and sees `API: online` (FND-05, AC 30) `[done]`
Proof: `task e2e -- -g "api online"`

### S6 - Harness para agentes · ~8 files · ~15 KB · ~4k

**C50** - Fed a PostToolUse payload for an edited `.go` file with bad formatting, the gofmt hook rewrites it to `gofmt` output and exits `0`; for a `.ts` file it changes nothing and exits `0` (FND-06, AC 37) `[done]`
Proof: `go -C app test ./cmd/agenthooks -run '^TestGofmtHook$'`

**C51** - Fed a Stop payload with `stop_hook_active: false` and a failing `task check:fast`, the Stop hook exits `2` with the failing output on stderr; with a passing command it exits `0`; with `stop_hook_active: true` it exits `0` without running the command (FND-06, AC 38) `[done]`
Proof: `go -C app test ./cmd/agenthooks -run '^TestStopHook$'`

**C52** - `.claude/settings.json` wires `PostToolUse` (matcher `Edit|Write|MultiEdit`) to the gofmt hook and `Stop` to the stop hook (FND-06, AC 37, AC 38) `[done]`
Proof: `go -C app test ./archtest -run '^TestClaudeSettings_WiresHooks$'`

**C53** - `.cursor/rules/agents.mdc` has front-matter `alwaysApply: true` and `.windsurf/rules/agents.md` has `trigger: always_on`, and both bodies reference `AGENTS.md` (FND-06, AC 39) `[done]`
Proof: `go -C app test ./archtest -run '^TestAgentRuleFiles$'`

**C54** - `.github/workflows/ci.yml` has a job whose steps run `task check` and `task e2e`, with no `continue-on-error: true` (FND-06, AC 40) `[done]`
Proof: `go -C app test ./archtest -run '^TestCIWorkflow_RunsCheckAndE2E$'`

**C55** - `task check:fast` runs format check, `go vet`, archtest and web typecheck, and exits non-zero when any fails (FND-06, AC 38) `[done]`
Proof: `go -C app test ./archtest -run '^TestTaskfile_CheckFastSteps$'`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| `GET /healthz` statuses (1) | 200 C1 | - |
| `GET /readyz` statuses (2) | 200 C2 · 503 C3 | - |
| `GET /api/openapi.json` statuses (1) | 200 C18 | - |
| `ANY /api/*` sem rota statuses (1) | 404 C12 | - |
| `GET /*` fora de `/api/` statuses (1) | 200 C44 | - |
| problem+json fields (5) | C11, table-driven over all 5 | - |
| error statuses carrying problem+json (3) | 404 C11 · 500 C11 · 503 C11 | - |
| request log keys (8) | C17, table-driven over all 8 | - |
| `X-Request-ID` cases (2) | present C15 · absent C16 | - |
| mutating methods requiring audit (4) | C20, table-driven over all 4 | - |
| `op.Register` decision (3) | missing permission C19 · mutation without audit C20 · valid specs accepted C21 | - |
| startup assembly (2 places) | `cmd/api` via `app.New` C22 · test harness via `app.New` C22 | - |
| shutdown paths (2) | drains in-flight C5 · bounded by timeout C6 | - |
| migration runs (3) | pending applied C7 · rerun no-op C8 · serve never migrates C9 | - |
| import rules (4) | cross-feature rejected C23 · platform->feature rejected C24 · same-feature allowed C25 · composition root allowed C25 | - |
| generated-code diffs (3) | sqlc C27 · OpenAPI C28 · TS schema C29 | - |
| `task check` steps (8) | C30, table-driven over all 8 | - |
| generator outcomes (5) | new slice C33 · new feature C34 · gate passes C35 · existing refused C37 · invalid refused C38 | - |
| generated files (3) | `endpoint.go` C33 · `queries.sql` C33 · `<n>_test.go` C33 | - |
| invalid generator inputs (4) | C38, table-driven over all 4 | - |
| screen `/` states (4) | loading C42 · online C39 · offline C40 · retry C41 | - |
| `webui` routing (3) | SPA fallback C44 · static file C45 · API/health excluded C46 | - |
| dev proxy prefixes (3) | C48, table-driven over all 3 | - |
| Stop hook branches (3) | failing C51 · passing C51 · `stop_hook_active` C51 | - |
| agent rule files (2) | Cursor C53 · Windsurf C53 | - |
| CI steps (2) | `task check` C54 · `task e2e` C54 | - |
| Landing doors (13) | 1 C31 · 2 C36 · 3 C11 · 4 C15 · 5 C19 · 6 C33 · 7 C23 · 8 C10 · 9 C44 · 10 C32 · 11 C22 · 12 C36 · 13 C31 | - |

- Claims naming a status code, route or response shape: C1-C3, C11-C13, C18, C36, C44-C46 - each proof issues a real HTTP request against a handler or the assembled server
- C49 is a second proof of AC 30 at the browser level; it does not stand in for C39-C42, which assert each state at the component level
- No other check claims more than the single case its proof exercises

## Test policy

O repositório não tem diretriz de testes; estas linhas são a régua deste build e vão para o `AGENTS.md` só se você aprovar.

| Code | Required proofs | Coverage expectation |
| --- | --- | --- |
| Decides, reached across a boundary | one at the boundary **and** one at its own layer | the contract at the boundary; one asserted case per row of the decision table at its own layer |
| Decides, not reached across a boundary | one at its own layer | one asserted case per row of the decision table |
| Entry point that decides nothing | one at the boundary | accepted input, each rejected input, each error path |
| Instrumentation, pass-throughs | none of its own | covered by its consumer's proof |

Evidence (planned code, by shape):

- `internal/platform/op`: decides over permission/public and 4 mutating methods, 3 branch points -> decides, reached across a boundary (C19-C21 own layer, C22 at assembly)
- `internal/platform/webui`: decides over API prefix / existing file / fallback, 3 branch points -> decides, reached across a boundary (C44-C45 own layer, C46 at assembly)
- `archtest`: decides over 4 import rules and the migration name pattern -> decides, not reached across a boundary (fixture modules, C23-C25, C10)
- `cmd/newslice`: decides over name validity, existing slice, existing feature, 4 branch points -> decides, reached across a boundary (C33-C38)
- `cmd/agenthooks`: decides over file extension and 3 Stop branches -> decides (C50-C51)
- `internal/platform/health`: `/healthz` forwards nothing -> entry point that decides nothing; `/readyz` decides on DB answer (C1-C3)
- closest analogue in the repo: none - greenfield

Cost: 14 proofs at their own layer across 6 files, beyond the boundary proofs. Without them the op guard and the SPA fallback are proven only by the paths the assembled server happens to take.

## Swept

- validation: C38, C4
- failure modes: C13, C14, C3
- idempotency: C8, C37
- authorization: C19, C20, C22 - only the registration contract; enforcement is the `rbac` feature
- concurrency: C5, C9
- data lifecycle: n/a - no domain data is stored; only goose bookkeeping
- dependency failure: C3, C40
- state transitions: n/a - no entity with states in this feature
- observability: C14, C16, C17

## Handoff

Intended split, written before any code: greenfield, so the estimate counts files this feature creates rather than files it reads.
S1 ~10k + S2 ~9k + S3 ~8k + S4 ~8k + S5 ~12k + S6 ~4k = ~51k, plus scaffolding output from `npm create vite` and `shadcn` (not hand-read) - under the 150k budget -> one builder.
