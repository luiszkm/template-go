# Hardening verification

**Verdict**: PASS
**Profile**: standard
**Diff range**: 4fafad4..7eab4b0
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier)

All 38 checks are green at `7eab4b0` and each has a located assertion. The fix commit `7eab4b0`
(`75ae156..7eab4b0`) touched only tests and `checks.md`: `app/internal/app/hardening_test.go` and
`app/internal/platform/httpx/hardening_test.go`. No production file changed. The five round-1 gaps
are closed:

1. A 204 is now sampled for headers (C16).
2. The 502/503 detail branch of `serverProblem` is asserted (C2, C4).
3. The panic `stack` attribute and the standalone `Recover` log are asserted (C37, C38).
4. Door 9 has a coverage row.
5. The `SESSION_SWEEP_INTERVAL` default is tied to C29.

Four faults were injected on the new assertion surfaces, and all four were killed.

## Binding sources

Carried from 75ae156. The fix did not touch the interface, so this did not run again.

Not applicable. `plan.md` marks no source as binding, and the feature is backend-only, so step 1
and step 5 do not run.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none marked binding (backend only, no binding source) | n/a | - | - |

## Checks

Verified at 7eab4b0. All proofs ran again in full at HEAD in one invocation:
`go -C app test -count=1 -v ./internal/platform/httpx ./internal/platform/auth ./internal/platform/op ./internal/platform/audit ./internal/platform/config ./internal/app ./internal/features/users/login ./cmd/api ./archtest -run '^(<44 names: every proof named in checks.md + TestMiddleware_NoDatabase503>)$'`.
It exited 0. The output has 44 top-level `--- PASS` lines, one per name, with no FAIL and no SKIP, and
all 9 packages report `ok`. C31 also ran `task vuln` (exit 0, "Your code is affected by 0 vulnerabilities").

The citations for the two touched files (`app/internal/platform/httpx/hardening_test.go` and
`app/internal/app/hardening_test.go`) were refreshed at 7eab4b0. Citations into untouched files are
carried from 75ae156 and marked that way. Those files did not change in `75ae156..7eab4b0`.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | plain error -> 500 problem+json, detail `internal server error`, no `errors`, no `secret` | `TestProblem_5xxHidesCause` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:104` `require.Equal(t, http.StatusInternalServerError, rec.Code)`; `:106` `require.Equal(t, "internal server error", body["detail"])`; `:107` `require.NotContains(t, body, "errors")`; `:108` `require.NotContains(t, rec.Body.String(), "secret")` | PASS |
| C2 | huma 500/502/503 with cause -> no `errors`, no `secret`; 500 detail replaced, 502/503 keep `x` | `TestProblem_No5xxCarriesErrors` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:117` `require.NotContains(t, body, "errors", path)`; `:118` NotContains `secret`; `:120` `require.Equal(t, "internal server error", body["detail"], path)` for 500; `:122` `require.Equal(t, "x", body["detail"], ...)` for 502/503 (map `:113`) | PASS |
| C3 | one access entry: ERROR, status 500, request_id = header, error = cause | `TestAccessLog_5xxLogsCauseAtError` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:130` onlyEntry (`:97` `require.Len(t, entries, 1)`); `:131` `"ERROR"`; `:132` status 500; `:133` request_id == header; `:134` `require.Equal(t, leakedCause, entry["error"])` | PASS |
| C4 | 503 without cause -> detail `db down`, one ERROR entry, no `error` key | `TestAccessLog_5xxWithoutCause` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:140` 503; `:141` `require.Equal(t, "db down", problemOf(t, rec)["detail"])`; `:143` `"ERROR"`; `:144` `require.NotContains(t, entry, "error")` | PASS |
| C5 | 404 and 422 -> INFO, no `error`; 422 keeps one `errors` entry with location | `TestAccessLog_Below500IsInfo` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:154` `require.Len(t, errs, 1)`; `:157` `"body.name"` location; `:160` 2 entries; `:162` `"INFO"`; `:163` NotContains `error` | PASS |
| C6 | lookup DB failure -> 500 `internal server error`, no `errors`; ERROR log with injected text | `TestMiddleware_LookupFailure500Logged` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:128` 500; `:130` detail; `:131` NotContains `errors`; `:136` `"ERROR"`; `:137` `require.Contains(t, entries[0]["error"], errInjected.Error())` | PASS |
| C7 | app.New: cross-site/same-site login -> 403 problem+json `cross-origin request rejected`, no login_attempts row | `TestCrossOrigin_RejectsCrossSiteMutation` PASS | verified at 7eab4b0 (lines unchanged): `app/internal/app/hardening_test.go:52` 403; `:53` problem+json; `:54` detail; `:58` `require.Zero(t, attempts, ...)` | PASS |
| C8 | PUT/PATCH/DELETE cross-site -> 403, handler not run | `TestCrossOrigin_RejectsEveryUnsafeMethod` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:171` `require.Equal(t, http.StatusForbidden, rec.Code, m)`; `:174` `require.Zero(t, e.calls.Load())` | PASS |
| C9 | no Sec-Fetch-Site + foreign Origin -> 403; matching Origin reaches handler | `TestCrossOrigin_OriginMustMatchHost` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:180` 403; `:181` zero calls; `:184` `require.Less(t, rec.Code, 300, ...)`; `:185` `require.EqualValues(t, 1, e.calls.Load())` | PASS |
| C10 | same-origin, none, headerless POST reach handler | `TestCrossOrigin_AllowsSameOriginAndHeaderless` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:196` `< 300` per case; `:198` `require.EqualValues(t, 3, e.calls.Load())` | PASS |
| C11 | GET/HEAD cross-site reach handler | `TestCrossOrigin_SafeMethodsPass` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:205` `< 300`; `:207` `require.EqualValues(t, 2, e.calls.Load())` | PASS |
| C12 | every served mutation documents 403; committed openapi.json == served | `TestOpenAPI_MutationsDocument403`, `TestOpenAPI_ServedMatchesCommitted` PASS | verified at 7eab4b0: `app/internal/app/hardening_test.go:71` `require.Contains(t, o.Responses, "403", ...)`; `:75` `require.Positive(t, mutations)`; carried from 75ae156: `app/internal/app/app_test.go:61` `require.Equal(t, string(committed), rec.Body.String())` | PASS |
| C13 | Public POST documents 403; Public GET neither 401 nor 403 | `TestDocumentedErrors_MutationsInclude403` PASS | carried from 75ae156: `app/internal/platform/op/op_test.go:119` `require.Contains(t, things.Post.Responses, "403")`; `:120`-`:121` NotContains `401`/`403` | PASS |
| C14 | timeouts default 30/30/120/10s; env 5/6/7s | `TestNewServer_Timeouts` PASS | carried from 75ae156: `app/cmd/api/hardening_test.go:24`-`:27`; `:33`-`:35`; `app/cmd/api/main.go:100` `srv := newServer(cfg, h)` (re-read at 7eab4b0) | PASS |
| C15 | invalid duration in each of 4 vars -> exit 1, stderr `config:` | `TestServe_InvalidDurationExits1` PASS | carried from 75ae156: `app/cmd/api/hardening_test.go:44` `require.Equal(t, 1, code, name)`; `:45` `strings.HasPrefix(errb.String(), "config:")` | PASS |
| C16 | app.New: 200 `/`, 404, 200 `/healthz`, 403 cross-site POST, 204 DELETE session each carry the 3 headers | `TestSecurityHeaders_OnEveryResponse` PASS | verified at 7eab4b0: `app/internal/app/hardening_test.go:99` case `{"DELETE /api/v1/users/session", logout(t, h, pool), http.StatusNoContent}`; `:102` `require.Equal(t, c.status, c.rec.Code, c.name)`; `:103` `nosniff`; `:104` `DENY`; `:105` `strict-origin-when-cross-origin` | PASS |
| C17 | panic 500 carries the 3 headers | `TestSecurityHeaders_OnPanic500` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:213` 500; `:214`-`:216` the three header values | PASS |
| C18 | `/` and `/users/123` carry exactly the SPA CSP | `TestSecurityHeaders_SPAPolicy` PASS | verified at 7eab4b0: `app/internal/app/hardening_test.go:114` `require.Equal(t, []string{spaPolicy}, rec.Header().Values("Content-Security-Policy"), p)`; literal `:17`-`:18` | PASS |
| C19 | API paths, /healthz, /readyz carry exactly the API CSP | `TestSecurityHeaders_APIPolicy` PASS | verified at 7eab4b0: `app/internal/app/hardening_test.go:121` `require.Equal(t, []string{apiPolicy}, ...)`; paths `:120`; literal `:19` | PASS |
| C20 | /api/docs 200 with one CSP containing unpkg | `TestSecurityHeaders_DocsKeepsHumaPolicy` PASS | verified at 7eab4b0: `app/internal/app/hardening_test.go:127` 200; `:129` `require.Len(t, policies, 1)`; `:130` `require.Contains(t, policies[0], "https://unpkg.com/")` | PASS |
| C21 | HSTS `max-age=31536000` iff CookieSecure | `TestSecurityHeaders_HSTSFollowsCookieSecure` PASS | verified at 7eab4b0: `app/internal/app/hardening_test.go:135` `require.Equal(t, "max-age=31536000", ...)`; `:136` `require.Empty(t, ...Values("Strict-Transport-Security"), p)` | PASS |
| C22 | valid ids echoed in header and problem | `TestRequestID_AcceptsValidFormat` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:223` `require.Equal(t, id, rec.Header().Get(httpx.HeaderRequestID))`; `:224` problem `request_id`; inputs `:221` | PASS |
| C23 | absent/invalid ids -> UUID in header, access log, problem | `TestRequestID_RejectsInvalidFormat` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:234` `require.NoError(t, err, ...)` (uuid.Parse); `:235` problem; `:236` `onlyEntry(t, e.log)["request_id"]`; inputs `:229` | PASS |
| C24 | `X-Request-ID: a b` -> audit request_id == header, UUID | `TestRecord_InvalidRequestIDReplaced` PASS | carried from 75ae156: `app/internal/platform/audit/audit_test.go:145` uuid.Parse NoError; `:148` `require.Equal(t, sent, stored)` | PASS |
| C25 | TTL 1h sweep deletes 61m/13h, keeps 59m/fresh | `TestSweepSessions_DeletesOnlyExpired` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:149` Eventually both gone; `:152` `require.True(t, exists(t, pool, inside))`; `:153` fresh | PASS |
| C26 | sweeps at start and each tick | `TestSweepSessions_RunsAtStartAndEveryInterval` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:165` Eventually 2s (every 1h `:164`); `:177` Eventually 200ms (every 20ms `:173`); `:178` `>= 2` | PASS |
| C27 | failing Exec -> >= 2 WARN with `error` in 200ms, keeps running | `TestSweepSessions_FailureLogsAndRetries` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:192` WARN + error predicate; `:198` Eventually `warns() >= 2`; `:199`-`:203` fatal if returned | PASS |
| C28 | returns within 100ms after cancel | `TestSweepSessions_ReturnsOnCancel` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:214`-`:217` select `done` vs `time.After(100 * time.Millisecond)` -> `t.Fatal` | PASS |
| C29 | serve deletes 2h-old session within 5s, keeps fresh, exits 0; default sweep interval 1h | `TestServe_SweepsExpiredSessions`, `TestDefaults_HardeningConfig` PASS | carried from 75ae156: `app/cmd/api/hardening_test.go:72` Eventually `count(old) == 0`; `:73` fresh == 1; `:75` `require.Equal(t, 0, <-done)`; `app/internal/platform/config/config_test.go:45` `require.Equal(t, time.Hour, cfg.SessionSweepInterval)`; wired at `app/cmd/api/main.go:104` (re-read at 7eab4b0) | PASS |
| C30 | no `go.opentelemetry.io/` dep; no direct otel require | `TestDependencies_NoOpenTelemetry` PASS | carried from 75ae156: `app/archtest/hardening_test.go:15` `require.False(t, strings.HasPrefix(pkg, "go.opentelemetry.io/"))`; `:20` `// indirect` suffix | PASS |
| C31 | `task vuln` exit 0; check lists vuln | `task vuln` exit 0 at 7eab4b0; `TestTaskfile_CheckRunsVuln` PASS | carried from 75ae156: `app/archtest/hardening_test.go:27` `require.Contains(t, calledTasks(t, tf, "check"), "vuln")`; `:28`-`:30`; `Taskfile.yml:56` | PASS |
| C32 | serve against 127.0.0.1:1 -> exit 1, `db: ping:`, no `listening` | `TestServe_UnreachableDatabaseExits1` PASS | carried from 75ae156: `app/cmd/api/hardening_test.go:82` exit 1; `:83` `db: ping:`; `:84` NotContains `listening` | PASS |
| C33 | create-admin against 127.0.0.1:1 -> exit 1, `db: ping:` | `TestCreateAdmin_UnreachableDatabaseExits1` PASS | carried from 75ae156: `app/cmd/api/hardening_test.go:89` exit 1; `:90` `require.Contains(t, stderr, "db: ping:")` | PASS |
| C34 | registration verifies once against DummyHash; first unknown login adds one | `TestRegister_WarmsDummyHash`, `TestLogin_UnknownEmailVerifiesDummyHash` PASS | carried from 75ae156: `app/internal/features/users/login/login_test.go:209` `require.Equal(t, []string{password.DummyHash()}, hashes, ...)`; `:212`; `:196` | PASS |
| C35 | one Lookup = one call; principal holds 3 perms | `TestLookup_SingleQuery` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:229` `require.EqualValues(t, 1, q.calls.Load())`; `:231` perms map | PASS |
| C36 | roleless -> empty perms, 403 Permission, 200 Authenticated; 401 unchanged | `TestLookup_UserWithoutRoles`, `TestMiddleware_Rejects401`, `TestMiddleware_Forbids403` PASS | carried from 75ae156: `app/internal/platform/auth/hardening_test.go:240` `require.Empty(t, p.Permissions)`; `:243` 403; `:244` 200; `:247` ErrNoSession; `app/internal/platform/auth/auth_test.go:80` 401; `:101` 403 | PASS |
| C37 | panic behind full chain -> one ERROR entry, `error` `panic: boom`, `stack` with `goroutine`; body has no `goroutine` | `TestAccessLog_PanicCarriesCauseAndStack` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:251` onlyEntry; `:252` `"ERROR"`; `:253` `require.Equal(t, "panic: boom", entry["error"])`; `:255` `require.True(t, ok, "stack attribute missing: %v", entry)`; `:256` `require.Contains(t, stack, "goroutine")`; `:257` `require.NotContains(t, rec.Body.String(), "goroutine")` | PASS |
| C38 | `Recover` without AccessLog logs `msg` `panic`, `panic` `boom`, `stack` with `goroutine`; answers 500 | `TestRecover_LogsPanicWithoutAccessLog` PASS | verified at 7eab4b0: `app/internal/platform/httpx/hardening_test.go:266` 500; `:267` onlyEntry; `:269` `require.Equal(t, "panic", entry["msg"])`; `:270` `require.Equal(t, "boom", entry["panic"])`; `:271` `require.Contains(t, entry["stack"], "goroutine")` | PASS |

## Coverage

The rows the fix touched were recomputed and verified at 7eab4b0: statuses sampled for headers,
5xx origin, 5xx `detail`, panic log destination, new config variables and doors. All other rows are
carried from 75ae156, where they were recomputed from authority and had nothing unproven.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| statuses sampled for headers (4) - verified at 7eab4b0 | plan `Surface` row 3: `200`, `204`, `404`, `500` | 200 C16 (`app/internal/app/hardening_test.go:95`, `:97`) · 204 C16 (`:99`, asserted `:102`-`:105`) · 404 C16 (`:96`) · 500 C17 (`app/internal/platform/httpx/hardening_test.go:213`-`:216`) | - |
| `detail` of 5xx (2 rows of `serverProblem`) - verified at 7eab4b0 | code: `app/internal/platform/httpx/problem.go:45`-`:48` | `== 500` -> `internal server error` C1 (`hardening_test.go:106`), C2 (`:120`) · other 5xx keep the caller's detail C2 (`:122`, 502 and 503), C4 (`:141`) | - |
| 5xx origin (5) - verified at 7eab4b0 | `problem.go` NewError hook, `auth.Install`, `Recover` | plain error C1 · huma 5xx with cause C2 · huma 5xx without cause C4 · auth lookup failure C6 · panic C37 (`:250`-`:257`), C17 | - |
| panic log destination (2) - verified at 7eab4b0 | code: `app/internal/platform/httpx/middleware.go:108`-`:110` (slot -> access line `stack`), `:126`-`:131` (no slot -> own `panic` line) | access log line with `error` + `stack` C37 (`hardening_test.go:253`, `:256`) · standalone `Recover` line C38 (`:269`-`:271`) | - |
| new config variables (4) - verified at 7eab4b0 | door 4; `app/internal/platform/config/config.go:24` | `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT` C14, C15 · `SESSION_SWEEP_INTERVAL` invalid C15, default `1h` C29 (`app/internal/platform/config/config_test.go:45`, wired `app/cmd/api/main.go:104`) | - |
| one-way doors (9) - verified at 7eab4b0 | plan `Landing` (9 doors); checks row now counts 9 | 1 C1, C3, C37 · 2 C7, C12 · 3 C16, C18 · 4 C14, C15 · 5 C22, C23 · 6 C25, C35 · 7 C30, C31 · 8 C32, C33 · 9 C31 (`task vuln` exit 0 at 7eab4b0) | - |
| mutation route statuses (2) - carried from 75ae156 | plan `Surface` row 1 | 403 C7, C8, C9 · processed normally C9, C10 | - |
| any API route `500` (1) - carried from 75ae156 | plan `Surface` row 2 | C1, C6 | - |
| access-log level (3) - carried from 75ae156 | AC 3-5 | >=500 with cause C3 · without cause C4 · <500 C5 | - |
| `Sec-Fetch-Site` on unsafe methods (5) - carried from 75ae156 | AC 7, AC 9 | cross-site C7, C8 · same-site C7 · same-origin C10 · none C10 · absent C9, C10 | - |
| `Origin` when `Sec-Fetch-Site` absent (3) - carried from 75ae156 | AC 8, AC 9 | mismatch C9 · match C9 · absent C10 | - |
| HTTP methods (6) - carried from 75ae156 | AC 7, AC 10 | POST C7 · PUT/PATCH/DELETE C8 · GET/HEAD C11 | - |
| OpenAPI 403 on mutations (2) - carried from 75ae156 | door 2 | assembled C12 · `op.Register` C13 | - |
| CSP by path (6) - carried from 75ae156 | AC 15-17 | `/` C18 · deep SPA C18 · `/api/*` C19 · `/healthz` C19 · `/readyz` C19 · `/api/docs` C20 | - |
| fixed headers (3) - carried from 75ae156 | AC 14 | nosniff, DENY, Referrer-Policy C16, C17 | - |
| HSTS by `COOKIE_SECURE` (2) - carried from 75ae156 | AC 18-19 | true C21 · false C21 | - |
| server timeouts (4) - carried from 75ae156 | AC 12 | Read, Write, Idle, ReadHeader C14 | - |
| request id inputs (5) - carried from 75ae156 | AC 20-21 | short valid C22 · 64 C22 · 65 C23 · bad char C23 · absent C23 | - |
| request id sinks (4) - carried from 75ae156 | AC 21 | header C22, C23 · access log C23 · problem C22, C23 · audit C24 | - |
| session age vs TTL (4) - carried from 75ae156 | AC 22-23 | 13h, 61m, 59m, fresh C25 | - |
| sweep lifecycle (4) - carried from 75ae156 | AC 22, 24, 25 | start C26, C29 · interval C26 · failure C27 · cancel C28, C29 | - |
| `db.Open` callers (2) - carried from 75ae156 | code: `app/cmd/api/main.go` serve, create-admin | serve C32 · create-admin C33 | - |
| `auth.Lookup` outcomes (3) - carried from 75ae156 | AC 30-31 | with perms C35 · without roles C36 · unknown/expired/deactivated C36 | - |
| middleware assembly sites (3) - carried from 75ae156 | code: `app/internal/app/app.go:75`, `app/internal/platform/testkit/http.go:42`, `:50` | app.New C7, C16 · NewAPI C6 · NewAPIWithoutDatabase `TestMiddleware_NoDatabase503` (ran at 7eab4b0, PASS) | - |

Findings from the recompute:

- The round-1 coverage gap (no proof for a `204`) is closed by C16 at `app/internal/app/hardening_test.go:99`.
- The round-1 artifact gap (door 9 had no row) is closed: `checks.md` now counts 9 doors, and door 9 is proven by C31.
- The round-1 precision gap (the `SESSION_SWEEP_INTERVAL` default was tied to no check) is closed: C29 now names `TestDefaults_HardeningConfig`, and `app/cmd/api/main.go:104` passes `cfg.SessionSweepInterval` to `SweepSessions`.
- Swept re-read (carried from 75ae156): no `Swept` row resolves to `existing`.

## Test policy rows

The rows that were unmet in round 1, and every row that classifies a file the fix touched, were
re-judged at 7eab4b0. The other rows are carried from 75ae156. They are judged against `AGENTS.md`
`## Test policy`.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary - verified at 7eab4b0 | `app/internal/platform/httpx/problem.go` (`>= 500` branch; `== 500` detail override) | own layer C1, C2, C4 · boundary C1, C6 via `httpx.Wrap` | yes - both `serverProblem` rows are asserted: 500 -> `internal server error` (`hardening_test.go:106`, `:120`) and 502/503 keep the detail (`:122`, `:141`). Fault (a) was killed |
| Decides, reached across a boundary - verified at 7eab4b0 | `app/internal/platform/httpx/middleware.go` (request-id regex, AccessLog level/error/stack, Recover slot/no-slot) | own layer C3-C5, C22, C23, C37, C38 · boundary C16, C24 | yes - `stack` on the access line (`hardening_test.go:256`) and the standalone `Recover` line (`:269`-`:271`) are asserted. Faults (b) and (c) were killed |
| Decides, reached across a boundary - verified at 7eab4b0 | `app/internal/platform/httpx/edge.go` (CrossOrigin, policyFor, HSTS, SecurityHeaders) | own layer C8-C11, C17 · boundary C7, C16, C18-C21 | yes - the boundary now includes a 204 (`app/internal/app/hardening_test.go:99`). Fault (d) was killed |
| Decides, reached across a boundary - carried from 75ae156 | `app/internal/platform/op/op.go` (`documentedErrors`) | own layer C13 · boundary C12 | yes |
| Decides, reached across a boundary - carried from 75ae156 | `app/internal/platform/auth/auth.go` (`Lookup`, `Install` 500) | own layer C35, C36 · boundary C6, C36 | yes |
| Decides, not reached across a boundary - carried from 75ae156 | `app/internal/platform/auth/auth.go` `SweepSessions` | own layer C25-C28 | yes |
| Entry point that decides nothing - carried from 75ae156 | `app/cmd/api/main.go`, `app/internal/platform/db/db.go`, `app/internal/platform/config/config.go` | boundary C14, C15, C29, C32, C33 | yes |
| Instrumentation, pass-throughs - carried from 75ae156 | `app/internal/app/app.go`, `app/internal/platform/testkit/http.go`, `app/internal/features/users/login/endpoint.go` | covered by consumers C7, C16, C34 | yes |

## Faults injected

Verified at 7eab4b0. The faults ran in an isolated worktree created with `git worktree add ../hardening-fault2 HEAD`.
Before the run, the real tree's `git status --porcelain` baseline was empty. Each fault was reverted
with `git checkout -- .` in the scratch tree, and the scratch tree's porcelain was confirmed empty
before the next fault. After `git worktree remove --force ../hardening-fault2`, the real tree's
porcelain still matched the baseline (`diff` was clean). The round-1 faults (5 of 5 killed) are
carried from 75ae156. Production code did not change since then.

| Mutation | Location | Killed |
| --- | --- | --- |
| (a) `serverProblem` overwrites detail for every 5xx: `if status == 500` -> `if status >= 500` | `app/internal/platform/httpx/problem.go:46` | yes - `TestProblem_No5xxCarriesErrors` FAIL (expected `x`, actual `internal server error`) and `TestAccessLog_5xxWithoutCause` FAIL (expected `db down`) |
| (b) remove the `stack` attr append in `AccessLog` | `app/internal/platform/httpx/middleware.go:108`-`:110` | yes - `TestAccessLog_PanicCarriesCauseAndStack` FAIL ("stack attribute missing") |
| (c) `Recover` never logs when there is no slot (`_ = recordPanic(...)`, standalone log dropped) | `app/internal/platform/httpx/middleware.go:126`-`:132` | yes - `TestRecover_LogsPanicWithoutAccessLog` FAIL (`"[]" should have 1 item(s)`); `TestAccessLog_PanicCarriesCauseAndStack` still PASS, as expected |
| (d) `SecurityHeaders` skips `X-Content-Type-Options` on `DELETE` (the 204 surface) | `app/internal/platform/httpx/edge.go:38` | yes - `TestSecurityHeaders_OnEveryResponse` FAIL on case `DELETE /api/v1/users/session` (expected `nosniff`, actual empty) |

## Gate

Run from the repository root at `7eab4b0`. As instructed, `task check` and the `./cmd/newslice` and `./cmd/rename` tests were not run.

- `go -C app test -count=1 -v <9 packages> -run '^(...44 names...)$'`: 44 passed, 0 failed (exit 0)
- `task vuln`: exit 0, "Your code is affected by 0 vulnerabilities"
- `task fmt:check`: exit 0
- `task lint`: exit 0, "0 issues."
