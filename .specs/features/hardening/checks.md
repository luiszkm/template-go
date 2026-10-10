# Hardening checks

Profile: standard
Plan: `.specs/features/hardening/plan.md`

## Intent

36 checks in 8 slices · 8 one-way doors · 0 open

Comandos reais do repositório: `go -C app test <pkg> -run '<regex>'`, `task <name>` e `go -C app list`. Testes que
tocam o banco usam `testkit.MigratedDB` (testcontainers, Postgres 17); nenhum banco é simulado. Um `auth.Querier`
que falha ou conta chamadas pode envolver o pool real - é a única dupla, e só onde o claim é sobre o número de
consultas ou sobre a reação a uma falha do banco.

A cadeia de middlewares passa a ser montada por uma única função de `platform/httpx`, usada por `app.New` e por
`testkit.NewAPI`/`NewAPIWithoutDatabase`, para que a montagem de produção e a dos testes não divirjam.

Consequência aprovada de AC 29: `TestLogin_UnknownEmailVerifiesDummyHash` (users C13) passa a medir as chamadas a
partir do primeiro request, porque o registro agora faz uma verificação de aquecimento. O claim de users C13
("exatamente uma verificação por login com e-mail desconhecido, contra o hash fictício") não muda.

Consequência aprovada das decisões do usuário (remover OTel; `govulncheck` no gate): foundation C32 perde o membro
`otel` (`TestDependencies_Declared` deixa de exigir `go.opentelemetry.io/otel`) e `TestTaskfile_CheckRunsAllSteps`
passa a esperar o passo `vuln`. Nenhum outro membro dessas asserções muda.

## Checks

### S1 - 5xx não vaza a causa · ~6 files · ~30 KB · ~8k

**C1** - A Huma operation whose handler returns `errors.New("pq: secret host=db user=app")` answers `500` `application/problem+json` with `detail` `internal server error`, no `errors` member, and a body that does not contain `secret` (HARD-01, AC 1, door 1) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestProblem_5xxHidesCause$'`

**C2** - `huma.Error500InternalServerError("x", errors.New("secret"))`, `huma.Error503ServiceUnavailable("x", errors.New("secret"))` and `huma.Error502BadGateway("x", errors.New("secret"))` returned by handlers each produce a body with no `errors` member and without `secret` (HARD-01, AC 2, door 1) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestProblem_No5xxCarriesErrors$'`

**C3** - For the C1 request, the access log has exactly one entry; it has `level` `ERROR`, `status` `500`, `request_id` equal to the response `X-Request-ID`, and `error` equal to `pq: secret host=db user=app` (HARD-01, AC 3, door 1) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestAccessLog_5xxLogsCauseAtError$'`

**C4** - A handler returning `huma.Error503ServiceUnavailable("db down")` yields one access log entry with `level` `ERROR` and no `error` key (HARD-01, AC 4) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestAccessLog_5xxWithoutCause$'`

**C5** - A `404` and a `422` (body failing a `minLength` validation) each yield an access log entry with `level` `INFO` and no `error` key; the `422` body keeps one `errors` entry with its `location` (HARD-01, AC 5) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestAccessLog_Below500IsInfo$'`

**C6** - With a `Querier` around the real pool whose `QueryRow` fails, a request carrying a session cookie to an `Authenticated` operation answers `500` with `detail` `internal server error` and no `errors`, and the access log entry has `level` `ERROR` with the injected error text in `error` (HARD-01, AC 6, door 1) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_LookupFailure500Logged$'`

### S2 - Mutação de outra origem recusada · ~5 files · ~25 KB · ~7k

**C7** - Through `app.New`, `POST /api/v1/users/session` with a valid JSON body answers `403` problem+json with `detail` `cross-origin request rejected` for `Sec-Fetch-Site` `cross-site` and for `same-site`, and the login handler records no `login_attempts` row (HARD-02, AC 7, door 2) `[done]`
Proof: `go -C app test ./internal/app -run '^TestCrossOrigin_RejectsCrossSiteMutation$'`

**C8** - Each of `PUT`, `PATCH`, `DELETE` to a registered operation with `Sec-Fetch-Site: cross-site` answers `403` and the handler is not run (HARD-02, AC 7) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestCrossOrigin_RejectsEveryUnsafeMethod$'`

**C9** - `POST` with no `Sec-Fetch-Site` and `Origin: https://evil.example` while `Host` is `app.local` answers `403`; with `Origin: http://app.local` it reaches the handler (HARD-02, AC 8, AC 9) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestCrossOrigin_OriginMustMatchHost$'`

**C10** - `POST` reaches the handler for `Sec-Fetch-Site` `same-origin`, for `none`, and with neither `Sec-Fetch-Site` nor `Origin` (HARD-02, AC 9) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestCrossOrigin_AllowsSameOriginAndHeaderless$'`

**C11** - `GET` and `HEAD` with `Sec-Fetch-Site: cross-site` reach the handler (HARD-02, AC 10) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestCrossOrigin_SafeMethodsPass$'`

**C12** - In the assembled server's OpenAPI document, every `post`, `put`, `patch` and `delete` operation lists response `403`, and the committed `app/openapi.json` matches the served one (HARD-02, AC 11, door 2) `[done]`
Proof: `go -C app test ./internal/app -run '^TestOpenAPI_MutationsDocument403$'`
Proof: `go -C app test ./internal/app -run '^TestOpenAPI_ServedMatchesCommitted$'`

**C13** - `op.Register` of a `POST` operation declaring `Public: true` and no `Errors` documents `403`; a `GET` `Public` operation documents neither `401` nor `403` (HARD-02, AC 11) `[done]`
Proof: `go -C app test ./internal/platform/op -run '^TestDocumentedErrors_MutationsInclude403$'`

### S3 - Timeouts do servidor · ~3 files · ~20 KB · ~5k

**C14** - With no timeout variables set, the server `serve` builds has `ReadTimeout` 30s, `WriteTimeout` 30s, `IdleTimeout` 120s, `ReadHeaderTimeout` 10s; with `HTTP_READ_TIMEOUT=5s`, `HTTP_WRITE_TIMEOUT=6s`, `HTTP_IDLE_TIMEOUT=7s` it has 5s, 6s, 7s (HARD-03, AC 12, door 4) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestNewServer_Timeouts$'`

**C15** - `api serve` with a valid `DATABASE_URL` exits `1` with stderr starting `config:` for each of `HTTP_READ_TIMEOUT=abc`, `HTTP_WRITE_TIMEOUT=abc`, `HTTP_IDLE_TIMEOUT=abc`, `SESSION_SWEEP_INTERVAL=abc` (HARD-03, AC 13, door 4) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestServe_InvalidDurationExits1$'`

### S4 - Cabeçalhos de segurança · ~4 files · ~25 KB · ~7k

**C16** - Through `app.New`, responses to `GET /` (200), `GET /api/does-not-exist` (404), `GET /healthz` (200) and a `POST /api/v1/users/session` with `Sec-Fetch-Site: cross-site` (403) each carry `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` and `Referrer-Policy: strict-origin-when-cross-origin` (HARD-04, AC 14, door 3) `[done]`
Proof: `go -C app test ./internal/app -run '^TestSecurityHeaders_OnEveryResponse$'`

**C17** - A `500` produced by a panicking handler carries the three AC 14 headers (HARD-04, AC 14) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestSecurityHeaders_OnPanic500$'`

**C18** - Through `app.New`, `GET /` and `GET /users/123` carry exactly the SPA `Content-Security-Policy` literal of AC 15 (HARD-04, AC 15, door 3) `[done]`
Proof: `go -C app test ./internal/app -run '^TestSecurityHeaders_SPAPolicy$'`

**C19** - Through `app.New`, `GET /api/does-not-exist`, `GET /api/openapi.json`, `GET /healthz` and `GET /readyz` carry exactly `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'` (HARD-04, AC 16, door 3) `[done]`
Proof: `go -C app test ./internal/app -run '^TestSecurityHeaders_APIPolicy$'`

**C20** - Through `app.New`, `GET /api/docs` answers `200` with one `Content-Security-Policy` header that contains `https://unpkg.com/` (HARD-04, AC 17) `[done]`
Proof: `go -C app test ./internal/app -run '^TestSecurityHeaders_DocsKeepsHumaPolicy$'`

**C21** - With `CookieSecure: true`, `GET /` and `GET /api/does-not-exist` carry `Strict-Transport-Security: max-age=31536000`; with `CookieSecure: false` neither carries `Strict-Transport-Security` (HARD-04, AC 18, AC 19) `[done]`
Proof: `go -C app test ./internal/app -run '^TestSecurityHeaders_HSTSFollowsCookieSecure$'`

### S5 - `X-Request-ID` validado · ~2 files · ~15 KB · ~4k

**C22** - Incoming ids `abc-123`, `A.b_C-9`, a UUID, and a 64-character `a` string are echoed unchanged in the response header and in the problem `request_id` (HARD-05, AC 20, door 5) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestRequestID_AcceptsValidFormat$'`

**C23** - An absent header and incoming ids of 65 characters, `a b`, `a\tb`, `a"b`, `ação` and `{"x":1}` are each replaced by a value that parses as a UUID, used in the response header, the access log `request_id` and the problem `request_id` (HARD-05, AC 21, door 5) `[done]`
Proof: `go -C app test ./internal/platform/httpx -run '^TestRequestID_RejectsInvalidFormat$'`

**C24** - A mutation sent with `X-Request-ID: a b` records an `audit_events` row whose `request_id` equals the response `X-Request-ID`, which parses as a UUID (HARD-05, AC 21) `[done]`
Proof: `go -C app test ./internal/platform/audit -run '^TestRecord_InvalidRequestIDReplaced$'`

### S6 - Sessões expiradas apagadas · ~4 files · ~25 KB · ~7k

**C25** - Against the real database with TTL `1h`, one sweep deletes a session created `61 minutes` ago and a session created `13 hours` ago, and keeps a session created `59 minutes` ago and one created now (HARD-06, AC 22, AC 23, door 6) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestSweepSessions_DeletesOnlyExpired$'`

**C26** - `SweepSessions` sweeps immediately on start (with `every` 1h, a session expired before start is gone within 2s) and again after each tick (with `every` 20ms, a session made expired after start is gone within 200ms) (HARD-06, AC 22, door 6) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestSweepSessions_RunsAtStartAndEveryInterval$'`

**C27** - With a `Querier` whose `Exec` fails, `SweepSessions` with `every` 10ms logs at least 2 entries with `level` `WARN` and an `error` key within 200ms and does not return (HARD-06, AC 24) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestSweepSessions_FailureLogsAndRetries$'`

**C28** - After its context is cancelled, `SweepSessions` returns within 100ms (HARD-06, AC 25) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestSweepSessions_ReturnsOnCancel$'`

**C29** - `api serve` started against a database holding a session created `2 hours` ago with `SESSION_TTL=1h` deletes it within 5s of answering `/healthz`, keeps a fresh session, and exits `0` on cancel (HARD-06, AC 22, AC 25) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestServe_SweepsExpiredSessions$'`

### S7 - Dependências e gate · ~4 files · ~10 KB · ~3k

**C30** - `go -C app list -deps ./cmd/api` lists no package with prefix `go.opentelemetry.io/`, and `app/go.mod` has no direct (non-`// indirect`) `require` line for any `go.opentelemetry.io/` module (HARD-07, AC 26, door 7) `[done]`
Proof: `go -C app test ./archtest -run '^TestDependencies_NoOpenTelemetry$'`

**C31** - `task vuln` runs `govulncheck ./...` in `app` from the tools module and exits `0`; `task check` lists `vuln` among its steps (HARD-07, AC 27, door 7) `[done]`
Proof: `task vuln` (comando único; o claim é o exit code dele)
Proof: `go -C app test ./archtest -run '^TestTaskfile_CheckRunsVuln$'`

### S8 - Subida e caminho quente · ~6 files · ~35 KB · ~9k

**C32** - `api serve` with `DATABASE_URL` pointing at `127.0.0.1:1` exits `1`, stderr contains `db: ping:`, and stdout has no `listening` entry (HARD-08, AC 28, door 8) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestServe_UnreachableDatabaseExits1$'`

**C33** - `api users create-admin` with a valid password and `DATABASE_URL` pointing at `127.0.0.1:1` exits `1` with stderr containing `db: ping:` (HARD-08, AC 28, door 8) `[done]`
Proof: `go -C app test ./cmd/api -run '^TestCreateAdmin_UnreachableDatabaseExits1$'`

**C34** - `login.RegisterWithVerifier` with a recording verifier records exactly one call, with hash `password.DummyHash()`, before any request; the first unknown-email login then adds exactly one more call against `password.DummyHash()` (HARD-08, AC 29) `[done]`
Proof: `go -C app test ./internal/features/users/login -run '^TestRegister_WarmsDummyHash$'`
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_UnknownEmailVerifiesDummyHash$'`

**C35** - With a counting `Querier` around the real pool, one `auth.Lookup` of a valid session for a user holding two roles with three distinct permissions issues exactly one call, and the principal holds those three permissions (HARD-08, AC 30, door 6) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestLookup_SingleQuery$'`

**C36** - `auth.Lookup` of a valid session for a user with no role returns a principal with an empty permission set and no error; through the middleware that user gets `403` on a `Permission` operation and `200` on an `Authenticated` one; the existing `401` cases still answer `401` (HARD-08, AC 31) `[done]`
Proof: `go -C app test ./internal/platform/auth -run '^TestLookup_UserWithoutRoles$'`
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_Rejects401$'`
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_Forbids403$'`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| `POST`/`PUT`/`PATCH`/`DELETE /api/v1/*` statuses (2) | 403 C7 · C8 · C9 · processed normally C9 · C10 | - |
| qualquer rota da API - status `500` (1) | 500 C1 · C6 | - |
| qualquer rota - statuses sampled for headers (4) | 200 C16 · 204 C21 · 404 C16 · 500 C17 | - |
| erro 5xx por origem (4) | handler plain error C1 · huma 5xx with cause C2 · huma 5xx without cause C4 · auth lookup failure C6 | - |
| nível do log de acesso (3) | `>= 500` with cause C3 · `>= 500` without cause C4 · `< 500` C5 | - |
| `Sec-Fetch-Site` values on unsafe methods (5) | `cross-site` C7 · `same-site` C7 · `same-origin` C10 · `none` C10 · absent C9 · C10 | - |
| `Origin` when `Sec-Fetch-Site` absent (3) | mismatched host C9 · matching host C9 · absent C10 | - |
| HTTP methods (6) | `POST` C7 · `PUT` C8 · `PATCH` C8 · `DELETE` C8 · `GET` C11 · `HEAD` C11 | - |
| OpenAPI 403 on mutations (2 places) | assembled document C12 · `op.Register` C13 | - |
| CSP by path (6) | `/` C18 · SPA deep path C18 · `/api/*` C19 · `/healthz` C19 · `/readyz` C19 · `/api/docs` C20 | - |
| fixed headers (3) | `X-Content-Type-Options` C16 · `X-Frame-Options` C16 · `Referrer-Policy` C16 | - |
| HSTS by `COOKIE_SECURE` (2) | `true` C21 · `false` C21 | - |
| server timeouts (4) | `ReadTimeout` C14 · `WriteTimeout` C14 · `IdleTimeout` C14 · `ReadHeaderTimeout` C14 | - |
| new config variables (4) | `HTTP_READ_TIMEOUT` C14 · C15 · `HTTP_WRITE_TIMEOUT` C14 · C15 · `HTTP_IDLE_TIMEOUT` C14 · C15 · `SESSION_SWEEP_INTERVAL` C15 · C29 | - |
| request id inputs (5 classes) | valid short C22 · exactly 64 C22 · 65 chars C23 · disallowed char C23 · absent C23 | - |
| request id sinks (4) | response header C22 · C23 · access log C23 · problem body C22 · C23 · audit event C24 | - |
| session age vs TTL (4 edges) | well past C25 · just past C25 · just inside C25 · fresh C25 | - |
| sweep lifecycle (4) | at start C26 · C29 · each interval C26 · on failure C27 · on cancel C28 · C29 | - |
| `db.Open` callers (2) | `serve` C32 · `create-admin` C33 | - |
| `auth.Lookup` outcomes (3) | user with permissions C35 · user without roles C36 · unknown/expired/deactivated session C36 | - |
| startup assembly of middleware chain (3 places) | `app.New` C7 · C16 · `testkit.NewAPI` C6 · `testkit.NewAPIWithoutDatabase` existing `TestMiddleware_NoDatabase503` - all three call the one `httpx` function | - |
| doors (8) | 1 C1 · C3 · 2 C7 · C12 · 3 C16 · C18 · 4 C14 · C15 · 5 C22 · C23 · 6 C25 · C35 · 7 C30 · C31 · 8 C32 · C33 | - |

- Claims naming a status code, route or response shape: C1, C2, C6, C7, C8, C9, C12, C16, C18, C19, C20, C21 - each proof crosses the HTTP boundary through a real handler chain
- C7, C12, C16, C18-C21 cross `app.New`, the production assembly; the `httpx`-level proofs C8-C11 prove the decision table at its own layer

## Test policy

O `AGENTS.md` (`## Test policy`) já responde as duas perguntas para todas as camadas tocadas; nenhuma linha nova.
Evidência: `platform/httpx` (request id, CSRF, CSP por caminho) decide e é alcançado pela fronteira HTTP -> prova em
`httpx` por linha da tabela (C8-C11, C22, C23) e na fronteira por `app.New` (C7, C16, C18-C21); `auth.SweepSessions`
decide (falha/tick/cancel) sem fronteira -> prova na própria camada (C25-C28); `cmd/api` é entry point -> C14,
C15, C29, C32, C33.

## Swept

- validation: C22, C23 (request id format), C15 (config durations)
- failure modes: C1, C2, C6 (5xx sem vazamento), C27 (varredura falha e segue), C32, C33 (banco fora na subida)
- idempotency: C25 - a varredura é um `DELETE` por idade, repetível; várias réplicas apagam o mesmo conjunto sem efeito duplo
- authorization: C7, C8, C9 (origem); permissões existentes inalteradas - C36 reexecuta `TestMiddleware_Rejects401` e `TestMiddleware_Forbids403`
- concurrency: n/a - nenhum estado compartilhado novo em memória; a varredura concorrente com um login só apaga linhas já expiradas, que o `Lookup` já rejeita
- data lifecycle: C25, C26, C29 (sessões expiradas apagadas)
- dependency failure: C6 (banco falha no `Lookup`), C27 (banco falha na varredura), C32 (banco fora na subida)
- state transitions: n/a - nenhuma entidade ganha estado novo
- observability: C3, C4, C5 (nível e causa no log de acesso), C27 (WARN da varredura), C23 (request id no log)

## Handoff

Intended split, with the arithmetic, written before any code:

- S1-S8 ≈ 50k somados (~34 arquivos, a maioria pequena em `platform/httpx`, `platform/auth`, `cmd/api`, `internal/app`) - abaixo do budget de 150k -> um único builder, sem handoff

- **Settled mid-build:** C30 corrigido antes de ficar verde: `otelhttp` permanece em `app/go.mod` como `// indirect` porque `testcontainers-go` (só testes) o importa; o claim passou de "nenhuma linha `require`" para "nenhum `require` direto". AC 26 (o binário não linka `go.opentelemetry.io/`) não muda. C26 reescrito para o tempo que o teste realmente mede (início com intervalo de 1h, em vez de 15ms). Door 9 acrescentada: o primeiro `task vuln` falhou com 23 vulnerabilidades da stdlib 1.26.2 e uma em `moby/go-archive`.
