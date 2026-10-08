# RBAC verification

**Verdict**: PASS
**Profile**: standard
**Diff range**: 54895ff..7d0cae6
**Round**: 3 - scoped
**Verifier**: independent sub-agent (author != verifier)

Scope: fix diff `7ad1473..7d0cae6` - `app/internal/features/rbac/assign_roles/assign_roles_test.go` (lines 202 and 212 replaced in place: gate lock mode `ACCESS EXCLUSIVE` -> `EXCLUSIVE`, waiter filter adds `backend_xid IS NOT NULL`) and the `checks.md` C21 text; plus the only round 2 verdict that was not PASS (C21). No production code changed (`git diff 7ad1473..7d0cae6 -- app web` touches only that test file).

## Binding sources

carried from fe13da6 - neither the round 2 nor the round 3 fix touched the interface.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none - the plan marks no source binding; profile is `standard`, step 1 does not apply | n/a | - | - |

## Checks

verified at 7d0cae6 - every proof re-ran in full at `7d0cae6`; each named test appears individually in the output:

- Go: `go -C app test -count=1 ./internal/features/rbac/... ./internal/app ./archtest ./migrations -run '^(<35 names taken from checks.md>)$' -v` exit 0 - 35 top-level `--- PASS` (each of the 35 names grepped individually, none missing) + 13 subtests (48), 0 `--- FAIL`
- Web: `npm --prefix web run test -- src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.test.tsx src/features/users/UserMenu.test.tsx src/routes/userDetail.test.tsx --reporter=verbose` exit 0 - 6 files, 40 passed; every `-t` name in checks.md shown with a pass mark
- e2e: `task e2e -- rbac.spec.ts` exit 0 - `rbac.spec.ts:5:1 admin grants a role and the user gets exactly its access` 1 passed
- `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0
- C21 additionally: `go -C app test -count=5 ./internal/features/rbac/assign_roles -run '^TestAssignRoles_ConcurrentLastAdmin$' -v` on the real tree - 5/5 `--- PASS`

Citations: the one file the fix touched (`assign_roles_test.go`) was re-read at 7d0cae6; the diff replaces lines 202 and 212 in place, so every cited line (`:93-97`, `:102-105`, `:117-119`, `:133-135`, `:141-142`, `:153-157`, `:184-186`, `:193`, `:202`, `:210-213`, `:218-220`, `:237-238`) was confirmed at its number. All other citations are carried from 7ad1473 (files byte-identical, `git diff 7ad1473..7d0cae6` empty for them).

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | catalogue sorted, has rbac:read/rbac:assign/zeta:do, no `*` | `TestListPermissions_ReturnsCatalogue` PASS | (carried) `app/internal/features/rbac/list_permissions/list_permissions_test.go:25` - `require.True(t, slices.IsSorted(items))`; `:26-29` Contains rbac:read, rbac:assign, zeta:do, NotContains `*` | PASS |
| C2 | order admin, Beta, Zeta; perms and counts | `TestListRoles_OrderedWithCounts` PASS | (carried) `app/internal/features/rbac/list_roles/list_roles_test.go:41` - `require.Equal(t, []string{"admin", "Beta", "Zeta"}, order)`; `:42-47` | PASS |
| C3 | GET role returns id, name, sorted perms, user_count | `TestGetRole_ReturnsRole` PASS | (carried) `app/internal/features/rbac/get_role/get_role_test.go:25` - `require.Equal(t, role.Role{ID: beta, Name: "Beta", ...})` | PASS |
| C4 | 404 random uuid / 422 `abc` on GET, PATCH, DELETE | `TestGetRole_404And422`, `TestUpdateRole_404And422`, `TestDeleteRole_404And422` PASS | (carried) `app/internal/features/rbac/get_role/get_role_test.go:36-37`; `app/internal/features/rbac/update_role/update_role_test.go:164-165`; `app/internal/features/rbac/delete_role/delete_role_test.go:102-103` - `http.StatusNotFound` / `http.StatusUnprocessableEntity` | PASS |
| C5 | 201 trimmed Leitor, one audit role.created | `TestCreateRole_Creates` PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:53-59` | PASS |
| C6 | duplicate name 409 on POST and PATCH, no change, no audit | `TestCreateRole_DuplicateNameIs409`, `TestUpdateRole_DuplicateNameIs409` PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:68-70`; `app/internal/features/rbac/update_role/update_role_test.go:123-125` - `StatusConflict`, `nameOf == "Outro"`, audit zero | PASS |
| C7 | concurrent same name: one 201, one 409 | `TestCreateRole_ConcurrentSameName` PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:82-83` | PASS |
| C8 | 6 rejected, 3 accepted, PATCH same 6 | `TestCreateRole_Validation`, `TestUpdateRole_Validation` PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:103,108,110,119`; `app/internal/features/rbac/update_role/update_role_test.go:154-155,157-159` | PASS |
| C9 | PATCH name only / perms only / both in one call, audit before/after | `TestUpdateRole_UpdatesSentFields`, `TestUpdateRole_UpdatesBothFields` PASS | (carried) `app/internal/features/rbac/update_role/update_role_test.go:84,86-87,92-98` (single fields); `:175-178` - `require.Equal(t, "Gestor", got.Name)`, perms `[]string{"users:create", "users:read"}` in response and DB; `:180-181` before/after snapshots; `:182` one audit event | PASS |
| C10 | permission edit applies next request, sessions unchanged | `TestUpdateRole_AppliesOnNextRequest` PASS | (carried) `app/internal/features/rbac/update_role/update_role_test.go:110,114-115` | PASS |
| C11 | admin PATCH/DELETE 409, unchanged, no audit | `TestUpdateRole_AdminIs409`, `TestDeleteRole_AdminIs409` PASS | (carried) `app/internal/features/rbac/update_role/update_role_test.go:131-135`; `app/internal/features/rbac/delete_role/delete_role_test.go:59-61` | PASS |
| C12 | delete unused role 204, perms gone, audit with before | `TestDeleteRole_Deletes` PASS | (carried) `app/internal/features/rbac/delete_role/delete_role_test.go:39-42,52-53` | PASS |
| C13 | delete in-use role 409, rows intact | `TestDeleteRole_InUseIs409` PASS | (carried) `app/internal/features/rbac/delete_role/delete_role_test.go:70-74` | PASS |
| C14 | delete waits for uncommitted assignment then 409 | `TestDeleteRole_WaitsForConcurrentAssignment` PASS | (carried) `app/internal/features/rbac/delete_role/delete_role_test.go:91,96-97` | PASS |
| C15 | user roles sorted by name; empty list | `TestGetUserRoles_ReturnsRoles` PASS | (carried) `app/internal/features/rbac/get_user_roles/get_user_roles_test.go:37-38` | PASS |
| C16 | replace set, sessions 0, cookies 401, one audit before/after | `TestAssignRoles_ReplacesAndRevokes` PASS | (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:93-97` - 204, `sorted(beta, gama)`, sessions zero, both probes 401; `:104-105` `{"roles":["Alfa"]}` / `{"roles":["Beta","Gama"]}` | PASS |
| C17 | same set reordered: 204, nothing changes | `TestAssignRoles_SameSetIsNoop` PASS | (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:117-119` - `require.Equal(t, was, f.state(t, user))`, audits zero | PASS |
| C18 | unknown/duplicate id -> 422 at body.role_ids | `TestAssignRoles_InvalidRoleIds` PASS | (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:133-135` - 422, `"location":"body.role_ids"`, state unchanged | PASS |
| C19 | 404 / 422 on GET and PUT users/{id}/roles | `TestGetUserRoles_404And422`, `TestAssignRoles_404And422` PASS | (carried) `app/internal/features/rbac/get_user_roles/get_user_roles_test.go:48-49`; (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:141-142` | PASS |
| C20 | last-admin guard, 5 cases | `TestAssignRoles_LastAdminGuard` PASS (5 subtests shown) | (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:184` - `require.Equal(t, c.want, rec.Code)`; `:186` state unchanged on 409; table `:153-157` | PASS |
| C21 | two concurrent admin removals held inside their transactions (gate `LOCK TABLE sessions IN EXCLUSIVE MODE` lets the auth `SELECT` through and blocks the session `DELETE`, released when `pg_stat_activity` shows two lock waiters holding a transaction id): one 204, one 409, one active admin; x3 | `TestAssignRoles_ConcurrentLastAdmin` PASS (full batch, and `-count=5` 5/5) | (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:202` - `LOCK TABLE sessions IN EXCLUSIVE MODE`; `:210-213` - `require.Eventually(... wait_event_type = 'Lock' AND backend_xid IS NOT NULL ... == 2 ...)`; `:218` - `require.Equal(t, []int{http.StatusNoContent, http.StatusConflict}, codes)`; `:219-220` exactly one active admin holder; `:193` `for range 3`. Mechanism observed in the scratch worktree with a `pg_stat_activity JOIN pg_locks (NOT granted)` dump at the release point, 3 iterations: each time exactly two waiters, both with `backend_xid` set - one `relation sessions RowExclusiveLock` on `DeleteUserSessions` (the request that won the admin row, `app/internal/features/rbac/assign_roles/endpoint.go:132`), one `transactionid ShareLock` on `LockAdminRole ... FOR UPDATE` (`app/internal/features/rbac/assign_roles/queries.sql:11`) waiting for the first. The auth lookup (`app/internal/platform/auth/auth.go:99-101`, plain `SELECT`, AccessShare) does not conflict with `EXCLUSIVE`. Claim and test agree; F1 killed 5/5 at `:218` | PASS |
| C22 | trigger failure -> 500, state unchanged | `TestAssignRoles_RollsBackTogether` PASS | (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:237-238` - `StatusInternalServerError`, `require.Equal(t, was, f.state(t, user))` | PASS |
| C23 | 8 ops declare permission, 401, 403 without own, no writes | `TestRBAC_OperationAccess` PASS (8 subtests shown) | (carried) `app/internal/app/rbac_test.go:68` - `require.Len(t, declared, len(operations))`; `:81` declared permission; `:84` 401; `:94-98` 403, audit zero, roles/user_roles unchanged | PASS |
| C24 | ADMIN insert unique violation; Financeiro stored | `TestSchema_RoleNameUniqueIgnoringCase` PASS | (carried) `app/migrations/schema_test.go:63,67` | PASS |
| C25 | 8 ops each document exactly Surface statuses plus `500`; schema.d.ts regenerated | `TestOpenAPI_RBACStatuses` PASS; `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0 | (carried) `app/internal/app/rbac_test.go:114-122` - `want` lists all 8 ops with the plan Surface statuses; `:127-128` - `require.Equal(t, slices.Sorted(slices.Values(append(statuses, "500"))), documented, key)` - set equality | PASS |
| C26 | archtest clean; web feature import rule with failing fixture | `TestImports_RepositoryIsClean`, `TestWebFeatures_DoNotImportEachOther` PASS | (carried) `app/archtest/imports_test.go:37`; `app/archtest/rbac_test.go:16-17,21` | PASS |
| C27 | lib/session.ts, lib/problems.ts exports; old files gone | `TestWebSharedSession_Moved` PASS | (carried) `app/archtest/rbac_test.go:27,31,35` | PASS |
| C28 | /roles table headers; `Todas`, `2` | `RolesList > shows the roles table` PASS | (carried) `web/src/features/rbac/RolesList.test.tsx:30,38-41` | PASS |
| C29 | role="status" while pending on /roles and /roles/$id | `RolesList > shows loading`, `RoleDetail > shows loading` PASS | (carried) `web/src/features/rbac/RolesList.test.tsx:47`; `web/src/features/rbac/RoleDetail.test.tsx:31` - `findByRole("status")` | PASS |
| C30 | 500 message and retry, both screens | `RolesList > shows error with retry`, `RoleDetail > shows error with retry` PASS | (carried) `web/src/features/rbac/RolesList.test.tsx:56,59`; `web/src/features/rbac/RoleDetail.test.tsx:41,44` | PASS |
| C31 | forbidden without rbac:read on 3 screens, no rbac call; API 403 on /roles, /roles/new, /roles/$id; menu link | `-t forbidden` on RolesList (2), RoleForm (2), RoleDetail (2); `UserMenu` roles link (2) PASS | (carried) `web/src/features/rbac/RolesList.test.tsx:65-66,75`; `web/src/features/rbac/RoleForm.test.tsx:88-89,95`; `web/src/features/rbac/RoleDetail.test.tsx:50-51,186`; `web/src/features/users/UserMenu.test.tsx:35,42` | PASS |
| C32 | create sends body, navigates | `RoleForm > creates and navigates` PASS | (carried) `web/src/features/rbac/RoleForm.test.tsx:35-36` | PASS |
| C33 | grouped headings and checkboxes, both screens | `RoleForm > groups permissions`, `RoleDetail > groups permissions` PASS | (carried) `web/src/features/rbac/RoleForm.test.tsx:49`; `web/src/features/rbac/RoleDetail.test.tsx:71` | PASS |
| C34 | 409 message in `name-error`, referenced by the name input's `aria-describedby`, create and edit | `RoleForm > shows name conflict`, `RoleDetail > shows name conflict` PASS | (carried) `web/src/features/rbac/RoleForm.test.tsx:61-62` and `web/src/features/rbac/RoleDetail.test.tsx:107-108` - `expect(conflict.id).toBe("name-error")`, `toHaveAttribute("aria-describedby", "name-error")` | PASS |
| C35 | 422 messages under their fields | `RoleForm > shows field errors`, `RoleDetail > shows field errors` PASS | (carried) `web/src/features/rbac/RoleForm.test.tsx:81-82`; `web/src/features/rbac/RoleDetail.test.tsx:129-130` | PASS |
| C36 | PATCH only changed fields; success text | `RoleDetail > patches only changed fields` PASS | (carried) `web/src/features/rbac/RoleDetail.test.tsx:93,95` | PASS |
| C37 | admin locked | `RoleDetail > locks the admin role` PASS | (carried) `web/src/features/rbac/RoleDetail.test.tsx:137-144` | PASS |
| C38 | delete dialog, cancel, confirm | `RoleDetail > confirms deletion` PASS | (carried) `web/src/features/rbac/RoleDetail.test.tsx:156,159,166-167` | PASS |
| C39 | in-use message, disabled Excluir | `RoleDetail > blocks deletion in use` PASS | (carried) `web/src/features/rbac/RoleDetail.test.tsx:174,176` | PASS |
| C40 | 404 shows `Papel não encontrado.` | `RoleDetail > shows not found` PASS | (carried) `web/src/features/rbac/RoleDetail.test.tsx:61` | PASS |
| C41 | checkbox per role, checked for held | `UserRoles > checks held roles` PASS | (carried) `web/src/features/rbac/UserRoles.test.tsx:45-46` | PASS |
| C42 | read only without rbac:assign | `UserRoles > read only without assign` PASS | (carried) `web/src/features/rbac/UserRoles.test.tsx:53-54` | PASS |
| C43 | confirm dialog, cancel, PUT body, success | `UserRoles > confirms and saves` PASS | (carried) `web/src/features/rbac/UserRoles.test.tsx:67-68,72,76,78-79` | PASS |
| C44 | 409 last-admin message | `UserRoles > shows last admin conflict` PASS | (carried) `web/src/features/rbac/UserRoles.test.tsx:94` | PASS |
| C45 | hidden without rbac:read, no request | `UserRoles > hidden without read` PASS | (carried) `web/src/features/rbac/UserRoles.test.tsx:104-105` | PASS |
| C46 | 500 on either roles request shows load error in section | `UserRoles > shows load error`, `shows load error for held roles` PASS | (carried) `web/src/features/rbac/UserRoles.test.tsx:115`; `:119,123` | PASS |
| C47 | /users/$id renders user heading and Papéis | `/users/$id route > composes user and roles` PASS | (carried) `web/src/routes/userDetail.test.tsx:22-23` | PASS |
| C48 | assembled e2e | `task e2e -- rbac.spec.ts` 1 passed | (carried) `web/e2e/rbac.spec.ts:34,40,42` | PASS |
| C49 | audit actions exactly the 4 names | C5/C9/C12/C16 proofs PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:58`; `app/internal/features/rbac/update_role/update_role_test.go:98`; `app/internal/features/rbac/delete_role/delete_role_test.go:44`; (refreshed) `app/internal/features/rbac/assign_roles/assign_roles_test.go:102` | PASS |
| C50 | /roles/new loading and load error with retry | `RoleForm > shows loading`, `RoleForm > shows error with retry` PASS | (carried) `web/src/features/rbac/RoleForm.test.tsx:101`; `:110`; `:111-113` | PASS |
| C51 | `Novo papel` with rbac:create, absent without | `RolesList > new role link only with rbac:create`, `hides new role link without rbac:create` PASS | (carried) `web/src/features/rbac/RolesList.test.tsx:84`; `:93-94` | PASS |

## Coverage

The round 3 fix touched one test file and no set's authority (no production code, route, state or branch). Rows are carried with their marks; the Landing doors row, whose door 2 cell referenced the C21 finding, is re-judged at 7d0cae6.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| route statuses, 8 routes (39 Surface + 8 `500` = 47) - carried from 7ad1473 | plan Surface + Surface note + `app/openapi.json` | contract: C25 set equality per op (`app/internal/app/rbac_test.go:128`); behaviour: every Surface status C1-C23; `500` behaviour PUT C22, others foundation (Swept `existing`) | - |
| access marker per RBAC operation (8) - carried from fe13da6 | `app/internal/features/rbac/register.go:22-29` | C23 | - |
| role validation inputs (9) - carried from fe13da6 | AC 8 | C8 | - |
| `PATCH` field combinations (3: name, permissions, both) - carried from 7ad1473 | AC 9 | name only C9 · permissions only C9 · both C9 | - |
| role guards (3) - carried from fe13da6 | AC 11, 13 | C11 · C13 · C14 | - |
| last-admin guard cases (5) - carried from fe13da6 | AC 19 | C20 | - |
| assignment outcomes (5) - carried from fe13da6 | AC 15-21 | C16 · C17 · C18 · C20/C21 · C22 | - |
| audit actions (4) - carried from fe13da6 | Landing door 3 | C5 · C9 · C12 · C16 | - |
| Relations edges (4) - carried from fe13da6 | plan Relations | C12 · C13 · C15/C16 · C16 | - |
| Landing doors (5) - verified at 7d0cae6 | plan Landing | door 1 C6, C7, C24 · door 2 C20, C21 (in-transaction interleave observed; F1 killed 5/5) · door 3 C23, C25, C49 · door 4 C26, C47 · door 5 C26, C27 | - |
| AC 27 forbidden by API 403 (3 screens) - carried from 7ad1473 | AC 27 | /roles C31 · /roles/new C31 · /roles/$id C31 | - |
| screen `/roles` states (4) + link `Novo papel` (2) - carried from 7ad1473 | Observable + `RolesList.tsx:19-21,27` | C28 · C29 · C30 · C31 · C51 shown · C51 hidden | - |
| screen `/roles/new` states (8) - carried from 7ad1473 | Observable + `RoleForm.tsx:37-41` + `roleErrors.ts:7-8` | C32 · C33 · C34 · C35 · C31 forbidden · C31 API 403 · C50 loading · C50 load error | - |
| screen `/roles/$id` states (12) - carried from 7ad1473 | Observable + `RoleDetail.tsx:81-84,97,134` | C29 · C30 · C31 · C31 API 403 · C33 · C34 · C35 · C36 · C37 · C38 · C39 · C40 | - |
| section `Papéis` states (7) - carried from 7ad1473 | Observable + `UserRoles.tsx:20,73,75` | C41 · C42 · C43 · C44 · C45 · C46 roles error · C46 held-roles error | - |
| menu link `Papéis` (2) - carried from fe13da6 | AC 27 | C31 both | - |
| startup assembly (2 places) - carried from fe13da6 | `app/internal/features/registry.go:17`, `app/internal/app/app.go:50` | C23, C25 · slice harness C1-C22 | - |
| auth-security rule 6 (1) - carried from fe13da6 | SKILL rule 6 | C16, C22 | - |

Round 2 open item, re-judged: door 2 lock (C21) -> proven with the interleave the claim states (both waiters in-transaction) and F1 killed 5/5.

## Test policy rows

Row 1 re-judged at 7d0cae6 (it classifies `assign_roles/endpoint.go`, whose proof the fix touched); rows 2-4 classify no touched file and are carried from 7ad1473 / fe13da6.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `features/rbac/{assign_roles,create_role,update_role,delete_role}/endpoint.go`, `features/rbac/role/role.go` | boundary C23, C25 · own layer C5-C22 | yes - every decision row asserted at its own layer; the lock row of `assign_roles` (`LockAdminRole ... FOR UPDATE`) has a proof whose stated interleave was observed and whose assertion kills F1 5/5 |
| Decides, not reached across a boundary | `web/src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.tsx`, `roleErrors.ts`, `web/src/features/users/UserMenu.tsx` | own layer C28-C47, C50, C51 | yes (carried from 7ad1473) |
| Entry point that decides nothing | `features/rbac/{list_permissions,list_roles,get_role,get_user_roles}/endpoint.go` | boundary C1-C4, C15, C19 | yes (carried from fe13da6) |
| Instrumentation, pass-throughs | `features/rbac/register.go`, `rbactest`, `web/src/features/rbac/{api,copy,States}.ts(x)`, routes | none of its own | yes (carried from fe13da6) |

Swept rows resolving to `existing` - carried from fe13da6: "dependency failure - database errors surface as `500` problem+json through foundation" holds (`app/internal/platform/httpx/problem.go:16`, `app/internal/platform/httpx/middleware.go:93`).

## Faults injected

verified at 7d0cae6 for F1; F2-F5 carried from 7ad1473. The fix touched one surface - the C21 proof (`assign_roles_test.go:202,212`), which guards `LockAdminRole ... FOR UPDATE` - so F1 was re-injected there. Scratch worktree `git worktree add --detach <scratchpad>/wt3 HEAD` (7d0cae6); baseline porcelain of the real tree `?? .claude/skills/auth-security/`, `?? .cursor/skills/auth-security/`; a temporary pg_locks diagnostic was added to the test in the scratch, run, then reverted with `git checkout` before F1; F1 regenerated sqlc with `go -C app run ./cmd/sqlcrun generate` in the scratch (generated `db/queries.sql.go` diff shows `FOR UPDATE` removed); worktree removed with `git worktree remove --force`; real-tree porcelain afterwards identical to baseline (`diff` empty).

| Mutation | Location | Killed |
| --- | --- | --- |
| F1 drop `FOR UPDATE` from `LockAdminRole` -> unmodified `TestAssignRoles_ConcurrentLastAdmin -count=5` - verified at 7d0cae6 | `app/internal/features/rbac/assign_roles/queries.sql:11` (`db/queries.sql.go`) | yes - 5/5 runs FAIL at `:218` `expected: []int{204, 409} actual: []int{204, 204}`; no `Condition never satisfied` in any run (the gate was met each time: both mutant requests park on the `sessions` DELETE inside their transactions) |
| F2 `roleErrors` maps 409 to `permissions` instead of `name` - carried from 7ad1473 | `web/src/features/rbac/roleErrors.ts:7` | yes |
| F3 `RoleForm` drops the API 403 branch - carried from 7ad1473 | `web/src/features/rbac/RoleForm.tsx:37` | yes |
| F4 `UserRoles` error state ignores `held.isError` - carried from 7ad1473 | `web/src/features/rbac/UserRoles.tsx:73` | yes |
| F5 `RolesList` shows `Novo papel` unconditionally - carried from 7ad1473 | `web/src/features/rbac/RolesList.tsx:27` | yes |

## Gate

verified at 7d0cae6.

`go -C app test -count=1 <35 rbac proofs> -v` - 48 passed (35 tests + 13 subtests), 0 failed
`go -C app test -count=5 ./internal/features/rbac/assign_roles -run '^TestAssignRoles_ConcurrentLastAdmin$'` - 5 passed, 0 failed
`npm --prefix web run test -- <6 files>` - 40 passed, 0 failed
`task e2e -- rbac.spec.ts` - 1 passed, 0 failed
`task gen:openapi:check` - exit 0; `npm --prefix web run gen:check` - exit 0
Full `task check` not run by the Verifier.

Ranked gaps: none.
