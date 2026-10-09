# Audit verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 89c27d1..73721e3
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier)

Inputs read: `.specs/features/audit/plan.md`, `.specs/features/audit/checks.md` (Profile: standard), `AGENTS.md`,
`.specs/STATE.md` AD-007 ("the `audit` feature only reads"), `.specs/LESSONS.md` confirmed L-010, and the diff
`89c27d1..73721e3` (41 files). All 34 checks hold with located evidence; the verdict is FAIL because the recompute
found unproven members and one mutant survived (see Coverage, Test policy rows, Faults injected).

## Binding sources

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none - the plan marks no source binding; profile is `standard`, step 1 does not apply | n/a | - | - |

## Checks

Proofs run at `73721e3`:
- Go, one invocation: `go -C app test -count=1 ./internal/features/audit/... ./internal/platform/op ./internal/app ./migrations ./archtest -run '^(TestListEvents_NewestFirstWithoutDiff|...|TestWebFeatures_DoNotImportEachOther)$' -v` - 22 named tests, each `--- PASS`, exit 0.
- Web, one invocation: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx src/features/audit/AuditDetail.test.tsx src/features/users/UserMenu.test.tsx --reporter=verbose` - 27 passed (12 list, 8 detail, 7 menu), each named test printed with a check mark, exit 0.
- `task e2e -- audit.spec.ts` - `1 passed`, `audit.spec.ts:5:1 admin finds the creation of a user in the audit log`.
- `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | newest first, exact 8 keys | `TestListEvents_NewestFirstWithoutDiff` PASS | `app/internal/features/audit/list_events/list_events_test.go:72` - `require.Equal(t, []int64{e3, e2, e1}, ids(...))`; `:84` - exact sorted key list `action, actor, id, ip, occurred_at, request_id, resource_id, resource_type` | PASS |
| C2 | actor `{id,email}`, current email, null actor | `TestListEvents_Actor` PASS | `list_events_test.go:102` - `require.Equal(t, &actor{ID: a, Email: "a@x.com"}, got[byA])`; `:103` - `require.Nil(t, got[bySystem])`; `:107` - `"b@x.com"` after the email update | PASS |
| C3 | default 50 + next = 50th; limit=2 next; limit=3 null | `TestListEvents_Pages` PASS | `list_events_test.go:119` - `require.Len(t, p.Items, 50)`; `:121` - `require.Equal(t, all[49], *p.Next)`; `:130` - `require.Equal(t, two.Items[1].ID, *two.Next)`; `:131` - `require.Nil(t, g.get(t, "limit=3").Next)` | PASS |
| C4 | `before` below newest and below oldest | `TestListEvents_Before` PASS | `list_events_test.go:140` - `[]int64{e2, e1}`; `:142` - `require.Empty(t, last.Items)`; `:143` - `require.Nil(t, last.Next)` | PASS |
| C5 | 422 naming each of 7 params; limit 1 and 100 accepted | `TestListEvents_Validation` PASS | `list_events_test.go:159` - `require.Equal(t, http.StatusUnprocessableEntity, rec.Code, c.query)`; `:160` - `"location":"`+c.location+`"` over the 7-row table at `:148-156`; `:163` - `f.get` asserts 200 for `limit=1`, `limit=100` | PASS |
| C6 | `action` filter | `TestListEvents_FilterAction` PASS | `list_events_test.go:172` - `require.Equal(t, []int64{c2, c1}, ids(f.get(t, "action=user.created")))` | PASS |
| C7 | `actor_id` filter | `TestListEvents_FilterActor` PASS | `list_events_test.go:182` - `require.Equal(t, []int64{byA}, ...)` | PASS |
| C8 | `resource_type` and `resource_id` | `TestListEvents_FilterResource` PASS | `list_events_test.go:190` - `[]int64{u2, u1}`; `:191` - `[]int64{u1}` | PASS |
| C9 | `from` inclusive, `to` exclusive, both | `TestListEvents_FilterPeriod` PASS | `list_events_test.go:203` - `[]int64{twelve, eleven}`; `:204` - `[]int64{eleven, ten}`; `:205` - `[]int64{eleven}` | PASS |
| C10 | filter + paging within filtered set | `TestListEvents_FiltersCombineAndPage` PASS | `list_events_test.go:216` - `[]int64{created[2], created[1]}`; `:219` - `[]int64{created[0]}`; `:220` - `require.Nil(t, second.Next)` | PASS |
| C11 | `op.AuditActions` sorted distinct; endpoint; assembled server | `TestAuditActions_SortedDistinct`, `TestListActions_ReturnsCatalogue`, `TestAudit_ActionsOfAssembledServer` PASS | `app/internal/platform/op/op_test.go:109` - `require.Equal(t, []string{"a.done", "b.done"}, op.AuditActions(api))`; `app/internal/features/audit/list_actions/list_actions_test.go:41` - `[]string{"a.done", "b.done"}` as `items`; `app/internal/app/audit_test.go:58` - `require.Contains(t, items, want)` over the 4 actions, `:60` - `slices.IsSorted(items)` | PASS |
| C12 | detail with before/after; null before | `TestGetEvent_ReturnsDiff` PASS | `app/internal/features/audit/get_event/get_event_test.go:38` - `map[string]any{"name": "Ana"}` = `body["before"]`; `:39` - `{"name": "Bia"}` = `body["after"]`; `:30-36` list fields; `:42-43` - `require.Contains(t, body, "before")`, `require.Nil(t, body["before"])` | PASS |
| C13 | 404 unused id; 422 for `abc`, `0`, `-1` | `TestGetEvent_404And422` PASS | `get_event_test.go:54` - `require.Equal(t, http.StatusNotFound, get("999999"))`; `:56` - `StatusUnprocessableEntity` over `abc`,`0`,`-1` | PASS |
| C14 | 3 ops: `audit:read`, 401, 403 for every other permission | `TestAudit_OperationAccess` PASS | `audit_test.go:68` - `ElementsMatch` of the 3 operations; `:85` - `op.Permission("audit:read")`; `:86` - `StatusUnauthorized`; `:87-88` - `StatusForbidden` with caller holding every `op.Permissions` except `audit:read` (`:71-77`) | PASS |
| C15 | no POST/PUT/PATCH/DELETE under `/api/v1/audit` | `TestAudit_ReadOnly` PASS | `audit_test.go:96` - `require.NotContains(t, []string{"POST", "PUT", "PATCH", "DELETE"}, method, key)`; operation set from the assembled `app.New` at `:23-30`, `:38` | PASS |
| C16 | 4 named indexes after migrate up | `TestSchema_AuditIndexes` PASS | `app/migrations/schema_test.go:82` - `require.Contains(t, names, want)` over the 4 names | PASS |
| C17 | committed statuses = Surface + 500; TS client regenerated | `TestOpenAPI_AuditStatuses` PASS; `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0 | `audit_test.go:118` - `require.Equal(t, slices.Sorted(... append(statuses, "500")), slices.Sorted(maps.Keys(operation.Responses)), key)` with the Surface table at `:109-113` | PASS |
| C18 | no cross-feature or `platform/audit` import; web features isolated | `TestImports_RepositoryIsClean`, `TestImports_AuditReadsOnly`, `TestWebFeatures_DoNotImportEachOther` PASS | `app/archtest/imports_test.go:37` - `require.Empty(t, v)`; `app/archtest/rbac_test.go:46` - `require.NotEqual(t, ".../internal/platform/audit", pkg)`, `:48` - `require.Contains(t, pkg, "/internal/features/audit", pkg)`; `rbac_test.go:23` - `require.Empty(t, v)` on the real `web/src` | PASS |
| C19 | table headers, date in TZ, email, `Sistema`, `user 42` | `shows the events table` PASS | `web/src/features/audit/AuditList.test.tsx:40` - headers `["Quando", "Ação", "Quem", "Recurso"]`; `:49-52` - rows `["08/10/2026 14:30:05", "user.created", "a@x.com", "user 42"]` and `[..., "Sistema", "user 42"]` | PASS |
| C20 | empty state | `shows empty state` PASS | `AuditList.test.tsx:62` - `findByText("Nenhum evento encontrado.")` | PASS |
| C21 | loading status | `shows loading` PASS | `AuditList.test.tsx:72` - `findByRole("status")` with the events request `pending` | PASS |
| C22 | 500 error + retry repeats request | `shows error with retry` PASS | `AuditList.test.tsx:82` - `"Não foi possível carregar a auditoria."`; `:84` - table after retry; `:85` - `eventRequests(requests)).toHaveLength(2)` | PASS |
| C23 | forbidden on both screens (no permission, API 403), menu link both rows | `forbidden without audit:read`, `forbidden on API 403` (list and detail), `audit link with audit:read`, `hides the audit link without audit:read` PASS | `AuditList.test.tsx:91-92` - forbidden text, `0` audit requests; `:102` - forbidden on 403; `web/src/features/audit/AuditDetail.test.tsx:85-86` and `:92` - same on `/audit/5`; `web/src/features/users/UserMenu.test.tsx:48` - link shown, `:55` - `not.toBeInTheDocument()` | PASS |
| C24 | `Mais antigos` sends `before=7`, appends, hidden when null | `loads older events` PASS | `AuditList.test.tsx:118` - `["user.created", "user.created", "role.created"]`; `:119` - `get("before")).toBe("7")`; `:120` - button gone | PASS |
| C25 | `Todas` + catalogue; URL and request carry `action`; `Todas` removes | `filters by action` PASS | `AuditList.test.tsx:132-136` - options `["Todas", "role.created", "user.created"]`; `:139` - `search toEqual({ action: "user.created" })`; `:140` - request `action`; `:143-144` - removed from URL and request | PASS |
| C26 | `from=2026-10-01T03:00:00.000Z`, `to=2026-10-03T03:00:00.000Z` | `filters by period` PASS | `AuditList.test.tsx:158` - `get("from")).toBe("2026-10-01T03:00:00.000Z")`; `:159` - `get("to")).toBe("2026-10-03T03:00:00.000Z")` | PASS |
| C27 | actor and resource click filters; `Limpar filtros` clears | `filters by actor and resource` PASS | `AuditList.test.tsx:171-172` - `actor_id` in URL and request; `:175-176` - cleared URL and empty query; `:180-185` - `resource_type=user`, `resource_id=42`; `:187` - `Limpar filtros` shown | PASS |
| C28 | URL filters on first request | `reads filters from the URL` PASS | `AuditList.test.tsx:199-202` - first request carries `action`, `actor_id`, `resource_type`, `resource_id=42` | PASS |
| C29 | `Quando` link opens `/audit/5` | `opens an event` PASS | `AuditList.test.tsx:214` - `pathname).toBe("/audit/5")` | PASS |
| C30 | detail fields, indented JSON, `—` for null block | `shows the event`, `shows a dash for a null block` PASS | `AuditDetail.test.tsx:34-39` - six `fieldValue` assertions; `:40` - `'{\n  "name": "Ana"\n}'`; `:41` - `'{\n  "name": "Bia"\n}'`; `:48` - `block("Antes")).toBe("—")` | PASS |
| C31 | 404 not found | `shows not found` PASS | `AuditDetail.test.tsx:61` - `"Evento não encontrado."` | PASS |
| C32 | detail loading; 500 error + retry | `shows loading`, `shows error with retry` PASS | `AuditDetail.test.tsx:67` - `findByRole("status")`; `:76` - error text; `:78` - heading after retry; `:79` - `toHaveLength(2)` | PASS |
| C33 | path id reaches the request | `requests the path id` PASS | `AuditDetail.test.tsx:55` - `requests.some((r) => r.path === path)).toBe(true)` with `path = "/api/v1/audit/events/5"` | PASS |
| C34 | built binary: create user, filter, open, see email in `Depois` | `task e2e -- audit.spec.ts` 1 passed | `web/e2e/audit.spec.ts:24` - `expect(row).toHaveCount(1)` on `user ${userId}` after `selectOption("user.created")`; `:30` - `Depois` block `toContainText(email)` | PASS |

## Coverage

Recomputed from the plan's Surface, Landing, Relations and criteria, and from the code at `73721e3`.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| `GET /api/v1/audit/events` statuses (5) | plan Surface + `500`; committed `app/openapi.json` | 200 C1 · 401 C14 · 403 C14 · 422 C5 · 500 documented C17 | - |
| `GET /api/v1/audit/events/{id}` statuses (6) | plan Surface + `500` | 200 C12 · 401 C14 · 403 C14 · 404 C13 · 422 C13 · 500 documented C17 | - |
| `GET /api/v1/audit/actions` statuses (4) | plan Surface + `500` | 200 C11 · 401 C14 · 403 C14 · 500 documented C17 | - |
| list item keys (8) | AC 1; `event.go:29-38` | C1 exact key set | - |
| `event.From` mapping rows (4) | code `app/internal/features/audit/event/event.go:52-61` (two conditionals) | actor present C2, C12 · actor null C2 · ip present C12 (`get_event_test.go:35`) · ip absent -> `null` (`event.go:58-60`) no proof | ip absent -> `ip: null` (`event.go:58`): no test inserts an event without ip (`rg` over audit tests finds none); mutant `if true` survived |
| page outcomes (3) | AC 3, door 1, `list_events/endpoint.go:59-64` | default 50 + next C3 · limit + next C3 · last page null C3 | - |
| cursor edges (2) | AC 4 | below newest C4 · below oldest C4 | - |
| rejected list inputs (7) | AC 5; `endpoint.go:21-28`, `:81-110` | C5 table over all 7 | - |
| filters (6) | AC 6-9; `queries.sql:6-12` | action C6 · actor_id C7 · resource_type C8 · resource_id C8 · from C9 · to C9 | - |
| period edges (2) | AC 9 | from inclusive C9 · to exclusive C9 | - |
| filter combination with paging (1) | AC 10 | C10 | - |
| detail id inputs (4) | AC 13 | unused 404 C13 · `abc` C13 · `0` C13 · `-1` C13 | - |
| detail before/after (2 x 2) | AC 12; `get_event/endpoint.go:60-65` (`orNull` shared) | before object C12 · after object C12 · before null C12 · after null via the same `orNull` call | - |
| access marker per audit operation (3) | AC 14-15; assembled server | C14 over all 3 | - |
| mutating methods under `/api/v1/audit` (4) | AC 16 | C15 over all 4 | - |
| URL filter params on first request (6) | AC 26; `web/src/routes/_authed/audit/index.tsx:6-13` | action, actor_id, resource_type, resource_id C28 · de, ate through the URL round-trip of C26 (the date inputs are controlled by `filters.de/ate` read back from the route search) | - |
| screen `/audit` states (11) | AC 17-27 | table C19 · empty C20 · loading C21 · error C22 · forbidden C23 · older C24 · action C25 · period C26 · actor/resource C27 · URL C28 · open C29 | - |
| screen `/audit/$id` field values (6 fields + fallbacks) | AC 28; `web/src/features/audit/AuditDetail.tsx:44-51` | six fields with values C30 · `Quem` fallback `Sistema` via shared `actorLabel` (proven on the list, C19) · `IP` fallback `e.ip ?? "—"` (`AuditDetail.tsx:49`) no proof | `IP` null -> `—` on `/audit/$id` (`AuditDetail.tsx:49`): no detail test sends `ip: null` (L-010 recurrence) |
| screen `/audit/$id` states (6) | AC 28-30, AC 21 | fields and JSON C30 · null block C30 · not found C31 · loading C32 · error C32 · forbidden C23 | - |
| menu link `Auditoria` (2) | AC 21; `UserMenu.tsx:23` | shown C23 · hidden C23 | - |
| startup assembly (2 places) | `app/cmd/api/main.go:88` calls `app.New` without `Register`; `app/internal/app/app.go:49-50` defaults to `features.Register`; `app/internal/features/registry.go:19` adds `audit.Register` | binary C11/C14/C15 via `app.New`, C34 via the built binary · slice harness `audittest.Serve` -> `testkit.NewAPI` C1-C13 | - |
| Landing doors (4) | plan Landing | 1 C3, C4, C17 · 2 C16 (names) · 3 C11 · 4 C14, C17 | door 2 literal shape: the column lists `(actor_id, id DESC)`, `(resource_type, resource_id, id DESC)`, `(action, id DESC)`, `(occurred_at)` are never asserted - C16 reads only `indexname` (`schema_test.go:82`); precision gap in the check |

- Level: every claim naming a status, route or response shape (C1-C15, C17) is proven by an HTTP request against the registered handler or the assembled server, or by reading the committed contract. No level gap.
- Swept `existing` rows re-read: failure modes - `app/internal/platform/httpx/middleware.go:93` writes `500` problem+json and huma maps unknown handler errors to `500`; dependency failure - no new dependency in the diff; observability - `middleware.go:69` logs every request with `request_id`. All three constraints are present.
- AD-007: `TestImports_AuditReadsOnly` (`rbac_test.go:46`) and C15 hold; no audit slice writes.

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `app/internal/platform/op/op.go` `AuditActions` | own layer C11 (`op_test.go:109`) · boundary C11 (`list_actions_test.go:41`, `audit_test.go:58`) | yes - dedup, skip empty (the `GET` with `""` metadata) and sort each asserted |
| Decides at the slice boundary (checks.md classification) | `features/audit/list_events/endpoint.go`, `queries.sql`, and the unclassified new mapper `features/audit/event/event.go` used by it and by `get_event` | HTTP test at the slice C1-C10 | gap - `event.From` ip-absent row (`event.go:58`) has no asserted case and its mutant survived; `event.go` is new decision code the checks.md evidence list never classified |
| Entry point that decides nothing | `features/audit/get_event/endpoint.go` | boundary C12, C13 | yes - accepted, 404, three 422 inputs |
| Instrumentation, pass-throughs | `features/audit/list_actions/endpoint.go`, `audittest`, routes `_authed/audit/*.tsx` | through consumers C11, C28, C33 | yes |
| Decides, not reached across a boundary | `web/src/features/audit/AuditList.tsx`, `AuditDetail.tsx`, `api.ts` (`periodOf`, `actorLabel`), `web/src/features/users/UserMenu.tsx` | one asserted case per branch C19-C33 | gap - `AuditDetail.tsx:49` `e.ip ?? "—"` fallback branch has no asserted case (confirmed lesson L-010: assert every branch, including fallback values) |

## Faults injected

Scratch worktree `git worktree add <scratchpad>/wt HEAD`; real-tree porcelain recorded before (`?? .claude/skills/auth-security/`, `?? .cursor/skills/auth-security/`) and identical after `git worktree remove --force` + `prune`. Web run used a junction to `web/node_modules`, removed with `rmdir` before the worktree.

| Mutation | Location | Killed |
| --- | --- | --- |
| `to` bound inclusive: `e.occurred_at < to_at` -> `<=`, sqlc regenerated | `app/internal/features/audit/list_events/queries.sql:12` | yes - `TestListEvents_FilterPeriod` expected `[2 1]`, got `[3 2 1]` |
| never set `next` (`page.Next = &last` -> `_ = last`) | `app/internal/features/audit/list_events/endpoint.go:62` | yes - `TestListEvents_Pages` "Expected value not to be nil" |
| `AuditActions` without dedupe (`return slices.Compact(out)` -> `return out`) | `app/internal/platform/op/op.go:114` | yes - `TestAuditActions_SortedDistinct` got `[a.done b.done b.done]` |
| ip never null (`if r.IP != ""` -> `if true`) | `app/internal/features/audit/event/event.go:58` | survived - every audit Go test and the assembled `internal/app` tests stayed `ok` |
| `to` without the extra day (`localMidnight(filters.ate, 1)` -> `localMidnight(filters.ate)`) | `web/src/features/audit/api.ts:28` | yes - `filters by period` expected `2026-10-03T03:00:00.000Z`, received `2026-10-02T03:00:00.000Z` |

## Gate

- `go -C app test -count=1 <7 pkgs> -run '<22 names>' -v` - 22 passed, 0 failed
- `npm --prefix web run test -- AuditList.test.tsx AuditDetail.test.tsx UserMenu.test.tsx` - 27 passed, 0 failed
- `task e2e -- audit.spec.ts` - 1 passed, 0 failed
- `task gen:openapi:check` - exit 0 (`TestOpenAPI_ServedMatchesCommitted` ok)
- `npm --prefix web run gen:check` - exit 0

## Ranked gaps

1. `ip` null mapping unproven, mutant survived - C1/C12 (AC 1 lists `ip` but no check states its null case) - `app/internal/features/audit/event/event.go:58`; add a slice case inserting an event with `ip` NULL and asserting `"ip": null` (list and detail), and classify `event.go` in the Test policy evidence.
2. Detail `IP` fallback `—` unasserted (L-010 recurrence) - C30 - `web/src/features/audit/AuditDetail.tsx:49`; add a detail case with `ip: null` asserting `fieldValue("IP")` is `—`.
3. Landing door 2 precision gap: C16 asserts index names only, not the literal column lists - `app/migrations/schema_test.go:82` vs `app/migrations/20261008180000_audit_indexes.sql:2-5`; assert `indexdef` per index.
