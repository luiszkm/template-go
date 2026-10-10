# Hardening verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 4fafad4..e435b73
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier)

All 36 checks are green at `e435b73` with located assertions, and 5 of 5 injected faults were killed.
The verdict is FAIL for two reasons. One `Coverage` member the plan's `Surface` names has no proof:
security headers on a `204`. Two `AGENTS.md` decision tables in changed `httpx` files have rows
that no test asserts.

## Binding sources

Not applicable: `plan.md` marks no source as binding (its `Sources` are conversation, AD records,
the users plan and library code), and the feature is backend-only, so step 1 and step 5 do not run.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none marked binding (backend only) | n/a | - | - |

## Checks

Proof run: one invocation at HEAD:
`go -C app test ./internal/platform/httpx ./internal/app ./internal/platform/op ./cmd/api ./internal/platform/audit ./internal/platform/auth ./archtest ./internal/features/users/login -count=1 -v -run '^(<all 40 named proofs + TestMiddleware_NoDatabase503>)$'`
exited 0. 41 `--- PASS` lines, one per named test, no FAIL or SKIP. All 8 packages report `ok`.
C31 also ran `task vuln` (exit 0).

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | plain handler error -> 500 problem+json, detail `internal server error`, no `errors`, no `secret` | `TestProblem_5xxHidesCause` PASS | `app/internal/platform/httpx/hardening_test.go:104` `require.Equal(t, http.StatusInternalServerError, rec.Code)`; `:105` problemOf asserts `application/problem+json` (`:90`); `:106` `require.Equal(t, "internal server error", body["detail"])`; `:107` `require.NotContains(t, body, "errors")`; `:108` `require.NotContains(t, rec.Body.String(), "secret")` | PASS |
| C2 | huma 500/502/503 with cause -> no `errors`, no `secret` | `TestProblem_No5xxCarriesErrors` PASS | `app/internal/platform/httpx/hardening_test.go:116` `require.NotContains(t, problemOf(t, rec), "errors", path)`; `:117` `require.NotContains(t, rec.Body.String(), "secret", path)` over the map at `:113` (500, 502, 503) | PASS |
| C3 | one access log entry: ERROR, status 500, request_id = header, error = cause | `TestAccessLog_5xxLogsCauseAtError` PASS | `app/internal/platform/httpx/hardening_test.go:124` onlyEntry (`:97` `require.Len(t, entries, 1)`); `:125` `"ERROR"`; `:126` status `500`; `:127` `request_id` == `X-Request-ID`; `:128` `require.Equal(t, leakedCause, entry["error"])` | PASS |
| C4 | 503 without cause -> one ERROR entry, no `error` key | `TestAccessLog_5xxWithoutCause` PASS | `app/internal/platform/httpx/hardening_test.go:133` 503; `:135` `require.Equal(t, "ERROR", entry["level"])`; `:136` `require.NotContains(t, entry, "error")` | PASS |
| C5 | 404 and 422 -> INFO, no `error`; 422 keeps one `errors` entry with location | `TestAccessLog_Below500IsInfo` PASS | `app/internal/platform/httpx/hardening_test.go:146` `require.Len(t, errs, 1)`; `:149` `require.Equal(t, "body.name", first["location"])`; `:152` 2 entries; `:154` `"INFO"`; `:155` `require.NotContains(t, entry, "error")` | PASS |
| C6 | lookup DB failure -> 500 `internal server error`, no `errors`; log ERROR with injected text | `TestMiddleware_LookupFailure500Logged` PASS | `app/internal/platform/auth/hardening_test.go:128` 500; `:130` detail; `:131` `require.NotContains(t, body, "errors")`; `:135` one entry; `:136` `"ERROR"`; `:137` `require.Contains(t, entries[0]["error"], errInjected.Error())` | PASS |
| C7 | app.New: cross-site and same-site login -> 403 problem+json `cross-origin request rejected`, no login_attempts row | `TestCrossOrigin_RejectsCrossSiteMutation` PASS | `app/internal/app/hardening_test.go:52` 403; `:53` problem+json; `:54` detail; `:58` `require.Zero(t, attempts, ...)` | PASS |
| C8 | PUT/PATCH/DELETE cross-site -> 403, handler not run | `TestCrossOrigin_RejectsEveryUnsafeMethod` PASS | `app/internal/platform/httpx/hardening_test.go:163` `require.Equal(t, http.StatusForbidden, rec.Code, m)`; `:166` `require.Zero(t, e.calls.Load())` | PASS |
| C9 | no Sec-Fetch-Site + foreign Origin -> 403; matching Origin reaches handler | `TestCrossOrigin_OriginMustMatchHost` PASS | `app/internal/platform/httpx/hardening_test.go:172` 403; `:173` zero calls; `:176` `require.Less(t, rec.Code, 300)`; `:177` `require.EqualValues(t, 1, e.calls.Load())` | PASS |
| C10 | same-origin, none, headerless POST reach handler | `TestCrossOrigin_AllowsSameOriginAndHeaderless` PASS | `app/internal/platform/httpx/hardening_test.go:188` `< 300` per case; `:190` `require.EqualValues(t, 3, e.calls.Load())` | PASS |
| C11 | GET/HEAD cross-site reach handler | `TestCrossOrigin_SafeMethodsPass` PASS | `app/internal/platform/httpx/hardening_test.go:197` `< 300`; `:199` `require.EqualValues(t, 2, e.calls.Load())` | PASS |
| C12 | every served mutation documents 403; committed openapi.json == served | `TestOpenAPI_MutationsDocument403`, `TestOpenAPI_ServedMatchesCommitted` PASS | `app/internal/app/hardening_test.go:71` `require.Contains(t, o.Responses, "403", ...)`; `:75` `require.Positive(t, mutations)`; `app/internal/app/app_test.go:61` `require.Equal(t, string(committed), rec.Body.String())` | PASS |
| C13 | Public POST documents 403; Public GET documents neither 401 nor 403 | `TestDocumentedErrors_MutationsInclude403` PASS | `app/internal/platform/op/op_test.go:119` `require.Contains(t, things.Post.Responses, "403")`; `:120`-`:121` NotContains `401`/`403` on Get | PASS |
| C14 | server timeouts default 30/30/120/10s; env 5/6/7s | `TestNewServer_Timeouts` PASS | `app/cmd/api/hardening_test.go:24`-`:27` (30s, 30s, 120s, 10s); `:33`-`:35` (5s, 6s, 7s); `serve` builds through the same function at `app/cmd/api/main.go:100` `srv := newServer(cfg, h)` | PASS |
| C15 | invalid duration in each of 4 vars -> exit 1, stderr starts `config:` | `TestServe_InvalidDurationExits1` PASS | `app/cmd/api/hardening_test.go:44` `require.Equal(t, 1, code, name)`; `:45` `strings.HasPrefix(errb.String(), "config:")` over the 4 names at `:40` | PASS |
| C16 | app.New: GET / 200, /api/does-not-exist 404, /healthz 200, cross-site POST 403 each carry the 3 fixed headers | `TestSecurityHeaders_OnEveryResponse` PASS | `app/internal/app/hardening_test.go:91` status per case; `:92` `nosniff`; `:93` `DENY`; `:94` `strict-origin-when-cross-origin` | PASS |
| C17 | panic 500 carries the 3 headers | `TestSecurityHeaders_OnPanic500` PASS | `app/internal/platform/httpx/hardening_test.go:205` 500; `:206`-`:208` the three header values | PASS |
| C18 | GET / and /users/123 carry exactly the SPA CSP | `TestSecurityHeaders_SPAPolicy` PASS | `app/internal/app/hardening_test.go:103` `require.Equal(t, []string{spaPolicy}, rec.Header().Values("Content-Security-Policy"), p)`; literal at `:17`-`:18` matches AC 15 | PASS |
| C19 | /api/does-not-exist, /api/openapi.json, /healthz, /readyz carry exactly the API CSP | `TestSecurityHeaders_APIPolicy` PASS | `app/internal/app/hardening_test.go:110` `require.Equal(t, []string{apiPolicy}, ...)`; literal `:19` | PASS |
| C20 | /api/docs 200 with one CSP containing unpkg | `TestSecurityHeaders_DocsKeepsHumaPolicy` PASS | `app/internal/app/hardening_test.go:116` 200; `:118` `require.Len(t, policies, 1)`; `:119` `require.Contains(t, policies[0], "https://unpkg.com/")` | PASS |
| C21 | HSTS `max-age=31536000` iff CookieSecure on / and /api/does-not-exist | `TestSecurityHeaders_HSTSFollowsCookieSecure` PASS | `app/internal/app/hardening_test.go:124` `require.Equal(t, "max-age=31536000", ...)`; `:125` `require.Empty(t, ...Values("Strict-Transport-Security"))` | PASS |
| C22 | valid ids echoed in header and problem request_id | `TestRequestID_AcceptsValidFormat` PASS | `app/internal/platform/httpx/hardening_test.go:215` `require.Equal(t, id, rec.Header().Get(httpx.HeaderRequestID))`; `:216` `require.Equal(t, id, problemOf(t, rec)["request_id"])`; inputs `:213` | PASS |
| C23 | absent/invalid ids -> UUID in header, access log, problem | `TestRequestID_RejectsInvalidFormat` PASS | `app/internal/platform/httpx/hardening_test.go:226` `uuid.Parse` no error; `:227` problem `request_id`; `:228` `onlyEntry(...)["request_id"]`; inputs `:221` | PASS |
| C24 | `X-Request-ID: a b` -> audit_events.request_id == response header, a UUID | `TestRecord_InvalidRequestIDReplaced` PASS | `app/internal/platform/audit/audit_test.go:145` `require.NoError(t, err, sent)` (uuid.Parse); `:148` `require.Equal(t, sent, stored)` | PASS |
| C25 | TTL 1h sweep deletes 61m and 13h, keeps 59m and fresh | `TestSweepSessions_DeletesOnlyExpired` PASS | `app/internal/platform/auth/hardening_test.go:149` Eventually `!exists(past) && !exists(long)`; `:152` `require.True(t, exists(t, pool, inside))`; `:153` fresh | PASS |
| C26 | sweeps at start (every 1h, gone within 2s) and each tick (every 20ms, gone within 200ms) | `TestSweepSessions_RunsAtStartAndEveryInterval` PASS | `app/internal/platform/auth/hardening_test.go:165` Eventually 2s with `every` 1h (`:164`); `:177` Eventually 200ms with `every` 20ms (`:173`); `:178` `>= 2` calls | PASS |
| C27 | failing Exec -> >= 2 WARN entries with `error` within 200ms, does not return | `TestSweepSessions_FailureLogsAndRetries` PASS | `app/internal/platform/auth/hardening_test.go:192` `e["level"] == "WARN" && e["error"] == errInjected.Error()`; `:198` Eventually `warns() >= 2` 200ms; `:199`-`:203` fatal if returned | PASS |
| C28 | returns within 100ms after cancel | `TestSweepSessions_ReturnsOnCancel` PASS | `app/internal/platform/auth/hardening_test.go:214`-`:217` select on `done` vs `time.After(100 * time.Millisecond)` -> `t.Fatal` | PASS |
| C29 | serve deletes 2h-old session (TTL 1h) within 5s of /healthz, keeps fresh, exits 0 on cancel | `TestServe_SweepsExpiredSessions` PASS | `app/cmd/api/hardening_test.go:72` Eventually `count(old) == 0` 5s; `:73` `require.Equal(t, 1, count(fresh.Cookie.Value))`; `:75` `require.Equal(t, 0, <-done)` | PASS |
| C30 | no `go.opentelemetry.io/` in deps of cmd/api; no direct otel require | `TestDependencies_NoOpenTelemetry` PASS | `app/archtest/hardening_test.go:15` `require.False(t, strings.HasPrefix(pkg, "go.opentelemetry.io/"))`; `:20` `require.True(t, strings.HasSuffix(line, "// indirect"))` | PASS |
| C31 | `task vuln` runs govulncheck in app from tools module, exit 0; check lists vuln | `task vuln` exit 0 ("Your code is affected by 0 vulnerabilities"); `TestTaskfile_CheckRunsVuln` PASS | `app/archtest/hardening_test.go:27` `require.Contains(t, calledTasks(t, tf, "check"), "vuln")`; `:28` `govulncheck ./...`; `:29` Dir `app`; `:30` tools/go.mod declares the tool; `Taskfile.yml:56` `{{.GOTOOL}} govulncheck ./...` | PASS |
| C32 | serve against 127.0.0.1:1 -> exit 1, stderr `db: ping:`, no `listening` | `TestServe_UnreachableDatabaseExits1` PASS | `app/cmd/api/hardening_test.go:82` `require.Equal(t, 1, code)`; `:83` `require.Contains(t, errb.String(), "db: ping:")`; `:84` `require.NotContains(t, out.String(), "listening")` | PASS |
| C33 | create-admin against 127.0.0.1:1 -> exit 1, stderr `db: ping:` | `TestCreateAdmin_UnreachableDatabaseExits1` PASS | `app/cmd/api/hardening_test.go:89` exit 1; `:90` `require.Contains(t, stderr, "db: ping:")` | PASS |
| C34 | registration verifies once against DummyHash; first unknown-email login adds exactly one | `TestRegister_WarmsDummyHash`, `TestLogin_UnknownEmailVerifiesDummyHash` PASS | `app/internal/features/users/login/login_test.go:209` `require.Equal(t, []string{password.DummyHash()}, hashes, ...)`; `:212` two DummyHash entries; `:196` after reset at `:193` | PASS |
| C35 | one Lookup = exactly one call; principal holds the 3 permissions | `TestLookup_SingleQuery` PASS | `app/internal/platform/auth/hardening_test.go:229` `require.EqualValues(t, 1, q.calls.Load())`; `:231` map of `a:read`, `b:write`, `c:delete` | PASS |
| C36 | roleless user -> empty perms, 403 on Permission, 200 on Authenticated; 401 cases unchanged | `TestLookup_UserWithoutRoles`, `TestMiddleware_Rejects401`, `TestMiddleware_Forbids403` PASS | `app/internal/platform/auth/hardening_test.go:240` `require.Empty(t, p.Permissions)`; `:243` 403; `:244` 200; `:247` `ErrNoSession`; `app/internal/platform/auth/auth_test.go:80` `require.Equal(t, http.StatusUnauthorized, rec.Code, path)`; `app/internal/platform/auth/auth_test.go:101` 403 | PASS |

## Coverage

Recomputed from authority. The route statuses and header sample come from the plan's `Surface`;
the CSP paths, HSTS, timeouts, request-id format and sweep from its ACs and `Landing`; the
assembly sites and `db.Open` callers from the code (grep).

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| mutation route statuses (2) | plan `Surface` row 1 | 403 C7 (`app/internal/app/hardening_test.go:52`), C8, C9 · processed normally C9 (`:177`), C10 | - |
| any API route `500` (1) | plan `Surface` row 2 | C1, C6 | - |
| statuses sampled for headers (4) | plan `Surface` row 3: `200`, `204`, `404`, `500` | 200 C16 (`app/internal/app/hardening_test.go:85`) · 404 C16 (`:86`) · 500 C17 (`app/internal/platform/httpx/hardening_test.go:205`) · 204: the table maps it to C21, but C21 only requests `/` (200) and `/api/does-not-exist` (404) (`app/internal/app/hardening_test.go:123`). A search for `StatusNoContent` or `204` across `*hardening_test.go`, and for `X-Content-Type-Options` or `Strict-Transport-Security` across all `*_test.go`, found no test that checks headers on a 204 | 204 |
| 5xx origin (5) | `problem.go` NewError hook, `auth.Install`, `Recover` | plain error C1 · huma 5xx with cause C2 · huma 5xx without cause C4 · auth lookup failure C6 · panic C17 (headers) + existing `TestRecover_LogsErrorWithRequestID` (one ERROR line, run, PASS) | - |
| access-log level (3) | AC 3-5 | >=500 with cause C3 · >=500 without cause C4 · <500 C5 | - |
| `Sec-Fetch-Site` on unsafe methods (5) | AC 7, AC 9 | cross-site C7, C8 · same-site C7 · same-origin C10 · none C10 · absent C9, C10 | - |
| `Origin` when `Sec-Fetch-Site` absent (3) | AC 8, AC 9 | mismatch C9 · match C9 · absent C10 | - |
| HTTP methods (6) | AC 7, AC 10 | POST C7, C9, C10 · PUT C8 · PATCH C8 · DELETE C8 · GET C11 · HEAD C11 | - |
| OpenAPI 403 on mutations (2 places) | door 2 | assembled C12 · `op.Register` C13 | - |
| CSP by path (6) | AC 15-17, door 3 | `/` C18 · SPA deep path C18 · `/api/*` C19 · `/healthz` C19 · `/readyz` C19 · `/api/docs` C20 | - |
| fixed headers (3) | AC 14 | nosniff, DENY, Referrer-Policy C16, C17 | - |
| HSTS by `COOKIE_SECURE` (2) | AC 18-19 | true C21 · false C21 | - |
| server timeouts (4) | AC 12 | Read, Write, Idle, ReadHeader C14 | - |
| new config variables (4) | door 4 | `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT` C14, C15 · `SESSION_SWEEP_INTERVAL` C15 (its `1h` default is asserted only by `app/internal/platform/config/config_test.go:45`, a test no check names; run here, PASS) | - |
| request id inputs (5 classes) | AC 20-21, door 5 | short valid C22 · exactly 64 C22 · 65 C23 · disallowed char C23 · absent C23 | - |
| request id sinks (4) | AC 21 | header C22, C23 · access log C23 · problem body C22, C23 · audit event C24 | - |
| session age vs TTL (4) | AC 22-23 | 13h, 61m, 59m, fresh C25 | - |
| sweep lifecycle (4) | AC 22, 24, 25 | start C26, C29 · interval C26 · failure C27 · cancel C28, C29 | - |
| `db.Open` callers (2) | code: `app/cmd/api/main.go:84`, `app/cmd/api/main.go:176` | serve C32 · create-admin C33 | - |
| `auth.Lookup` outcomes (3) | AC 30-31 | with perms C35 · without roles C36 · unknown/expired/deactivated C36 (`auth_test.go:80`, `hardening_test.go:247`) | - |
| middleware assembly sites (3) | code: `app/internal/app/app.go:75`, `app/internal/platform/testkit/http.go:42`, `app/internal/platform/testkit/http.go:50`, all `httpx.Wrap(...)` | app.New C7, C16 · NewAPI C6 path family · NewAPIWithoutDatabase `TestMiddleware_NoDatabase503` (run, PASS) | - |
| one-way doors (9) | plan `Landing` has 9 doors; the checks' table says 8 | 1 C1, C3 · 2 C7, C12 · 3 C16, C18 · 4 C14, C15 · 5 C22, C23 · 6 C25, C35 · 7 C30, C31 · 8 C32, C33 · 9 C31 (`task vuln` exit 0 is the outcome door 9 exists for; literals at `app/go.mod:3`, `app/tools/go.mod:3`, `app/go.mod:49`) | - |

Findings from the recompute:

- **Coverage gap (FAIL):** the plan's `Surface` samples header statuses `200, 204, 404, 500`. No
  proof covers `204`, and the checks' table points it at C21, which never produces a 204.
- **Artifact gap (not failing):** the checks' doors row counts 8, but `Landing` has 9. Door 9
  (Go 1.26.9 / `moby/go-archive v0.3.0`) is added in Handoff but has no row. Its outcome is proven
  by C31's `task vuln` exit 0. No test asserts the literal versions.
- **Precision gap:** the `SESSION_SWEEP_INTERVAL` default of `1h` (AC 22, door 4) is asserted only
  by `TestDefaults_HardeningConfig` (`app/internal/platform/config/config_test.go:45`), which no
  check names. C29 relies on the start sweep and never sets or reads the interval.
- **Swept re-read:** no `Swept` row resolves to `existing`. The authorization row cites C36 re-running
  `TestMiddleware_Rejects401` and `TestMiddleware_Forbids403`, and both ran and passed. The
  `concurrency` and `state transitions` rows are `n/a`, which is approved policy.

## Test policy rows

Judged against `AGENTS.md` `## Test policy`, as `checks.md` `## Test policy` says to.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `app/internal/platform/httpx/edge.go` (CrossOrigin, policyFor, HSTS) | own layer C8-C11, C17 · boundary C7, C16, C18-C21 | yes |
| Decides, reached across a boundary | `app/internal/platform/httpx/problem.go` (`>= 500` branch; `== 500` detail override) | own layer C1, C2 · boundary C1 via `httpx.Wrap` | no - `serverProblem` at `app/internal/platform/httpx/problem.go:45` has 2 rows. `status == 500` -> `internal server error` is asserted (C1). `502`/`503` keep the caller's detail is asserted nowhere: C2 (`hardening_test.go:115-117`) and C4 (`:133-136`) never read `detail`, so a mutant that always overwrites the detail would pass. The checks also leave the 502/503 detail unspecified (precision gap) |
| Decides, reached across a boundary | `app/internal/platform/httpx/middleware.go` (request-id regex, AccessLog level/error/stack, Recover slot/no-slot) | own layer C3-C5, C22, C23 · boundary C24, C16 | no - two rows have no asserted case. The `stack` attribute on a panic's access line (`middleware.go:108-109`) is never asserted (a search for `"stack"` finds no test assertion). Recover's no-AccessLog row, where it writes its own `panic` line (`middleware.go:126-131`), is never exercised: the only standalone `httpx.Recover` test (`httpx_test.go:162`) takes the re-panic path |
| Decides, reached across a boundary | `app/internal/platform/op/op.go` (`documentedErrors` mutation -> 403) | own layer C13 · boundary C12 | yes |
| Decides, reached across a boundary | `app/internal/platform/auth/auth.go` (`Lookup` single query; `Install` 500 with cause) | own layer C35, C36 · boundary C6, C36 middleware | yes |
| Decides, not reached across a boundary | `app/internal/platform/auth/auth.go` `SweepSessions` (failure/tick/cancel, age bound) | own layer C25-C28 | yes |
| Entry point that decides nothing | `app/cmd/api/main.go`, `app/internal/platform/db/db.go` ping, `app/internal/platform/config/config.go` | boundary C14, C15, C29, C32, C33 | yes |
| Instrumentation, pass-throughs | `app/internal/app/app.go`, `app/internal/platform/testkit/http.go`, `app/internal/features/users/login/endpoint.go` warm-up | covered by consumers C7, C16, C34 | yes |

## Faults injected

Isolated in `git worktree add ../hardening-fault HEAD`. The real tree's baseline
`git status --porcelain` was empty before. After `git worktree remove --force` it was still empty
and matched the baseline (`diff` clean). Each fault was reverted in the scratch tree
(`git checkout -- .`) before the next one.

| Mutation | Location | Killed |
| --- | --- | --- |
| `if status >= http.StatusInternalServerError` -> `if status > 599` (5xx keeps `errors`) | `app/internal/platform/httpx/problem.go:21` | yes - `TestProblem_5xxHidesCause`, `TestProblem_No5xxCarriesErrors` FAIL |
| drop `\|\| path == "/readyz"` from API CSP selection | `app/internal/platform/httpx/edge.go:50` | yes - `TestSecurityHeaders_APIPolicy` FAIL (`hardening_test.go:110`) |
| request-id bound `{1,64}` -> `{1,65}` | `app/internal/platform/httpx/middleware.go:18` | yes - `TestRequestID_RejectsInvalidFormat` FAIL |
| sweep bound `make_interval(secs => $1)` -> `$1 - 300` (TTL shifted 5 min) | `app/internal/platform/auth/auth.go:127` | yes - `TestSweepSessions_DeletesOnlyExpired` FAIL (`hardening_test.go:152`) |
| `documentedErrors`: drop `\|\| mutates(s.Method)` | `app/internal/platform/op/op.go:127` | yes - `TestDocumentedErrors_MutationsInclude403` FAIL |

Cap of 5 reached. The `CrossOrigin` deny path and the `db.Open` ping were not mutated.

## Gate

Run from the repository root at `e435b73`. `task check` was not run, as instructed.

- `go -C app test <8 packages> -count=1 -v -run '^(...41 names...)$'` - 41 passed, 0 failed (exit 0)
- `go -C app test ./internal/platform/httpx ./internal/platform/config -run '^(TestRecover_LogsErrorWithRequestID|TestRecover_500WithoutStack|TestDefaults_HardeningConfig)$'` - 3 passed, 0 failed (supporting evidence)
- `task vuln` - exit 0, "Your code is affected by 0 vulnerabilities"
- `task fmt:check` - exit 0
- `task lint` - exit 0, "0 issues."
