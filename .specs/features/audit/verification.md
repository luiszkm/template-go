# Audit verification

**Verdict**: PASS
**Profile**: standard
**Diff range**: 89c27d1..8ba925d
**Round**: 3 - scoped
**Verifier**: independent sub-agent (author != verifier)

Inputs read: `.specs/features/audit/plan.md`, `.specs/features/audit/checks.md` (Profile: standard, now C1-C38), the
round 2 report (FAIL at `363bbf6`, one gap: the `event.From` Test policy row was unmet), `AGENTS.md` `## Test policy`
and the fix diff `363bbf6..8ba925d`. The fix's code files are `event/event_test.go` (new) and
`get_event/get_event_test.go` (appended at `:76-87`). No production code changed; the rest of the diff is specs and
lessons. The fix closes the round 2 gap. `event.From` now has an own-layer test with one asserted case per decision
row (4), and `GET /api/v1/audit/events/{id}` asserts `"actor": null` for an actor-less event. All 38 checks hold with
located evidence, and all 3 faults injected on the fix's surfaces were killed.

## Binding sources

carried from 363bbf6 (originally 73721e3) - the fix touched no interface.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none - the plan marks no source binding; profile is `standard`, step 1 does not apply | n/a | - | - |

## Checks

Proofs re-run in full at `8ba925d`:
- Go, one invocation: `go -C app test -count=1 ./internal/features/audit/... ./internal/platform/op ./internal/app ./migrations ./archtest -run '^(<the 26 names in checks.md>)$' -v`. Exit 0. Each of the 26 named tests printed `--- PASS`, including the new `TestFrom_ActorAndIPArms` (each of its 4 subtests `--- PASS`) and `TestGetEvent_ActorNull`.
- Web, one invocation: `npm --prefix web run test -- src/features/audit/AuditList.test.tsx src/features/audit/AuditDetail.test.tsx src/features/users/UserMenu.test.tsx --reporter=verbose`. Exit 0, `Tests 29 passed (29)`, each named test printed with a check mark.
- `task e2e -- audit.spec.ts`: `1 passed`, `e2e\audit.spec.ts:5:1 › admin finds the creation of a user in the audit log`.
- `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0.

Citations were refreshed at 8ba925d for the files the fix touched. `event_test.go` is new. `get_event_test.go` was only appended to, so its old lines `:1-75` did not move. All other citations are carried from 363bbf6, because those files are unchanged in `363bbf6..8ba925d`.

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
| C12 | detail before/after; null before | `TestGetEvent_ReturnsDiff` PASS | `app/internal/features/audit/get_event/get_event_test.go:38` - `{"name": "Ana"}`; `:39` - `{"name": "Bia"}`; `:42-43` - `Contains "before"`, `Nil(body["before"])` (verified at 8ba925d, unmoved: the fix only appended at `:76-87`) | PASS |
| C13 | 404 unused; 422 `abc`,`0`,`-1` | `TestGetEvent_404And422` PASS | `get_event_test.go:54` - `StatusNotFound`; `:56` - `StatusUnprocessableEntity` (verified at 8ba925d, unmoved: the fix only appended at `:76-87`) | PASS |
| C14 | 3 ops `audit:read`, 401, 403 | `TestAudit_OperationAccess` PASS | `audit_test.go:68`, `:85`, `:86`, `:87-88` (carried) | PASS |
| C15 | no mutating method under `/api/v1/audit` | `TestAudit_ReadOnly` PASS | `audit_test.go:96` (carried) | PASS |
| C16 | 4 indexes with their column lists | `TestSchema_AuditIndexes` PASS | `app/migrations/schema_test.go:82-87` - the want map `audit_events_actor_idx: "(actor_id, id DESC)"`, `resource_idx: "(resource_type, resource_id, id DESC)"`, `action_idx: "(action, id DESC)"`, `occurred_idx: "(occurred_at)"`; `:89` - `require.Contains(t, defs, name)`; `:90` - `require.True(t, strings.HasSuffix(defs[name], "USING btree "+columns))` (carried from 363bbf6) | PASS |
| C17 | committed statuses = Surface + 500; TS regenerated | `TestOpenAPI_AuditStatuses` PASS; both gen checks exit 0 | `audit_test.go:118` over `:109-113` (carried) | PASS |
| C18 | import isolation | `TestImports_RepositoryIsClean`, `TestImports_AuditReadsOnly`, `TestWebFeatures_DoNotImportEachOther` PASS | `app/archtest/imports_test.go:37`; `app/archtest/rbac_test.go:46`, `:48`, `:23` (carried) | PASS |
| C19 | table headers, date, email, `Sistema`, `user 42` | `shows the events table` PASS | `web/src/features/audit/AuditList.test.tsx:40`; `:49-52` (carried from 363bbf6) | PASS |
| C20 | empty state | `shows empty state` PASS | `AuditList.test.tsx:62` (carried from 363bbf6) | PASS |
| C21 | loading | `shows loading` PASS | `AuditList.test.tsx:72` (carried from 363bbf6) | PASS |
| C22 | 500 + retry | `shows error with retry` PASS | `AuditList.test.tsx:82`, `:84`, `:85` (carried from 363bbf6) | PASS |
| C23 | forbidden both screens; menu link | list/detail `forbidden ...`, `audit link with audit:read`, `hides the audit link without audit:read` PASS | `AuditList.test.tsx:91-92`, `:102`; `web/src/features/audit/AuditDetail.test.tsx:96` - forbidden text, `:97` - `toHaveLength(0)` audit requests, `:103` - forbidden on API 403 (carried from 363bbf6); `web/src/features/users/UserMenu.test.tsx:48`, `:55` (carried) | PASS |
| C24 | `Mais antigos` | `loads older events` PASS | `AuditList.test.tsx:118`, `:119`, `:120` (carried from 363bbf6) | PASS |
| C25 | action select | `filters by action` PASS | `AuditList.test.tsx:132-136`, `:139`, `:140`, `:143-144` (carried from 363bbf6) | PASS |
| C26 | period request params | `filters by period` PASS | `AuditList.test.tsx:158` - `get("from")).toBe("2026-10-01T03:00:00.000Z")`; `:159` - `get("to")).toBe("2026-10-03T03:00:00.000Z")` (carried from 363bbf6) | PASS |
| C27 | actor/resource click filters; clear | `filters by actor and resource` PASS | `AuditList.test.tsx:171-172`, `:175-176`, `:180-185`, `:187` (carried from 363bbf6) | PASS |
| C28 | URL filters on first request | `reads filters from the URL` PASS | `AuditList.test.tsx:199-202` (carried from 363bbf6) | PASS |
| C29 | `Quando` opens `/audit/5` | `opens an event` PASS | `AuditList.test.tsx:229` - `expect(router.state.location.pathname).toBe("/audit/5")` (carried from 363bbf6) | PASS |
| C30 | detail fields, JSON, `—` null block | `shows the event`, `shows a dash for a null block` PASS | `AuditDetail.test.tsx:34-39` six `fieldValue`; `:40` - `'{\n  "name": "Ana"\n}'`; `:41`; `:48` - `block("Antes")).toBe("—")` (carried from 363bbf6) | PASS |
| C31 | 404 not found | `shows not found` PASS | `AuditDetail.test.tsx:72` - `findByText("Evento não encontrado.")` (carried from 363bbf6) | PASS |
| C32 | detail loading; 500 + retry | `shows loading`, `shows error with retry` PASS | `AuditDetail.test.tsx:78` - `findByRole("status")`; `:87` - error text; `:89` - heading after retry; `:90` - `toHaveLength(2)` (carried from 363bbf6) | PASS |
| C33 | path id reaches the request | `requests the path id` PASS | `AuditDetail.test.tsx:66` - `requests.some((r) => r.path === path)).toBe(true)` (carried from 363bbf6) | PASS |
| C34 | built binary flow | `task e2e -- audit.spec.ts` 1 passed | `web/e2e/audit.spec.ts:24`, `:30` (carried) | PASS |
| C35 | null ip -> `"ip": null`, `10.0.0.7` -> string, list and detail | `TestListEvents_IPPresentOrNull`, `TestGetEvent_IPPresentOrNull` PASS | `list_events_test.go:239` - `require.Equal(t, "10.0.0.7", *ips[withIP])`; `:240-241` - `require.Contains(t, ips, withoutIP)`, `require.Nil(t, ips[withoutIP])`; `:242` - `require.Contains(t, rec.Body.String(), `"ip":null`)`; `get_event_test.go:73` - `require.Equal(t, "10.0.0.7", ipOf(withIP))`; `:74` - `require.Nil(t, ipOf(withoutIP))` with `:70` `require.Contains(t, body, "ip")` (get_event lines verified at 8ba925d, unmoved; list_events lines carried from 363bbf6) | PASS |
| C36 | detail `IP` `—` and `Quem` `Sistema` | `shows fallbacks for a missing ip and actor` PASS | `AuditDetail.test.tsx:58` - `expect(fieldValue("IP")).toBe("—")`; `:59` - `expect(fieldValue("Quem")).toBe("Sistema")` (carried from 363bbf6) | PASS |
| C37 | period from URL on the first request, inputs filled | `reads the period from the URL` PASS | `AuditList.test.tsx:214` - `first?.get("from")).toBe("2026-10-01T03:00:00.000Z")`; `:215` - `get("to")).toBe("2026-10-03T03:00:00.000Z")`; `:216-217` - `getByLabelText("De")).toHaveValue("2026-10-01")`, `"Até"` `"2026-10-02"` (carried from 363bbf6) | PASS |
| C38 | `event.From` 4 rows at own layer; detail `"actor": null` | `TestFrom_ActorAndIPArms` (subtests actor_present, actor_null, ip_present, ip_null) and `TestGetEvent_ActorNull` PASS | `app/internal/features/audit/event/event_test.go:34` - `require.Equal(t, c.wantActor, got.Actor)`; `:35` - `require.Equal(t, c.wantIP, got.IP)` over the table `:23-27` (present -> `&event.AuditActor{ID: actor, Email: "a@x.com"}`, actor null -> `nil`, ip `10.0.0.7` -> `&ip`, empty ip -> `nil`); `app/internal/features/audit/get_event/get_event_test.go:85` - `require.Contains(t, body, "actor")`; `:86` - `require.Nil(t, body["actor"])` (verified at 8ba925d) | PASS |

## Coverage

The rows the fix touched were recomputed from the code at `8ba925d`: `event/event.go:52-62` (two decisions, `:55` actor and `:58` ip), used by `get_event/endpoint.go:52` and `list_events/endpoint.go:65`. Every other row is carried from 363bbf6 with `-` unproven.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| route statuses: events (5), event (6), actions (4) | carried from 363bbf6 | as round 2, C1, C5, C11-C14, C17 | - |
| list item keys (8), page outcomes (3), cursor edges (2), rejected list inputs (7), filters (6), period edges (2), filter+paging (1), detail id inputs (4), detail before/after (2 x 2), access per op (3), mutating methods (4), screen `/audit/$id` states (6), screen `/audit/$id` field fallbacks (2), index definitions (4), Landing doors (4), URL filter params on the first request (6), screen `/audit` states (12), menu link (2), startup assembly (2) | carried from 363bbf6 | as round 2 | - |
| actor shapes (2) | verified at 8ba925d; `event.go:55-57` | user C2 (`list_events_test.go:102`), C12 (`get_event_test.go:32`), C38 (`event_test.go:34`, row `:23-24`) · null C2 (`list_events_test.go:103`), C38 (`event_test.go:34`, row `:25`; `get_event_test.go:85-86`) | - |
| ip shapes (2) | verified at 8ba925d; `event.go:58-60` | present C35 (`list_events_test.go:239`, `get_event_test.go:73`), C38 (`event_test.go:35`, row `:26`) · null C35 (`list_events_test.go:241-242`, `get_event_test.go:74`), C38 (`event_test.go:35`, row `:27`) | - |
| `event.From` rows at own layer (4) | verified at 8ba925d; `event.go:55`, `:58` | actor present · actor null · ip present · ip null: C38, table-driven `event_test.go:23-27`, asserted `:34-35`; flip-actor and ip-`if true` faults both killed here | - |
| `event.From` arms per consumer (4 arms x 2 slices) | verified at 8ba925d; `get_event/endpoint.go:52`, `list_events/endpoint.go:65` | list: actor present C2 · actor null C2 · ip present C35 · ip null C35. detail: actor present C12 (`get_event_test.go:32`) · actor null C38 (`get_event_test.go:85-86`) · ip present C35 (`:73`) · ip null C35 (`:74`). The round 2 unproven member is now proven, and its `Valid: true` fault was killed | - |

Note, not a gap: the own-layer rows `actor null` (`event_test.go:25`) and `ip present` (`:26`) use the same input, `event.Row{IP: ip}`. Both cases still assert actor and ip (`:34-35`), so all 4 decision rows are covered. The duplicate adds no information.

- Level: C38 reaches `event.From` directly at its own layer, and reaches `GET /api/v1/audit/events/{id}` over real HTTP against Postgres. Neither has a level gap.
- Swept `existing` rows: carried from 363bbf6. The fix touched no production code.

## Test policy rows

The `event.From` row (unmet in round 2) and the row classifying the touched `get_event` slice were re-judged at 8ba925d. The rest are carried from 363bbf6.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `app/internal/platform/op/op.go` `AuditActions` | own layer C11 (`op_test.go:109`) · boundary C11 (`list_actions_test.go:41`, `audit_test.go:58`) | yes (carried from 363bbf6) |
| Decides, reached across a boundary | `app/internal/features/audit/event/event.go` `From`, 4 rows (actor present/null, ip present/null) | own layer C38 (`event_test.go:34-35` over `:23-27`, one case per row) · boundary: list_events C2, C35 · get_event C12, C35, C38 (`get_event_test.go:85-86`) | yes (verified at 8ba925d) - each decision row is asserted at its own layer, and each arm is asserted through both boundaries; mutants on both conditions were killed at the own layer |
| Decides at the slice boundary | `features/audit/list_events/endpoint.go`, `queries.sql` | HTTP test at the slice C1-C10, C35 | yes (carried from 363bbf6) |
| Entry point that decides nothing | `features/audit/get_event/endpoint.go` | boundary C12, C13, C35, C38 | yes (verified at 8ba925d) - accepted, 404, three 422 inputs, and the row build's actor-null contract |
| Instrumentation, pass-throughs | `list_actions/endpoint.go`, `audittest/audittest.go`, routes `_authed/audit/*.tsx` | through consumers C11, C28, C33, C35, C37, C38 | yes (carried from 363bbf6) |
| Decides, not reached across a boundary | `web/src/features/audit/AuditList.tsx`, `AuditDetail.tsx`, `api.ts`, `web/src/features/users/UserMenu.tsx` | one asserted case per branch C19-C33, C36, C37 | yes (carried from 363bbf6) |

## Faults injected

Verified at 8ba925d. Faults were re-injected on the surfaces the fix created (C38's two tests). Scratch: `git worktree add <scratchpad long path>/wt3 HEAD`, outside the repo. The real-tree porcelain was recorded first (`?? .claude/skills/auth-security/`, `?? .cursor/skills/auth-security/`). It was identical after `git worktree remove --force` + `prune` (empty `diff`), and `git worktree list` shows only the main tree. `git stash` was not used. Each fault was reverted with `git checkout -- .` in the scratch before the next one.

| Mutation | Location | Killed |
| --- | --- | --- |
| detail row built with `ActorID: pgtype.UUID{Bytes: r.ActorID.Bytes, Valid: true}` | `app/internal/features/audit/get_event/endpoint.go:52` | yes - `TestGetEvent_ActorNull` "Expected nil, but got: map[string]interface {}{email: '', id: 00000000-0000-0000-0000-000000000000}" at `get_event_test.go:86` |
| actor condition flipped `if r.ActorID.Valid` -> `if !r.ActorID.Valid` | `app/internal/features/audit/event/event.go:55` | yes - `TestFrom_ActorAndIPArms` subtests actor_present, actor_null and ip_present FAIL "Not equal" at `event_test.go:34` (own layer) |
| ip condition `if r.IP != ""` -> `if true` | `app/internal/features/audit/event/event.go:58` | yes - `TestFrom_ActorAndIPArms/ip_null` FAIL "Not equal" at `event_test.go:35`; the other 3 subtests PASS, so the ip-null row is what kills it |

Round 2's five faults (ip `if true` through both boundaries, the index columns, the detail `IP` and `Quem` fallbacks, `De` read from the URL) are carried from 363bbf6. The fix changed none of their production code or proofs.

## Gate

- `go -C app test -count=1 <7 pkgs> -run '<26 names>' -v` - 26 passed, 0 failed
- `npm --prefix web run test -- AuditList.test.tsx AuditDetail.test.tsx UserMenu.test.tsx` - 29 passed, 0 failed
- `task e2e -- audit.spec.ts` - 1 passed, 0 failed
- `task gen:openapi:check` - exit 0
- `npm --prefix web run gen:check` - exit 0
