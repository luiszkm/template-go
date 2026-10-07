# Foundation

Sources:

- conversation (2026-10-07) - stack, VSA, harness para LLM, escopo users/RBAC/auditoria
- `.specs/STATE.md` AD-001..AD-010 - restrições de projeto que este plano segue
- `AGENTS.md` - regras e comandos que este plano torna reais

## Problem

O repositório não tem código, padrão de slice, forma de erro, setup de testes nem gate. Um agente que começa
`users`, `rbac` ou `audit` hoje decide sozinho a estrutura da pasta, o formato do erro, como registrar a rota e
como testar contra o banco - e o próximo agente copia o que vier primeiro, seja acerto ou acidente. As regras do
`AGENTS.md` (sem import entre features, permissão e auditoria por operação, gerados não editados à mão) são
só texto: nada falha quando são violadas.

Quando isto entra, existe um esqueleto que sobe, um gerador que cria slices no padrão e um único comando
(`task check`) que falha quando qualquer regra do `AGENTS.md` é quebrada.

## Out of scope

| Excluded | Why |
| --- | --- |
| Autenticação, sessões, usuários | feature `users` |
| Checagem real de permissão e gravação de auditoria | features `rbac` e `audit`; aqui só o contrato de registro que as exige |
| Rate limiting e security headers (CSP, HSTS) | responsabilidade do reverse proxy na implantação; entra como feature se o template precisar |
| Deploy, imagem publicada, infraestrutura | o template não presume ambiente |
| i18n | UI só em pt-BR |
| Renomear o template (`task rename`) | só faz sentido depois que o esqueleto existir |

## Assumptions

| Assumption | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Caminho do módulo Go | `github.com/luiszkm/template-go` | definido pelo usuário | y |
| CI | GitHub Actions `.github/workflows/ci.yml` rodando `task check` e `task e2e` | repositório no GitHub (`luiszkm/template-go`); o workflow só chama o Taskfile, então trocar de CI é trocar um arquivo | y |
| Task runner | Taskfile v3 (`task`) instalado globalmente; demais ferramentas Go via diretiva `tool` no `go.mod` | roda igual em Windows, Linux e macOS; `make` não existe no Windows por padrão || y |
| Quando as migrations rodam | comando explícito `api migrate up`; o `serve` nunca migra | duas réplicas migrando no boot disputam o mesmo lock || y |
| Idioma | código, rotas e logs em inglês; textos de UI em pt-BR | identificadores nunca são traduzidos; usuários finais falam português || y |
| E2E no gate | Playwright fica em `task e2e`, fora de `task check` | precisa da stack inteira de pé; `task check` roda a cada mudança || y |
| Hook de Stop do Claude Code | roda `task check:fast` (format, vet, archtest, typecheck), não o gate completo | o gate completo sobe containers e o hook roda a cada resposta || y |
| Git | `git init` com branch `main` no início do build | o build precisa de commits; nada é enviado a remoto || y |

**Open questions:** none - all resolved or logged above.

## Criteria

### S1: O backend sobe, responde saúde e migra (P1)

**Acceptance Criteria**

1. WHEN `GET /healthz` is requested THEN the system SHALL return `200` with body `{"status":"ok"}` without opening a database connection
2. WHEN `GET /readyz` is requested and the database answers `SELECT 1` within 2 seconds THEN the system SHALL return `200` with body `{"status":"ready"}`
3. IF the database does not answer `SELECT 1` within 2 seconds on `GET /readyz` THEN the system SHALL return `503` with an `application/problem+json` body
4. IF `DATABASE_URL` is unset at startup THEN the process SHALL exit with code `1` and write a line containing `DATABASE_URL` to stderr
5. WHEN the process receives an interrupt or SIGTERM THEN the system SHALL stop accepting new connections, let in-flight requests finish for at most 10 seconds, and exit with code `0`
6. WHEN `api migrate up` runs against a database with pending migrations THEN the system SHALL apply every pending file in `app/migrations` and exit with code `0`
7. WHEN `api migrate up` runs against a database with no pending migrations THEN the system SHALL apply nothing and exit with code `0`

**Independent test:** `docker compose up -d db`, `api migrate up`, `api serve`; `/healthz` e `/readyz` retornam `200`; parar o banco e `/readyz` retorna `503`.

### S2: Contrato HTTP compartilhado (P1)

**Acceptance Criteria**

8. The system SHALL return every error response with content type `application/problem+json` and the fields `type`, `title`, `status`, `detail` and `request_id`
9. IF a request path under `/api/` matches no route THEN the system SHALL return `404` with an `application/problem+json` body
10. IF a handler panics THEN the system SHALL return `500` with an `application/problem+json` body whose `detail` contains no stack trace
11. IF a handler panics THEN the system SHALL log one entry at level `ERROR` carrying the same `request_id` returned to the client
12. WHEN a request carries an `X-Request-ID` header THEN the system SHALL return the same value in the response `X-Request-ID` header
13. IF a request carries no `X-Request-ID` header THEN the system SHALL generate a UUID and return it in the response `X-Request-ID` header
14. The system SHALL write one JSON log line per request with the keys `time`, `level`, `msg`, `request_id`, `method`, `path`, `status` and `duration_ms`
15. WHEN `GET /api/openapi.json` is requested THEN the system SHALL return an OpenAPI 3.1 document byte-identical to the committed `app/openapi.json`
16. IF an operation is registered without a permission and without the public marker THEN startup SHALL fail with an error naming the operation ID
17. IF an operation with method `POST`, `PUT`, `PATCH` or `DELETE` is registered without an audit action THEN startup SHALL fail with an error naming the operation ID

**Independent test:** registrar uma operação sem permissão num teste de boot e ver a falha; chamar `/api/nao-existe` e ver `404` problem+json com `request_id` igual ao header.

### S3: Regras de arquitetura executáveis (P1)

**Acceptance Criteria**

18. IF a package under `app/internal/features/<a>` imports a package under `app/internal/features/<b>` with `<a>` different from `<b>` THEN `task check` SHALL exit non-zero and print both import paths
19. IF a package under `app/internal/platform` imports a package under `app/internal/features` THEN `task check` SHALL exit non-zero and print both import paths
20. IF the committed sqlc output differs from what `sqlc generate` produces THEN `task check` SHALL exit non-zero
21. IF the committed `app/openapi.json` differs from the document the code exports THEN `task check` SHALL exit non-zero
22. IF the committed generated TypeScript schema in `web/src/api/` differs from what is generated from `app/openapi.json` THEN `task check` SHALL exit non-zero
23. The `task check` command SHALL run format check, golangci-lint, generated-code diffs, archtest, `go test`, web typecheck, web lint and web unit tests, and SHALL exit non-zero when any step fails

**Independent test:** adicionar um import `features/a -> features/b` e ver `task check` falhar citando os dois caminhos; reverter e ver passar.

### S4: Gerador de slice (P1)

**Acceptance Criteria**

24. WHEN `task new:slice FEATURE=<f> NAME=<n>` runs with valid names THEN the generator SHALL create `app/internal/features/<f>/<n>/` containing `endpoint.go`, `queries.sql` and `<n>_test.go`, and register the slice in `app/internal/features/<f>/register.go`
25. WHEN the generator creates a slice THEN `task check` SHALL exit `0` on the resulting tree without manual edits
26. WHEN the generator creates a slice THEN the generated endpoint SHALL answer `501` with an `application/problem+json` body and the generated test SHALL assert that `501`
27. WHEN `<f>` has no folder yet THEN the generator SHALL create `app/internal/features/<f>/register.go` and register the feature in `app/internal/features/registry.go`
28. IF `app/internal/features/<f>/<n>/` already exists THEN the generator SHALL exit non-zero and modify no file
29. IF `FEATURE` or `NAME` does not match `^[a-z][a-z0-9_]*$` THEN the generator SHALL exit non-zero and create no file

**Independent test:** `task new:slice FEATURE=demo NAME=get_thing`, `task check` verde, o endpoint gerado responde `501`; rodar de novo e ver a recusa.

### S5: Web shell (P1)

**Acceptance Criteria**

30. WHEN `/` loads and `GET /readyz` returns `200` THEN the web app SHALL show the text `API: online`
31. IF `GET /readyz` fails or returns a status other than `200` THEN the web app SHALL show the text `API: offline` and a button labelled `Tentar novamente`
32. WHEN the user clicks `Tentar novamente` THEN the web app SHALL request `GET /readyz` again
33. WHILE the `GET /readyz` request is pending the web app SHALL show an element with `role="status"`
34. IF the browser path matches no web route THEN the web app SHALL show the text `Página não encontrada` and a link to `/`
35. WHEN the binary built by `task build` receives a `GET` for a path outside `/api/`, `/healthz` and `/readyz` that matches no static file THEN the system SHALL return `200` with the SPA `index.html`
36. The web app SHALL call the backend only through the client generated from `app/openapi.json`

**Independent test:** `task dev`, abrir `/` e ver `API: online`; parar o banco e recarregar para ver `API: offline`.

### S6: Harness para agentes (P2)

**Acceptance Criteria**

37. WHEN Claude Code writes or edits a `.go` file THEN the PostToolUse hook SHALL run `gofmt -w` on that file
38. WHEN Claude Code finishes a turn THEN the Stop hook SHALL run `task check:fast` and report a non-zero exit back to the agent
39. The repository SHALL contain `.cursor/rules/agents.mdc` with `alwaysApply: true` and `.windsurf/rules/agents.md` with `trigger: always_on`, each directing the agent to `AGENTS.md`
40. The `.github/workflows/ci.yml` workflow SHALL run `task check` and `task e2e` and fail when either exits non-zero

**Independent test:** editar um `.go` mal formatado pelo Claude Code e ver o arquivo formatado; abrir o projeto no Cursor e ver a regra carregada.

## Traceability

| ID | Slice | Criteria | Status |
| --- | --- | --- | --- |
| FND-01 | S1 | 1, 2, 3, 4, 5, 6, 7 | Pending |
| FND-02 | S2 | 8, 9, 10, 11, 12, 13, 14, 15, 16, 17 | Pending |
| FND-03 | S3 | 18, 19, 20, 21, 22, 23 | Pending |
| FND-04 | S4 | 24, 25, 26, 27, 28, 29 | Pending |
| FND-05 | S5 | 30, 31, 32, 33, 34, 35, 36 | Pending |
| FND-06 | S6 | 37, 38, 39, 40 | Pending |

## Observable

| Surface | Decision | Landing |
| --- | --- | --- |
| API `GET /healthz` | response shape | AC 1 |
| API `GET /healthz` | error shape and codes | n/a - não depende de nada que possa falhar além do próprio processo |
| API `GET /healthz` | who may call it | AC 16 - registrada com o marcador público |
| API `GET /readyz` | response shape | AC 2 |
| API `GET /readyz` | error shape and codes | AC 3, AC 8 |
| API `GET /readyz` | who may call it | AC 16 - registrada com o marcador público |
| API `GET /api/openapi.json` | response shape | AC 15 |
| API `GET /api/openapi.json` | who may call it | n/a - documento público do contrato, sem dado de negócio |
| API `/api/*` | error shape and codes | AC 8, AC 9, AC 10 |
| API `/api/*` | versioning | Landing door 2 - prefixo `/api/v1` |
| API `/api/*` | rate limit behaviour | n/a - fora do escopo, fica no reverse proxy |
| screen `/` | loading state | AC 33 |
| screen `/` | error state | AC 31, AC 32 |
| screen `/` | empty state | n/a - a tela mostra só o status da API, não há coleção que possa estar vazia |
| screen `/` | unauthorised state | n/a - não existe autenticação até a feature `users` |
| screen `*` (rota inexistente) | error state | AC 34 |
| command `task new:slice` | flags and defaults | AC 24, AC 27, AC 29 - `FEATURE` e `NAME` obrigatórios, sem default |
| command `task new:slice` | exit codes and halfway failure | AC 28, AC 29 - valida tudo antes de escrever o primeiro arquivo |
| command `task check` | output and exit codes | AC 23 |
| command `api migrate up` | exit codes | AC 6, AC 7 |
| command `api serve` | failure at startup | AC 4, AC 16, AC 17 |
| document `AGENTS.md` | what the reader does next | existing - seções Workflow e Architecture rules já escritas; este plano torna os comandos reais |

## Flow

Não há nada no repositório para reaproveitar: tudo é novo. O que se reaproveita é o que as bibliotecas já fazem, em vez de reescrever: Problem Details e OpenAPI do Huma, migrations do goose, geração do sqlc.

Requisição HTTP:

1. request -> `platform/httpx` (new, door 4) - request id, log JSON, recover de panic
2. `platform/op` (new, door 5) - registro único de operações; recusa no boot operação sem permissão/marcador público, ou mutação sem ação de auditoria
3. `features/<f>/<slice>` (new, door 6) - handler do slice; acessa o banco via código gerado por sqlc no próprio slice
4. `platform/db` (new, door 8) - pool `pgx` e `WithTx`
5. out: JSON do slice ou `application/problem+json` (door 3); fora de `/api/`, `platform/webui` (new, door 9) entrega o SPA

Gerador e gate:

6. `task new:slice` -> `app/cmd/newslice` (new, door 6) - valida nomes, escreve slice, `register.go`, entrada no `sqlc.yaml`, regenera sqlc e OpenAPI
7. `task check` -> `app/archtest` (new, door 7), golangci-lint, diffs de gerados, testes Go e web; sai não-zero na primeira falha

## Relations

None - nenhuma entidade de domínio; o único dado persistido é a tabela de controle de versão do próprio goose.

## Surface

| Route | In | Out | Status |
| --- | --- | --- | --- |
| `GET /healthz` | nada | `status` | `200` |
| `GET /readyz` | nada | `status` · problem+json | `200`, `503` |
| `GET /api/openapi.json` | nada | documento OpenAPI 3.1 | `200` |
| `ANY /api/*` sem rota | qualquer | problem+json | `404` |
| `GET /*` fora de `/api/` | path do SPA | `index.html` ou arquivo estático | `200` |

## Landing

| One-way door | Literal shape | Alternative rejected |
| --- | --- | --- |
| 1. layout de módulos | um módulo Go em `app/go.mod` (`github.com/luiszkm/template-go`); `web/` é um pacote npm separado; ferramentas Go (`sqlc`, `goose`, `golangci-lint`) via diretiva `tool` no `go.mod` | `go.mod` na raiz - mistura o toolchain Go com `web/` e `node_modules` |
| 2. prefixo e versão da API | rotas de feature em `/api/v1/<feature>/...`; `/healthz` e `/readyz` na raiz; contrato em `/api/openapi.json`, docs em `/api/docs` | sem versão no path - a primeira quebra de contrato não tem para onde ir |
| 3. forma do erro | RFC 9457 `{type, title, status, detail, instance, request_id, errors?}`; `errors` é a lista `{location, message, value}` do Huma para validação | `{error:{code,message}}` próprio - nenhuma ferramenta conhece, e cada handler inventa um campo |
| 4. middleware HTTP | cadeia fixa `requestid -> slog logger -> recover` em `platform/httpx`, header `X-Request-ID`, UUID gerado quando ausente | middleware por feature - cada feature loga diferente e o `request_id` se perde |
| 5. contrato de registro de operação | `op.Register(api, op.Spec{ID, Method, Path, Permission or Public: true, AuditAction}, handler)`; boot falha sem permissão/marcador público, ou mutação sem `AuditAction` | `huma.Register` direto - esquecer permissão ou auditoria passa em silêncio |
| 6. layout do slice | `features/<f>/<slice>/{endpoint.go, queries.sql, db/ (sqlc), <slice>_test.go}`, `features/<f>/register.go`, `features/<f>/permissions.go`, `features/registry.go`; uma entrada no `sqlc.yaml` por slice | camadas `handlers/ services/ repositories/` - uma mudança atravessa três pastas e acopla features |
| 7. regra de imports | `app/archtest` usando `golang.org/x/tools/go/packages`: `features/<a>` nunca importa `features/<b>`, `platform` nunca importa `features` | só `depguard` - não expressa "outra feature que não a própria" |
| 8. migrations | goose SQL em `app/migrations/YYYYMMDDHHMMSS_<nome>.sql`, embutidas no binário, aplicadas por `api migrate up` | numeração sequencial - branches paralelas colidem no mesmo número; migrar no boot - réplicas disputam o lock |
| 9. entrega do front | build do Vite copiado para `app/internal/platform/webui/dist` e servido por `embed` com fallback para `index.html`; em dev, proxy do Vite de `/api`, `/healthz` e `/readyz` para `:8080` | nginx separado - quebra o same-origin que o cookie de sessão (AD-005) exige |
| 10. dependências novas | Go: `huma/v2`, `pgx/v5`, `goose/v3`, `sqlc`, `caarlos0/env`, `testify`, `testcontainers-go`, `otel`; web: `react@19`, `vite`, `@tanstack/react-router`, `@tanstack/react-query`, `tailwindcss@4`, shadcn/ui, `zod`, `react-hook-form`, `openapi-typescript`, `openapi-fetch`, `vitest`, `@testing-library/react`, `@playwright/test`, `@biomejs/biome` | alternativas já rejeitadas em AD-002, AD-003 e AD-008 |
| 11. raiz de composição | `app/internal/app.New(cfg)` monta config, pool, middleware, `op` e `features/registry.go`; é o único pacote, junto com `app/cmd/...`, autorizado a importar `features`; `cmd/api` e os testes montam o servidor pela mesma função | montar em `platform` - violaria a regra 19; montar só em `main` - o teste montaria outro servidor e provaria a montagem errada |

- Nothing else in this change is hard to reverse

## Impact

| Front | What changes |
| --- | --- |
| domain | new term: `feature` - pasta de topo em `app/internal/features/<f>`, unidade de isolamento de imports |
| domain | new term: `slice` - um caso de uso, uma pasta dentro de uma feature, com endpoint, SQL e teste |
| domain | new term: `op.Spec` - declaração de uma operação HTTP com `Permission`/`Public` e `AuditAction`, consumida depois por `rbac` e `audit` |
| stored data | nothing to migrate - banco vazio |
| docs | `AGENTS.md` passa a citar comandos que existem; `.specs/STATE.md` ganha o handoff da foundation |
