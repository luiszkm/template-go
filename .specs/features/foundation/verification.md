# Foundation verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: a927e8b..14cada0 (HEAD)
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier)

All 55 checks are proven at HEAD with located assertions, all 5 injected faults were killed, and `task check` is green.
The FAIL comes from the standard-profile recompute. The code has members that no check asserts, and one of them
contradicts AC 8: a non-Huma error response (`webui` with no `index.html`) is `text/plain`, not
`application/problem+json`. Several `Test policy` decision-table rows are also unasserted. See **Ranked gaps**.

## Binding sources

There are none. The plan marks no source as binding and there is no UI design, so step 1 does not apply.

## Proof runs (all at HEAD 14cada0)

Each named test appears individually in `-v` / `--reporter=verbose` output as `--- PASS` / `✓`. Every name was confirmed to exist in the tree (cited below at file:line).

- **B1** `go -C app test ./archtest -run '^(<16 names>)$' -v -count=1`: exit 0, 16/16 `--- PASS` (68.3s)
- **B2** `go -C app test ./internal/platform/{health,config,httpx,op,webui} ./internal/app ./cmd/api ./cmd/agenthooks -run '^(<27 names>)$' -v -count=1`: exit 0, 27/27 `--- PASS`, including subtests POST/PUT/PATCH/DELETE and 3 Stop subtests
- **B3** `go -C app test ./cmd/newslice -run '^(<5 names>)$' -v -count=1`: exit 0, 5/5 `--- PASS`, including 4 invalid-name subtests
- **B4** `go -C app test ./cmd/newslice -run '^TestGenerated_TaskCheckPasses$' -v -count=1`: exit 0, `--- PASS (268.05s)`
- **B5** `npm --prefix web run test -- <5 files> --reporter=verbose -t "<8 names>"`: exit 0, 9 passed. 1 skipped (`committed schema passes`, outside the filter, and covered by B6)
- **B6** `task gen:sqlc:check`: exit 0. It printed `sqlcrun: no sqlc packages configured; nothing to do`. `task gen:openapi:check`: exit 0. `npm --prefix web run gen:check`: exit 0
- **B7** `task e2e -- -g "api online"`: exit 0, `✓ [chromium] › e2e\status.spec.ts:4:1 › api online`, 1 passed
- **B8** (supporting, not named by any check) `TestReadyz_503WhenDatabaseHangs`, `TestHandler_RootServesIndex`, `TestRegister_RejectsPermissionAndPublicTogether` and `TestGenerate_MissingMarkerWritesNothing`: exit 0, 4/4 `--- PASS`

Every proof file is in the diff `a927e8b..HEAD`. There are no pre-existing tests (greenfield).

## Checks

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `/healthz` 200 `{"status":"ok"}` without DB | B2 `TestHealthz_OKWithoutDatabase` PASS | `app/internal/platform/health/health_test.go:44` - `require.Equal(t, http.StatusOK, rec.Code)`; `:45` - `require.JSONEq(t, `{"status":"ok"}`, ...)`; `untouchable` fake fails the test on any DB call (`:23-24`) | PASS |
| C2 | `/readyz` 200 `{"status":"ready"}` on live PG | B2 `TestReadyz_ReadyWhenDatabaseAnswers` PASS | `health_test.go:60` - `require.Equal(t, http.StatusOK, rec.Code)`; `:61` - `require.JSONEq(t, `{"status":"ready"}`, ...)` | PASS |
| C3 | `/readyz` 503 problem+json within 3s when DB stopped | B2 `TestReadyz_503WhenDatabaseDown` PASS | `health_test.go:72` - `require.Less(t, time.Since(start), 3*time.Second)`; `:73` - `require.Equal(t, http.StatusServiceUnavailable, rec.Code)`; `:74` - `require.Equal(t, "application/problem+json", ...)` | PASS |
| C4 | built binary, `DATABASE_URL` unset: exit 1, stderr names it | B2 `TestServe_MissingDatabaseURLExits1` PASS | `app/cmd/api/main_test.go:46` - `require.Equal(t, 1, exitErr.ExitCode())`; `:47` - `require.Contains(t, stderr.String(), "DATABASE_URL")` | PASS |
| C5 | cancel drains in-flight 200, refuses new conn, returns nil | B2 `TestRun_ShutdownDrainsInFlight` PASS | `app/internal/app/app_test.go:118-125` - `require.Eventually(... DialTimeout err != nil ...)`; `:128` - `require.Equal(t, http.StatusOK, <-status)`; `:131` - `require.NoError(t, err)` | PASS |
| C6 | default shutdown 10s; blocked handler does not hold `Run` | B2 `TestDefaults_ShutdownTimeoutIs10s`, `TestRun_ShutdownTimeoutBoundsDrain` PASS | `app/internal/platform/config/config_test.go:16` - `require.Equal(t, 10*time.Second, cfg.ShutdownTimeout)`; `app_test.go:162` - `require.Less(t, time.Since(start), 2*time.Second)` (timeout 200ms) | PASS |
| C7 | `migrate up` applies all, version = newest, exit 0 | B2 `TestMigrateUp_AppliesPending` PASS | `main_test.go:85` - `require.Equal(t, 0, code, stderr)`; `:86` - `require.Equal(t, newestMigrationVersion(t), gooseVersion(t, pg.URL))` | PASS |
| C8 | rerun applies nothing, exit 0 | B2 `TestMigrateUp_RerunIsNoop` PASS | `main_test.go:103` - `require.Equal(t, 0, code, stderr)`; `:104` - `require.Equal(t, before, gooseVersion(...))`; `:107` - `require.Equal(t, rowsBefore, rowsAfter)` | PASS |
| C9 | `serve` creates no `goose_db_version` | B2 `TestServe_DoesNotMigrate` PASS | `main_test.go:121` - `require.Equal(t, 0, <-done)`; `:128` - `require.False(t, exists)` | PASS |
| C10 | bad migration name fails naming file; real dir clean | B1 `TestMigrationNames_RejectsNonTimestamp`, `TestMigrationNames_RepositoryIsClean` PASS | `app/archtest/imports_test.go:48` - `require.Len(t, bad, 1)`; `:49` - `require.Contains(t, bad[0], "001_bad.sql")`; `:56` - `require.Empty(t, bad)` | PASS |
| C11 | 404/500/503 problem+json with 5 fields, `request_id` == header | B2 `TestProblem_HasRequiredFields` PASS | `app/internal/platform/httpx/httpx_test.go:69` - content type `application/problem+json`; `:72-73` - `require.Contains(t, body, field)` over 5 fields; `:76` - `require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), body["request_id"], path)` | PASS |
| C12 | `GET /api/does-not-exist` 404 problem+json | B2 `TestRouting_UnknownAPIPathIs404Problem` PASS | `app_test.go:44` - `require.Equal(t, http.StatusNotFound, rec.Code)`; `:45` - content type `application/problem+json` (assembled via `app.New`) | PASS |
| C13 | panic -> 500, detail free of `goroutine` / `.go:` | B2 `TestRecover_500WithoutStack` PASS | `httpx_test.go:85` - `require.Equal(t, http.StatusInternalServerError, rec.Code)`; `:90` - `require.NotContains(t, detail, "goroutine")`; `:91` - `require.NotContains(t, detail, ".go:")` | PASS |
| C14 | panic -> exactly one ERROR log with same `request_id` | B2 `TestRecover_LogsErrorWithRequestID` PASS | `httpx_test.go:105` - `require.Len(t, errs, 1)`; `:106` - `require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), errs[0]["request_id"])` | PASS |
| C15 | `X-Request-ID: abc-123` echoed | B2 `TestRequestID_EchoesIncoming` PASS | `httpx_test.go:113` - `require.Equal(t, "abc-123", rec.Header().Get(httpx.HeaderRequestID))` | PASS |
| C16 | absent header -> UUID, distinct across 2 requests | B2 `TestRequestID_GeneratesUUID` PASS | `httpx_test.go:121-124` - `uuid.Parse(a)` / `uuid.Parse(b)` NoError; `:125` - `require.NotEqual(t, a, b)` | PASS |
| C17 | one JSON log line with all 8 keys | B2 `TestAccessLog_HasAllKeys` PASS | `httpx_test.go:133` - `require.Len(t, entries, 1)`; `:135-136` - `require.Contains(t, e, k)` over the 8 keys | PASS |
| C18 | `/api/openapi.json` 200, byte-identical, `openapi` 3.1 | B2 `TestOpenAPI_ServedMatchesCommitted` PASS | `app_test.go:61` - `require.Equal(t, http.StatusOK, rec.Code)`; `:64` - `require.Equal(t, string(committed), rec.Body.String(), ...)`; `:69` - `require.True(t, strings.HasPrefix(doc.OpenAPI, "3.1"))` | PASS |
| C19 | no permission + not public -> error naming op ID | B2 `TestRegister_RejectsMissingPermission` PASS | `app/internal/platform/op/op_test.go:27` - `require.Error(t, err)`; `:28` - `require.Contains(t, err.Error(), "list-things")` | PASS |
| C20 | POST/PUT/PATCH/DELETE without audit -> error naming ID | B2 `TestRegister_RejectsMutationWithoutAudit` + 4 subtests PASS | `op_test.go:36` - `require.Error(t, err)`; `:37` - `require.Contains(t, err.Error(), "mutate-"+m)` (table over 4 methods, `:33`) | PASS |
| C21 | 3 valid specs register | B2 `TestRegister_AcceptsValidSpecs` PASS | `op_test.go:45-47` - `require.NoError(t, op.Register(...))` x3; `:48-50` - `require.NotNil(t, api.OpenAPI().Paths[...].Get/Post)` | PASS |
| C22 | `app.New` fails naming op ID; `cmd/api` uses `app.New` | B2 `TestNew_FailsOnInvalidOperation`, B1 `TestCompositionRoot_CmdAPIUsesAppNew` PASS | `app_test.go:80` - `require.Error(t, err)`; `:81` - `require.Contains(t, err.Error(), "no-permission-op")`; `app/archtest/composition_test.go:17` - `require.Contains(t, imports, ".../internal/app")`; `:19-20` - no `/internal/features` nor `huma/v2` import. Assembly read directly: `app/cmd/api/main.go:86` `app.New(app.Options{...})` | PASS |
| C23 | `features/a` -> `features/b` reported with both paths | B1 `TestImports_RejectsCrossFeature` PASS | `imports_test.go:16` - `require.Len(t, v, 1)`; `:17-18` - `require.Contains(t, v[0], ".../features/a")` and `".../features/b"` | PASS |
| C24 | `platform/x` -> `features/a` reported with both paths | B1 `TestImports_RejectsPlatformToFeature` PASS | `imports_test.go:25` - `require.Len(t, v, 1)`; `:26-27` - `require.Contains(t, v[0], ".../platform/x")` and `".../features/a"` | PASS |
| C25 | same-feature and composition root allowed | B1 `TestImports_AllowsSameFeatureAndCompositionRoot` PASS | `imports_test.go:34` - `require.Empty(t, v)` over fixture `testdata/allowed` (slice->a, internal/app->a) | PASS |
| C26 | real `app` module clean | B1 `TestImports_RepositoryIsClean` PASS | `imports_test.go:41` - `require.Empty(t, v)` | PASS |
| C27 | sqlc drift fails in copy; committed tree passes | B1 `TestGenCheck_SqlcDriftFails` PASS; B6 `task gen:sqlc:check` exit 0 | `app/archtest/gencheck_test.go:56` - `require.NoError(t, err, "freshly generated code must pass")`; `:64` - `require.Error(t, err, "edited generated code must fail")`. Proof runs `go run ./cmd/sqlcrun diff`, the exact command of `gen:sqlc:check` (`Taskfile.yml:107`) | PASS |
| C28 | openapi.json drift fails in copy; committed passes | B1 `TestGenCheck_OpenAPIDriftFails` PASS; B6 `task gen:openapi:check` exit 0 | `gencheck_test.go:76` - `require.Error(t, err, "edited openapi.json must fail the check")` | PASS |
| C29 | schema.d.ts drift fails; committed passes | B5 `gen:check > drift fails` ✓; B6 `npm run gen:check` exit 0 | `web/src/api/genCheck.test.ts:20` - `expect(run(spec, edited).status).not.toBe(0)` | PASS |
| C30 | `check` runs the 8 steps as sequential cmds | B1 `TestTaskfile_CheckRunsAllSteps` PASS | `app/archtest/harness_test.go:76` - `require.Equal(t, []string{"fmt:check","lint","gen:check","archtest","test","web:typecheck","web:lint","web:test"}, steps)`; `:80-90` - each step's command | PASS |
| C31 | module path; tools module with 3 `tool`s; no root go.mod | B1 `TestModuleLayout` PASS | `harness_test.go:102` - `HasPrefix(..., "module github.com/luiszkm/template-go\n")`; `:109` - `require.Contains(t, tools, "\t"+tool+"\n")` x3; `:111` - `require.NoFileExists(t, .../go.mod)` | PASS |
| C32 | Go and web dependencies declared | B1 `TestDependencies_Declared` PASS | `harness_test.go:122` - `require.Contains(t, gomod, "\t"+mod)` x7; `:142` - `require.Contains(t, all, name)` x13; `:144` react `19.`; `:145` tailwindcss `4.` | PASS |
| C33 | slice files, register.go entry, one sqlc entry | B3 `TestGenerate_CreatesSliceAndRegisters` PASS | `app/cmd/newslice/newslice_test.go:75` - `require.FileExists` x3; `:79` - `Contains(register, "createuser.Register(api, d),")`; `:81` - `require.Equal(t, entriesBefore+1, ...Count("- engine:"))` | PASS |
| C34 | new feature: register.go + registry entry | B3 `TestGenerate_NewFeatureRegistersFeature` PASS | `newslice_test.go:93` - `Contains(register, "package billing")`; `:96-97` - registry imports `.../features/billing` and `billing.Register(api, d),` | PASS |
| C35 | after `task new:slice`, `task check` exits 0 | B4 `TestGenerated_TaskCheckPasses` PASS (268s) | `app/cmd/newslice/repo_test.go:128` - `require.NoError(t, err, out)` on `task check` in the copy | PASS |
| C36 | generated endpoint at `/api/v1/demo/`, 501 problem+json, generated test passes | B3 `TestGenerated_SliceAnswers501` PASS | `repo_test.go:148` - `require.Equal(t, []string{"/api/v1/demo/get-thing"}, demo)`; `:156` - test file contains `http.StatusNotImplemented`; `:160` - `Contains(out, "--- PASS: TestEndpoint_NotImplemented")`; template `app/cmd/newslice/main.go:279-280` asserts 501 and `application/problem+json` | PASS |
| C37 | existing slice: non-zero, every file hash unchanged | B3 `TestGenerate_RefusesExistingSlice` PASS | `newslice_test.go:106` - `require.ErrorContains(t, err, "already exists")`; `:107` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C38 | 4 invalid inputs: non-zero, no file | B3 `TestGenerate_RejectsInvalidNames` + 4 subtests PASS | `newslice_test.go:123` - `require.Error(t, err)`; `:124` - `require.Equal(t, before, treeHash(t, root))` (table over 4, `:112-117`) | PASS |
| C39 | readyz 200 -> `API: online` | B5 `shows online when readyz answers 200` ✓ | `web/src/features/status/ApiStatus.test.tsx:21` - `expect(await screen.findByText("API: online")).toBeInTheDocument()` | PASS |
| C40 | 503 and network error -> `API: offline` + retry button | B5 `...answers 503` ✓, `...fails at the network` ✓ | `ApiStatus.test.tsx:29`, `:37` - `findByText("API: offline")`; `:30`, `:38` - `getByRole("button", { name: "Tentar novamente" })` | PASS |
| C41 | retry issues a second readyz (1 -> 2) | B5 `retry requests readyz again` ✓ | `ApiStatus.test.tsx:22` - `toHaveLength(1)` (first load); `:47` - `expect(readyz(calls)).toHaveLength(2)` | PASS |
| C42 | pending shows `role="status"`, gone after | B5 `shows a status element while readyz is pending` ✓ | `ApiStatus.test.tsx:60` - `findByRole("status")`; `:68` - `expect(screen.queryByRole("status")).not.toBeInTheDocument()` | PASS |
| C43 | `/nao-existe` -> `Página não encontrada` + link `/` | B5 `not found page links back to /` ✓ | `web/src/routes/notFound.test.tsx:8` - `findByText("Página não encontrada")`; `:9` - `toHaveAttribute("href", "/")` | PASS |
| C44 | webui `/users/42` -> 200 index bytes | B2 `TestHandler_FallsBackToIndex` PASS | `app/internal/platform/webui/webui_test.go:28` - `require.Equal(t, http.StatusOK, rec.Code)`; `:29` - `require.Equal(t, "<html>index</html>", rec.Body.String())` | PASS |
| C45 | webui `/assets/app.js` -> its own bytes | B2 `TestHandler_ServesStaticFile` PASS | `webui_test.go:35` - 200; `:36` - `require.Equal(t, "console.log('app')", rec.Body.String())` | PASS |
| C46 | assembled server: `/api/x`, `/healthz`, `/readyz` never SPA | B2 `TestRouting_APIAndHealthNotSwallowedBySPA` PASS | `app_test.go:53` - `require.NotContains(t, rec.Body.String(), "<html>spa</html>", p)` over 3 paths; `:55` - control path does get the SPA | PASS |
| C47 | no `fetch(`/axios/XHR outside `web/src/api/` | B5 `only generated client calls the backend` ✓ | `web/src/api/boundary.test.ts:22` - `expect(offenders).toEqual([])`. Independent `grep -rnE` over `web/src` finds `fetch(` only in `src/api/client.ts:10` | PASS |
| C48 | Vite proxies 3 prefixes to `:8080` | B5 `proxies backend paths to the api server` ✓ | `web/vite.config.test.ts:8` - `expect(proxy[path]).toBe("http://localhost:8080")` over 3 paths | PASS |
| C49 | Playwright on running stack sees `API: online` | B7 `api online` ✓ | `web/e2e/status.spec.ts:6` - `await expect(page.getByText("API: online")).toBeVisible()` | PASS |
| C50 | gofmt hook rewrites bad `.go`, leaves `.ts`, exit 0 | B2 `TestGofmtHook` PASS | `app/cmd/agenthooks/main_test.go:34` - `require.Equal(t, 0, gofmtHook(...))`; `:37` - `require.Equal(t, "package x\n\nfunc F() {\n\treturn\n}\n", string(got))`; `:42` - `require.Equal(t, tsSrc, string(got))` | PASS |
| C51 | Stop: failing -> 2 + stderr; passing -> 0; active -> 0, no run | B2 `TestStopHook` + 3 subtests PASS | `main_test.go:68` - `require.Equal(t, 2, stopHook(...fail))`; `:69` - stderr contains output; `:74` - `require.Equal(t, 0, ...pass)`; `:80-81` - 0 with `stop_hook_active: true` and empty stderr. Real command read directly: `app/cmd/agenthooks/main.go:31` `[]string{"task", "check:fast"}` | PASS |
| C52 | settings wires PostToolUse matcher + Stop | B1 `TestClaudeSettings_WiresHooks` PASS | `harness_test.go:163` - matcher equals `Edit`/`Write`/`MultiEdit` alternation; `:165` - `Contains(..., "./cmd/agenthooks gofmt")`; `:170` - `Contains(..., "./cmd/agenthooks stop")` | PASS |
| C53 | Cursor `alwaysApply: true`, Windsurf `trigger: always_on`, both cite AGENTS.md | B1 `TestAgentRuleFiles` PASS | `harness_test.go:186` - `require.Equal(t, true, fm["alwaysApply"])`; `:190` - `require.Equal(t, "always_on", fm["trigger"])`; `:187`, `:191` - `Contains(body, "AGENTS.md")` | PASS |
| C54 | CI runs `task check` and `task e2e`, no continue-on-error | B1 `TestCIWorkflow_RunsCheckAndE2E` PASS | `harness_test.go:208`, `:210` - `require.NotEqual(t, true, ...ContinueOnError)`; `:216` - `require.ElementsMatch(t, []string{"task check", "task e2e"}, found)` | PASS |
| C55 | `check:fast` = fmt, vet, archtest, typecheck | B1 `TestTaskfile_CheckFastSteps` PASS | `harness_test.go:96` - `require.Equal(t, []string{"fmt:check","vet","archtest","web:typecheck"}, calledTasks(...))`; `:97` - vet runs `go vet ./...` | PASS |

## Coverage

Members come from the authority over each set: plan `Surface` / `Landing` / ACs for contract sets, and the code for decision branches and assemblies.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| `GET /healthz` statuses (1) | plan Surface; `health.go:32-34` | 200 C1 | - |
| `GET /readyz` statuses (2) | plan Surface; `health.go:37-50` | 200 C2 · 503 C3 | - |
| `/readyz` 503 triggers (3) | AC 3; `health.go:41`, `:46`, `:63` | DB refused C3 · DB hangs past timeout `health_test.go:83` (B8) · nil DB (`health.go:41-42`) | nil-DB branch: no assertion |
| `GET /api/openapi.json` statuses (1) | plan Surface | 200 C18 | - |
| `ANY /api/*` unrouted statuses (1) | plan Surface; `app.go:66` (method-agnostic `/api/`) | 404 C12 (GET), C11 (GET) | - |
| `GET /*` outside `/api/` statuses (2) | plan Surface lists 200; code `webui.go:25-43` emits 200 and 404 | 200 C44, C45 · 404 `text/plain` when `index.html` is absent (`webui.go:37`, reproduced in a scratch probe) | 404 missing-build: not in Surface, unproven, and not problem+json |
| problem+json fields per AC 8 (5) | AC 8 | C11, table over all 5 (`httpx_test.go:72-73`) | - |
| Landing door 3 error shape (7) | plan Landing door 3 `{type,title,status,detail,instance,request_id,errors?}` | 5 fields C11 · `errors?` optional, no input-validated op at HEAD · `instance` | `instance`: never emitted (scratch probe body has no `instance`), no check |
| error-response emitters (5) | AC 8 ("every error response"); code `problem.go:50-70` (WriteProblem, NotFound), `middleware.go:85-104`, `huma.NewError` `problem.go:18-39`, `webui.go:37` | router 404 C11, C12 · panic 500 C11, C13 · Huma 503 C3, C11 · Huma 501 generated slice C36 (`main.go:280` template) · webui 404 | webui 404 `text/plain` contradicts AC 8 |
| request log keys (8) | AC 14 | C17, table over all 8 | - |
| `X-Request-ID` cases (2) | AC 12-13; `middleware.go:35-37` | present C15 · absent C16 | - |
| mutating methods requiring audit (4) | AC 17; `op.go:49-54` | C20, table over all 4 | - |
| `op.Register` decision rows (5) | `op.go:40-55` | empty ID (`op.go:42`) · no permission and not public C19 · permission and public `op_test.go:55` (B8) · mutation without audit C20 · valid C21 | empty-ID branch: no assertion |
| startup assembly (3 places) | read directly: `cmd/api/main.go:86` (serve), `cmd/api/main.go:114` (openapi export), `internal/app/app_test.go:30` (tests) | all three call `app.New`; C22 (`composition_test.go:17`); export == served is enforced by C18/C28 | - |
| shutdown paths (2) | AC 5; `app.go:94-105` | drains in-flight C5 · bounded by timeout C6 | - |
| migration runs (3) | AC 6-7, door 8 | pending applied C7 · rerun no-op C8 · serve never migrates C9 | - |
| import rules (4) | AC 18-19, door 7, door 11; `archtest.go:49-64` | cross-feature C23 · platform->feature C24 · same feature C25 · composition root C25 | - |
| generated-code diffs (3) | AC 20-22 | sqlc C27 · OpenAPI C28 · TS schema C29 | - |
| `task check` steps (8) | AC 23; `Taskfile.yml:17-26` | C30, table over all 8 | - |
| `task check:fast` steps (4) | plan Assumptions; `Taskfile.yml:28-34` | C55, table over all 4 | - |
| generator outcomes (6) | AC 24-29; `newslice/main.go:44-141` | new slice C33 · new feature C34 · gate passes C35 · existing refused C37 · invalid refused C38 · missing marker writes nothing `newslice_test.go:135-136` (B8) | - |
| generated files (3) | AC 24, door 6 | `endpoint.go`, `queries.sql`, `<n>_test.go` C33 (`newslice_test.go:75`) | - |
| invalid generator inputs (4) | checks C38 | C38, table over all 4 | - |
| screen `/` states (4) | AC 30-33, Observable | loading C42 · online C39 · offline C40 · retry C41 | - |
| web routes (2) | AC 30, AC 34 | `/` C39, C49 · `*` C43 | - |
| `webui` routing (4) | door 9; `webui.go:28-40` | SPA fallback C44 · static file C45 · API/health excluded C46 · `/` serves index `webui_test.go:42` (B8) | - |
| dev proxy prefixes (3) | door 9 | C48, table over all 3 | - |
| Stop hook branches (3) | AC 38; `agenthooks/main.go:75-90` | failing C51 · passing C51 · `stop_hook_active` C51 | - |
| gofmt hook branches (5) | AC 37; `agenthooks/main.go:41-71` | `.go` reformatted C50 · non-`.go` untouched C50 · bad payload (`:47-49`) · unreadable file (`:56-58`) · `.go` that does not parse (`:60-63`) | bad payload, unreadable file, unparseable `.go`: no assertion |
| agent rule files (2) | AC 39 | Cursor C53 · Windsurf C53 | - |
| CI steps (2) | AC 40 | `task check` C54 · `task e2e` C54 | - |
| `cmd/api` outcomes (6) | `cmd/api/main.go:41-69` | serve ok exit 0 C9 · migrate ok exit 0 C7, C8 · openapi export (run by `task gen` in C35/C36, `repo_test.go:148`) · config missing exit 1 C4 · unknown subcommand exit 2 (`main.go:46-48`) · serve/migrate runtime error exit 1 (`main.go:63-66`) | unknown subcommand exit 2; runtime-error exit 1: no assertion |
| Landing doors (13) | plan Landing | 1 C31 · 2 C36 (feature prefix), C1/C2 (root probes), C18 (contract path), `/api/docs` · 3 C11 (see door-3 row) · 4 C15, C14 (chain order `middleware.go:29`) · 5 C19 · 6 C33 · 7 C23 · 8 C10, C9 · 9 C44, C48 · 10 C32 · 11 C22 · 12 C36 (`repo_test.go:152`) · 13 C31 | door 2: `/api/docs` (`httpx/api.go:23`) has no proof |

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `internal/platform/op/op.go` | own layer C19-C21 · boundary C22 | not met - empty-ID row (`op.go:42`) has no asserted case |
| Decides, reached across a boundary | `internal/platform/webui/webui.go` | own layer C44, C45 · boundary C46 | not met - missing-`index.html` row (`webui.go:35-38`, 404 text/plain) has no asserted case |
| Decides, reached across a boundary | `internal/platform/health/health.go` (`/readyz`) | own layer and HTTP C2, C3, hang test | not met - nil-DB row (`health.go:41-42`) has no asserted case |
| Decides, reached across a boundary | `internal/platform/httpx/middleware.go`, `problem.go` (not classified by checks.md) | own layer C11, C13-C17 · boundary C12 | not met - `http.ErrAbortHandler` re-panic row (`middleware.go:92-93`) has no asserted case |
| Decides, reached across a boundary | `cmd/newslice/main.go` | own layer C33, C34, C37, C38 · boundary C35, C36 | yes |
| Decides, not reached across a boundary | `archtest/archtest.go` | own layer C23-C26, C10 (fixture modules) | yes |
| Decides, not reached across a boundary | `cmd/agenthooks/main.go` | own layer C50, C51 | not met - gofmt rows for bad payload, unreadable file and unparseable `.go` (`main.go:47`, `:56`, `:60-63`) have no asserted case |
| Entry point that decides nothing | `health.go` (`/healthz`) | boundary C1 | yes |
| Entry point that decides nothing | `cmd/api/main.go` (`run`) | boundary C4, C7-C9 | not met - rejected input (unknown subcommand, exit 2, `main.go:46-48`) and runtime error path (exit 1, `main.go:63-66`) have no proof |
| Instrumentation, pass-throughs | `platform/db/db.go`, `platform/deps/deps.go`, `platform/testkit/*`, otel wrapper `app.go:69`, `cmd/fmtcheck`, `cmd/sqlcrun`, `cmd/webcopy`, `platform/config/config.go` | none of its own | yes - covered by consumers (C2/C3, C4, C6, C27, gate `fmt:check`, C49 build) |

## Faults injected

Isolation: `git worktree add --detach <scratchpad>/wt HEAD`. The real-tree `git status --porcelain` baseline was ` M .specs/STATE.md`. The 5 faults went into independent packages and each was run against its narrowest proof. The web proof ran from the long (non-8.3) path, with `web/node_modules` as a junction that was removed before cleanup. Discard: `git worktree remove --force` + `prune`. Porcelain afterwards was ` M .specs/STATE.md`, identical to the baseline (`diff` empty). `web/node_modules` is intact.

| Mutation | Location | Killed |
| --- | --- | --- |
| dropped `http.MethodDelete` from the mutating-method case | `app/internal/platform/op/op.go:50` | yes - `TestRegister_RejectsMutationWithoutAudit/DELETE` FAIL |
| cross-feature comparison `!=` -> `==` | `app/archtest/archtest.go:59` | yes - `TestImports_RejectsCrossFeature` FAIL (`"[]" should have 1 item(s)`) |
| request id always regenerated (`if id == ""` -> `if true`) | `app/internal/platform/httpx/middleware.go:36` | yes - `TestRequestID_EchoesIncoming` FAIL |
| SPA fallback reads a missing file instead of `index.html` | `app/internal/platform/webui/webui.go:35` | yes - `TestHandler_FallsBackToIndex` FAIL |
| removed the `Tentar novamente` button from the error state | `web/src/features/status/ApiStatus.tsx:20-22` | yes - both C40 tests FAIL (`Unable to find ... role "button" and name "Tentar novamente"`) |

The first web run from the 8.3 short path failed at suite load (`Cannot find module '/src/test/setup.ts'`, "no tests"). That was not counted as a kill; the rerun from the long path failed on the assertion itself.

## Swept existing re-read

- validation (C38, C4): `newslice/main.go:25` regex `^[a-z][a-z0-9_]*$`, `:45-50` checked before any write; `config.go:8` `env:"DATABASE_URL,required,notEmpty"`. Present.
- failure modes (C13, C14, C3): `middleware.go:85-104` Recover logs ERROR with stack and writes Problem; `health.go:46-47` 503. Present.
- idempotency (C8, C37): goose provider `cmd/api/main.go:105-110`; `newslice/main.go:58-62` refuses an existing slice before planning writes. Present.
- authorization (C19, C20, C22): `op.go:44`, `:49-54`; `app.go:62-64` propagates the error. Present (registration contract only, as scoped).
- concurrency (C5, C9): `app.go:84-105` Shutdown/Close; `serve` has no goose call (`main.go:79-97`). Present.
- dependency failure (C3, C40), observability (C14, C16, C17): present at the lines cited in Checks.
- data lifecycle and state transitions are `n/a` (user-approved policy).

## Precision notes

These do not fail the feature on their own:

- C55 claims `check:fast` "exits non-zero when any fails", but the proof asserts only the step list (`harness_test.go:96-97`). Exit propagation rests on Task's sequential-`cmds` semantics, which no test exercises.
- C27, committed-tree half: `task gen:sqlc:check` passes vacuously at HEAD (`sqlcrun: no sqlc packages configured; nothing to do`, `sqlc.yaml` has `sql: []`). Real drift detection is exercised only in the temp copy (`gencheck_test.go:64`).
- C9 waits on `time.Sleep(time.Second)` (`main_test.go:119`) without a readiness signal. `serve` is proven to exit 0 after cancel (`:121`) but not to have finished starting first.
- C37/C38 claim "exits non-zero" but are asserted on `Generate()`'s error (`newslice_test.go:106`, `:123`). The process mapping was read directly at `newslice/main.go:33-36` (`os.Exit(1)`).
- C47's scanner also excludes `src/test/` and `*.test.ts(x)` (`boundary.test.ts:18-20`), which is broader than the claim's single exclusion `web/src/api/`. At HEAD no file in those excluded paths calls `fetch(` either.
- C54 collects `task check` / `task e2e` across all jobs (`harness_test.go:205-216`), so it does not prove that both sit in one job. At HEAD both are in job `check`.
- C3 exercises only the refused-connection cause. AC 3's literal "does not answer within 2 seconds" is proven only by the unlisted `TestReadyz_503WhenDatabaseHangs` (`health_test.go:81-83`).
- Observation, not a criterion: non-GET requests to `/healthz` / `/readyz` fall through to the `/` SPA handler (a scratch probe got `DELETE /healthz` -> 200 `text/html`), because `app.go:67` mounts `webui` for every method.

## Ranked gaps

1. **AC 8 contradiction / unproven Surface status**: `GET /*` outside `/api/` returns `404 text/plain` ("web build not found") whenever the embedded `dist` has no `index.html`. That is the case for `go run ./cmd/api serve` (`task dev:api`) and for a plain `go build` without `task build`, since `dist` holds only `.gitkeep` at HEAD. AC 8 requires every error response to be `application/problem+json`. No check covers it, and Surface lists only `200` for this route. - C44/C11 - `app/internal/platform/webui/webui.go:37`
2. **Landing door 3 member unproven**: the literal shape names `instance` as non-optional, but no response carries it (the `Problem` literal at `problem.go:51-59` never sets it; a scratch probe of `/readyz` confirmed it absent) and no check mentions it. - C11 - `app/internal/platform/httpx/problem.go:51-59`
3. **Landing door 2 member unproven**: docs at `/api/docs` are configured but nothing proves they are served. - C36 - `app/internal/platform/httpx/api.go:23`
4. **Test policy, decides-across-boundary rows not met**: `op` empty-ID branch (`op.go:42`); `/readyz` nil-DB branch (`health.go:41-42`); `httpx` `ErrAbortHandler` re-panic (`middleware.go:92-93`). None has an asserted case. - C19-C21, C2-C3, C13 - files as cited
5. **Test policy, entry-point row not met for `cmd/api`**: unknown subcommand exit `2` (`cmd/api/main.go:46-48`) and serve/migrate runtime-error exit `1` (`main.go:63-66`) have no proof. - C4 - `app/cmd/api/main.go:46`, `:63`
6. **Test policy, `agenthooks` decision rows not met**: bad payload, unreadable file and unparseable `.go` have no asserted case. - C50 - `app/cmd/agenthooks/main.go:47`, `:56`, `:60`

## Gate

`task check` (repo root, HEAD 14cada0): exit 0 in 7m57s.
- `fmt:check`, `lint` (golangci-lint), `gen:check` (sqlc / OpenAPI / web) and `archtest` all clean.
- `go test -count=1 ./...`: 11 packages ok, 0 failed.
- `web:typecheck` clean; `web:lint` reported "Checked 28 files ... No fixes applied."
- `web:test`: 5 files, 10 passed, 0 failed.

Also `task e2e -- -g "api online"`: 1 passed, 0 failed.
