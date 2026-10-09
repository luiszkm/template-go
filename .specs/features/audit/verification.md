# Audit verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 89c27d1..363bbf6
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier)

Inputs read: `.specs/features/audit/plan.md`, `.specs/features/audit/checks.md` (Profile: standard, C1-C37), the
round 1 report (FAIL at `73721e3`), `AGENTS.md` `## Test policy`, and the fix diff `73721e3..363bbf6` (code files:
`audittest/audittest.go`, `get_event/get_event_test.go`, `list_events/list_events_test.go`,
`migrations/schema_test.go`, `AuditDetail.test.tsx`, `AuditList.test.tsx`; no production code changed). The fix
closes all three round 1 gaps: the ip mutant is now killed, the detail `IP` and `Quem` fallbacks are asserted, and
C16 asserts each `indexdef`. All 37 checks hold with located evidence and every counted fault was killed. The
verdict is still FAIL. The fix added a Test policy classification for `event.From`, and the tree does not meet it:
the row requires an own-layer proof and none exists, and the classification's own claim ("each arm asserted
through both boundaries") is contradicted for actor null through `get_event`. See Coverage and Test policy rows.

## Binding sources

carried from 73721e3 - the fix touched no interface.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none - the plan marks no source binding; profile is `standard`, step 1 does not apply | n/a | - | - |

## Checks

Proofs re-run in full at `363bbf6`:
- Go, one invocation: `go -C app test -count=1 ./internal/features/audit/... ./internal/platform/op ./internal/app ./migrations ./archtest -run '^(<the 24 names in checks.md>)$' -v`. Exit 0, 24 named tests each printed `--- PASS`, including the new `TestListEvents_IPPresentOrNull` and `TestGetEvent_IPPresentOrNull`.
- Web, one invocation: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx src/features/audit/AuditDetail.test.tsx src/features/users/UserMenu.test.tsx --reporter=verbose`. Exit 0, `Tests 29 passed (29)` (13 list, 9 detail, 7 menu), each printed with a check mark, including `shows fallbacks for a missing ip and actor` and `reads the period from the URL`.
- `task e2e -- audit.spec.ts`: `1 passed`, `e2e\audit.spec.ts:5:1 › admin finds the creation of a user in the audit log`.
- `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0.

Citations: verified at 363bbf6 for the files the fix touched (`list_events_test.go` and `get_event_test.go` were only appended to, so their old lines did not move; `schema_test.go`, `AuditDetail.test.tsx` and `AuditList.test.tsx` were refreshed). All other citations are carried from 73721e3 because those files are unchanged in `73721e3..363bbf6`.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | newest first, exact 8 keys | `TestListEvents_NewestFirstWithoutDiff` PASS | `app/internal/features/audit/list_events/list_events_test.go:72` - `require.Equal(t, []int64{e3, e2, e1}, ids(...))`; `:84` - exact sorted key list (carried) | PASS |
| C2 | actor `{id,email}`, current email, null actor | `TestListEvents_Actor` PASS | `list_events_test.go:102` - `&actor{ID: a, Email: "a@x.com"}`; `:103` - `require.Nil(t, got[bySystem])`; `:107` - `"b@x.com"` (carried) | PASS |
| C3 | default 50 + next; limit=2 next; limit=3 null | `TestListEvents_Pages` PASS | `list_events_test.go:119` - `require.Len(t, p.Items, 50)`; `:121` - `all[49]`; `:130`; `:131` - `Nil(...Next)` (carried) | PASS |
| C4 | `before` below newest and oldest | `TestListEvents_Before` PASS | `list_events_test.go:140` - `[]int64{e2, e1}`; `:142` - `Empty`; `:143` - `Nil(last.Next)` (carried) | PASS |
| C5 | 422 per each of 7 params; 1 and 100 accepted | `TestListEvents_Validation` PASS | `list_events_test.go:159` - `StatusUnprocessableEntity`; `:160` - `"location":"`+c.location+`"` over `:148-156`; `:163` (carried) | PASS |
| C6 | `action` filter | `TestListEvents_FilterAction` PASS | `list_events_test.go:172` - `[]int64{c2, c1}` (carried) | PASS |
| C7 | `actor_id` filter | `TestListEvents_FilterActor` PASS | `list_events_test.go:182` - `[]int64{byA}` (carried) | PASS |
| C8 | `resource_type`, `resource_id` | `TestListEvents_FilterResource` PASS | `list_events_test.go:190` - `[]int64{u2, u1}`; `:191` - `[]int64{u1}` (carried) | PASS |
| C9 | `from` inclusive, `to` exclusive | `TestListEvents_FilterPeriod` PASS | `list_events_test.go:203`, `:204`, `:205` (carried) | PASS |
| C10 | filter + paging | `TestListEvents_FiltersCombineAndPage` PASS | `list_events_test.go:216`, `:219`, `:220` (carried) | PASS |
| C11 | `op.AuditActions`; endpoint; assembled | `TestAuditActions_SortedDistinct`, `TestListActions_ReturnsCatalogue`, `TestAudit_ActionsOfAssembledServer` PASS | `app/internal/platform/op/op_test.go:109`; `app/internal/features/audit/list_actions/list_actions_test.go:41`; `app/internal/app/audit_test.go:58`, `:60` (carried) | PASS |
| C12 | detail before/after; null before | `TestGetEvent_ReturnsDiff` PASS | `app/internal/features/audit/get_event/get_event_test.go:38` - `{"name": "Ana"}`; `:39` - `{"name": "Bia"}`; `:42-43` - `Contains "before"`, `Nil(body["before"])` (verified at 363bbf6, unmoved) | PASS |
| C13 | 404 unused; 422 `abc`,`0`,`-1` | `TestGetEvent_404And422` PASS | `get_event_test.go:54` - `StatusNotFound`; `:56` - `StatusUnprocessableEntity` (verified at 363bbf6, unmoved) | PASS |
| C14 | 3 ops `audit:read`, 401, 403 | `TestAudit_OperationAccess` PASS | `audit_test.go:68`, `:85`, `:86`, `:87-88` (carried) | PASS |
| C15 | no mutating method under `/api/v1/audit` | `TestAudit_ReadOnly` PASS | `audit_test.go:96` (carried) | PASS |
| C16 | 4 indexes with their column lists | `TestSchema_AuditIndexes` PASS | `app/migrations/schema_test.go:82-87` - the want map `audit_events_actor_idx: "(actor_id, id DESC)"`, `resource_idx: "(resource_type, resource_id, id DESC)"`, `action_idx: "(action, id DESC)"`, `occurred_idx: "(occurred_at)"`; `:89` - `require.Contains(t, defs, name)`; `:90` - `require.True(t, strings.HasSuffix(defs[name], "USING btree "+columns))` (verified at 363bbf6) | PASS |
| C17 | committed statuses = Surface + 500; TS regenerated | `TestOpenAPI_AuditStatuses` PASS; both gen checks exit 0 | `audit_test.go:118` over `:109-113` (carried) | PASS |
| C18 | import isolation | `TestImports_RepositoryIsClean`, `TestImports_AuditReadsOnly`, `TestWebFeatures_DoNotImportEachOther` PASS | `app/archtest/imports_test.go:37`; `app/archtest/rbac_test.go:46`, `:48`, `:23` (carried) | PASS |
| C19 | table headers, date, email, `Sistema`, `user 42` | `shows the events table` PASS | `web/src/features/audit/AuditList.test.tsx:40`; `:49-52` (unmoved, insertion is at `:205`) | PASS |
| C20 | empty state | `shows empty state` PASS | `AuditList.test.tsx:62` (unmoved) | PASS |
| C21 | loading | `shows loading` PASS | `AuditList.test.tsx:72` (unmoved) | PASS |
| C22 | 500 + retry | `shows error with retry` PASS | `AuditList.test.tsx:82`, `:84`, `:85` (unmoved) | PASS |
| C23 | forbidden both screens; menu link | list/detail `forbidden ...`, `audit link with audit:read`, `hides the audit link without audit:read` PASS | `AuditList.test.tsx:91-92`, `:102`; `web/src/features/audit/AuditDetail.test.tsx:96` - forbidden text, `:97` - `toHaveLength(0)` audit requests, `:103` - forbidden on API 403 (refreshed); `web/src/features/users/UserMenu.test.tsx:48`, `:55` (carried) | PASS |
| C24 | `Mais antigos` | `loads older events` PASS | `AuditList.test.tsx:118`, `:119`, `:120` (unmoved) | PASS |
| C25 | action select | `filters by action` PASS | `AuditList.test.tsx:132-136`, `:139`, `:140`, `:143-144` (unmoved) | PASS |
| C26 | period request params | `filters by period` PASS | `AuditList.test.tsx:158` - `get("from")).toBe("2026-10-01T03:00:00.000Z")`; `:159` - `get("to")).toBe("2026-10-03T03:00:00.000Z")` (unmoved) | PASS |
| C27 | actor/resource click filters; clear | `filters by actor and resource` PASS | `AuditList.test.tsx:171-172`, `:175-176`, `:180-185`, `:187` (unmoved) | PASS |
| C28 | URL filters on first request | `reads filters from the URL` PASS | `AuditList.test.tsx:199-202` (unmoved) | PASS |
| C29 | `Quando` opens `/audit/5` | `opens an event` PASS | `AuditList.test.tsx:229` - `expect(router.state.location.pathname).toBe("/audit/5")` (refreshed) | PASS |
| C30 | detail fields, JSON, `—` null block | `shows the event`, `shows a dash for a null block` PASS | `AuditDetail.test.tsx:34-39` six `fieldValue`; `:40` - `'{\n  "name": "Ana"\n}'`; `:41`; `:48` - `block("Antes")).toBe("—")` (unmoved) | PASS |
| C31 | 404 not found | `shows not found` PASS | `AuditDetail.test.tsx:72` - `findByText("Evento não encontrado.")` (refreshed) | PASS |
| C32 | detail loading; 500 + retry | `shows loading`, `shows error with retry` PASS | `AuditDetail.test.tsx:78` - `findByRole("status")`; `:87` - error text; `:89` - heading after retry; `:90` - `toHaveLength(2)` (refreshed) | PASS |
| C33 | path id reaches the request | `requests the path id` PASS | `AuditDetail.test.tsx:66` - `requests.some((r) => r.path === path)).toBe(true)` (refreshed) | PASS |
| C34 | built binary flow | `task e2e -- audit.spec.ts` 1 passed | `web/e2e/audit.spec.ts:24`, `:30` (carried) | PASS |
| C35 | null ip -> `"ip": null`, `10.0.0.7` -> string, list and detail | `TestListEvents_IPPresentOrNull`, `TestGetEvent_IPPresentOrNull` PASS | `list_events_test.go:239` - `require.Equal(t, "10.0.0.7", *ips[withIP])`; `:240-241` - `require.Contains(t, ips, withoutIP)`, `require.Nil(t, ips[withoutIP])`; `:242` - `require.Contains(t, rec.Body.String(), `"ip":null`)`; `get_event_test.go:73` - `require.Equal(t, "10.0.0.7", ipOf(withIP))`; `:74` - `require.Nil(t, ipOf(withoutIP))` with `:70` `require.Contains(t, body, "ip")` (verified at 363bbf6) | PASS |
| C36 | detail `IP` `—` and `Quem` `Sistema` | `shows fallbacks for a missing ip and actor` PASS | `AuditDetail.test.tsx:58` - `expect(fieldValue("IP")).toBe("—")`; `:59` - `expect(fieldValue("Quem")).toBe("Sistema")` (verified at 363bbf6) | PASS |
| C37 | period from URL on the first request, inputs filled | `reads the period from the URL` PASS | `AuditList.test.tsx:214` - `first?.get("from")).toBe("2026-10-01T03:00:00.000Z")`; `:215` - `get("to")).toBe("2026-10-03T03:00:00.000Z")`; `:216-217` - `getByLabelText("De")).toHaveValue("2026-10-01")`, `"Até"` `"2026-10-02"` (verified at 363bbf6) | PASS |

## Coverage

Rows the fix touched were recomputed from the code at `363bbf6` (`event/event.go:52-62`, `get_event/endpoint.go:52`, `list_events/endpoint.go:65`, `AuditDetail.tsx:44-51`, `migrations/20261008180000_audit_indexes.sql:2-5`, `routes/_authed/audit/index.tsx:6-13`). Every other row is carried from 73721e3 with `-` unproven.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| route statuses: events (5), event (6), actions (4) | carried from 73721e3 | as round 1, C1, C5, C11-C14, C17 | - |
| list item keys (8), page outcomes (3), cursor edges (2), rejected list inputs (7), filters (6), period edges (2), filter+paging (1), detail id inputs (4), detail before/after (2 x 2), access per op (3), mutating methods (4), screen `/audit/$id` states (6), menu link (2), startup assembly (2) | carried from 73721e3 | as round 1 | - |
| ip shapes (2) | verified at 363bbf6; `event.go:58-60` | present C35 (`list_events_test.go:239`, `get_event_test.go:73`) · null C35 (`list_events_test.go:241-242`, `get_event_test.go:74`); the `if true` mutant is now killed | - |
| `event.From` arms per consumer (4 arms x 2 slices), the set named by checks.md Test policy evidence "each arm asserted through both boundaries (C2, C35)" | verified at 363bbf6; `event.go:55-60`, used at `get_event/endpoint.go:52` and `list_events/endpoint.go:65` | list: actor present C2 · actor null C2 · ip present C35 · ip null C35. detail: actor present C12 (`get_event_test.go:32`) · ip present C35 · ip null C35 · actor null: none | actor null through `GET /api/v1/audit/events/{id}`. `rg -n actor get_event_test.go` hits only `:19`, `:20` and `:32` (the present case). `TestGetEvent_IPPresentOrNull` inserts an actor-less event and asserts only `200` (`:68`), never `actor: null`. So a wrong `get_event` row build (e.g. `Valid: true`) would pass. The member is named in the artifact's own prose |
| screen `/audit/$id` field fallbacks (2) | verified at 363bbf6; `AuditDetail.tsx:47` `actorLabel(e)`, `:49` `e.ip ?? "—"` | `Quem` `Sistema` C36 (`AuditDetail.test.tsx:59`) · `IP` `—` C36 (`:58`); both faults killed | - |
| index definitions (4) | verified at 363bbf6; `20261008180000_audit_indexes.sql:2-5` | actor, resource, action, occurred: C16, each `HasSuffix(indexdef, "USING btree "+cols)` (`schema_test.go:83-90`); column fault killed | - |
| Landing doors (4) | verified at 363bbf6 | 1 C3, C4, C17 · 2 C16 (now literal columns) · 3 C11 · 4 C14, C17 | - |
| URL filter params on the first request (6) | verified at 363bbf6; `routes/_authed/audit/index.tsx:6-13` | action, actor_id, resource_type, resource_id C28 · de, ate C37 (`AuditList.test.tsx:214-217`) | - |
| screen `/audit` states (12) | verified at 363bbf6 | as round 1 plus URL period C37 | - |

Round 1's three unproven members are now proven: ip absent -> `null` (C35), detail `IP` `—` (C36) and the door 2 literal column lists (C16). One new unproven member comes from the fix's own Test policy evidence, shown above.

- Level: C35 issues real HTTP requests against both registered slices, and C16 reads `pg_indexes` after the real migrations. Neither has a level gap.
- Swept `existing` rows: carried from 73721e3. The fix touched no production code.

## Test policy rows

`event.From` and the touched files were re-judged at 363bbf6. The other rows are carried from 73721e3.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `app/internal/platform/op/op.go` `AuditActions` | own layer C11 (`op_test.go:109`) · boundary C11 (`list_actions_test.go:41`, `audit_test.go:58`) | yes (carried from 73721e3) |
| Decides, reached across a boundary (checks.md classification added in the fix) | `app/internal/features/audit/event/event.go` `From`, 4 rows (actor present/null, ip present/null) | boundary: list_events C2, C35 · get_event C12, C35 · own layer: none (`app/internal/features/audit/event/` holds only `event.go`) | no - the row demands one asserted case per decision row at its own layer and there is none; the evidence line's claim that each arm is asserted through both boundaries is contradicted for actor null at `get_event` (asserted only as `200`, `get_event_test.go:68`) |
| Decides at the slice boundary | `features/audit/list_events/endpoint.go`, `queries.sql` | HTTP test at the slice C1-C10, C35 | yes - round 1's ip-absent row is now asserted and its mutant killed |
| Entry point that decides nothing | `features/audit/get_event/endpoint.go` | boundary C12, C13, C35 | yes - accepted, 404, three 422 inputs |
| Instrumentation, pass-throughs | `list_actions/endpoint.go`, `audittest/audittest.go` (fix added `NoIP`), routes `_authed/audit/*.tsx` | through consumers C11, C28, C33, C35, C37 | yes |
| Decides, not reached across a boundary | `web/src/features/audit/AuditList.tsx`, `AuditDetail.tsx`, `api.ts`, `web/src/features/users/UserMenu.tsx` | one asserted case per branch C19-C33, C36, C37 | yes - the `AuditDetail.tsx:49` `—` branch (C36 `:58`) and the detail `actorLabel` null branch (C36 `:59`) are now asserted |

## Faults injected

Verified at 363bbf6. Scratch: `git worktree add <scratchpad>/wt HEAD`. The real-tree porcelain was recorded first (`?? .claude/skills/auth-security/`, `?? .cursor/skills/auth-security/`) and was identical after `git worktree remove --force` + `prune`. Web runs used a junction to the real `web/node_modules`, removed with `rmdir` before the worktree. Vitest had to run from the long path, because the 8.3 short path breaks `setupFiles` resolution.

| Mutation | Location | Killed |
| --- | --- | --- |
| ip never null (`if r.IP != ""` -> `if true`) | `app/internal/features/audit/event/event.go:58` | yes - `TestListEvents_IPPresentOrNull` "Expected nil, but got: (*string)" and `TestGetEvent_IPPresentOrNull` "Expected nil, but got: \"\"" |
| actor index columns `(actor_id, id DESC)` -> `(actor_id, id)` | `app/migrations/20261008180000_audit_indexes.sql:2` | yes - `TestSchema_AuditIndexes` "Should be true", indexdef `... USING btree (actor_id, id)` at `schema_test.go:90` |
| detail IP fallback `e.ip ?? "—"` -> `e.ip ?? ""` | `web/src/features/audit/AuditDetail.tsx:49` | yes - `shows fallbacks for a missing ip and actor` "expected '' to be '—'" |
| detail Quem `actorLabel(e)` -> `e.actor?.email ?? ""` | `web/src/features/audit/AuditDetail.tsx:47` | yes - same test, "expected '' to be 'Sistema'" |
| `De` input not read from the URL (`value={filters.de ?? ""}` -> `value=""`) | `web/src/features/audit/AuditList.tsx:63` | yes - `reads the period from the URL` "expect(element).toHaveValue(2026-10-01)", received empty |

One more mutant was tried and is not counted. Removing `de` and `ate` from `validateSearch` (`routes/_authed/audit/index.tsx:11-12`) left `reads the period from the URL` green. That is an equivalent mutant, not a weak test. TanStack Router merges the raw parsed search into the route search, so both strings still reach `AuditList`. The same test asserts the observable outcomes (request `from`/`to` and the input values), and they were unchanged. It was replaced by the `AuditList.tsx:63` fault above.

## Gate

- `go -C app test -count=1 <7 pkgs> -run '<24 names>' -v` - 24 passed, 0 failed
- `npm --prefix web run test -- AuditList.test.tsx AuditDetail.test.tsx UserMenu.test.tsx` - 29 passed, 0 failed
- `task e2e -- audit.spec.ts` - 1 passed, 0 failed
- `task gen:openapi:check` - exit 0
- `npm --prefix web run gen:check` - exit 0

## Ranked gaps

1. `event.From` Test policy row unmet. It is classified "Decides, reached across a boundary" but has no own-layer proof (`app/internal/features/audit/event/` has no test), and the evidence claim "each arm asserted through both boundaries (C2, C35)" is contradicted. Actor null through `get_event` is never asserted: `get_event_test.go:68` checks only `200`. Close it by adding an `event` package test over the 4 rows and asserting `actor: null` on `GET /api/v1/audit/events/{id}` for an actor-less event. Alternatively, reclassify `event.From` honestly, but the get_event actor-null assertion is still owed by the claim.
