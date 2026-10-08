# Foundation verification

**Verdict**: PASS
**Profile**: standard
**Diff range**: a927e8b..b5b76c1 (HEAD); fix under review 071b520..b5b76c1
**Round**: 4 - scoped (a 4th round beyond the 3-round bound, authorized by the user to close round 3's single gap)
**Verifier**: independent sub-agent (author != verifier)

All 75 checks are proven at HEAD b5b76c1. Each has a located assertion, and every named test was re-run in full at
b5b76c1 and appears individually as PASS. Round 3's only ranked gap (the `gofmt` and `stop` dispatch arms of
`cmd/agenthooks`) is closed by C74 and C75:

- C74 runs the built binary with `gofmt` and asserts the file's bytes equal the gofmt output, and that the exit is 0.
- C75 runs the built binary with `stop` against a stub `task` that exits 3. It asserts exit 2 and that the stub
  received exactly `check:fast`.

All 4 faults injected on the new surfaces were killed. The coverage recompute of `app/cmd/agenthooks/main.go` finds
every dispatch arm, gofmt branch and Stop branch proven. The `cmd/agenthooks` Test policy row, the only row unmet
in round 3, is now met. `task check` exits 0 at b5b76c1, and `task e2e -- -g "api online"` passes.

The fix (`git diff --stat 071b520..b5b76c1 -- app web`) touches one file, `app/cmd/agenthooks/main_test.go`. It
extracts `buildAgenthooks` from C72 and appends C74 and C75. No production code changed.

## Binding sources

carried from 071b520: none. The plan marks no source as binding, and the fix touched no UI or interface, so step 1
does not apply.

## Proof runs (all at HEAD b5b76c1, verified at b5b76c1)

Each named test appears individually in the `-v` / `--reporter=verbose` output as `--- PASS` / `✓`.

- **Names exist.** `rg -n "^func (<68 names>)\("` over `app/` returned 68 hits, including
  `app/cmd/agenthooks/main_test.go:163` `TestMain_GofmtArmFormats` and `:179` `TestMain_StopArmRunsTaskCheckFast`.
  `rg -n "^\s*(it|test)\("` over `web/src`, `web/vite.config.test.ts` and `web/e2e` found all 10 vitest names and
  the e2e `api online` (`web/e2e/status.spec.ts:4`).
- **Load.** Every Go batch ran while `task check` was running concurrently.

Batches:

- **B1** `go -C app test ./archtest -run '^(<17 names>)$' -v -count=1 -timeout 30m`: exit 0, `ok ... 96.541s`.
  17/17 `--- PASS`, including subtests `TestTaskfile_GatesFailOnFailingStep/check` and `/check:fast`.
- **B2** `go -C app test ./internal/platform/{health,config,httpx,op,webui} ./internal/app ./cmd/api ./cmd/agenthooks -run '^(<44 names>)$' -v -count=1`:
  exit 0. 44/44 top-level `--- PASS`, 8 packages `ok`.
  - New in this round: `TestMain_GofmtArmFormats (12.72s)` and `TestMain_StopArmRunsTaskCheckFast (9.14s)`.
  - `TestMain_DispatchFailuresExit1` also passed after its refactor, with subtests `/no_argument` and `/unknown_hook`.
  - Also passed: the 4 POST/PUT/PATCH/DELETE subtests, 3 `TestStopHook` subtests and 3 `NeverBlocks` subtests.
- **B3** `go -C app test ./cmd/newslice -run '^(<7 names>)$' -v -count=1 -timeout 30m`: exit 0, `ok ... 473.084s`.
  7/7 `--- PASS`, with 4 invalid-name subtests and 2 `TestMain_FailuresExit1` subtests.
  `TestGenerated_TaskCheckPasses` passed first time (437.20s) under concurrent load.
- **B5** `npm --prefix web run test -- <6 files> --reporter=verbose -t "online|offline|retry|pending|not found|drift fails|only generated client|proxies backend paths|async timeout"`:
  exit 0. 10 passed. 1 was skipped (`committed schema passes`), which is outside the filter and covered by B6.
- **B6** `task gen:sqlc:check` exit 0 (`sqlcrun: no sqlc packages configured; nothing to do`).
  `task gen:openapi:check` exit 0 (`ok ... 1.109s`). `npm --prefix web run gen:check` exit 0.
- **B7** `task e2e -- -g "api online"`: exit 0, `✓ 1 [chromium] › e2e\status.spec.ts:4:1 › api online (25.4s)`,
  1 passed.

Every proof file is in the diff `a927e8b..HEAD`. The two new proofs (C74, C75) are in `071b520..b5b76c1`.

## Checks

The two citations that need a pipe character write it as `¦`, because a table cell cannot hold a literal pipe.

Rows citing `app/cmd/agenthooks/main_test.go`, the only source file the fix touched, are marked
`verified at b5b76c1`:

- The fix starts at line 136. C50, C51, C64 and C71 cite lines above it, and those lines were re-read unchanged.
- C72's lines moved. C74 and C75 are new.

All other rows are marked `carried from 071b520`. Their citations are unchanged, because their files are not in
`071b520..b5b76c1`, and their proofs were re-run at b5b76c1 above.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `/healthz` 200 `{"status":"ok"}` without DB | B2 `TestHealthz_OKWithoutDatabase` PASS | carried from 071b520: `app/internal/platform/health/health_test.go:44` - `require.Equal(t, http.StatusOK, rec.Code)`; `:45` - `require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())` | PASS |
| C2 | `/readyz` 200 `{"status":"ready"}` on live PG | B2 `TestReadyz_ReadyWhenDatabaseAnswers` PASS | carried from 071b520: `health_test.go:60` - `require.Equal(t, http.StatusOK, rec.Code)`; `:61` - `require.JSONEq(t, `{"status":"ready"}`, ...)` | PASS |
| C3 | `/readyz` 503 problem+json within 3s, DB stopped | B2 `TestReadyz_503WhenDatabaseDown` PASS | carried from 071b520: `health_test.go:72` - `require.Less(t, time.Since(start), 3*time.Second)`; `:73` - `require.Equal(t, http.StatusServiceUnavailable, rec.Code)`; `:74` - content type `application/problem+json` | PASS |
| C4 | built binary, `DATABASE_URL` unset: exit 1, stderr names it | B2 `TestServe_MissingDatabaseURLExits1` PASS | carried from 071b520: `app/cmd/api/main_test.go:49` - `require.Equal(t, 1, exitErr.ExitCode())`; `:50` - `require.Contains(t, stderr.String(), "DATABASE_URL")` | PASS |
| C5 | cancel drains in-flight 200, refuses new conn, nil | B2 `TestRun_ShutdownDrainsInFlight` PASS | carried from 071b520: `app/internal/app/app_test.go:130-138` - `require.Eventually(... DialContext err != nil ...)`; `:141` - `require.Equal(t, http.StatusOK, <-status)`; `:144` - `require.NoError(t, err)` | PASS |
| C6 | default shutdown 10s; blocked handler does not hold `Run` | B2 `TestDefaults_ShutdownTimeoutIs10s`, `TestRun_ShutdownTimeoutBoundsDrain` PASS | carried from 071b520: `app/internal/platform/config/config_test.go:16` - `require.Equal(t, 10*time.Second, cfg.ShutdownTimeout)`; `app_test.go:175` - `require.Less(t, time.Since(start), 2*time.Second)` | PASS |
| C7 | `migrate up` applies all, version = newest, exit 0 | B2 `TestMigrateUp_AppliesPending` PASS | carried from 071b520: `main_test.go:88` - `require.Equal(t, 0, code, stderr)`; `:89` - `require.Equal(t, newestMigrationVersion(t), gooseVersion(t, pg.URL))` | PASS |
| C8 | rerun applies nothing, exit 0 | B2 `TestMigrateUp_RerunIsNoop` PASS | carried from 071b520: `main_test.go:106` - `require.Equal(t, 0, code, stderr)`; `:107` - `require.Equal(t, before, gooseVersion(t, pg.URL))`; `:110` - `require.Equal(t, rowsBefore, rowsAfter)` | PASS |
| C9 | `serve` creates no `goose_db_version` | B2 `TestServe_DoesNotMigrate` PASS | carried from 071b520: `main_test.go:123` - `waitServing(t, "http://"+addr+"/healthz", done)`; `:125` - `require.Equal(t, 0, <-done)`; `:132` - `require.False(t, exists)` | PASS |
| C10 | bad migration name fails naming file; real dir clean | B1 `TestMigrationNames_RejectsNonTimestamp`, `TestMigrationNames_RepositoryIsClean` PASS | carried from 071b520: `app/archtest/imports_test.go:48` - `require.Len(t, bad, 1)`; `:49` - `require.Contains(t, bad[0], "001_bad.sql")`; `:56` - `require.Empty(t, bad)` | PASS |
| C11 | 404/500/503 problem+json, 5 fields, `request_id` == header | B2 `TestProblem_HasRequiredFields` PASS | carried from 071b520: `app/internal/platform/httpx/httpx_test.go:79` - content type; `:83` - `require.Contains(t, body, field, ...)` over the 5 fields; `:85` - status; `:86` - `require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), body["request_id"], path)` | PASS |
| C12 | `GET /api/does-not-exist` 404 problem+json | B2 `TestRouting_UnknownAPIPathIs404Problem` PASS | carried from 071b520: `app_test.go:44` - `require.Equal(t, http.StatusNotFound, rec.Code)`; `:45` - content type `application/problem+json` | PASS |
| C13 | panic -> 500, detail free of `goroutine` / `.go:` | B2 `TestRecover_500WithoutStack` PASS | carried from 071b520: `httpx_test.go:95` - 500; `:100` - `require.NotContains(t, detail, "goroutine")`; `:101` - `require.NotContains(t, detail, ".go:")` | PASS |
| C14 | panic -> exactly one ERROR log, same `request_id` | B2 `TestRecover_LogsErrorWithRequestID` PASS | carried from 071b520: `httpx_test.go:115` - `require.Len(t, errs, 1)`; `:116` - `require.Equal(t, rec.Header().Get(httpx.HeaderRequestID), errs[0]["request_id"])` | PASS |
| C15 | `X-Request-ID: abc-123` echoed | B2 `TestRequestID_EchoesIncoming` PASS | carried from 071b520: `httpx_test.go:123` - `require.Equal(t, "abc-123", rec.Header().Get(httpx.HeaderRequestID))` | PASS |
| C16 | absent header -> UUID, distinct across 2 | B2 `TestRequestID_GeneratesUUID` PASS | carried from 071b520: `httpx_test.go:132`, `:134` - `uuid.Parse` NoError; `:135` - `require.NotEqual(t, a, b)` | PASS |
| C17 | one JSON log line with all 8 keys | B2 `TestAccessLog_HasAllKeys` PASS | carried from 071b520: `httpx_test.go:143` - `require.Len(t, entries, 1)`; `:146` - `require.Contains(t, e, k)` over the 8 keys | PASS |
| C18 | `/api/openapi.json` 200, byte-identical, `openapi` 3.1 | B2 `TestOpenAPI_ServedMatchesCommitted` PASS | carried from 071b520: `app_test.go:61` - 200; `:64` - `require.Equal(t, string(committed), rec.Body.String(), ...)`; `:69` - `HasPrefix(doc.OpenAPI, "3.1")` | PASS |
| C19 | no permission + not public -> error naming op ID | B2 `TestRegister_RejectsMissingPermission` PASS | carried from 071b520: `app/internal/platform/op/op_test.go:28` - `require.Error(t, err)`; `:29` - `require.Contains(t, err.Error(), "list-things")` | PASS |
| C20 | POST/PUT/PATCH/DELETE without audit -> error naming ID | B2 `TestRegister_RejectsMutationWithoutAudit` + 4 subtests PASS | carried from 071b520: `op_test.go:37` - `require.Error(t, err)`; `:38` - `require.Contains(t, err.Error(), "mutate-"+m)` | PASS |
| C21 | 3 valid specs register | B2 `TestRegister_AcceptsValidSpecs` PASS | carried from 071b520: `op_test.go:46-48` - `require.NoError(t, op.Register(...))` x3; `:49-51` - `require.NotNil(t, api.OpenAPI().Paths[...])` | PASS |
| C22 | `app.New` fails naming op ID; `cmd/api` uses `app.New` | B2 `TestNew_FailsOnInvalidOperation`, B1 `TestCompositionRoot_CmdAPIUsesAppNew` PASS | carried from 071b520: `app_test.go:80` - `require.Error(t, err)`; `:81` - `require.Contains(t, err.Error(), "no-permission-op")`; `app/archtest/composition_test.go:16` - imports `.../internal/app`; `:18-19` - no `/internal/features`, no `huma/v2` | PASS |
| C23 | `features/a` -> `features/b` reported with both paths | B1 `TestImports_RejectsCrossFeature` PASS | carried from 071b520: `imports_test.go:16` - `require.Len(t, v, 1)`; `:17-18` - both paths contained | PASS |
| C24 | `platform/x` -> `features/a` reported with both paths | B1 `TestImports_RejectsPlatformToFeature` PASS | carried from 071b520: `imports_test.go:25` - `require.Len(t, v, 1)`; `:26-27` - both paths contained | PASS |
| C25 | same-feature and composition root allowed | B1 `TestImports_AllowsSameFeatureAndCompositionRoot` PASS | carried from 071b520: `imports_test.go:34` - `require.Empty(t, v)` | PASS |
| C26 | real `app` module clean | B1 `TestImports_RepositoryIsClean` PASS | carried from 071b520: `imports_test.go:41` - `require.Empty(t, v)` | PASS |
| C27 | sqlc drift fails in copy; committed passes | B1 `TestGenCheck_SqlcDriftFails` PASS; B6 `task gen:sqlc:check` exit 0 | carried from 071b520: `app/archtest/gencheck_test.go:62` - `require.NoError(t, err, "freshly generated code must pass: %s", out)`; `:70` - `require.Error(t, err, "edited generated code must fail: %s", out)` | PASS |
| C28 | openapi.json drift fails in copy; committed passes | B1 `TestGenCheck_OpenAPIDriftFails` PASS; B6 `task gen:openapi:check` exit 0 | carried from 071b520: `gencheck_test.go:82` - `require.Error(t, err, "edited openapi.json must fail the check: %s", out)` | PASS |
| C29 | schema.d.ts drift fails; committed passes | B5 `gen:check > drift fails` ✓; B6 `npm run gen:check` exit 0 | carried from 071b520: `web/src/api/genCheck.test.ts:20` - `expect(run(spec, edited).status).not.toBe(0)` | PASS |
| C30 | `check` runs the 8 steps as sequential cmds | B1 `TestTaskfile_CheckRunsAllSteps` PASS | carried from 071b520: `app/archtest/harness_test.go:79` - `require.Equal(t, []string{"fmt:check",...,"web:test"}, ...)`; `:83-93` - each step's command | PASS |
| C31 | module path; tools module with 3 `tool`s; no root go.mod | B1 `TestModuleLayout` PASS | carried from 071b520: `harness_test.go:105` - module prefix; `:112` - `require.Contains(t, tools, "\t"+tool+"\n", ...)`; `:114` - `require.NoFileExists(t, filepath.Join(repoRoot, "go.mod"))` | PASS |
| C32 | Go and web dependencies declared | B1 `TestDependencies_Declared` PASS | carried from 071b520: `harness_test.go:125` - `require.Contains(t, gomod, "\t"+mod, ...)`; `:141` - `require.Contains(t, all, name)`; `:143` react `19.`; `:144` tailwindcss `4.` | PASS |
| C33 | slice files, register.go entry, one sqlc entry | B3 `TestGenerate_CreatesSliceAndRegisters` PASS | carried from 071b520: `app/cmd/newslice/newslice_test.go:77` - `require.FileExists` x3; `:81` - `require.Contains(t, register, "createuser.Register(api, d),")`; `:83` - `require.Equal(t, entriesBefore+1, strings.Count(..., "- engine:"))` | PASS |
| C34 | new feature: register.go + registry entry | B3 `TestGenerate_NewFeatureRegistersFeature` PASS | carried from 071b520: `newslice_test.go:95` - `"package billing"`; `:98-99` - registry imports `.../features/billing` and `billing.Register(api, d),` | PASS |
| C35 | after `task new:slice`, `task check` exits 0 | B3 `TestGenerated_TaskCheckPasses` PASS first time under load | carried from 071b520: `app/cmd/newslice/repo_test.go:130` - `require.NoError(t, err, out)` on `task check` in the copy | PASS |
| C36 | generated endpoint at `/api/v1/demo/`, 501 problem+json, generated test passes | B3 `TestGenerated_SliceAnswers501` PASS | carried from 071b520: `repo_test.go:150` - `require.Equal(t, []string{"/api/v1/demo/get-thing"}, demo)`; `:158` - test contains `http.StatusNotImplemented`; `:162` - `Contains(out, "--- PASS: TestEndpoint_NotImplemented")` | PASS |
| C37 | existing slice: non-zero, every file hash unchanged | B3 `TestGenerate_RefusesExistingSlice` PASS | carried from 071b520: `newslice_test.go:108` - `require.ErrorContains(t, err, "already exists")`; `:109` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C38 | 4 invalid inputs: non-zero, no file | B3 `TestGenerate_RejectsInvalidNames` + 4 subtests PASS | carried from 071b520: `newslice_test.go:125` - `require.Error(t, err)`; `:126` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C39 | readyz 200 -> `API: online` | B5 `shows online when readyz answers 200` ✓ (first run, under load) | carried from 071b520: `web/src/features/status/ApiStatus.test.tsx:21` - `expect(await screen.findByText("API: online")).toBeInTheDocument()`; timeout now from `web/src/test/setup.ts:6` (C73) | PASS |
| C40 | 503 and network error -> `API: offline` + retry button | B5 both offline tests ✓ | carried from 071b520: `ApiStatus.test.tsx:29`, `:37` - `findByText("API: offline")`; `:30`, `:38` - `getByRole("button", { name: "Tentar novamente" })` | PASS |
| C41 | retry issues a second readyz (1 -> 2) | B5 `retry requests readyz again` ✓ | carried from 071b520: `ApiStatus.test.tsx:22` - `toHaveLength(1)`; `:47` - `expect(readyz(calls)).toHaveLength(2)` | PASS |
| C42 | pending shows `role="status"`, gone after | B5 `shows a status element while readyz is pending` ✓ | carried from 071b520: `ApiStatus.test.tsx:60` - `findByRole("status")`; `:68` - `expect(screen.queryByRole("status")).not.toBeInTheDocument()` | PASS |
| C43 | `/nao-existe` -> `Página não encontrada` + link `/` | B5 `not found page links back to /` ✓ | carried from 071b520: `web/src/routes/notFound.test.tsx:8` - `findByText("Página não encontrada")`; `:9` - `toHaveAttribute("href", "/")` | PASS |
| C44 | webui `/users/42` -> 200 index bytes | B2 `TestHandler_FallsBackToIndex` PASS | carried from 071b520: `app/internal/platform/webui/webui_test.go:29` - `require.Equal(t, http.StatusOK, rec.Code)`; `:30` - `require.Equal(t, "<html>index</html>", rec.Body.String())` | PASS |
| C45 | webui `/assets/app.js` -> its own bytes | B2 `TestHandler_ServesStaticFile` PASS | carried from 071b520: `webui_test.go:36` - 200; `:37` - `require.Equal(t, "console.log('app')", rec.Body.String())` | PASS |
| C46 | assembled server: `/api/x`, `/healthz`, `/readyz` never SPA | B2 `TestRouting_APIAndHealthNotSwallowedBySPA` PASS | carried from 071b520: `app_test.go:53` - `require.NotContains(t, rec.Body.String(), "<html>spa</html>", p)`; `:55` - control path gets the SPA | PASS |
| C47 | no `fetch(`/axios/XHR outside `web/src/api/` | B5 `only generated client calls the backend` ✓ | carried from 071b520: `web/src/api/boundary.test.ts:22` - `expect(offenders).toEqual([])` | PASS |
| C48 | Vite proxies 3 prefixes to `:8080` | B5 `proxies backend paths to the api server` ✓ | carried from 071b520: `web/vite.config.test.ts:8` - `expect(proxy[path]).toBe("http://localhost:8080")` | PASS |
| C49 | Playwright on running stack sees `API: online` | B7 `api online` ✓ | carried from 071b520: `web/e2e/status.spec.ts:6` - `await expect(page.getByText("API: online")).toBeVisible()` | PASS |
| C50 | gofmt hook rewrites bad `.go`, leaves `.ts`, exit 0 | B2 `TestGofmtHook` PASS | verified at b5b76c1 (lines unchanged): `app/cmd/agenthooks/main_test.go:37` - `require.Equal(t, 0, gofmtHook(...))`; `:40` - `require.Equal(t, "package x\n\nfunc F() {\n\treturn\n}\n", string(got))`; `:42`, `:45` - exit 0 and `require.Equal(t, tsSrc, string(got))` | PASS |
| C51 | Stop: failing -> 2 + stderr; passing -> 0; active -> 0 | B2 `TestStopHook` + 3 subtests PASS | verified at b5b76c1 (lines unchanged): `main_test.go:71` - `require.Equal(t, 2, stopHook(...))`; `:72` - stderr contains the output; `:77` - 0 on pass; `:83-84` - 0 and empty stderr when active. The command is always `helperCmd(...)`; the literal `task check:fast` at the dispatch layer is now C75 | PASS |
| C52 | settings wires PostToolUse matcher + Stop | B1 `TestClaudeSettings_WiresHooks` PASS | carried from 071b520: `harness_test.go:162` - `require.Equal(t, "Edit¦Write¦MultiEdit", post[0].Matcher)`; `:164` - `"./cmd/agenthooks gofmt"`; `:169` - `"./cmd/agenthooks stop"` | PASS |
| C53 | Cursor `alwaysApply: true` citing AGENTS.md; no `.windsurf/` | B1 `TestAgentRuleFiles` PASS | carried from 071b520: `harness_test.go:185` - `require.Equal(t, true, fm["alwaysApply"])`; `:186` - `require.Contains(t, body, "AGENTS.md")`; `:189` - `require.NoDirExists(t, filepath.Join(repoRoot, ".windsurf"))` | PASS |
| C54 | CI runs `task check` and `task e2e`, no continue-on-error | B1 `TestCIWorkflow_RunsCheckAndE2E` PASS | carried from 071b520: `harness_test.go:207`, `:210` - `require.NotEqual(t, true, ...ContinueOnError)`; `:220` - `require.ElementsMatch(t, []string{"task check", "task e2e"}, found)`; `:221` - `require.True(t, sameJob, ...)` | PASS |
| C55 | `check:fast` = fmt, vet, archtest, typecheck; non-zero on failure | B1 `TestTaskfile_CheckFastSteps` PASS (+ C65) | carried from 071b520: `harness_test.go:99` - `require.Equal(t, []string{"fmt:check", "vet", "archtest", "web:typecheck"}, ...)`; exit propagation `gencheck_test.go:102-103` (C65) | PASS |
| C56 | webui with no `index.html`: `/users/42` -> 404 problem+json | B2 `TestHandler_MissingBuildIs404Problem` PASS | carried from 071b520: `webui_test.go:51` - `require.Equal(t, http.StatusNotFound, rec.Code)`; `:52` - `require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))` | PASS |
| C57 | `instance` == request path for 404, 500, 503 | B2 `TestProblem_InstanceIsRequestPath` PASS | carried from 071b520: `httpx_test.go:157` - cases `{404: "/api/nope", 500: "/api/panic", 503: "/api/unavailable"}`; `:160` - status; `:164` - `require.Equal(t, path, body["instance"], path)` | PASS |
| C58 | assembled `GET /api/docs` 200 `text/html` | B2 `TestRouting_DocsServed` PASS | carried from 071b520: `app_test.go:184` - `require.Equal(t, http.StatusOK, rec.Code)`; `:185` - `require.Contains(t, rec.Header().Get("Content-Type"), "text/html")` | PASS |
| C59 | empty `ID` -> error naming method and path | B2 `TestRegister_RejectsMissingID` PASS | carried from 071b520: `op_test.go:62` - `require.Error(t, err)`; `:63` - `http.MethodGet`; `:64` - `"/things/{id}"` | PASS |
| C60 | `/readyz` with no DB -> 503 problem+json | B2 `TestReadyz_503WhenDatabaseNotConfigured` PASS | carried from 071b520: `health_test.go:98` - 503; `:99` - content type `application/problem+json`; `:104` - body `status` 503 | PASS |
| C61 | `ErrAbortHandler` re-panicked: no 500 body, no ERROR log | B2 `TestRecover_RepanicsAbortHandler` PASS | carried from 071b520: `httpx_test.go:183` - `require.Equal(t, http.ErrAbortHandler, got, ...)`; `:184` - `require.Empty(t, rec.Body.String(), ...)`; `:187` - `require.NotEqual(t, "ERROR", e["level"], ...)` | PASS |
| C62 | unknown subcommand exits 2 with usage on stderr | B2 `TestRun_UnknownCommandExits2` PASS | carried from 071b520: `main_test.go:175` - `require.Equal(t, 2, code)`; `:176` - `require.Contains(t, errb.String(), "usage: api serve ¦ api migrate up ¦ api openapi")` | PASS |
| C63 | `migrate up` on unreachable DB exits 1, `migrate up:` on stderr | B2 `TestMigrateUp_UnreachableDatabaseExits1` PASS | carried from 071b520: `main_test.go:182` - `require.Equal(t, 1, code, stderr)`; `:183` - `require.Contains(t, stderr, "migrate up:")` | PASS |
| C64 | gofmt hook: bad payload, missing file, unparseable `.go` each exit 0, stderr, files unchanged | B2 `TestGofmtHook_NeverBlocks` + 3 subtests PASS | verified at b5b76c1 (lines unchanged): `agenthooks/main_test.go:92-93` - exit 0 and `NotEmpty(stderr)`; `:100-105` - exit 0, stderr, `NoFileExists(missing)`, dir `Empty`; `:113-117` - exit 0, stderr, `require.Equal(t, src, got)` | PASS |
| C65 | temp copy with unformatted `.go`: `task check` and `check:fast` non-zero | B1 `TestTaskfile_GatesFailOnFailingStep` + 2 subtests PASS | carried from 071b520: `gencheck_test.go:102` - `require.ErrorAs(t, err, &exitErr, ...)`; `:103` - `require.NotEqual(t, 0, exitErr.ExitCode())`; `:104` - `require.Contains(t, string(out), "unformatted.go", ...)` | PASS |
| C66 | built `newslice` exits 1 for invalid name and existing slice | B3 `TestMain_FailuresExit1` + 2 subtests PASS | carried from 071b520: `newslice_test.go:159` - `require.ErrorAs(t, err, &exitErr, ...)`; `:160` - `require.Equal(t, 1, exitErr.ExitCode(), ...)` | PASS |
| C67 | hanging DB -> 503 once the configured timeout elapses | B2 `TestReadyz_503WhenDatabaseHangs` PASS | carried from 071b520: `health_test.go:89` - 503; `:90` - problem+json; `:91` - `require.GreaterOrEqual(t, elapsed, timeout, ...)`; `:92` - `require.Less(t, elapsed, 2*time.Second, ...)` | PASS |
| C68 | `POST` outside `/api/` -> 405 problem+json, never SPA | B2 `TestRouting_NonGetOutsideAPIIs405Problem` PASS | carried from 071b520: `app_test.go:194` - `require.Equal(t, http.StatusMethodNotAllowed, rec.Code, p)`; `:195` - content type `application/problem+json`; `:196` - `require.NotContains(t, rec.Body.String(), "<html>spa</html>", p)` | PASS |
| C69 | `api openapi`, stdout rejects writes: exit 1, `openapi` on stderr | B2 `TestOpenAPI_WriteFailureExits1` PASS | carried from 071b520: `app/cmd/api/main_test.go:186-188` - `failingWriter.Write` returns `errors.New("disk full")`; `:193` - `run(..., []string{"openapi"}, ..., failingWriter{}, &errb)`; `:194` - `require.Equal(t, 1, code)`; `:195` - `require.Contains(t, errb.String(), "openapi")` | PASS |
| C70 | own-layer webui `POST /users` -> 405, problem+json, `Allow: GET, HEAD` | B2 `TestHandler_NonGetIs405Problem` PASS | carried from 071b520: `webui_test.go:58` - `webui.Handler(site)` with `http.MethodPost, "/users"`; `:59` - `require.Equal(t, http.StatusMethodNotAllowed, rec.Code)`; `:60` - `require.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))`; `:61` - `require.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))` | PASS |
| C71 | read-only unformatted `.go`: exit 0, stderr, bytes unchanged | B2 `TestGofmtHook_WriteFailureNeverBlocks` PASS | verified at b5b76c1 (lines unchanged): `agenthooks/main_test.go:126` - `os.Chmod(goFile, 0o444)`; `:130` - `require.Equal(t, 0, gofmtHook(payload(t, edit(goFile)), &stderr))`; `:131` - `require.NotEmpty(t, stderr.String())`; `:134` - `require.Equal(t, src, got)` | PASS |
| C72 | built `agenthooks` exits 1 with no argument and with `nope` | B2 `TestMain_DispatchFailuresExit1` + `/no_argument`, `/unknown_hook` PASS | verified at b5b76c1: `app/cmd/agenthooks/main_test.go:139` - `bin := buildAgenthooks(t)` (`:151-160`, `go build -o bin .`); `:141` - cases `{"no argument": nil, "unknown hook": {"nope"}}`; `:145` - `require.ErrorAs(t, err, &exitErr)`; `:146` - `require.Equal(t, 1, exitErr.ExitCode())` | PASS |
| C73 | test setup sets `asyncUtilTimeout` to 5000 ms | B5 `async timeout is 5000 ms for every findBy query` ✓ | carried from 071b520: `web/src/test/setup.ts:6` - `configure({ asyncUtilTimeout: 5000 })`, loaded for every file by `web/vite.config.ts` `setupFiles`; `web/src/test/setup.test.ts:6` - `expect(getConfig().asyncUtilTimeout).toBe(5000)` | PASS |
| C74 | built `agenthooks gofmt`, PostToolUse payload on stdin naming a bad `.go`: file rewritten to gofmt output, exit 0 | B2 `TestMain_GofmtArmFormats` PASS | verified at b5b76c1: `app/cmd/agenthooks/main_test.go:164` - built binary; `:166` - writes `"package x\nfunc  F( ) {\nreturn}\n"`; `:168-169` - `exec.CommandContext(..., bin, "gofmt")` with `cmd.Stdin = payload(t, edit(goFile))`; `:171` - `require.NoError(t, err, string(out))` (nil error from `CombinedOutput` = exit 0); `:175` - `require.Equal(t, "package x\n\nfunc F() {\n\treturn\n}\n", string(got))` | PASS |
| C75 | built `agenthooks stop`, `{"stop_hook_active":false}`, stub `task` first on PATH exiting 3: invoked with exactly `check:fast`, exit 2 | B2 `TestMain_StopArmRunsTaskCheckFast` PASS | verified at b5b76c1: `main_test.go:183-189` - stub `task.bat` (Windows) / `task` (sh) writes its args to `args.txt` and exits `3`; `:191-193` - `bin stop`, stdin `{"stop_hook_active":false}`, `PATH=stubDir` + separator + `PATH`; `:196` - `require.ErrorAs(t, err, &exitErr)`; `:197` - `require.Equal(t, 2, exitErr.ExitCode())`; `:201` - `require.Equal(t, "check:fast", strings.TrimSpace(string(args)))` | PASS |

## Round-3 ranked gap re-judged

verified at b5b76c1

| # | Round-3 gap | Closing check | Assertion targets the check-defined value | Closed |
| --- | --- | --- | --- | --- |
| 1a | `gofmt` dispatch arm (`main.go:29-30`) unproven | C74 | yes. `main_test.go:168` runs the built binary with the literal argument `gofmt` and the payload on stdin. `:171` asserts a nil error from `CombinedOutput`, which means exit 0. `:175` asserts the file now holds exactly the gofmt output of the bad source written at `:166`. Two faults were killed: routing `gofmt` to `stopHook` (arm swap) and to a no-op `code = 0`. | yes |
| 1b | `stop` dispatch arm (`main.go:31-32`) and its literal `task check:fast` unproven | C75 | yes. `:191` runs the built binary with `stop`. `:193` puts a stub `task` first on `PATH`; the stub records its arguments and exits `3` (`:183-189`). `:197` asserts exit `2`. `:201` asserts the recorded arguments are exactly `check:fast`. Three faults were killed: the arm swap, `"check:fast"` -> `"check"`, and the arm discarding `stopHook`'s code. | yes |

## Coverage

The rows whose authority the fix touched, all in `app/cmd/agenthooks/main.go`, are recomputed at b5b76c1 from the
code. The other rows are carried from 071b520, because the fix changed no production file.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| `agenthooks` dispatch arms (4) - verified at b5b76c1 | `cmd/agenthooks/main.go:22-38` | no argument exit 1 (`:23-26`) C72 · unknown hook exit 1 (`:33-35`) C72 · `gofmt` -> `gofmtHook` (`:29-30`) C74 · `stop` -> `stopHook` with `[]string{"task", "check:fast"}` (`:31-32`), exit passed through `:37` C75 | - |
| gofmt hook branches (6) - verified at b5b76c1 | `cmd/agenthooks/main.go:42-72` | `.go` reformatted (`:66-67`) C50 (own layer), C74 (binary) · non-`.go` untouched (`:53-55`) C50 · bad payload (`:48-51`) C64 · unreadable file (`:56-60`) C64 · unparseable `.go` (`:61-65`) C64 · rewrite fails (`:67-69`) C71 | - |
| Stop hook branches (3) - verified at b5b76c1 | `cmd/agenthooks/main.go:76-90` | `stop_hook_active` -> 0 without running (`:81-83`) C51 · command fails -> 2 + output on stderr (`:86-88`) C51, C75 (binary, exit 2) · command passes -> 0 (`:90`) C51 | - |
| `api` command outcomes (8), `webui` routing (6) - carried from 071b520 | `cmd/api/main.go:41-69`, `:114-126`; `webui/webui.go:30-51` | C62, C4, C7, C8, C9, C63, C69, `openapi` ok via C35/C36 · C70, C68, `app_test.go:202`, C45, C44, `webui_test.go:42-43`, C56 | - |
| `/readyz` branches (4), Recover branches (3), problem+json fields (7), error-response emitters (6) - carried from 071b520 | `health.go:93-101`; `middleware.go:85-104`; `problem.go` | C2, C3, C67, C60 · C11-C17, C13, C14, C61 · C11, C57 · C11-C13, C3, C57, C60, C36, C56, C68/C70 | - |
| `GET /*` outside `/api/` statuses (3) - carried from 071b520 | plan Surface + `webui.go:33`, `:45` | 200 C44, C45 · 404 C56 · 405 C68, C70 | - |
| gate failure exit (2), `newslice` process exits (3), `op.Register` decision (5) - carried from 071b520 | `Taskfile.yml`; `newslice/main.go:33-36`; `op.go:40-55` | C65 · C66, C35 · C59, C19, `op_test.go:56`, C20, C21 | - |
| Landing doors (13), agent rule files (1), CI steps (2) - carried from 071b520 | plan Landing; AC 39; AC 40 | 1 C31 · 2 C36, C58, C18 · 3 C11, C57 · 4 C15, C14 · 5 C19 · 6 C33 · 7 C23 · 8 C10, C9 · 9 C44, C48, C68, C70 · 10 C32 · 11 C22 · 12 C36 · 13 C31 · C53 · C54 | - |
| route statuses, request log keys (8), `X-Request-ID` (2), mutating methods (4), startup assembly, shutdown, migrations, import rules, generated-code diffs, `task check` steps, `check:fast` steps, generator outcomes/files/inputs, screen `/` states, web routes, dev proxy prefixes - carried from 071b520 | as round 2 | C1-C3, C12, C18 · C17 · C15, C16 · C20 · C22 · C5, C6 · C7-C9 · C23-C25 · C27-C29 · C30 · C55 · C33-C38 · C39-C43 · C48 | - |

Sweep of `app/cmd/agenthooks/main.go` for members still left out:

- **Exits.** `rg -n 'case "|os.Exit|return [0-9]' app/cmd/agenthooks/main.go` lists 2 `case` arms plus `default`,
  2 `os.Exit` sites (`:25`, `:37`) and 9 `return` sites (`:50`, `:54`, `:59`, `:64`, `:71`, `:82`, `:88`, `:90`).
  Each one maps to a member above.
- **Arm literals.** `rg -n 'bin, "gofmt"|bin, "stop"|"check:fast"' app --glob '*_test.go'` now hits
  `main_test.go:168`, `:191` and `:201`, the binary-level proofs round 3 found missing.
- **Already-formatted `.go`** (`bytes.Equal` true at `:66`): this path skips the write. Its observable outcome
  (bytes unchanged, exit 0) cannot be told apart from rewriting the same bytes, so it is not a separate member.
  Inverting the condition would leave C50's bad file unformatted and fail `main_test.go:40`.
- **Stop payload decode error** (`:80`, `_ =`): the error is discarded unconditionally, so there is no branch.
  An undecodable payload behaves like `stop_hook_active: false`.
- `checks.md`'s row is now "`agenthooks` dispatch arms (4)", sized by the switch's arms rather than its exits.

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, not reached across a boundary | `cmd/agenthooks/main.go` (test touched; verified at b5b76c1) | own layer C50, C51, C64, C71, C72, C74, C75 | yes - dispatch table 4/4 rows asserted (C72 x2, C74, C75), gofmt branches 6/6, Stop branches 3/3 |
| Decides, reached across a boundary | `internal/platform/webui/webui.go`, `internal/platform/op/op.go`, `health.go` `/readyz`, `httpx/*`, `cmd/newslice/main.go` (carried from 071b520) | own layer C44, C45, C56, C70 · C19-C21, C59 · C2, C3, C60, C67 · C11-C17, C57, C61 · C33-C38, C66; boundary C46, C68, C22, C35, C36 | yes |
| Decides, not reached across a boundary | `archtest/archtest.go` (carried from 071b520) | own layer C23-C26, C10 | yes |
| Entry point that decides nothing | `cmd/api/main.go` `run`, `health.go` `/healthz` (carried from 071b520) | boundary C4, C7-C9, C62, C63, C69 · C1 | yes |
| Instrumentation, pass-throughs | `web/src/test/setup.ts`, `platform/db`, `platform/deps`, `platform/testkit/*`, `cmd/fmtcheck`, `cmd/sqlcrun`, `cmd/webcopy`, `platform/config` (carried from 071b520) | none of its own | yes - covered by consumers (C73, C39-C42; C2/C3, C4, C6, C27, C65, gate) |

## Faults injected

verified at b5b76c1.

- **Isolation:** `git worktree add --detach <scratchpad>/wt HEAD` (b5b76c1). The `cmd/agenthooks` package
  imports only the standard library, so the worktree needed no extra setup.
- **Baseline:** the real-tree `git status --porcelain` was recorded first: `?? .claude/skills/auth-security/` and
  `?? .cursor/skills/auth-security/`, both pre-existing.
- **Method:** each fault was applied with `sed` in the worktree and confirmed by `git diff`. It was then reverted
  with `git checkout -- app/cmd/agenthooks/main.go` inside the worktree before the next one. `git stash` was never
  used. Each narrowest proof failed on its assertion, not on compilation.
- **Cleanup:** `git worktree remove --force` and `git worktree prune`. `git worktree list` then showed only the main
  tree, and `diff` of the real-tree porcelain before and after was empty.

| Mutation | Location | Killed |
| --- | --- | --- |
| swap the arms: `gofmt` -> `stopHook(..., {"task","check:fast"})`, `stop` -> `gofmtHook` | `app/cmd/agenthooks/main.go:30`, `:32` | yes - `TestMain_StopArmRunsTaskCheckFast` FAIL at `main_test.go:196` (`An error is expected but got nil`); `TestMain_GofmtArmFormats` FAIL at `:171` (it ran the real `task check:fast`, 165.73s) |
| stop command args `"check:fast"` -> `"check"` | `app/cmd/agenthooks/main.go:32` | yes - `TestMain_StopArmRunsTaskCheckFast` FAIL at `main_test.go:201` (`expected: "check:fast" actual: "check"`) |
| route `gofmt` to a no-op: `code = gofmtHook(...)` -> `code = 0` | `app/cmd/agenthooks/main.go:30` | yes - `TestMain_GofmtArmFormats` FAIL at `main_test.go:175` (file left as `"package x\nfunc  F( ) {\nreturn}\n"`) |
| `stop` arm discards the hook's code: `code = stopHook(...)` -> `_ = stopHook(...)` | `app/cmd/agenthooks/main.go:32` | yes - `TestMain_StopArmRunsTaskCheckFast` FAIL at `main_test.go:196` (exit 0, `An error is expected but got nil`) |

Four faults cover the two new proofs and each of their distinct assertions:

- C74: the file bytes (`:175`) and, under the swap, the exit (`:171`).
- C75: the exit code (`:196`/`:197`) and the recorded arguments (`:201`).

A fifth fault would only re-run a proof that has already been made to fail. The round-3 faults on C69-C73 target
surfaces this fix did not touch and carry from 071b520, apart from C72's binary build. C72's build was refactored
into `buildAgenthooks`, and its proof passes at b5b76c1 (B2).

## Swept existing re-read

carried from 071b520. The fix changed no production code, so every cited constraint is where round 3 found it:

- **validation**: `newslice/main.go:25`, `:45-50`.
- **failure modes**: `middleware.go:85-104`; `health.go:98-99`.
- **idempotency**: `cmd/api/main.go:100-112`; `newslice/main.go:57-59`.
- **authorization**: `op.go:40-55`.
- **concurrency**: `app.go` Run/Shutdown; `serve` has no goose call (`main.go:79-98`).
- **data lifecycle** and **state transitions** are `n/a` (user-approved).

## Precision notes

These are not findings against the checks.

- **C75's argument comparison is on the joined string.**
  - `:201` compares `strings.TrimSpace` of the stub's `echo %*` / `echo "$@"` output. A single argument
    `check:fast` and an invocation that splits it across arguments would print the same text. The splits that
    could produce that output are implausible, and the realistic regressions (`check`, extra flags, the arm swap)
    all change the text.
  - On Windows the stub is `task.bat`. It wins only because `stubDir` is first on `PATH`, ahead of the real
    `task.exe`.
  - `cmd.Env` appends `PATH=` after `os.Environ()`'s `Path`. Go's `exec` dedups variable names case-insensitively
    on Windows and the last value wins, which is the stub-first value. The proof passed on Windows (B2) and the
    `check` fault was caught, so the stub was reached.
- **C74 under the arm-swap fault runs the real `task check:fast`** (165.73s in the worktree), because `stopHook`
  then runs the literal command. The fault is still killed, but slowly. This is a property of the mutant, not of
  the proof.
- **C73 / C39 per-query ceiling equals the test-level ceiling.** Carried from 071b520: Vitest's default
  `testTimeout` is 5000 ms and is not overridden.
- **C71 relies on `chmod 0444` making the file unwritable.** This fails when the tests run as root. Carried from
  071b520.
- **C27's committed-tree half still passes vacuously** (`sqlc.yaml` `sql: []`). Carried from 071b520.
- **C47's scanner exclusions are broader than the claim** (`src/test/`, `*.test.ts(x)`). Carried from 071b520.

## Ranked gaps

None. Round 3's single gap (C51, C52, C72 - `app/cmd/agenthooks/main.go:29-32`) is closed by C74 and C75, as shown
above.

## Gate

verified at b5b76c1. `task check` was run from the repo root at HEAD and exited 0 (`EXIT=0`, 0 lines matching
`FAIL`). The proof batches ran concurrently with it.

- `fmt:check` passed. `lint` reported `0 issues.`
- `gen:sqlc:check`, `gen:openapi:check` (`ok ... 0.539s`) and `gen:web:check` passed.
- `archtest` passed (`ok ... 98.988s`).
- `go test -count=1 ./...` reported 10 packages `ok` and 0 failed:
  - `cmd/agenthooks` 60.729s.
  - `cmd/newslice` 407.242s.
- `web:typecheck` was clean. `web:lint` reported `Checked 29 files in 21ms. No fixes applied.`
- `web:test` reported `Test Files 6 passed (6)`, `Tests 11 passed (11)`.

`task e2e -- -g "api online"` exited 0 with `✓ 1 [chromium] › e2e\status.spec.ts:4:1 › api online` (1 passed).
