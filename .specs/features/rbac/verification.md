# RBAC verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 54895ff..fe13da6
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier)

## Binding sources

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none - the plan marks no source binding; profile is `standard`, step 1 does not apply | n/a | - | - |

## Checks

Proof runs at `fe13da6` (each named test appears individually in the output):

- Go: `go -C app test -count=1 ./internal/features/rbac/... ./internal/app ./archtest ./migrations -run '^(<34 names>)$' -v` exit 0 - 34 top-level tests + 13 subtests `--- PASS`, 0 `--- FAIL`
- Web: `npm --prefix web run test -- src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.test.tsx src/features/users/UserMenu.test.tsx src/routes/userDetail.test.tsx --reporter=verbose` exit 0 - 6 files, 33 passed
- e2e: `task e2e -- rbac.spec.ts` exit 0 - `rbac.spec.ts:5:1 admin grants a role and the user gets exactly its access` 1 passed
- `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | catalogue sorted, has rbac:read/rbac:assign/zeta:do, no `*` | `TestListPermissions_ReturnsCatalogue` PASS | `app/internal/features/rbac/list_permissions/list_permissions_test.go:25` - `require.True(t, slices.IsSorted(items))`; `:26-29` Contains rbac:read, rbac:assign, zeta:do, NotContains `*` | PASS |
| C2 | order admin, Beta, Zeta; perms and counts | `TestListRoles_OrderedWithCounts` PASS | `app/internal/features/rbac/list_roles/list_roles_test.go:41` - `require.Equal(t, []string{"admin", "Beta", "Zeta"}, order)`; `:42-47` admin `["*"]`, Beta `["users:create","users:read"]`/2, Zeta `[]`/0 | PASS |
| C3 | GET role returns id, name, sorted perms, user_count | `TestGetRole_ReturnsRole` PASS | `app/internal/features/rbac/get_role/get_role_test.go:25` - `require.Equal(t, role.Role{ID: beta, Name: "Beta", Permissions: []string{"users:create", "users:read"}, UserCount: 1}, ...)` | PASS |
| C4 | 404 random uuid / 422 `abc` on GET, PATCH, DELETE | `TestGetRole_404And422`, `TestUpdateRole_404And422`, `TestDeleteRole_404And422` PASS | `app/internal/features/rbac/get_role/get_role_test.go:36-37`; `app/internal/features/rbac/update_role/update_role_test.go:164-165`; `app/internal/features/rbac/delete_role/delete_role_test.go:102-103` - `require.Equal(t, http.StatusNotFound, ...)` / `http.StatusUnprocessableEntity` | PASS |
| C5 | 201 trimmed Leitor, perms, user_count 0, one audit role.created | `TestCreateRole_Creates` PASS | `app/internal/features/rbac/create_role/create_role_test.go:53-55` - name `"Leitor"`, perms, `user_count` 0; `:56-59` one `audit_events` row with `action = 'role.created' AND resource_type = 'role' AND resource_id = $1 AND actor_id = $2` | PASS |
| C6 | POST LEITOR and PATCH " leitor " -> 409, nothing changed, no audit | `TestCreateRole_DuplicateNameIs409`, `TestUpdateRole_DuplicateNameIs409` PASS | `app/internal/features/rbac/create_role/create_role_test.go:68-70` - `StatusConflict`, roles count unchanged, audit zero; `app/internal/features/rbac/update_role/update_role_test.go:123-125` - `StatusConflict`, `nameOf == "Outro"`, audit zero | PASS |
| C7 | concurrent same name: one 201, one 409, one row | `TestCreateRole_ConcurrentSameName` PASS | `app/internal/features/rbac/create_role/create_role_test.go:82-83` - `require.Equal(t, []int{http.StatusCreated, http.StatusConflict}, codes)`; one `Paralelo` | PASS |
| C8 | 6 rejected (422 at location), 3 accepted, PATCH same 6 | `TestCreateRole_Validation`, `TestUpdateRole_Validation` PASS | `app/internal/features/rbac/create_role/create_role_test.go:103,108` - 422 and `require.Contains(t, locations, c.location)`; `:110` no row; `:119` 201 for 1, 50 chars, `[]`; `app/internal/features/rbac/update_role/update_role_test.go:154-155,157-159` - 422 at location, nothing changed | PASS |
| C9 | PATCH name only / perms only, audit before/after | `TestUpdateRole_UpdatesSentFields` PASS | `app/internal/features/rbac/update_role/update_role_test.go:84,86-87` - perms kept, before/after snapshots; `:92-97` name kept, perms exactly `["users:create"]`, snapshots; `:98` two `role.updated` | PASS |
| C10 | permission edit applies next request, sessions unchanged | `TestUpdateRole_AppliesOnNextRequest` PASS | `app/internal/features/rbac/update_role/update_role_test.go:110,114-115` - probe 200 then `http.StatusForbidden`, sessions count equal | PASS |
| C11 | admin PATCH/DELETE 409, unchanged, no audit | `TestUpdateRole_AdminIs409`, `TestDeleteRole_AdminIs409` PASS | `app/internal/features/rbac/update_role/update_role_test.go:131-135` - 409 twice, name `admin`, perms `[]string{"*"}`, audit zero; `app/internal/features/rbac/delete_role/delete_role_test.go:59-61` - 409, `*` row remains, audit zero | PASS |
| C12 | delete unused role 204, perms gone, audit role.deleted with before | `TestDeleteRole_Deletes` PASS | `app/internal/features/rbac/delete_role/delete_role_test.go:39-42` - 204, role and `role_permissions` zero, one audit; `:52-53` before name `Leitor`, perms `["users:create","users:read"]` | PASS |
| C13 | delete in-use role 409, rows intact, no audit | `TestDeleteRole_InUseIs409` PASS | `app/internal/features/rbac/delete_role/delete_role_test.go:70-74` - 409, role, permission, `user_roles` rows 1, audit zero | PASS |
| C14 | delete waits for uncommitted assignment then 409 | `TestDeleteRole_WaitsForConcurrentAssignment` PASS | `app/internal/features/rbac/delete_role/delete_role_test.go:91` - `t.Fatalf("delete answered %d while the assignment was uncommitted")`; `:96-97` 409, assignment intact | PASS |
| C15 | user roles sorted by name; empty list | `TestGetUserRoles_ReturnsRoles` PASS | `app/internal/features/rbac/get_user_roles/get_user_roles_test.go:37` - `require.Equal(t, []held{{beta, "Beta"}, {zeta, "Zeta"}}, get(user))`; `:38` `[]held{}` | PASS |
| C16 | replace set, sessions 0, both cookies 401, one audit with before/after | `TestAssignRoles_ReplacesAndRevokes` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:92-96` - 204, roles `sorted(beta, gama)`, sessions zero, both probes 401; `:98,103-104` one audit, `{"roles":["Alfa"]}` / `{"roles":["Beta","Gama"]}` | PASS |
| C17 | same set reordered: 204, nothing changes, no audit | `TestAssignRoles_SameSetIsNoop` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:116-118` - 204, `require.Equal(t, was, f.state(t, user))` (roles, sessions, audits), audits zero | PASS |
| C18 | unknown id or duplicate -> 422 at body.role_ids, unchanged | `TestAssignRoles_InvalidRoleIds` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:132-134` - 422, `"location":"body.role_ids"`, state equal | PASS |
| C19 | 404 / 422 on GET and PUT users/{id}/roles | `TestGetUserRoles_404And422`, `TestAssignRoles_404And422` PASS | `app/internal/features/rbac/get_user_roles/get_user_roles_test.go:48-49`; `app/internal/features/rbac/assign_roles/assign_roles_test.go:140-141` | PASS |
| C20 | last-admin guard, 5 cases; 409 leaves state | `TestAssignRoles_LastAdminGuard` PASS (5 subtests shown) | `app/internal/features/rbac/assign_roles/assign_roles_test.go:183` - `require.Equal(t, c.want, rec.Code)`; `:185` state unchanged on 409; table `:152-156` | PASS |
| C21 | concurrent removal from 2 admins: one 204 one 409, x10 | `TestAssignRoles_ConcurrentLastAdmin` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:206-208` - `[]int{204, 409}` and one active admin holder; see Faults: the proof does not reliably detect removal of the lock | PASS |
| C22 | trigger failure -> 500, roles/sessions/audit unchanged | `TestAssignRoles_RollsBackTogether` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:225-226` - `StatusInternalServerError`, state equal | PASS |
| C23 | 8 ops declare permission, 401 no cookie, 403 others, no writes | `TestRBAC_OperationAccess` PASS (8 subtests shown) | `app/internal/app/rbac_test.go:66` - `require.Len(t, declared, len(operations))`; `:79` declared permission; `:82` 401; `:92-96` 403, audit zero, roles/user_roles unchanged | PASS |
| C24 | ADMIN insert unique violation; Financeiro stored | `TestSchema_RoleNameUniqueIgnoringCase` PASS | `app/migrations/schema_test.go:63` - `require.ErrorContains(t, err, "23505")`; `:67` `"Financeiro"` | PASS |
| C25 | openapi.json has the 8 ops with **exactly** the Surface statuses; schema.d.ts regenerated | `TestOpenAPI_RBACStatuses` PASS; `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0 | `app/internal/app/rbac_test.go:126` - `require.Contains(t, operation.Responses, status, key)` asserts a subset only; the committed `app/openapi.json` carries `500` on all 8 RBAC operations, a status absent from the plan's Surface, so "exactly" is both unasserted and untrue | FAIL |
| C26 | archtest clean on real module; web feature imports rule with failing fixture | `TestImports_RepositoryIsClean`, `TestWebFeatures_DoNotImportEachOther` PASS | `app/archtest/imports_test.go:37` - `require.Empty(t, v)`; `app/archtest/rbac_test.go:16-17` - fixture yields 1 violation `"rbac" imports feature "users"`; `:21` real `web/src` empty | PASS |
| C27 | lib/session.ts and lib/problems.ts exports; old files gone | `TestWebSharedSession_Moved` PASS | `app/archtest/rbac_test.go:27,31` - Contains `export const meQuery`, `export function useMe`, `export function can`, `export function fieldErrors`, `export const forbidden`; `:35` `require.ErrorIs(t, err, os.ErrNotExist, gone)` | PASS |
| C28 | /roles table headers; admin `Todas`, 2-perm role `2` | `RolesList > shows the roles table` PASS | `web/src/features/rbac/RolesList.test.tsx:30` - headers `["Nome", "Permissões", "Usuários"]`; `:38-41` rows `["admin","Todas","1"]`, `["Leitor","2","3"]` | PASS |
| C29 | role="status" while pending on /roles and /roles/$id | `RolesList > shows loading`, `RoleDetail > shows loading` PASS | `web/src/features/rbac/RolesList.test.tsx:47`; `web/src/features/rbac/RoleDetail.test.tsx:31` - `findByRole("status")` | PASS |
| C30 | 500 shows message and retry repeats request, both screens | `RolesList > shows error with retry`, `RoleDetail > shows error with retry` PASS | `web/src/features/rbac/RolesList.test.tsx:56,59` - message, 2 requests; `web/src/features/rbac/RoleDetail.test.tsx:41,44` | PASS |
| C31 | forbidden on 3 screens without rbac:read, no rbac call; API 403 on /roles; menu link | `-t forbidden` on RolesList/RoleForm/RoleDetail, `UserMenu -t "roles link"` PASS (2 menu tests shown) | `web/src/features/rbac/RolesList.test.tsx:65-66,75`; `web/src/features/rbac/RoleForm.test.tsx:86-87`; `web/src/features/rbac/RoleDetail.test.tsx:50-51`; `web/src/features/users/UserMenu.test.tsx:35,42` | PASS |
| C32 | create sends body, navigates to /roles/<id> | `RoleForm > creates and navigates` PASS | `web/src/features/rbac/RoleForm.test.tsx:35` - pathname `/roles/${created}`; `:36` body `{name: "Leitor", permissions: ["users:read"]}` | PASS |
| C33 | grouped headings and checkboxes in order, both screens | `RoleForm > groups permissions`, `RoleDetail > groups permissions` PASS | `web/src/features/rbac/RoleForm.test.tsx:49`; `web/src/features/rbac/RoleDetail.test.tsx:71` - `["# rbac", "rbac:read", "# users", "users:create", "users:read"]` | PASS |
| C34 | 409 shows message **under the name field**, create and edit | `RoleForm > shows name conflict`, `RoleDetail > shows name conflict` PASS | `web/src/features/rbac/RoleForm.test.tsx:60`; `web/src/features/rbac/RoleDetail.test.tsx:106` - only `findByText("Já existe um papel com este nome.")`; placement under the name field is not asserted (no `aria-describedby`/`name-error` check, unlike C35), so mapping 409 to the permissions field would pass | FAIL |
| C35 | 422 messages under their fields, both screens | `RoleForm > shows field errors`, `RoleDetail > shows field errors` PASS | `web/src/features/rbac/RoleForm.test.tsx:79-80`; `web/src/features/rbac/RoleDetail.test.tsx:127-128` - `aria-describedby` `name-error`, `permissions-error` id | PASS |
| C36 | PATCH only changed fields; `Alterações salvas.` each | `RoleDetail > patches only changed fields` PASS | `web/src/features/rbac/RoleDetail.test.tsx:95` - `[{ name: "Consulta" }, { permissions: ["users:create", "users:read"] }]`; `:89,96` message | PASS |
| C37 | admin locked message, disabled fields, no Salvar/Excluir | `RoleDetail > locks the admin role` PASS | `web/src/features/rbac/RoleDetail.test.tsx:135-142` | PASS |
| C38 | delete dialog text, cancel sends nothing, confirm DELETE + navigate | `RoleDetail > confirms deletion` PASS | `web/src/features/rbac/RoleDetail.test.tsx:154,157,164-165` | PASS |
| C39 | in-use message and disabled Excluir | `RoleDetail > blocks deletion in use` PASS | `web/src/features/rbac/RoleDetail.test.tsx:172,174` | PASS |
| C40 | 404 shows `Papel não encontrado.` | `RoleDetail > shows not found` PASS | `web/src/features/rbac/RoleDetail.test.tsx:61` | PASS |
| C41 | one checkbox per role, checked for held | `UserRoles > checks held roles` PASS | `web/src/features/rbac/UserRoles.test.tsx:45-46` - admin checked, Leitor not checked | PASS |
| C42 | without rbac:assign all disabled, no Salvar papéis | `UserRoles > read only without assign` PASS | `web/src/features/rbac/UserRoles.test.tsx:53-54` | PASS |
| C43 | confirm dialog text, cancel no PUT, confirm PUT body, success text | `UserRoles > confirms and saves` PASS | `web/src/features/rbac/UserRoles.test.tsx:67-68,72,76,78-79` | PASS |
| C44 | 409 shows last-admin message | `UserRoles > shows last admin conflict` PASS | `web/src/features/rbac/UserRoles.test.tsx:94` | PASS |
| C45 | no section and no rbac request without rbac:read | `UserRoles > hidden without read` PASS | `web/src/features/rbac/UserRoles.test.tsx:104-105` | PASS |
| C46 | 500 on **either** roles request shows load error in section | `UserRoles > shows load error` PASS | `web/src/features/rbac/UserRoles.test.tsx:115` - only `GET /api/v1/rbac/roles` returns 500; the `GET /api/v1/rbac/users/{id}/roles` failure (`held.isError`, `web/src/features/rbac/UserRoles.tsx:73`) has no proof - a claim about 2 cases proven on 1 | FAIL |
| C47 | /users/$id renders user heading and Papéis section | `/users/$id route > composes user and roles` PASS | `web/src/routes/userDetail.test.tsx:22-23` | PASS |
| C48 | assembled: create role, assign, user sees /users, forbidden on /roles | `task e2e -- rbac.spec.ts` 1 passed | `web/e2e/rbac.spec.ts:34` - `Papéis atualizados.`; `:40` users table contains reader; `:42` permission message | PASS |
| C49 | audit actions exactly the 4 names, each by its own check | C5/C9/C12/C16 proofs PASS | `app/internal/features/rbac/create_role/create_role_test.go:58`; `app/internal/features/rbac/update_role/update_role_test.go:98`; `app/internal/features/rbac/delete_role/delete_role_test.go:44`; `app/internal/features/rbac/assign_roles/assign_roles_test.go:101` | PASS |

## Coverage

Recomputed from the plan's Surface, Landing, Relations and AC text, and from the code (`app/openapi.json`, `op.Spec.Errors`, the web components' branches).

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| route statuses, 8 routes (39 Surface members) | plan Surface + `app/openapi.json` | every Surface status proven as listed in checks.md (C1-C23); `500` additionally emitted on all 8 ops: PUT by C22, others by foundation (Swept `existing`) | `openapi.json` set != Surface set (extra `500` x8), so C25's "exactly" is unproven |
| access marker per RBAC operation (8) | `app/internal/features/rbac/register.go:22-29` | C23, table over 8 | - |
| role validation inputs (9) | AC 8 | C8 | - |
| `PATCH` field combinations (3: name, permissions, both) | AC 9 "`name`, `permissions` or both" | name only C9 · permissions only C9 | both fields in one PATCH - no proof (checks.md replaced it with "duplicate name") |
| role guards (3) | AC 11, 13 + delete_role code | C11 · C13 · C14 | - |
| last-admin guard cases (5) | AC 19 + `assign_roles/endpoint.go:150` | C20 | - |
| assignment outcomes (5) | AC 15-21 | C16 · C17 · C18 · C20/C21 · C22 | - |
| audit actions (4) | Landing door 3 | C5 · C9 · C12 · C16 | - |
| Relations edges (4) | plan Relations | role_permissions cascade C12 · user_roles refuse delete C13 · users-user_roles C15/C16 · sessions deleted on change C16 | - |
| Landing doors (5) | plan Landing | door 1 C6, C7, C24 · door 2 C20, C21 · door 3 C23, C25, C49 · door 4 C26, C47 · door 5 C26, C27 | door 2 literal shape "`FOR UPDATE` before counting" - fault F1 survived; C21 does not reliably detect it |
| AC 27 forbidden by API 403 (3 screens) | AC 27 "or the API answers 403 THEN /roles, /roles/new and /roles/$id" | /roles C31 | /roles/new API 403 (`web/src/features/rbac/RoleForm.tsx:37`); /roles/$id API 403 (`web/src/features/rbac/RoleDetail.tsx:81`) - the latter named in checks.md's own Test policy prose ("detail maps 403/...") |
| screen `/roles` states (4) | Observable | C28 · C29 · C30 · C31 | - |
| screen `/roles/new` states (5) | Observable | C32 · C33 · C34 · C35 · C31 | 409 placement under name field (C34 FAIL) |
| screen `/roles/$id` states (11) | Observable | C29 · C30 · C31 · C33 · C34 · C35 · C36 · C37 · C38 · C39 · C40 | 409 placement under name field (C34 FAIL) |
| section `Papéis` states (6, error state has 2 sources) | Observable + `UserRoles.tsx:73` | C41 · C42 · C43 · C44 · C45 · C46 (roles list only) | load error from `GET /rbac/users/{id}/roles` |
| menu link `Papéis` (2) | AC 27 | C31 both | - |
| startup assembly (2 places) | `app/internal/features/registry.go:17`, `app/internal/app/app.go:50` | C23, C25 · slice harness C1-C22 | - |
| auth-security rule 6 (1) | SKILL rule 6 | C16, C22 (F2 killed) | - |

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `features/rbac/{assign_roles,create_role,update_role,delete_role}/endpoint.go`, `features/rbac/role/role.go` | boundary C23, C25 · own layer C5-C22 | no - gap: door-2 lock row of `assign_roles` survives fault F1; PATCH "both fields" row of `update_role` unasserted |
| Decides, not reached across a boundary | `web/src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.tsx`, `web/src/features/users/UserMenu.tsx` | own layer C28-C47 | no - unasserted rows: `RoleForm.tsx:37` API 403, `:40` loading, `:41` load error; `RoleDetail.tsx:81` API 403; `UserRoles.tsx:73` held-roles error; RoleForm/RoleDetail 409 placement (C34); `RolesList.tsx:27` `Novo papel` link with `rbac:create` |
| Entry point that decides nothing | `features/rbac/{list_permissions,list_roles,get_role,get_user_roles}/endpoint.go` | boundary C1-C4, C15, C19 | yes |
| Instrumentation, pass-throughs | `features/rbac/register.go`, `rbactest`, `web/src/features/rbac/{api,copy,States}.ts(x)`, routes | none of its own | yes - covered by C23, C28-C47 |

Swept rows resolving to `existing`: "dependency failure - database errors surface as `500` problem+json through foundation C3, C13" - constraint present: `app/internal/platform/httpx/problem.go:16` overrides `huma.NewError` (problem+json), `app/internal/platform/httpx/middleware.go:93` writes `500` problem on panic; exercised for RBAC by C22. Holds.

## Faults injected

Scratch worktree `git worktree add --detach <scratchpad>/wt HEAD`; baseline porcelain `?? .claude/skills/auth-security/`, `?? .cursor/skills/auth-security/`; each fault reverted with `git checkout -- app|web/src` before the next; worktree removed with `git worktree remove --force`; real-tree porcelain afterwards identical to baseline. Web fault ran with a junction to the real `web/node_modules` (removed before worktree removal).

| Mutation | Location | Killed |
| --- | --- | --- |
| F1 drop `FOR UPDATE` from `LockAdminRole` (lock-then-count, door 2) -> `TestAssignRoles_ConcurrentLastAdmin` | `app/internal/features/rbac/assign_roles/queries.sql:11` (`db/queries.sql.go:67`) | no - survived the first run (10 iterations PASS); `-count=3` rerun: PASS, FAIL (`[]int{204, 204}`), PASS - killed in 1 of 4 runs |
| F2 skip `DeleteUserSessions` -> `TestAssignRoles_ReplacesAndRevokes` | `app/internal/features/rbac/assign_roles/endpoint.go:121` | yes - "Should be zero, but was 2" |
| F3 drop `FOR UPDATE` from delete `LockRole` -> `TestDeleteRole_WaitsForConcurrentAssignment` | `app/internal/features/rbac/delete_role/queries.sql:2` (`db/queries.sql.go:35`) | yes - expected 409, actual 204 |
| F4 `role.CheckPermissions` accepts `*` -> `TestCreateRole_Validation` | `app/internal/features/rbac/role/role.go:51` | yes - expected 422, actual 201 |
| F5 `RolesList` drops `Todas` (always the count) -> `RolesList > shows the roles table` | `web/src/features/rbac/RolesList.tsx:11` | yes - `"Todas"` expected, `"1"` received |

## Gate

`go -C app test -count=1 <rbac proofs> -v` - 47 passed (34 tests + 13 subtests), 0 failed
`npm --prefix web run test -- <6 files>` - 33 passed, 0 failed
`task e2e -- rbac.spec.ts` - 1 passed, 0 failed
`task gen:openapi:check` - exit 0; `npm --prefix web run gen:check` - exit 0
Full `task check` not run by the Verifier.

Ranked gaps:

1. F1 surviving mutant - C21 / Landing door 2 - `app/internal/features/rbac/assign_roles/assign_roles_test.go:206`: removing the admin-role lock is caught in about 1 of 4 runs; the race window is too narrow for 10 iterations (needs a forced interleave, e.g. hold the admin row lock from a test transaction, or a barrier between count and delete).
2. C25 - `app/internal/app/rbac_test.go:126`: `Contains` does not prove "exactly", and `app/openapi.json` carries an extra `500` on all 8 RBAC ops; either assert set equality and add `500` to the Surface, or restate the claim.
3. AC 27 API 403 on `/roles/new` and `/roles/$id` unproven - `web/src/features/rbac/RoleForm.tsx:37`, `web/src/features/rbac/RoleDetail.tsx:81`.
4. C46 - `web/src/features/rbac/UserRoles.test.tsx:115`: only one of the two roles requests fails in the proof; `UserRoles.tsx:73` `held.isError` unproven.
5. C34 - `web/src/features/rbac/RoleForm.test.tsx:60`, `web/src/features/rbac/RoleDetail.test.tsx:106`: "under the name field" not asserted.
6. AC 9 PATCH with both `name` and `permissions` unproven (coverage set row dropped the member).
7. Web decision rows without an asserted case: RoleForm loading / load error (`RoleForm.tsx:40-41`), `Novo papel` link gated by `rbac:create` (`RolesList.tsx:27`).
