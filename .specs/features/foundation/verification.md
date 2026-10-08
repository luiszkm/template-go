# Foundation verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: a927e8b..11c5c13 (HEAD); fix under review 14cada0..11c5c13
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier)

All 68 checks are proven at HEAD 11c5c13. Each one has a located assertion, and every named test was re-run and
shows individually as PASS. All 6 round-1 ranked gaps are closed by checks C56-C68, and each of those checks
asserts the value the check defines. All 5 faults injected on the fix's surfaces were killed. `task check` exits 0
at 11c5c13 (see Gate).

The FAIL comes from the scoped coverage and test-policy recompute. It found three members that no proof asserts.
None of them contradicts an AC:

1. The error exit of `api openapi`.
2. The new webui `405` row, which is asserted only at the assembled server and never at its own layer.
3. Three `agenthooks` branches: write failure, the no-argument exit and the unknown-hook exit.

The rows below show them. Under the report's own rules they fail the feature, and they are not softened here.

## Binding sources

verified at 11c5c13: none. The plan marks no source as binding and the fix touched no UI, so step 1 does not apply.

## Proof runs (all at HEAD 11c5c13, verified at 11c5c13)

Each named test appears individually in the `-v` / `--reporter=verbose` output as `--- PASS` / `✓`. Every name was
confirmed to exist with `rg -n "^func Test|^\s*(it|test)\("` over the test files (hits cited in Checks).

- **B1** `go -C app test ./archtest -run '^(<17 names>)$' -v -count=1`: exit 0. 17/17 `--- PASS`, including
  `TestTaskfile_GatesFailOnFailingStep/check` and `/check:fast`.
- **B2** `go -C app test ./internal/platform/{health,config,httpx,op,webui} ./internal/app ./cmd/api ./cmd/agenthooks -run '^(<38 names>)$' -v -count=1`:
  exit 0. 38/38 `--- PASS`, including the subtests POST/PUT/PATCH/DELETE, 3 Stop subtests and 3 `NeverBlocks`
  subtests.
- **B3** `go -C app test ./cmd/newslice -run '^(<7 names>)$' -v -count=1`: 6/7 `--- PASS`, including 4
  invalid-name subtests and 2 `TestMain_FailuresExit1` subtests.
  - `TestGenerated_TaskCheckPasses` FAILED in this run, after 420s. The nested `task check` failed at `web:test`
    on the same C39 test (`× shows online when readyz answers 200`, `Unable to find an element with the text:
    API: online`) while B1, B2 and the outer `task check` loaded the machine.
  - **B3b**: the same test re-run alone with `go -C app test ./cmd/newslice -run '^TestGenerated_TaskCheckPasses$' -v -count=1`
    exited 0, `--- PASS (293.18s)`.
  - The outer `task check` also ran `cmd/newslice`, including this nested gate, and it passed (`ok ... 394.984s`).
  - See ranked gap 4.
- **B5** `npm --prefix web run test -- <5 files> --reporter=verbose -t "online|offline|retry|pending|not found|drift fails|only generated client|proxies backend paths"`:
  exit 0, 9 passed. 1 was skipped (`committed schema passes`), which is outside the filter and covered by B6.
  The first attempt ran while B1-B3 and `task check` loaded the machine, and `shows online when readyz answers 200`
  failed: `findByText` hit its 1s default timeout with the skeleton still rendered. The immediate rerun passed. See
  Precision notes.
- **B6** `task gen:sqlc:check` exit 0 (`sqlcrun: no sqlc packages configured; nothing to do`).
  `task gen:openapi:check` exit 0. `npm --prefix web run gen:check` exit 0.
- **B7** `task e2e -- -g "api online"`: see Gate.
- **B8** (supporting, not named by any check) `TestRegister_RejectsPermissionAndPublicTogether` and
  `TestHandler_RootServesIndex`: exit 0, 2/2 `--- PASS`.

Every proof file is in the diff `a927e8b..HEAD`, and every new proof (C56-C68) is in `14cada0..11c5c13`.

## Checks

Rows for files the fix touched are marked `verified at 11c5c13`. Rows for untouched files are marked
`carried from 14cada0`: their citation is unchanged, and their proof was re-run at 11c5c13.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `/healthz` 200 `{"status":"ok"}` without DB | B2 `TestHealthz_OKWithoutDatabase` PASS | verified at 11c5c13: `app/internal/platform/health/health_test.go:44` - `require.Equal(t, http.StatusOK, rec.Code)`; `:45` - `require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())` | PASS |
| C2 | `/readyz` 200 `{"status":"ready"}` on live PG | B2 `TestReadyz_ReadyWhenDatabaseAnswers` PASS | verified at 11c5c13: `health_test.go:60` - `require.Equal(t, http.StatusOK, rec.Code)`; `:61` - `require.JSONEq(t, `{"status":"ready"}`, ...)` | PASS |
| C3 | `/readyz` 503 problem+json within 3s, DB stopped | B2 `TestReadyz_503WhenDatabaseDown` PASS | verified at 11c5c13: `health_test.go:72` - `require.Less(t, time.Since(start), 3*time.Second)`; `:73` - `require.Equal(t, http.StatusServiceUnavailable, rec.Code)`; `:74` - content type `application/problem+json` | PASS |
| C4 | built binary, `DATABASE_URL` unset: exit 1, stderr names it | B2 `TestServe_MissingDatabaseURLExits1` PASS | verified at 11c5c13: `app/cmd/api/main_test.go:48` - `require.Equal(t, 1, exitErr.ExitCode())`; `:49` - `require.Contains(t, stderr.String(), "DATABASE_URL")` | PASS |
| C5 | cancel drains in-flight 200, refuses new conn, nil | B2 `TestRun_ShutdownDrainsInFlight` PASS | verified at 11c5c13: `app/internal/app/app_test.go:130-138` - `require.Eventually(... DialContext err != nil ...)`; `:141` - `require.Equal(t, http.StatusOK, <-status)`; `:144` - `require.NoError(t, err)` | PASS |
| C6 | default shutdown 10s; blocked handler does not hold `Run` | B2 `TestDefaults_ShutdownTimeoutIs10s`, `TestRun_ShutdownTimeoutBoundsDrain` PASS | carried from 14cada0: `app/internal/platform/config/config_test.go:16` - `require.Equal(t, 10*time.Second, cfg.ShutdownTimeout)`; verified at 11c5c13: `app_test.go:175` - `require.Less(t, time.Since(start), 2*time.Second)` | PASS |
| C7 | `migrate up` applies all, version = newest, exit 0 | B2 `TestMigrateUp_AppliesPending` PASS | verified at 11c5c13: `main_test.go:87` - `require.Equal(t, 0, code, stderr)`; `:88` - `require.Equal(t, newestMigrationVersion(t), gooseVersion(t, pg.URL))` | PASS |
| C8 | rerun applies nothing, exit 0 | B2 `TestMigrateUp_RerunIsNoop` PASS | verified at 11c5c13: `main_test.go:105` - `require.Equal(t, 0, code, stderr)`; `:106` - `require.Equal(t, before, gooseVersion(...))`; `:109` - `require.Equal(t, rowsBefore, rowsAfter)` | PASS |
| C9 | `serve` creates no `goose_db_version` | B2 `TestServe_DoesNotMigrate` PASS | verified at 11c5c13: `main_test.go:122` - `waitServing(t, "http://"+addr+"/healthz", done)` (readiness, replaces round-1 sleep); `:124` - `require.Equal(t, 0, <-done)`; `:131` - `require.False(t, exists)` | PASS |
| C10 | bad migration name fails naming file; real dir clean | B1 `TestMigrationNames_RejectsNonTimestamp`, `TestMigrationNames_RepositoryIsClean` PASS | carried from 14cada0: `app/archtest/imports_test.go:48` - `require.Len(t, bad, 1)`; `:49` - `require.Contains(t, bad[0], "001_bad.sql")`; `:56` - `require.Empty(t, bad)` | PASS |
| C11 | 404/500/503 problem+json, 5 fields, `request_id` == header | B2 `TestProblem_HasRequiredFields` PASS | verified at 11c5c13: `app/internal/platform/httpx/httpx_test.go:79` - content type `application/problem+json`; `:83` - `require.Contains(t, body, field, ...)` over the 5 fields; `:85` - `require.Equal(t, status, intField(t, body, "status"), path)`; `:86` - `require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), body["request_id"], path)` | PASS |
| C12 | `GET /api/does-not-exist` 404 problem+json | B2 `TestRouting_UnknownAPIPathIs404Problem` PASS | verified at 11c5c13: `app_test.go:44` - `require.Equal(t, http.StatusNotFound, rec.Code)`; `:45` - content type `application/problem+json` | PASS |
| C13 | panic -> 500, detail free of `goroutine` / `.go:` | B2 `TestRecover_500WithoutStack` PASS | verified at 11c5c13: `httpx_test.go:95` - `require.Equal(t, http.StatusInternalServerError, rec.Code)`; `:100` - `require.NotContains(t, detail, "goroutine")`; `:101` - `require.NotContains(t, detail, ".go:")` | PASS |
| C14 | panic -> exactly one ERROR log, same `request_id` | B2 `TestRecover_LogsErrorWithRequestID` PASS | verified at 11c5c13: `httpx_test.go:115` - `require.Len(t, errs, 1)`; `:116` - `require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), errs[0]["request_id"])` | PASS |
| C15 | `X-Request-ID: abc-123` echoed | B2 `TestRequestID_EchoesIncoming` PASS | verified at 11c5c13: `httpx_test.go:123` - `require.Equal(t, "abc-123", rec.Header().Get(httpx.HeaderRequestID))` | PASS |
| C16 | absent header -> UUID, distinct across 2 | B2 `TestRequestID_GeneratesUUID` PASS | verified at 11c5c13: `httpx_test.go:132`, `:134` - `uuid.Parse` NoError for both; `:135` - `require.NotEqual(t, a, b)` | PASS |
| C17 | one JSON log line with all 8 keys | B2 `TestAccessLog_HasAllKeys` PASS | verified at 11c5c13: `httpx_test.go:143` - `require.Len(t, entries, 1)`; `:146` - `require.Contains(t, e, k)` over the 8 keys | PASS |
| C18 | `/api/openapi.json` 200, byte-identical, `openapi` 3.1 | B2 `TestOpenAPI_ServedMatchesCommitted` PASS | verified at 11c5c13: `app_test.go:61` - 200; `:64` - `require.Equal(t, string(committed), rec.Body.String(), ...)`; `:69` - `require.True(t, strings.HasPrefix(doc.OpenAPI, "3.1"), ...)` | PASS |
| C19 | no permission + not public -> error naming op ID | B2 `TestRegister_RejectsMissingPermission` PASS | verified at 11c5c13: `app/internal/platform/op/op_test.go:28` - `require.Error(t, err)`; `:29` - `require.Contains(t, err.Error(), "list-things")` | PASS |
| C20 | POST/PUT/PATCH/DELETE without audit -> error naming ID | B2 `TestRegister_RejectsMutationWithoutAudit` + 4 subtests PASS | verified at 11c5c13: `op_test.go:37` - `require.Error(t, err)`; `:38` - `require.Contains(t, err.Error(), "mutate-"+m)` | PASS |
| C21 | 3 valid specs register | B2 `TestRegister_AcceptsValidSpecs` PASS | verified at 11c5c13: `op_test.go:46-48` - `require.NoError(t, op.Register(...))` x3; `:49-51` - `require.NotNil(t, api.OpenAPI().Paths[...])` | PASS |
| C22 | `app.New` fails naming op ID; `cmd/api` uses `app.New` | B2 `TestNew_FailsOnInvalidOperation`, B1 `TestCompositionRoot_CmdAPIUsesAppNew` PASS | verified at 11c5c13: `app_test.go:80` - `require.Error(t, err)`; `:81` - `require.Contains(t, err.Error(), "no-permission-op")`; `app/archtest/composition_test.go:16` - `require.Contains(t, imports, ".../internal/app")`; `:18-19` - `require.NotContains(t, imp, "/internal/features")`, `"huma/v2"` | PASS |
| C23 | `features/a` -> `features/b` reported with both paths | B1 `TestImports_RejectsCrossFeature` PASS | carried from 14cada0: `imports_test.go:16` - `require.Len(t, v, 1)`; `:17-18` - both paths contained | PASS |
| C24 | `platform/x` -> `features/a` reported with both paths | B1 `TestImports_RejectsPlatformToFeature` PASS | carried from 14cada0: `imports_test.go:25` - `require.Len(t, v, 1)`; `:26-27` - both paths contained | PASS |
| C25 | same-feature and composition root allowed | B1 `TestImports_AllowsSameFeatureAndCompositionRoot` PASS | carried from 14cada0: `imports_test.go:34` - `require.Empty(t, v)` | PASS |
| C26 | real `app` module clean | B1 `TestImports_RepositoryIsClean` PASS | carried from 14cada0: `imports_test.go:41` - `require.Empty(t, v)`. At HEAD `webui` -> `platform/httpx` is platform->platform, which the rule allows | PASS |
| C27 | sqlc drift fails in copy; committed passes | B1 `TestGenCheck_SqlcDriftFails` PASS; B6 `task gen:sqlc:check` exit 0 | verified at 11c5c13: `app/archtest/gencheck_test.go:62` - `require.NoError(t, err, "freshly generated code must pass: %s", out)`; `:70` - `require.Error(t, err, "edited generated code must fail: %s", out)` | PASS |
| C28 | openapi.json drift fails in copy; committed passes | B1 `TestGenCheck_OpenAPIDriftFails` PASS; B6 `task gen:openapi:check` exit 0 | verified at 11c5c13: `gencheck_test.go:82` - `require.Error(t, err, "edited openapi.json must fail the check: %s", out)` | PASS |
| C29 | schema.d.ts drift fails; committed passes | B5 `gen:check > drift fails` ✓; B6 `npm run gen:check` exit 0 | carried from 14cada0: `web/src/api/genCheck.test.ts:20` - `expect(run(spec, edited).status).not.toBe(0)` | PASS |
| C30 | `check` runs the 8 steps as sequential cmds | B1 `TestTaskfile_CheckRunsAllSteps` PASS | verified at 11c5c13: `app/archtest/harness_test.go:79` - `require.Equal(t, []string{"fmt:check",...,"web:test"}, ...)`; `:83-93` - each step's command | PASS |
| C31 | module path; tools module with 3 `tool`s; no root go.mod | B1 `TestModuleLayout` PASS | verified at 11c5c13: `harness_test.go:105` - `HasPrefix(..., "module github.com/luiszkm/template-go\n")`; `:112` - `require.Contains(t, tools, "\t"+tool+"\n", ...)`; `:114` - `require.NoFileExists(t, filepath.Join(repoRoot, "go.mod"))` | PASS |
| C32 | Go and web dependencies declared | B1 `TestDependencies_Declared` PASS | verified at 11c5c13: `harness_test.go:125` - `require.Contains(t, gomod, "\t"+mod, ...)`; `:141` - `require.Contains(t, all, name)`; `:143` react `19.`; `:144` tailwindcss `4.` | PASS |
| C33 | slice files, register.go entry, one sqlc entry | B3 `TestGenerate_CreatesSliceAndRegisters` PASS | verified at 11c5c13: `app/cmd/newslice/newslice_test.go:77` - `require.FileExists` x3; `:81` - `require.Contains(t, register, "createuser.Register(api, d),")`; `:83` - `require.Equal(t, entriesBefore+1, strings.Count(..., "- engine:"))` | PASS |
| C34 | new feature: register.go + registry entry | B3 `TestGenerate_NewFeatureRegistersFeature` PASS | verified at 11c5c13: `newslice_test.go:95` - `"package billing"`; `:98-99` - registry imports `.../features/billing` and `billing.Register(api, d),` | PASS |
| C35 | after `task new:slice`, `task check` exits 0 | B3b `TestGenerated_TaskCheckPasses` PASS (failed once under load in B3; see gap 4) | verified at 11c5c13: `app/cmd/newslice/repo_test.go:130` - `require.NoError(t, err, out)` on `task check` in the copy | PASS |
| C36 | generated endpoint at `/api/v1/demo/`, 501 problem+json, generated test passes | B3 `TestGenerated_SliceAnswers501` PASS | verified at 11c5c13: `repo_test.go:150` - `require.Equal(t, []string{"/api/v1/demo/get-thing"}, demo)`; `:158` - test contains `http.StatusNotImplemented`; `:162` - `Contains(out, "--- PASS: TestEndpoint_NotImplemented")`; template `app/cmd/newslice/main.go:279-280` asserts 501 and `application/problem+json` | PASS |
| C37 | existing slice: non-zero, every file hash unchanged | B3 `TestGenerate_RefusesExistingSlice` PASS | verified at 11c5c13: `newslice_test.go:108` - `require.ErrorContains(t, err, "already exists")`; `:109` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C38 | 4 invalid inputs: non-zero, no file | B3 `TestGenerate_RejectsInvalidNames` + 4 subtests PASS | verified at 11c5c13: `newslice_test.go:125` - `require.Error(t, err)`; `:126` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C39 | readyz 200 -> `API: online` | B5 `shows online when readyz answers 200` ✓ (rerun; see Precision notes) | carried from 14cada0: `web/src/features/status/ApiStatus.test.tsx:21` - `expect(await screen.findByText("API: online")).toBeInTheDocument()` | PASS |
| C40 | 503 and network error -> `API: offline` + retry button | B5 both offline tests ✓ | carried from 14cada0: `ApiStatus.test.tsx:29`, `:37` - `findByText("API: offline")`; `:30`, `:38` - `getByRole("button", { name: "Tentar novamente" })` | PASS |
| C41 | retry issues a second readyz (1 -> 2) | B5 `retry requests readyz again` ✓ | carried from 14cada0: `ApiStatus.test.tsx:22` - `toHaveLength(1)`; `:47` - `expect(readyz(calls)).toHaveLength(2)` | PASS |
| C42 | pending shows `role="status"`, gone after | B5 `shows a status element while readyz is pending` ✓ | carried from 14cada0: `ApiStatus.test.tsx:60` - `findByRole("status")`; `:68` - `expect(screen.queryByRole("status")).not.toBeInTheDocument()` | PASS |
| C43 | `/nao-existe` -> `Página não encontrada` + link `/` | B5 `not found page links back to /` ✓ | carried from 14cada0: `web/src/routes/notFound.test.tsx:8` - `findByText("Página não encontrada")`; `:9` - `toHaveAttribute("href", "/")` | PASS |
| C44 | webui `/users/42` -> 200 index bytes | B2 `TestHandler_FallsBackToIndex` PASS | verified at 11c5c13: `app/internal/platform/webui/webui_test.go:29` - `require.Equal(t, http.StatusOK, rec.Code)`; `:30` - `require.Equal(t, "<html>index</html>", rec.Body.String())` | PASS |
| C45 | webui `/assets/app.js` -> its own bytes | B2 `TestHandler_ServesStaticFile` PASS | verified at 11c5c13: `webui_test.go:36` - 200; `:37` - `require.Equal(t, "console.log('app')", rec.Body.String())` | PASS |
| C46 | assembled server: `/api/x`, `/healthz`, `/readyz` never SPA | B2 `TestRouting_APIAndHealthNotSwallowedBySPA` PASS | verified at 11c5c13: `app_test.go:53` - `require.NotContains(t, rec.Body.String(), "<html>spa</html>", p)`; `:55` - control path gets the SPA | PASS |
| C47 | no `fetch(`/axios/XHR outside `web/src/api/` | B5 `only generated client calls the backend` ✓ | carried from 14cada0: `web/src/api/boundary.test.ts:22` - `expect(offenders).toEqual([])` | PASS |
| C48 | Vite proxies 3 prefixes to `:8080` | B5 `proxies backend paths to the api server` ✓ | carried from 14cada0: `web/vite.config.test.ts:8` - `expect(proxy[path]).toBe("http://localhost:8080")` | PASS |
| C49 | Playwright on running stack sees `API: online` | B7 `api online` (see Gate) | carried from 14cada0: `web/e2e/status.spec.ts:6` - `await expect(page.getByText("API: online")).toBeVisible()` | PASS |
| C50 | gofmt hook rewrites bad `.go`, leaves `.ts`, exit 0 | B2 `TestGofmtHook` PASS | verified at 11c5c13: `app/cmd/agenthooks/main_test.go:35` - `require.Equal(t, 0, gofmtHook(...))`; `:38` - `require.Equal(t, "package x\n\nfunc F() {\n\treturn\n}\n", string(got))`; `:43` - `require.Equal(t, tsSrc, string(got))` | PASS |
| C51 | Stop: failing -> 2 + stderr; passing -> 0; active -> 0 | B2 `TestStopHook` + 3 subtests PASS | verified at 11c5c13: `main_test.go:69` - `require.Equal(t, 2, stopHook(...))`; `:70` - stderr contains the output; `:75` - 0 on pass; `:81-82` - 0 and empty stderr when active | PASS |
| C52 | settings wires PostToolUse matcher + Stop | B1 `TestClaudeSettings_WiresHooks` PASS | verified at 11c5c13: `harness_test.go:162` - `require.Equal(t, "Edit|Write|MultiEdit", post[0].Matcher)`; `:164` - `"./cmd/agenthooks gofmt"`; `:169` - `"./cmd/agenthooks stop"` | PASS |
| C53 | Cursor `alwaysApply: true` citing AGENTS.md; no `.windsurf/` | B1 `TestAgentRuleFiles` PASS | verified at 11c5c13: `harness_test.go:185` - `require.Equal(t, true, fm["alwaysApply"])`; `:186` - `require.Contains(t, body, "AGENTS.md")`; `:189` - `require.NoDirExists(t, filepath.Join(repoRoot, ".windsurf"))`. AC 39 was amended in `plan.md` by the fix (user approval 2026-10-08) and the check matches it. `.windsurf/` is absent at HEAD | PASS |
| C54 | CI runs `task check` and `task e2e`, no continue-on-error | B1 `TestCIWorkflow_RunsCheckAndE2E` PASS | verified at 11c5c13: `harness_test.go:207`, `:210` - `require.NotEqual(t, true, ...ContinueOnError)`; `:220` - `require.ElementsMatch(t, []string{"task check", "task e2e"}, found)`; `:221` - `require.True(t, sameJob, ...)` | PASS |
| C55 | `check:fast` = fmt, vet, archtest, typecheck; non-zero on failure | B1 `TestTaskfile_CheckFastSteps` PASS (+ C65) | verified at 11c5c13: `harness_test.go:99` - `require.Equal(t, []string{"fmt:check", "vet", "archtest", "web:typecheck"}, ...)`; `:100` - `go vet ./...`; exit propagation now `gencheck_test.go:102-103` (C65) | PASS |
| C56 | webui with no `index.html`: `/users/42` -> 404 problem+json | B2 `TestHandler_MissingBuildIs404Problem` PASS | verified at 11c5c13: `webui_test.go:51` - `require.Equal(t, http.StatusNotFound, rec.Code)`; `:52` - `require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))` | PASS |
| C57 | `instance` == request path for 404, 500, 503 | B2 `TestProblem_InstanceIsRequestPath` PASS | verified at 11c5c13: `httpx_test.go:157` - cases `{404: "/api/nope", 500: "/api/panic", 503: "/api/unavailable"}`; `:160` - `require.Equal(t, status, rec.Code, path)`; `:164` - `require.Equal(t, path, body["instance"], path)` | PASS |
| C58 | assembled `GET /api/docs` 200 `text/html` | B2 `TestRouting_DocsServed` PASS | verified at 11c5c13: `app_test.go:184` - `require.Equal(t, http.StatusOK, rec.Code)`; `:185` - `require.Contains(t, rec.Header().Get("Content-Type"), "text/html")` | PASS |
| C59 | empty `ID` -> error naming method and path | B2 `TestRegister_RejectsMissingID` PASS | verified at 11c5c13: `op_test.go:62` - `require.Error(t, err)`; `:63` - `require.Contains(t, err.Error(), http.MethodGet)`; `:64` - `require.Contains(t, err.Error(), "/things/{id}")` | PASS |
| C60 | `/readyz` with no DB -> 503 problem+json | B2 `TestReadyz_503WhenDatabaseNotConfigured` PASS | verified at 11c5c13: `health_test.go:98` - `require.Equal(t, http.StatusServiceUnavailable, rec.Code)`; `:99` - content type `application/problem+json`; `:104` - body `status` 503 | PASS |
| C61 | `ErrAbortHandler` re-panicked: no 500 body, no ERROR log | B2 `TestRecover_RepanicsAbortHandler` PASS | verified at 11c5c13: `httpx_test.go:183` - `require.Equal(t, http.ErrAbortHandler, got, ...)`; `:184` - `require.Empty(t, rec.Body.String(), ...)`; `:187` - `require.NotEqual(t, "ERROR", e["level"], ...)` | PASS |
| C62 | unknown subcommand exits 2 with usage on stderr | B2 `TestRun_UnknownCommandExits2` PASS | verified at 11c5c13: `main_test.go:174` - `require.Equal(t, 2, code)`; `:175` - `require.Contains(t, errb.String(), "usage: api serve | api migrate up | api openapi")` | PASS |
| C63 | `migrate up` on unreachable DB exits 1, `migrate up:` on stderr | B2 `TestMigrateUp_UnreachableDatabaseExits1` PASS | verified at 11c5c13: `main_test.go:181` - `require.Equal(t, 1, code, stderr)`; `:182` - `require.Contains(t, stderr, "migrate up:")` | PASS |
| C64 | gofmt hook: bad payload, missing file, unparseable `.go` each exit 0, stderr, files unchanged | B2 `TestGofmtHook_NeverBlocks` + 3 subtests PASS | verified at 11c5c13: `agenthooks/main_test.go:90-91` - exit 0 and `NotEmpty(stderr)`; `:98-103` - exit 0, stderr, `NoFileExists(missing)`, dir `Empty`; `:111-115` - exit 0, stderr, `require.Equal(t, src, got)` | PASS |
| C65 | temp copy with unformatted `.go`: `task check` and `check:fast` non-zero | B1 `TestTaskfile_GatesFailOnFailingStep` + 2 subtests PASS | verified at 11c5c13: `gencheck_test.go:102` - `require.ErrorAs(t, err, &exitErr, ...)`; `:103` - `require.NotEqual(t, 0, exitErr.ExitCode())`; `:104` - `require.Contains(t, string(out), "unformatted.go", ...)` | PASS |
| C66 | built `newslice` exits 1 for invalid name and existing slice | B3 `TestMain_FailuresExit1` + 2 subtests PASS | verified at 11c5c13: `newslice_test.go:159` - `require.ErrorAs(t, err, &exitErr, ...)`; `:160` - `require.Equal(t, 1, exitErr.ExitCode(), ...)` | PASS |
| C67 | hanging DB -> 503 once the configured timeout elapses | B2 `TestReadyz_503WhenDatabaseHangs` PASS | verified at 11c5c13: `health_test.go:89` - 503; `:90` - problem+json; `:91` - `require.GreaterOrEqual(t, elapsed, timeout, ...)`; `:92` - `require.Less(t, elapsed, 2*time.Second, ...)` | PASS |
| C68 | `POST` outside `/api/` -> 405 problem+json, never SPA | B2 `TestRouting_NonGetOutsideAPIIs405Problem` PASS | verified at 11c5c13: `app_test.go:194` - `require.Equal(t, http.StatusMethodNotAllowed, rec.Code, p)`; `:195` - content type `application/problem+json`; `:196` - `require.NotContains(t, rec.Body.String(), "<html>spa</html>", p)` over `/healthz`, `/users` | PASS |

## Round-1 ranked gaps re-judged

verified at 11c5c13

| # | Round-1 gap | Closing check | Assertion targets the check-defined value | Closed |
| --- | --- | --- | --- | --- |
| 1 | webui missing build 404 `text/plain` (AC 8) | C56 | `webui_test.go:52` content type equals `application/problem+json`; code `webui.go:45` uses `httpx.WriteProblem` | yes |
| 2 | door 3 `instance` never emitted | C57 | `httpx_test.go:164` `body["instance"]` equals the request path, over 404/500/503; code `problem.go:47-49`, `:62` | yes |
| 3 | door 2 `/api/docs` unproven | C58 | `app_test.go:184-185` 200 and `text/html` on the assembled server | yes |
| 4 | op empty-ID, readyz nil-DB, `ErrAbortHandler` | C59, C60, C61 | `op_test.go:63-64`; `health_test.go:98-99`; `httpx_test.go:183-187` | yes |
| 5 | `cmd/api` unknown subcommand exit 2, runtime-error exit 1 | C62, C63 | `main_test.go:174`, `:181` | yes |
| 6 | agenthooks bad payload / unreadable / unparseable | C64 | `agenthooks/main_test.go:90-115` | yes |

Round-1 precision notes resolved by the fix: C55 exit propagation (C65), C9 sleep replaced by `waitServing`
(`main_test.go:122`), C37/C38 process exit (C66), AC 3 hang case named (C67), C54 same-job requirement
(`harness_test.go:221`), and non-GET to `/healthz` reaching the SPA (C68).

## Coverage

Rows whose authority the fix touched are recomputed at 11c5c13 from the code and the plan. The other rows are
carried from 14cada0.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| `webui` routing (6) - verified at 11c5c13 | `webui.go:31-50`, door 9, AC 35 | non-GET/HEAD -> 405 C68 (assembly only) · HEAD served `app_test.go:202` (C68) · static file C45 · SPA fallback C44 · `/` -> index `webui_test.go:43` (B8) · missing build -> 404 problem C56 | - (405 row lacks an own-layer case; see Test policy) |
| `GET /*` outside `/api/` statuses (3) - verified at 11c5c13 | plan Surface (200) + code `webui.go:33`, `:45` | 200 C44, C45 · 404 C56 · 405 C68 | - |
| problem+json fields, door 3 (7) - verified at 11c5c13 | plan Landing door 3; `problem.go:31-37`, `:47-49`, `:56-64` | `type`, `title`, `status`, `detail`, `request_id` C11 · `instance` C57 · `errors?` optional, no input-validated operation at HEAD | - |
| error-response emitters (6) - verified at 11c5c13 | AC 8; `problem.go:55-75`, `middleware.go:85-104`, Huma via `problem.go:18-39`, `webui.go:33`, `:45` | router 404 C11, C12, C57 · panic 500 C11, C13, C57 · Huma 503 C3, C11, C57, C60 · Huma 501 generated slice C36 · webui 404 C56 · webui 405 C68 | - |
| `/readyz` branches (4) - verified at 11c5c13 | `health.go:93-101` | ready C2 · DB down C3 · DB hangs C67 · DB nil C60 | - |
| Recover branches (3) - verified at 11c5c13 | `middleware.go:88-101` | no panic (every C11-C17 request) · panic -> 500 C13, C14 · `ErrAbortHandler` re-panicked C61 | - |
| `api` command outcomes (7) - verified at 11c5c13 | `cmd/api/main.go:41-69`, `:113-126` | unknown subcommand 2 C62 · config error 1 C4 · migrate ok 0 C7, C8 · serve ok 0 C9 · migrate/serve runtime error 1 (shared branch `:63-66`) C63 · `openapi` ok 0 (exercised by `task gen` inside C35/C36 `genErr`) · `openapi` failure 1 (`main.go:122-125`) | `openapi` failure exit 1 (`cmd/api/main.go:122-125`): no proof; checks.md row "api command exits (4)" omits the `openapi` subcommand |
| gofmt hook branches (6) - verified at 11c5c13 | `agenthooks/main.go:41-71` | `.go` reformatted C50 · non-`.go` untouched C50 · bad payload C64 · unreadable file C64 · unparseable `.go` C64 · rewrite fails (`main.go:66-68`, stderr, exit 0) | write-failure branch `agenthooks/main.go:66-68`: no asserted case |
| `agenthooks` process dispatch (4) - verified at 11c5c13 | `agenthooks/main.go:22-38` | `gofmt` dispatch and `stop` dispatch wired by C52 (strings only) · no argument exit 1 (`:23-26`) · unknown hook exit 1 (`:34-36`) | no-argument and unknown-hook exits: no proof |
| Stop hook branches (3) - verified at 11c5c13 | `agenthooks/main.go:75-90` | failing C51 · passing C51 · `stop_hook_active` C51 | - |
| gate failure exit (2) - verified at 11c5c13 | AC 23, AC 38; `Taskfile.yml` sequential `cmds` | `task check` C65 · `task check:fast` C65 | - |
| `newslice` process exits (3) - verified at 11c5c13 | `newslice/main.go:33-36` | invalid name 1 C66 · existing slice 1 C66 · success 0 (C35 `task new:slice` `genErr` NoError) | - |
| `op.Register` decision (5) - verified at 11c5c13 | `op.go:40-55` | empty ID C59 · no permission and not public C19 · permission and public `op_test.go:56` (B8) · mutation without audit C20 · valid C21 | - |
| Landing doors (13) - verified at 11c5c13 | plan Landing | 1 C31 · 2 C36, C58, C18 · 3 C11, C57 · 4 C15, C14 · 5 C19 · 6 C33 · 7 C23 · 8 C10, C9 · 9 C44, C48, C68 · 10 C32 · 11 C22 · 12 C36 · 13 C31 | - |
| agent rule files (1) - verified at 11c5c13 | AC 39 as amended | Cursor C53 · `.windsurf/` absent C53 | - |
| CI steps (2) - verified at 11c5c13 | AC 40 | `task check` C54 · `task e2e` C54, same job `harness_test.go:221` | - |
| `GET /healthz` (1), `GET /readyz` (2), `GET /api/openapi.json` (1), `ANY /api/*` (1) statuses | carried from 14cada0 | 200 C1 · 200 C2, 503 C3 · 200 C18 · 404 C12 | - |
| request log keys (8), `X-Request-ID` (2), mutating methods (4) | carried from 14cada0 | C17 · C15, C16 · C20 | - |
| startup assembly (3), shutdown paths (2), migration runs (3) | carried from 14cada0 | C22 · C5, C6 · C7, C8, C9 | - |
| import rules (4), generated-code diffs (3), `task check` steps (8), `check:fast` steps (4) | carried from 14cada0 | C23-C25 · C27-C29 · C30 · C55 | - |
| generator outcomes (6), generated files (3), invalid inputs (4) | carried from 14cada0 | C33-C38, `newslice_test.go:137-138` · C33 · C38 | - |
| screen `/` states (4), web routes (2), dev proxy prefixes (3) | carried from 14cada0 | C39-C42 · C39, C43 · C48 | - |

Sweep for new members the fix added:

- The `webui` method guard (`webui.go:31-35`) adds 405 and HEAD members, covered by C68 at the assembly.
- `ProblemTransformer`'s keep-existing-`instance` branch (`problem.go:47`) has no reachable producer at HEAD:
  `InstallProblems` never sets `Instance`, so it is not a member.
- The remaining production changes are lint-driven refactors with no new branch: `exec.CommandContext`,
  `ListenConfig`, `slog.DiscardHandler`, `SplitSeq` and `errors.AsType`.
- `app/.golangci.yml` adds 10 linters, which the gate exercises.

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `internal/platform/op/op.go` (verified at 11c5c13) | own layer C19-C21, C59, `op_test.go:56` · boundary C22 | yes - all 5 rows asserted |
| Decides, reached across a boundary | `internal/platform/webui/webui.go` (touched; verified at 11c5c13) | own layer C44, C45, C56 · boundary C46, C68 | not met - the new 405 row (`webui.go:31-35`) has its case only at the assembled server (`app_test.go:194`); `webui_test.go` has none |
| Decides, reached across a boundary | `internal/platform/health/health.go` `/readyz` (verified at 11c5c13) | own layer and HTTP C2, C3, C60, C67 | yes - all 4 rows asserted |
| Decides, reached across a boundary | `internal/platform/httpx/middleware.go`, `problem.go` (touched; verified at 11c5c13) | own layer C11, C13-C17, C57, C61 · boundary C12 | yes - every row asserted |
| Decides, reached across a boundary | `cmd/newslice/main.go` (touched; verified at 11c5c13) | own layer C33, C34, C37, C38 · boundary C35, C36, C66 | yes |
| Decides, not reached across a boundary | `archtest/archtest.go` (carried from 14cada0) | own layer C23-C26, C10 | yes |
| Decides, not reached across a boundary | `cmd/agenthooks/main.go` (touched; verified at 11c5c13) | own layer C50, C51, C64 | not met - write-failure row (`main.go:66-68`), no-argument exit (`:23-26`) and unknown-hook exit (`:34-36`) lack an asserted case |
| Entry point that decides nothing | `health.go` `/healthz` (carried from 14cada0) | boundary C1 | yes |
| Entry point that decides nothing | `cmd/api/main.go` `run` (touched; verified at 11c5c13) | boundary C4, C7-C9, C62, C63 | not met - error path `api openapi` failure exit 1 (`main.go:122-125`) has no proof |
| Instrumentation, pass-throughs | `platform/db`, `platform/deps`, `platform/testkit/*` (touched: `log.go`), `cmd/fmtcheck`, `cmd/sqlcrun` (touched), `cmd/webcopy`, `platform/config` | none of its own | yes - covered by consumers (C2/C3, C4, C6, C27, C65, gate) |

## Faults injected

verified at 11c5c13. Isolation used `git worktree add --detach <scratchpad>/wt HEAD`. The real-tree
`git status --porcelain` baseline was `?? .claude/skills/auth-security/` and `?? .cursor/skills/auth-security/`.
Each fault was applied in the worktree and reverted with `git checkout -- <file>` inside it before the next one,
and `git stash` was never used. Each narrowest proof failed on its assertion, not on compilation.

Cleanup used `git worktree remove --force` and `git worktree prune`. `git worktree list` then showed only the main
tree, and the real-tree porcelain matched the baseline exactly (`diff` empty).

| Mutation | Location | Killed |
| --- | --- | --- |
| method guard also lets `POST` through to the SPA | `app/internal/platform/webui/webui.go:31` | yes - `TestRouting_NonGetOutsideAPIIs405Problem` FAIL (`expected: 405 actual: 200`) |
| missing-build path back to `http.Error` (`text/plain`) | `app/internal/platform/webui/webui.go:45` | yes - `TestHandler_MissingBuildIs404Problem` FAIL (`expected: "application/problem+json" actual: "text/plain; charset=utf-8"`) |
| `ProblemTransformer` stops stamping `Instance` | `app/internal/platform/httpx/problem.go:48` | yes - `TestProblem_InstanceIsRequestPath` FAIL (`expected: "/api/unavailable" actual: <nil>`) |
| `Recover` swallows `http.ErrAbortHandler` instead of re-panicking | `app/internal/platform/httpx/middleware.go:93` | yes - `TestRecover_RepanicsAbortHandler` FAIL (`expected: ...abort Handler actual: <nil>`) |
| unknown subcommand returns `1` instead of `2` | `app/cmd/api/main.go:48` | yes - `TestRun_UnknownCommandExits2` FAIL (`expected: 2 actual: 1`) |

The cap of 5 was reached. The three production surfaces the fix created (`webui.go:31`, `webui.go:45`,
`problem.go:48`) were each made to fail once. Two of the test-only surfaces it added (C61, C62) were also made to
fail once. C59, C60, C63, C64, C65, C66 and C67 were not mutated in this round.

## Swept existing re-read

verified at 11c5c13. Every constraint cited is still present:

- **validation** (C38, C4): `newslice/main.go:25` holds the regex, checked at `:45-50` before any write.
- **failure modes** (C13, C14, C3): `middleware.go:85-104`; `health.go:98-99`.
- **idempotency** (C8, C37): `cmd/api/main.go:106-111`; `newslice/main.go:57-59`.
- **authorization** (C19, C20, C22): `op.go:40-55`.
- **concurrency** (C5, C9): `app.go` Run/Shutdown; `serve` has no goose call (`main.go:79-98`).
- **dependency failure** (C3, C40) and **observability** (C14, C16, C17): unchanged.
- **data lifecycle** and **state transitions** are `n/a` (user-approved).

## Precision notes

These do not fail the feature on their own:

- **C39 is timing-sensitive under load.** This is raised to ranked gap 4.
- **C27's committed-tree half still passes vacuously.** `sqlc.yaml` has `sql: []`, so the output is
  `sqlcrun: no sqlc packages configured`. Real drift is exercised only in the temp copy (`gencheck_test.go:70`).
- **C47's scanner exclusions are broader than the claim.** They also cover `src/test/` and `*.test.ts(x)`; this is
  carried from 14cada0.

## Ranked gaps

1. **Entry-point error path unproven: `api openapi` failure exits 1 with no proof.** The `run` dispatch reaches
   `exportOpenAPI`, which returns 1 when `app.New`/`app.OpenAPI` fails or stdout fails. That error path backs
   `task gen:openapi`. checks.md's "`api` command exits (4)" row omits the `openapi` subcommand entirely. -
   Test policy, `cmd/api` row - `app/cmd/api/main.go:122-125`
2. **Test policy, webui own-layer row not met.** The fix's new 405 decision (`webui.go:31-35`) is asserted only
   through the assembled server (C68, `app_test.go:194`). `webui_test.go` has no case for the 405 row. - C68 -
   `app/internal/platform/webui/webui.go:31`
3. **Test policy, agenthooks rows not met.**
   - The gofmt write-failure branch (stderr, exit 0) has no asserted case.
   - The process dispatch exits have no proof: no argument (exit 1) and unknown hook (exit 1).

   - C64, C52 - `app/cmd/agenthooks/main.go:66-68`, `:23-26`, `:34-36`
4. **Flaky proof that leaks into the gate.**
   - The C39 assertion `findByText("API: online")` uses Testing Library's 1s default timeout.
   - It failed 2 times in this round under concurrent load: once in B5, and once inside C35's nested `task check`,
     which turned C35 red.
   - It passed in isolation both times.
   - Because `web:test` is a `task check` step, the gate itself can go red on an unchanged commit when the
     machine is busy, for example a loaded CI runner.
   - C39, C35 - `web/src/features/status/ApiStatus.test.tsx:21`

## Gate

verified at 11c5c13. `task check` was run from the repo root and exited 0 in 590s:

- `fmt:check`, `lint` (golangci-lint with the 10 linters the fix added), `gen:sqlc:check`, `gen:openapi:check`,
  `gen:web:check` and `archtest` (`ok ... 128.354s`) all passed.
- `go test -count=1 ./...` reported 10 packages `ok` and 0 failed, `cmd/newslice` included (394.984s).
- `web:typecheck` was clean. `web:lint` reported `Checked 28 files ... No fixes applied.`
- `web:test` reported 5 files and 10 passed, 0 failed.

`task e2e -- -g "api online"` exited 0 with `✓ [chromium] › e2e\status.spec.ts:4:1 › api online` (1 passed).

The gate is green at this commit, but its `web:test` step carries the flake described in ranked gap 4.
