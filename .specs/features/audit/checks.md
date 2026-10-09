# Audit checks

Profile: standard
Plan: `.specs/features/audit/plan.md`

## Intent

38 checks in 9 slices · 4 one-way doors · 0 open

Comandos reais do repositório: `go -C app test <pkg> -run '<regex>'`, `npm --prefix web run test -- <file> -t "<name>"`
e `npm --prefix web run e2e -- <spec>`. Testes Go que tocam o banco usam `testkit.MigratedDB` (Postgres real). Os
slices de `audit` não importam `platform/audit` nem outra feature: os testes inserem linhas em `audit_events`
direto por SQL. Os testes web rodam com `TZ=America/Sao_Paulo` (fixado na configuração do Vitest) para que as
datas exibidas e os limites de `De`/`Até` tenham valor exato. Lição confirmada aplicada: L-010 (todo ramo de
componente web que decide tem caso afirmado).

## Checks

### S1 - Listar eventos · ~6 files · ~25 KB · ~6k

**C1** - With events `e1 < e2 < e3` by `id`, `GET /api/v1/audit/events` by a user holding `audit:read` returns `200` with `items` in the order `e3, e2, e1`; each item has exactly the keys `id`, `occurred_at`, `action`, `actor`, `resource_type`, `resource_id`, `ip`, `request_id` (AUD-01, AC 1) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_NewestFirstWithoutDiff$'`

**C2** - An event whose actor is user `a@x.com` returns `actor: {id: <a id>, email: "a@x.com"}`, and after that user's email changes to `b@x.com` it returns `b@x.com`; an event with a null actor returns `actor: null` (AUD-01, AC 2) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_Actor$'`

**C3** - With 51 matching events and no `limit`, the response has 50 items and `next` equal to the `id` of the 50th; with 3 events and `limit=2`, `next` is the `id` of the second item; with `limit=3` on 3 events, `next` is `null` (AUD-01, AC 3, door 1) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_Pages$'`

**C4** - With events `e1 < e2 < e3`, `before=<e3 id>` returns exactly `e2, e1`, and `before=<e1 id>` returns `items: []` with `next: null` (AUD-01, AC 4, door 1) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_Before$'`

**C5** - `GET /api/v1/audit/events` returns `422` with an `errors` entry whose `location` names the parameter for each of 7 inputs: `limit=0`, `limit=101`, `before=0`, `from=ontem`, `to=ontem`, `actor_id=abc`, `resource_id=x` without `resource_type`; it accepts `limit=1` and `limit=100` (AUD-01, AC 5) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_Validation$'`

### S2 - Filtrar eventos · ~4 files · ~20 KB · ~5k

**C6** - With events of actions `user.created`, `user.updated` and `user.created`, `action=user.created` returns exactly the two `user.created` events (AUD-02, AC 6) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_FilterAction$'`

**C7** - With events by actors `A`, `B` and none, `actor_id=<A>` returns exactly the events by `A` (AUD-02, AC 7) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_FilterActor$'`

**C8** - With events on `user/1`, `user/2` and `role/1`, `resource_type=user` returns the two user events, and `resource_type=user&resource_id=1` returns only the `user/1` event (AUD-02, AC 8) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_FilterResource$'`

**C9** - With events at `10:00`, `11:00` and `12:00` UTC of one day, `from=<11:00>` returns the `11:00` and `12:00` events, `to=<12:00>` returns the `10:00` and `11:00` events, and both together return only `11:00` (AUD-02, AC 9) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_FilterPeriod$'`

**C10** - With 3 `user.created` events interleaved with 3 other events, `action=user.created&limit=2` returns the two newest `user.created` with `next`, and `before=<next>` with the same filter returns the third and `next: null` (AUD-02, AC 10) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_FiltersCombineAndPage$'`

**C11** - `op.AuditActions` over operations declaring `b.done`, `a.done`, `b.done` and one `GET` with no action returns `["a.done", "b.done"]`; `GET /api/v1/audit/actions` returns the same list as `items`; on the assembled server the list contains `user.created`, `session.created`, `role.created` and `user.roles_changed` (AUD-02, AC 11, door 3) `[done]`
Proof: `go -C app test ./internal/platform/op -run '^TestAuditActions_SortedDistinct$'`
Proof: `go -C app test ./internal/features/audit/list_actions -run '^TestListActions_ReturnsCatalogue$'`
Proof: `go -C app test ./internal/app -run '^TestAudit_ActionsOfAssembledServer$'`

### S3 - Ver um evento · ~3 files · ~10 KB · ~3k

**C12** - `GET /api/v1/audit/events/{id}` of an event with `before {"name":"Ana"}` and `after {"name":"Bia"}` returns `200` with the list fields plus those two objects; an event with null `before` returns `before: null` (AUD-03, AC 12) `[done]`
Proof: `go -C app test ./internal/features/audit/get_event -run '^TestGetEvent_ReturnsDiff$'`

**C13** - An unused id returns `404`; `abc`, `0` and `-1` return `422` (AUD-03, AC 13) `[done]`
Proof: `go -C app test ./internal/features/audit/get_event -run '^TestGetEvent_404And422$'`

### S4 - Acesso · ~2 files · ~8 KB · ~2k

**C14** - Table-driven over the 3 audit operations of the assembled server: each declares exactly `audit:read`, answers `401` without a cookie and `403` to a user holding every `users:*` and `rbac:*` permission but not `audit:read` (AUD-04, AC 14, AC 15, door 4) `[done]`
Proof: `go -C app test ./internal/app -run '^TestAudit_OperationAccess$'`

**C15** - The assembled OpenAPI has no operation with method `POST`, `PUT`, `PATCH` or `DELETE` under `/api/v1/audit` (AUD-04, AC 16) `[done]`
Proof: `go -C app test ./internal/app -run '^TestAudit_ReadOnly$'`

### S5 - Esquema, contrato e regras · ~4 files · ~15 KB · ~4k

**C16** - After `migrate up`, `pg_indexes` lists on `audit_events` the indexes `audit_events_actor_idx` on `(actor_id, id DESC)`, `audit_events_resource_idx` on `(resource_type, resource_id, id DESC)`, `audit_events_action_idx` on `(action, id DESC)` and `audit_events_occurred_idx` on `(occurred_at)`, each asserted against its `indexdef` (door 2; columns added in verification round 1) `[done]`
Proof: `go -C app test ./migrations -run '^TestSchema_AuditIndexes$'`

**C17** - The committed `app/openapi.json` documents exactly the statuses of the plan's `Surface` plus `500` for the 3 audit operations, and `web/src/api/schema.d.ts` is regenerated from it (Surface, door 1, door 4) `[done]`
Proof: `go -C app test ./internal/app -run '^TestOpenAPI_AuditStatuses$'`
Proof: `task gen:openapi:check`
Proof: `npm --prefix web run gen:check`

**C18** - archtest reports no violation: no package under `features/audit` imports another feature or `platform/audit`, and no file under `web/src/features/audit/` imports another web feature (AD-001, AD-007) `[done]`
Proof: `go -C app test ./archtest -run '^TestImports_RepositoryIsClean$'`
Proof: `go -C app test ./archtest -run '^TestImports_AuditReadsOnly$'`
Proof: `go -C app test ./archtest -run '^TestWebFeatures_DoNotImportEachOther$'`

### S6 - Web: lista de eventos · ~8 files · ~40 KB · ~10k

**C19** - `/audit` for a user with `audit:read` shows a table with headers `Quando`, `Ação`, `Quem`, `Recurso`; an event at `2026-10-08T17:30:05Z` by `a@x.com` on `user` `42` shows `08/10/2026 14:30:05`, `a@x.com`, `user 42`; an event with null actor shows `Sistema` (AUD-05, AC 17) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "shows the events table"`

**C20** - An empty `items` shows `Nenhum evento encontrado.` (AUD-05, AC 18) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "shows empty state"`

**C21** - While the events request is pending an element with `role="status"` is visible (AUD-05, AC 19) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "shows loading"`

**C22** - A `500` shows `Não foi possível carregar a auditoria.` and `Tentar novamente`, which repeats the request (AUD-05, AC 20) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "shows error with retry"`

**C23** - Without `audit:read`, `/audit` and `/audit/$id` show `Você não tem permissão para acessar esta página.` and call no `/api/v1/audit` route; with `audit:read` and an API `403`, both show the same text; the menu shows `Auditoria` with `audit:read` and hides it without (AUD-05, AC 21) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "forbidden"`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "forbidden"`
Proof: `npm --prefix web run test -- src/features/users/UserMenu.test.tsx -t "audit link"`

**C24** - With `next: 7`, `Mais antigos` sends `before=7` and appends the returned rows below the current ones; with `next: null` no `Mais antigos` button is shown (AUD-05, AC 22) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "loads older events"`

**C25** - The `Ação` select lists `Todas` followed by the catalogue items; choosing `user.created` puts `action=user.created` in the URL and the next events request carries `action=user.created`; choosing `Todas` removes it (AUD-05, AC 23) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "filters by action"`

**C26** - Filling `De` with `2026-10-01` and `Até` with `2026-10-02` sends `from=2026-10-01T03:00:00.000Z` and `to=2026-10-03T03:00:00.000Z` (AUD-05, AC 24) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "filters by period"`

**C27** - Clicking the actor email puts `actor_id` in the URL and the request; clicking `user 42` puts `resource_type=user&resource_id=42`; each shows `Limpar filtros`, which removes every filter from the URL and the next request (AUD-05, AC 25) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "filters by actor and resource"`

**C28** - Opening `/audit?action=user.created&actor_id=<uuid>&resource_type=user&resource_id=42` makes the first events request carry those four parameters (AUD-05, AC 26) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "reads filters from the URL"`

**C29** - Clicking the `Quando` text of event `5` navigates to `/audit/5` (AUD-05, AC 27) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "opens an event"`

### S7 - Web: detalhe do evento · ~4 files · ~15 KB · ~4k

**C30** - `/audit/5` shows `Ação`, `Quando`, `Quem`, `Recurso`, `IP` and `Request ID` with the event's values, `Antes` containing `{\n  "name": "Ana"\n}` and `Depois` containing `{\n  "name": "Bia"\n}`; with `before: null` the `Antes` block shows `—` (AUD-06, AC 28) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "shows the event"`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "shows a dash for a null block"`

**C31** - A `404` shows `Evento não encontrado.` (AUD-06, AC 29) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "shows not found"`

**C32** - While pending an element with `role="status"` is visible; a `500` shows `Não foi possível carregar a auditoria.` and `Tentar novamente`, which repeats the request (AUD-06, AC 30) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "shows loading"`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "shows error with retry"`

**C33** - The route `/audit/$id` passes the path id to the detail screen: `/audit/5` requests `GET /api/v1/audit/events/5` (door 1) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "requests the path id"`

**C34** - Against the built binary, an admin creates a user, opens `/audit`, filters `Ação` by `user.created`, sees that user's id in `Recurso`, opens the event and sees the new email in `Depois` (AUD-01, AUD-02, AUD-03, AUD-05, AUD-06 assembled) `[done]`
Proof: `npm --prefix web run e2e -- audit.spec.ts`

### S8 - Lacunas da verificação, rodada 1 · ~6 files · ~10 KB · ~3k

**C35** - An event stored with a null `ip` returns `"ip": null` in the list and in the detail, and an event with `10.0.0.7` returns that string in both (AUD-01, AC 1; AUD-03, AC 12) `[done]`
Proof: `go -C app test ./internal/features/audit/list_events -run '^TestListEvents_IPPresentOrNull$'`
Proof: `go -C app test ./internal/features/audit/get_event -run '^TestGetEvent_IPPresentOrNull$'`

**C36** - On `/audit/5`, an event with `ip: null` and `actor: null` shows `IP` as `—` and `Quem` as `Sistema` (AUD-06, AC 28; Assumption - ator nulo) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditDetail.test.tsx -t "shows fallbacks for a missing ip and actor"`

**C37** - Opening `/audit?de=2026-10-01&ate=2026-10-02` makes the first events request carry `from=2026-10-01T03:00:00.000Z` and `to=2026-10-03T03:00:00.000Z`, and fills `De` and `Até` with those dates (AUD-05, AC 24, AC 26) `[done]`
Proof: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx -t "reads the period from the URL"`

### S9 - Lacunas da verificação, rodada 2 · ~2 files · ~4 KB · ~1k

**C38** - `event.From` maps its 4 rows at its own layer: actor present gives `{id, email}`, actor absent gives `nil`, ip `10.0.0.7` gives that string, empty ip gives `nil`; and `GET /api/v1/audit/events/{id}` of an event with no actor returns `"actor": null` (AUD-01, AC 2; AUD-03, AC 12; Test policy - `event.From`) `[done]`
Proof: `go -C app test ./internal/features/audit/event -run '^TestFrom_ActorAndIPArms$'`
Proof: `go -C app test ./internal/features/audit/get_event -run '^TestGetEvent_ActorNull$'`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| `GET /api/v1/audit/events` statuses (5) | 200 C1 · 401 C14 · 403 C14 · 422 C5 · 500 documented C17 | - |
| `GET /api/v1/audit/events/{id}` statuses (6) | 200 C12 · 401 C14 · 403 C14 · 404 C13 · 422 C13 · 500 documented C17 | - |
| `GET /api/v1/audit/actions` statuses (4) | 200 C11 · 401 C14 · 403 C14 · 500 documented C17 | - |
| list item keys (8) | C1, table-driven over all 8 (exact key set) | - |
| actor shapes (2) | user C2, C12, C38 · null C2, C38 | - |
| ip shapes (2) | present C35, C38 · null C35, C38 | - |
| `event.From` rows at own layer (4) | C38, table-driven over all 4 | - |
| page outcomes (3) | default 50 with next C3 · limit with next C3 · last page null C3 | - |
| cursor edges (2) | below the newest C4 · below the oldest C4 | - |
| rejected list inputs (7) | C5, table-driven over all 7 | - |
| filters (6) | `action` C6 · `actor_id` C7 · `resource_type` C8 · `resource_id` C8 · `from` C9 · `to` C9 | - |
| period edges (2) | `from` inclusive C9 · `to` exclusive C9 | - |
| filter combination with paging (1) | C10 | - |
| detail id inputs (4) | unused 404 C13 · `abc` C13 · `0` C13 · `-1` C13 | - |
| access marker per audit operation (3) | C14, table-driven over all 3 | - |
| mutating methods under `/api/v1/audit` (4) | C15, table-driven over all 4 | - |
| screen `/audit` states (12) | table C19 · empty C20 · loading C21 · error C22 · forbidden C23 · older C24 · action C25 · period C26 · actor and resource C27 · URL filters C28 · URL period C37 · open C29 | - |
| screen `/audit/$id` states (8) | fields and JSON C30 · null block C30 · null ip C36 · null actor C36 · not found C31 · loading C32 · error C32 · forbidden C23 | - |
| menu link `Auditoria` (2) | shown C23 · hidden C23 | - |
| startup assembly (2 places) | `features.Register` through `app.New` C11, C14, C15 · test harness `testkit.NewAPI` C1-C13 | - |
| index definitions (4) | C16, table-driven over all 4 | - |
| Landing doors (4) | 1 C3, C4, C17 · 2 C16 · 3 C11 · 4 C14, C17 | - |

- Claims naming a status code, route or response shape: C1-C15, C17 - each proof issues a real HTTP request against the slice's registered handler or the assembled server, or reads the committed contract
- C14 and C34 prove the assembled path a second time; they do not stand in for the slice proofs C1-C13 nor for the component proofs C19-C33
- No other check claims more than the single case its proof exercises

## Test policy

O `AGENTS.md` já responde às duas perguntas (`## Test policy`); estas são as linhas dele, aplicadas.

| Code | Required proofs | Coverage expectation |
| --- | --- | --- |
| Decides, reached across a boundary | one at the boundary **and** one at its own layer | the contract at the boundary; one asserted case per row of the decision table at its own layer |
| Decides, not reached across a boundary | one at its own layer | one asserted case per row of the decision table |
| Entry point that decides nothing | one at the boundary | accepted input, each rejected input, each error path |
| Instrumentation, pass-throughs | none of its own | covered by its consumer's proof |

Evidence (planned code, by shape):

- `features/audit/list_events`: 7 input validations, 6 optional filters, cursor and `next` decision -> decides at the slice boundary; its HTTP test is its own layer (C1-C10)
- `features/audit/get_event`: lookup or `404` -> entry point (C12, C13)
- `features/audit/event.From` (added in the build, classified in verification round 1): maps actor present/absent and ip present/absent -> decides, reached across two slices; each arm asserted at its own layer (C38) and through the list and detail boundaries (C2, C35, C38)
- `platform/op.AuditActions`: dedup, skip empty, sort -> decides, reached across a boundary (C11 own layer and through `list_actions` and the assembled server)
- `features/audit/list_actions`: forwards `op.AuditActions` -> instrumentation, proven through C11
- `web/src/features/audit/{AuditList,AuditDetail}.tsx` and the date-to-range helper: map status, permission and filters to screen state and request parameters -> decides, not reached across a boundary; one asserted case per branch (C19-C33)
- `web/src/features/users/UserMenu.tsx`: one more permission branch -> both rows asserted (C23)
- closest analogue in the repo: `features/users/list_users` (bounds proven per rejected input) and `web/src/features/rbac/RoleDetail.tsx` (one case per branch)

Cost: 18 slice-level Go proofs, 3 assembled, 1 schema, 15 component proofs.

## Swept

- validation: C5, C13
- failure modes: existing - a read has no partial state; database errors surface as `500` problem+json (foundation C3, C13) and C22, C32 cover the screen
- idempotency: n/a - every operation is a read with no side effect
- authorization: C14, C15, C23
- concurrency: C4, C10 - the cursor reads `id < before`, so events inserted while paging never shift or repeat the next page
- data lifecycle: n/a - the feature only reads; retention and purge are out of scope and the append-only trigger stays (users door 8)
- dependency failure: existing - no new external dependency; the database path is the foundation's
- state transitions: n/a - events are immutable rows
- observability: existing - reads are logged by the request middleware with `request_id` (foundation); reading the audit is not itself audited (Out of scope)

## Handoff

Intended split, written before any code. Existing files read as patterns (`rbac` slices and tests, `op`, `app`
tests, `migrations`, `archtest`, web `rbac` and `users` screens, `test/render.tsx`, e2e) = ~100 KB = ~25k. Written
per slice: S1 ~6k · S2 ~5k · S3 ~3k · S4 ~2k · S5 ~4k · S6 ~10k · S7 ~4k = ~34k, plus ~25k read = ~59k, under the
150k budget -> one builder.
