# Hardening

Sources:

- conversation (2026-10-09) - revisão do backend Go: sete pontos (vazamento de erro 5xx, CSRF, timeouts do servidor, cabeçalhos de segurança, `X-Request-ID`, limpeza de sessões, OTel) e quatro menores (ping do pool, aquecimento do hash fictício, `govulncheck`, `auth.Lookup` em uma consulta); o usuário pediu para implementar todos
- conversation (2026-10-09) - decisões do usuário: remover `otelhttp` em vez de configurar o SDK; `govulncheck` dentro do `task check`
- `.specs/STATE.md` AD-004, AD-005, AD-009 - forma do erro, sessão revogável no servidor, gate único
- `.specs/features/users/plan.md` - doors 4 (sessão), 9 (IP do cliente) e 10 (rate limit); assumption de CSRF que este plano endurece
- `huma/v2@v2.39.1` `huma.go:1243`, `error.go:265` - todo erro comum de handler vira `NewErrorWithContext(ctx, 500, "unexpected error occurred", err)`; `api.go:699` - a página `/api/docs` define o próprio `Content-Security-Policy`
- Go 1.26 `net/http/csrf.go` - `http.CrossOriginProtection` (`Handler`, `SetDenyHandler`, `Check`)

## Problem

Quando um handler devolve um erro comum, o texto do erro vai para `errors[].message` da resposta `500`
(`httpx.InstallProblems` copia `err.Error()`), e nada no servidor registra a causa: o log de acesso mostra só
`status: 500`. O cliente lê `db: begin: failed to connect to host=... user=...`; o operador não lê nada.

Além disso, o servidor HTTP não tem timeout de leitura, escrita nem ociosidade; a proteção contra CSRF depende só
de `SameSite=Lax`, que não barra outro subdomínio do mesmo site; nenhuma resposta leva cabeçalhos de segurança
(CSP, `nosniff`, anti-framing, HSTS); o `X-Request-ID` do cliente é aceito sem limite e vai para log, cabeçalho e
`audit_events`; sessões expiradas nunca são apagadas; `otelhttp` instrumenta cada requisição sem ter para onde
exportar; `serve` sobe sem saber se o banco responde; a primeira tentativa de login com e-mail inexistente paga o
custo do hash fictício; não há verificação de vulnerabilidades conhecidas nas dependências; cada requisição
autenticada faz duas consultas para montar o `Principal`.

Não há evidência numérica: é um template sem tráfego real. O custo é de quem copiar o template para produção.

Quando isto entra, um 5xx devolve só `internal server error` e o log de acesso daquela requisição sai em `ERROR`
com a causa; mutações vindas de outra origem recebem `403`; o servidor fecha conexões lentas; o navegador recebe
CSP e anti-framing; o gate falha numa dependência vulnerável.

## Out of scope

| Excluded | Why |
| --- | --- |
| Limpeza periódica de `login_attempts` | já é limitada: toda falha registrada apaga as linhas fora da janela de 15 min (users door 10); o que sobra é no máximo uma janela |
| Configurar o SDK OpenTelemetry (traces/métricas) | decisão do usuário: remover; nenhum coletor existe no compose |
| Origens confiáveis extras para CSRF (`TRUSTED_ORIGINS`) | SPA e API são o mesmo origin (foundation door 9); o proxy do Vite chega como `same-origin` |
| Expiração deslizante de sessão | já fora de escopo em `users` |
| CSP com nonce ou sem `'unsafe-inline'` em `style-src` | o Radix (`react-remove-scroll`) injeta elementos `<style>` em tempo de execução |
| Rate limit fora do login | nenhum endpoint além do login é alvo de força bruta hoje |

## Assumptions

| Assumption | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| OpenTelemetry | remover `otelhttp` e as dependências diretas de `go.opentelemetry.io` | sem provider configurado ele só custa; reintroduzir é uma linha | y |
| `govulncheck` | roda dentro de `task check`, instalado como `tool` no módulo `app/tools` | gate único (AD-009); exige rede | y |
| Texto do 5xx | `detail` = `internal server error` (o mesmo do `Recover`), sem `errors` | um único texto para toda falha interna | n |
| Nível do log de 5xx | a linha de acesso de toda resposta `>= 500` sai em `ERROR`; com `error` quando houver causa | um 5xx é sempre algo para o operador olhar; um campo só, sem segunda linha | n |
| Valores dos timeouts | `HTTP_READ_TIMEOUT=30s`, `HTTP_WRITE_TIMEOUT=30s`, `HTTP_IDLE_TIMEOUT=120s`; `ReadHeaderTimeout` continua `10s` | cobre o argon2 com folga; nenhum endpoint faz streaming | n |
| CSP do SPA | `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'` | o build do Vite só emite arquivos `'self'`; `'unsafe-inline'` em estilo pelo Radix | n |
| CSP da API | `default-src 'none'; frame-ancestors 'none'` em `/api/*` e `/healthz`/`/readyz`; `/api/docs` mantém a CSP que o Huma define | JSON não carrega nada; a página de docs precisa do unpkg | n |
| HSTS | `Strict-Transport-Security: max-age=31536000` só com `COOKIE_SECURE=true`; sem `includeSubDomains` | `COOKIE_SECURE` já é o sinal de "atrás de HTTPS"; `includeSubDomains` afetaria domínios que o template não conhece | n |
| Cabeçalhos fixos | `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin` em toda resposta | valores consensuais (OWASP Secure Headers) | n |
| Formato do `X-Request-ID` | `^[A-Za-z0-9._-]{1,64}$`; fora disso, gera UUID e ignora o recebido | cobre UUID, ULID e ids de proxies; impede injeção em log e linhas enormes | n |
| Intervalo da limpeza de sessões | `SESSION_SWEEP_INTERVAL=1h`, primeira execução ao subir | sessão expirada já é rejeitada pelo `Lookup`; a limpeza é só higiene | n |
| Réplicas e limpeza | toda réplica limpa; o `DELETE` é idempotente | sem lock nem eleição de líder | n |
| Ping na subida | `db.Open` faz `Ping`; falha encerra `serve`, `migrate up` não usa o pool | falhar cedo é mais claro que `readyz` vermelho indefinidamente; orquestrador reinicia | n |
| Aquecimento do hash fictício | o registro da operação de login roda uma verificação contra o hash fictício | aquece o hash e o caminho do argon2 antes da primeira requisição | n |
| `403` de CSRF no contrato | toda operação `POST`/`PUT`/`PATCH`/`DELETE` documenta `403` no OpenAPI | o contrato diz a verdade; o cliente TS passa a tipar o caso | n |

**Open questions:** none - all resolved or logged above.

## Criteria

### S1: 5xx não vaza a causa e a causa vai para o log (P1)

**Acceptance Criteria**

1. IF a handler returns an error that is not a Huma `StatusError` THEN the system SHALL answer `500` `application/problem+json` with `detail` `internal server error` and no `errors` member
2. The system SHALL never include an `errors` member in a problem+json response whose `status` is `>= 500`
3. WHEN a response with status `>= 500` is produced from an error carrying a cause THEN the access log entry of that request SHALL have level `ERROR`, its `request_id`, and an `error` attribute holding the cause text
4. WHEN a response has status `>= 500` and no cause THEN the access log entry SHALL have level `ERROR` and no `error` attribute
5. WHEN a response has status `< 500` THEN the access log entry SHALL have level `INFO` and the problem body SHALL keep its `errors` entries unchanged
6. IF the session lookup fails with a database error THEN the system SHALL answer `500` per AC 1 and log its cause per AC 3

**Independent test:** derrubar o Postgres com o servidor rodando, chamar `GET /api/v1/users/me` com cookie: corpo sem texto do pgx, linha `ERROR` no log com `error` e o mesmo `request_id` do cabeçalho.

### S2: Mutação de outra origem é recusada (P1)

**Acceptance Criteria**

7. IF a `POST`, `PUT`, `PATCH` or `DELETE` request carries `Sec-Fetch-Site` with a value other than `same-origin` or `none` THEN the system SHALL answer `403` problem+json with `detail` `cross-origin request rejected` and SHALL NOT run the handler
8. IF a `POST`, `PUT`, `PATCH` or `DELETE` request has no `Sec-Fetch-Site` and an `Origin` whose host differs from the `Host` header THEN the system SHALL answer `403` per AC 7
9. WHEN a mutating request carries `Sec-Fetch-Site: same-origin` or `none`, or neither `Sec-Fetch-Site` nor `Origin` THEN the system SHALL process it normally
10. WHEN a `GET` or `HEAD` request carries `Sec-Fetch-Site: cross-site` THEN the system SHALL process it normally
11. The OpenAPI document SHALL list `403` among the responses of every `POST`, `PUT`, `PATCH` and `DELETE` operation

**Independent test:** `curl -X POST -H 'Sec-Fetch-Site: cross-site' .../api/v1/users/session` responde `403` problem+json; o mesmo com `same-origin` chega ao handler (`422` sem corpo válido).

### S3: Timeouts do servidor HTTP (P2)

**Acceptance Criteria**

12. WHEN `api serve` builds its `http.Server` THEN it SHALL set `ReadTimeout` from `HTTP_READ_TIMEOUT` (default `30s`), `WriteTimeout` from `HTTP_WRITE_TIMEOUT` (default `30s`), `IdleTimeout` from `HTTP_IDLE_TIMEOUT` (default `120s`) and `ReadHeaderTimeout` `10s`
13. IF `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT` or `SESSION_SWEEP_INTERVAL` is not a Go duration THEN `api serve` SHALL exit `1` with a stderr line starting with `config:`

**Independent test:** `HTTP_READ_TIMEOUT=abc api serve` sai `1` com `config:`; teste de unidade lê os campos do `http.Server` montado.

### S4: Cabeçalhos de segurança (P1)

**Acceptance Criteria**

14. The system SHALL send `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY` and `Referrer-Policy: strict-origin-when-cross-origin` on every response, including problem+json and 404
15. WHEN the web UI handler answers a path outside `/api/` THEN the response SHALL carry `Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'`
16. WHEN a request to `/api/` other than `/api/docs`, or to `/healthz` or `/readyz`, is answered THEN the response SHALL carry `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'`
17. WHEN `/api/docs` is answered THEN the response SHALL carry the `Content-Security-Policy` Huma sets for its docs renderer, which contains `https://unpkg.com/`
18. WHERE `COOKIE_SECURE=true` the system SHALL send `Strict-Transport-Security: max-age=31536000` on every response
19. WHILE `COOKIE_SECURE=false` the system SHALL send no `Strict-Transport-Security` header

**Independent test:** `curl -I /`, `/api/v1/users/me`, `/api/docs` e `/api/nope` mostram os cabeçalhos esperados; com `COOKIE_SECURE=false` o HSTS some.

### S5: `X-Request-ID` validado (P2)

**Acceptance Criteria**

20. WHEN the incoming `X-Request-ID` matches `^[A-Za-z0-9._-]{1,64}$` THEN the system SHALL echo it unchanged in the response header and use it as the request id
21. IF the incoming `X-Request-ID` is longer than 64 characters or contains a character outside `[A-Za-z0-9._-]` THEN the system SHALL use a newly generated UUID as the request id in the response header, the access log, the problem body and the audit event

**Independent test:** `curl -H 'X-Request-ID: a b'` devolve um UUID no cabeçalho; `-H 'X-Request-ID: abc-123'` ecoa.

### S6: Sessões expiradas são apagadas (P2)

**Acceptance Criteria**

22. WHEN `api serve` starts, and every `SESSION_SWEEP_INTERVAL` after (default `1h`) THEN the system SHALL delete every `sessions` row created more than `SESSION_TTL` ago
23. The sweep SHALL NOT delete a session created less than `SESSION_TTL` ago
24. IF a sweep fails THEN the system SHALL log at `WARN` with an `error` attribute and run again at the next interval
25. WHEN the `serve` context is cancelled THEN the sweep loop SHALL return before `serve` returns

**Independent test:** inserir uma sessão com `created_at = now() - interval '13 hours'` e outra nova; rodar uma varredura com TTL `12h`; só a antiga some.

### S7: Dependências e gate (P3)

**Acceptance Criteria**

26. The `api` binary SHALL link no package under `go.opentelemetry.io/` (`go list -deps ./cmd/api` lists none)
27. WHEN `task check` runs THEN it SHALL run `govulncheck ./...` over the `app` module and fail when it reports a vulnerability

**Independent test:** `go list -deps ./cmd/api | grep -c opentelemetry` imprime `0`; `task vuln` sai `0`.

### S8: Subida e caminho quente de autenticação (P3)

**Acceptance Criteria**

28. IF the database does not answer a ping when `api serve` or `api users create-admin` opens the pool THEN the command SHALL exit `1` before listening or writing, with the connection error on stderr
29. WHEN the login operation is registered THEN the system SHALL run exactly one password verification against the dummy hash before serving any request
30. WHEN an authenticated request is checked THEN the system SHALL resolve the session, the user's active state and the user's permissions with exactly one SQL statement
31. WHEN a valid session belongs to a user holding no role THEN the lookup SHALL return a principal with an empty permission set (the request proceeds to the `403` check, not `401`)

**Independent test:** `DATABASE_URL` apontando para porta fechada: `api serve` sai `1` sem logar `listening`; um `Querier` contador em volta do pool conta 1 chamada por `Lookup`.

## Traceability

| ID | Slice | Criteria | Status |
| --- | --- | --- | --- |
| HARD-01 | S1 | 1, 2, 3, 4, 5, 6 | Implementing |
| HARD-02 | S2 | 7, 8, 9, 10, 11 | Implementing |
| HARD-03 | S3 | 12, 13 | Implementing |
| HARD-04 | S4 | 14, 15, 16, 17, 18, 19 | Implementing |
| HARD-05 | S5 | 20, 21 | Implementing |
| HARD-06 | S6 | 22, 23, 24, 25 | Implementing |
| HARD-07 | S7 | 26, 27 | Implementing |
| HARD-08 | S8 | 28, 29, 30, 31 | Implementing |

## Observable

| Surface | Decision | Landing |
| --- | --- | --- |
| API (todas as rotas) | error shape and codes | AC 1, AC 2, AC 6 - forma é a foundation door 3; AC 7, AC 8 acrescentam `403` às mutações |
| API (todas as rotas) | response shape | AC 14, AC 15, AC 16, AC 17, AC 18, AC 19 - só cabeçalhos; corpos inalterados |
| API (todas as rotas) | who may call it | AC 7, AC 8, AC 9, AC 10 - origem; permissões inalteradas (users door 5) |
| API (todas as rotas) | versioning | n/a - nenhum campo de corpo muda; `403` novo é aditivo no OpenAPI (AC 11) |
| API (todas as rotas) | rate limit behaviour | existing - users AC 12, AC 13; nada muda |
| API cabeçalho `X-Request-ID` | formato aceito | AC 20, AC 21 |
| command `api serve` | flags and defaults | AC 12, AC 13, AC 22 - variáveis de ambiente novas com default |
| command `api serve` | exit codes | AC 13, AC 28 |
| command `api serve` | what it prints when it fails halfway | AC 28 - erro de conexão no stderr, nada escutando; AC 24 - varredura falha loga e continua |
| command `api serve` | output format | AC 3, AC 4, AC 5, AC 24 - log JSON; muda o nível das linhas 5xx |
| command `api users create-admin` | exit codes | AC 28 |
| command `task check` | output and exit codes | AC 27 |
| scheduled task varredura de sessões | output format and verbosity | AC 24 - só loga em falha |
| scheduled task varredura de sessões | what it prints when it fails halfway | AC 24 |
| scheduled task varredura de sessões | flags and defaults | AC 22 |
| screen (web) | empty, loading, error states | n/a - nenhuma tela muda; a CSP (AC 15) é verificada pelo e2e existente carregando as telas |

## Flow

Reaproveita a cadeia `platform/httpx` (request id, log de acesso, recover), o hook de erro do Huma que
`httpx.InstallProblems` já sobrescreve, o `Querier` de `platform/auth`, o `http.CrossOriginProtection` da
biblioteca padrão e o módulo `app/tools` para o `govulncheck`. Nada reimplementa Problem Details.

Requisição:

1. request -> `platform/httpx.SecurityHeaders` (door 3) - grava os cabeçalhos fixos, HSTS se seguro e a CSP pelo prefixo do caminho; o handler pode sobrescrever (docs do Huma)
2. `platform/httpx.WithClientIP` (exists) - IP do cliente
3. `platform/httpx.RequestID` (exists, door 5 restringe) - aceita o id recebido só no formato; senão UUID
4. `platform/httpx.AccessLog` (exists, door 1 muda) - põe no contexto um registro de causa vazio; ao final loga `INFO` ou `ERROR` + `error`
5. `platform/httpx.Recover` (exists) - um pânico vira a causa (com `stack`) da linha de acesso, mantendo uma única linha `ERROR` por pânico (foundation C14); sem `AccessLog` na cadeia, loga a própria linha como antes
6. `http.CrossOriginProtection` (door 2, em volta do mux) - mutação de outra origem -> `403` problem+json pelo deny handler
7. mux -> Huma -> `platform/auth` (exists, door 6 muda a consulta) - uma consulta resolve sessão, usuário ativo e permissões; erro de banco -> `huma.WriteErr(..., 500, ..., err)`
8. handler do slice (exists) - erro comum -> Huma -> `huma.NewErrorWithContext` sobrescrito por `httpx` (door 1): status `>= 500` guarda a causa no registro do contexto e devolve problem sem `errors`
9. out: resposta com cabeçalhos; linha de log de acesso

Processo `api serve`:

10. `cmd/api` (exists) -> `platform/config` (exists, door 4 acrescenta variáveis) -> `platform/db.Open` (exists, door 8: `Ping`) -> `internal/app.New` (exists) -> `features/users/login` (exists, AC 29: verificação de aquecimento no registro)
11. `cmd/api serve` inicia `auth.SweepSessions` (door 6) numa goroutine ligada ao contexto do `serve`, monta o `http.Server` com timeouts (door 4) e espera a goroutine terminar antes de retornar

Gate:

12. `task check` -> `task vuln` (door 7) -> `go tool govulncheck ./...` no módulo `app`

## Relations

None - no stored-data shape change. A varredura apaga linhas de `sessions` que o `Lookup` já trata como inexistentes.

## Surface

| Route | In | Out | Status |
| --- | --- | --- | --- |
| `POST`/`PUT`/`PATCH`/`DELETE /api/v1/*` (toda mutação) | cabeçalhos `Sec-Fetch-Site`, `Origin`, `Host` | problem+json `detail: cross-origin request rejected` | `403` acrescentado aos status já existentes de cada rota |
| qualquer rota da API | erro interno | problem+json `detail: internal server error`, sem `errors` | `500` |
| qualquer rota | `X-Request-ID` | `X-Request-ID` (ecoado ou UUID) · `X-Content-Type-Options` · `X-Frame-Options` · `Referrer-Policy` · `Content-Security-Policy` · `Strict-Transport-Security` | `200`, `204`, `404`, `500` (amostra: cabeçalhos independem do status) |

Operador (não é API, forma literal na door 4): `api serve` lê `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`,
`HTTP_IDLE_TIMEOUT`, `SESSION_SWEEP_INTERVAL`; exit `1` em config inválida ou banco sem resposta.

## Landing

| One-way door | Literal shape | Alternative rejected |
| --- | --- | --- |
| 1. causa de 5xx e forma do log | `httpx.NewAPI` sobrescreve `huma.NewErrorWithContext`: status `>= 500` -> `Problem{detail: "internal server error"}` sem `errors`, causa gravada em `httpx.causeSlot` posto no contexto por `AccessLog`; `AccessLog` loga `level=ERROR` + `error=<causa>` quando `status >= 500`; `auth.Install` passa o erro em `huma.WriteErr(api, ctx, 500, "internal server error", err)` | logger global em `huma.NewError` - os testes sobem vários `app.New` em paralelo e o hook é global do pacote; log em cada handler - todo slice repete e o esquecido vaza em silêncio |
| 2. defesa CSRF | `cop := http.NewCrossOriginProtection(); cop.SetDenyHandler(<httpx.WriteProblem 403 "cross-origin request rejected">)`, envolvendo o `mux` em `app.New`; nenhuma origem confiável configurada; `op.documentedErrors` acrescenta `403` a todo método mutante | token CSRF double-submit - exige mudar o front e guardar segredo por sessão; `gorilla/csrf` - dependência nova para o que a biblioteca padrão faz |
| 3. cabeçalhos de segurança | `httpx.SecurityHeaders(next http.Handler, hsts bool) http.Handler`, mais externo da cadeia, grava antes de chamar `next` com os valores literais de AC 14-18; CSP escolhida por prefixo: `/api/` e `/healthz`/`/readyz` -> `default-src 'none'; frame-ancestors 'none'`, demais -> CSP do SPA; Huma sobrescreve em `/api/docs` | CSP só no `webui.Handler` - 404 e problem+json fora do SPA ficariam sem cabeçalho; `style-src 'self'` estrito - quebra o diálogo do Radix |
| 4. contrato de operador | `config.Config` ganha `HTTPReadTimeout env:"HTTP_READ_TIMEOUT" envDefault:"30s"`, `HTTPWriteTimeout env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`, `HTTPIdleTimeout env:"HTTP_IDLE_TIMEOUT" envDefault:"120s"`, `SessionSweepInterval env:"SESSION_SWEEP_INTERVAL" envDefault:"1h"` | constantes no código - quem roda atrás de proxy com timeout diferente precisaria recompilar |
| 5. formato do request id | `^[A-Za-z0-9._-]{1,64}$`; fora disso `uuid.NewString()`, o recebido é descartado | truncar ou escapar o recebido - a correlação com o proxy quebra em silêncio de qualquer jeito, e o valor parcial engana |
| 6. primeiro job em segundo plano e consulta de sessão | `auth.SweepSessions(ctx context.Context, q auth.Querier, ttl, every time.Duration, log *slog.Logger)` bloqueia até `ctx.Done()`, roda já e a cada `every` `DELETE FROM sessions WHERE created_at <= now() - make_interval(secs => $1)`; iniciado por `cmd/api serve`, que espera seu retorno; `auth.Lookup` em uma consulta `SELECT s.user_id, coalesce(array_agg(DISTINCT rp.permission) FILTER (WHERE rp.permission IS NOT NULL), '{}') ... LEFT JOIN user_roles ... LEFT JOIN role_permissions ... GROUP BY s.user_id` | `pg_cron` - extensão que o Postgres do template não traz; apagar no login - deixa as sessões de quem nunca volta; fila/worker separado - processo novo para um `DELETE` |
| 7. dependências | remove `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` de `app/go.mod`; acrescenta `tool golang.org/x/vuln/cmd/govulncheck` em `app/tools/go.mod`; `task vuln` (`go tool -modfile=tools/go.mod govulncheck ./...`) entra em `task check` | configurar o SDK OTLP - decisão do usuário; `govulncheck` só no CI - decisão do usuário (AD-009) |
| 8. ping na abertura do pool | `db.Open` chama `pool.Ping(ctx)` e fecha o pool se falhar, devolvendo `db: ping: <err>` | ping só no `serve` - `create-admin` falharia mais tarde, no meio da transação |

| 9. versão do Go e dependência vulnerável (descoberta no build: o primeiro `govulncheck` reportou 23 vulnerabilidades alcançáveis na stdlib 1.26.2 e GO-2026-6253 em `moby/go-archive@v0.2.0`) | `go 1.26.9` em `app/go.mod` e `app/tools/go.mod` (o `GOTOOLCHAIN=auto` baixa a versão; CI lê `go-version-file`); `github.com/moby/go-archive v0.3.0` | excluir os achados do gate (`-exclude`/allowlist) - esconde exatamente o que AC 27 existe para expor; ficar em 1.26.2 - o gate falha |

- O padrão "job em segundo plano dentro do `serve`" (door 6) é precedente para as próximas features: vira AD-012 em `.specs/STATE.md`.
- Nothing else in this change is hard to reverse.

## Impact

| Front | What changes |
| --- | --- |
| domain | termo existente `request_id`: era qualquer string recebida, agora só o formato da door 5 - quem lê `audit_events.request_id` e o log passa a ver UUID onde antes via o valor bruto inválido |
| logs | linhas de acesso com status `>= 500` passam de `INFO` para `ERROR`, com `error` - alertas ou dashboards que filtram por nível passam a disparar nesses casos |
| API | `detail` de 5xx de handler muda de `unexpected error occurred` para `internal server error` e perde `errors`; o web não lê esse texto (mostra mensagem própria por status) |
| API | toda mutação documenta `403`; `app/openapi.json` e `web/src/api/schema.d.ts` são regenerados |
| clientes da API | um cliente não-navegador que envie `Origin` de outro host numa mutação passa a receber `403`; `curl` e testes sem `Origin` seguem aceitos |
| testes | testes que montam `httpx.Chain`/`app.New` passam pelos cabeçalhos e pela proteção de origem; nenhum envia `Sec-Fetch-Site` hoje |
| stored data | nada a migrar; a primeira varredura apaga sessões já expiradas, que o `Lookup` já rejeitava |
| dependências | sai `otelhttp` (e transitivas só dele); entra `govulncheck` no módulo de ferramentas |
