# RBAC verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 54895ff..7ad1473
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier)

Scope: fix diff `fe13da6..7ad1473` (tests in `app/internal/app/rbac_test.go`, `assign_roles_test.go`, `update_role_test.go`, `web/src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.test.tsx`; `checks.md` C9, C21, C25, C31, C34, C46, C50, C51; `plan.md` Surface note) plus every round 1 verdict that was not PASS. No production code changed in the fix diff.

## Binding sources

carried from fe13da6 - the fix did not touch the interface.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| none - the plan marks no source binding; profile is `standard`, step 1 does not apply | n/a | - | - |

## Checks

verified at 7ad1473 - every proof re-ran in full at `7ad1473`; each named test appears individually in the output:

- Go: `go -C app test -count=1 ./internal/features/rbac/... ./internal/app ./archtest ./migrations -run '^(<35 names>)$' -v` exit 0 - 35 top-level tests + 13 subtests `--- PASS` (48), 0 `--- FAIL`; includes the new `TestUpdateRole_UpdatesBothFields`
- Web: `npm --prefix web run test -- src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.test.tsx src/features/users/UserMenu.test.tsx src/routes/userDetail.test.tsx --reporter=verbose` exit 0 - 6 files, 40 passed; includes the new `RoleForm > forbidden on API 403`, `shows loading`, `shows error with retry`, `RoleDetail > forbidden on API 403`, `RolesList > new role link only with rbac:create`, `hides new role link without rbac:create`, `UserRoles > shows load error for held roles`
- e2e: `task e2e -- rbac.spec.ts` exit 0 - `rbac.spec.ts:5:1 admin grants a role and the user gets exactly its access` 1 passed
- `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0

Citations: files the fix touched were refreshed at 7ad1473; citations in untouched files (`list_permissions`, `list_roles`, `get_role`, `create_role`, `delete_role`, `get_user_roles` tests, `migrations/schema_test.go`, `archtest`, `UserMenu.test.tsx`, `userDetail.test.tsx`, `e2e/rbac.spec.ts`) are carried from fe13da6 (files byte-identical, `git diff fe13da6..7ad1473` empty for them).

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | catalogue sorted, has rbac:read/rbac:assign/zeta:do, no `*` | `TestListPermissions_ReturnsCatalogue` PASS | (carried) `app/internal/features/rbac/list_permissions/list_permissions_test.go:25` - `require.True(t, slices.IsSorted(items))`; `:26-29` Contains rbac:read, rbac:assign, zeta:do, NotContains `*` | PASS |
| C2 | order admin, Beta, Zeta; perms and counts | `TestListRoles_OrderedWithCounts` PASS | (carried) `app/internal/features/rbac/list_roles/list_roles_test.go:41` - `require.Equal(t, []string{"admin", "Beta", "Zeta"}, order)`; `:42-47` | PASS |
| C3 | GET role returns id, name, sorted perms, user_count | `TestGetRole_ReturnsRole` PASS | (carried) `app/internal/features/rbac/get_role/get_role_test.go:25` - `require.Equal(t, role.Role{ID: beta, Name: "Beta", ...})` | PASS |
| C4 | 404 random uuid / 422 `abc` on GET, PATCH, DELETE | `TestGetRole_404And422`, `TestUpdateRole_404And422`, `TestDeleteRole_404And422` PASS | `app/internal/features/rbac/get_role/get_role_test.go:36-37`; `app/internal/features/rbac/update_role/update_role_test.go:164-165`; `app/internal/features/rbac/delete_role/delete_role_test.go:102-103` - `http.StatusNotFound` / `http.StatusUnprocessableEntity` | PASS |
| C5 | 201 trimmed Leitor, one audit role.created | `TestCreateRole_Creates` PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:53-59` | PASS |
| C6 | duplicate name 409 on POST and PATCH, no change, no audit | `TestCreateRole_DuplicateNameIs409`, `TestUpdateRole_DuplicateNameIs409` PASS | `app/internal/features/rbac/create_role/create_role_test.go:68-70`; `app/internal/features/rbac/update_role/update_role_test.go:123-125` - `StatusConflict`, `nameOf == "Outro"`, audit zero | PASS |
| C7 | concurrent same name: one 201, one 409 | `TestCreateRole_ConcurrentSameName` PASS | (carried) `app/internal/features/rbac/create_role/create_role_test.go:82-83` | PASS |
| C8 | 6 rejected, 3 accepted, PATCH same 6 | `TestCreateRole_Validation`, `TestUpdateRole_Validation` PASS | `app/internal/features/rbac/create_role/create_role_test.go:103,108,110,119`; `app/internal/features/rbac/update_role/update_role_test.go:154-155,157-159` | PASS |
| C9 | PATCH name only / perms only / both in one call, audit before/after | `TestUpdateRole_UpdatesSentFields`, `TestUpdateRole_UpdatesBothFields` PASS | `app/internal/features/rbac/update_role/update_role_test.go:84,86-87,92-98` (single fields); `:175-178` - `require.Equal(t, "Gestor", got.Name)`, perms `[]string{"users:create", "users:read"}` in response and DB; `:180-181` before `snapshot{"Leitor", ["users:read"]}` / after `snapshot{"Gestor", [...]}`; `:182` `require.Equal(t, 1, ... audit_events)` | PASS |
| C10 | permission edit applies next request, sessions unchanged | `TestUpdateRole_AppliesOnNextRequest` PASS | `app/internal/features/rbac/update_role/update_role_test.go:110,114-115` | PASS |
| C11 | admin PATCH/DELETE 409, unchanged, no audit | `TestUpdateRole_AdminIs409`, `TestDeleteRole_AdminIs409` PASS | `app/internal/features/rbac/update_role/update_role_test.go:131-135`; `app/internal/features/rbac/delete_role/delete_role_test.go:59-61` | PASS |
| C12 | delete unused role 204, perms gone, audit with before | `TestDeleteRole_Deletes` PASS | (carried) `app/internal/features/rbac/delete_role/delete_role_test.go:39-42,52-53` | PASS |
| C13 | delete in-use role 409, rows intact | `TestDeleteRole_InUseIs409` PASS | (carried) `app/internal/features/rbac/delete_role/delete_role_test.go:70-74` | PASS |
| C14 | delete waits for uncommitted assignment then 409 | `TestDeleteRole_WaitsForConcurrentAssignment` PASS | (carried) `app/internal/features/rbac/delete_role/delete_role_test.go:91,96-97` | PASS |
| C15 | user roles sorted by name; empty list | `TestGetUserRoles_ReturnsRoles` PASS | (carried) `app/internal/features/rbac/get_user_roles/get_user_roles_test.go:37-38` | PASS |
| C16 | replace set, sessions 0, cookies 401, one audit before/after | `TestAssignRoles_ReplacesAndRevokes` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:93-97` - 204, `sorted(beta, gama)`, sessions zero, both probes 401; `:102,104-105` `{"roles":["Alfa"]}` / `{"roles":["Beta","Gama"]}` | PASS |
| C17 | same set reordered: 204, nothing changes | `TestAssignRoles_SameSetIsNoop` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:117-119` - `require.Equal(t, was, f.state(t, user))` | PASS |
| C18 | unknown/duplicate id -> 422 at body.role_ids | `TestAssignRoles_InvalidRoleIds` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:133-135` | PASS |
| C19 | 404 / 422 on GET and PUT users/{id}/roles | `TestGetUserRoles_404And422`, `TestAssignRoles_404And422` PASS | `app/internal/features/rbac/get_user_roles/get_user_roles_test.go:48-49`; `app/internal/features/rbac/assign_roles/assign_roles_test.go:141-142` | PASS |
| C20 | last-admin guard, 5 cases | `TestAssignRoles_LastAdminGuard` PASS (5 subtests shown) | `app/internal/features/rbac/assign_roles/assign_roles_test.go:184` - `require.Equal(t, c.want, rec.Code)`; `:186` state unchanged on 409; table `:153-157` | PASS |
| C21 | two concurrent admin removals "both held **inside their transactions** until each is waiting on a lock": one 204, one 409, one active admin; x3 | `TestAssignRoles_ConcurrentLastAdmin` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:218` - `require.Equal(t, []int{http.StatusNoContent, http.StatusConflict}, codes)`; `:219-220` one active admin holder. Outcome holds and F1 is now killed (13/13 runs). The stated precondition is false: the gate `:202` takes `LOCK TABLE sessions IN ACCESS EXCLUSIVE MODE`, which also blocks the session lookup `SELECT ... FROM sessions` in `app/internal/platform/auth/auth.go:99-101`, run before `platformdb.WithTx` (`app/internal/features/rbac/assign_roles/endpoint.go:68`). Diagnostic in the scratch worktree: both `pg_stat_activity` waiters counted at `:210-213` had `backend_xid` NULL, i.e. neither had reached `LockUser ... FOR UPDATE`; they wait in authentication, outside the transaction. The interleave is a simultaneous release, not the forced one the claim and the assertion message ("both requests must be inside their transactions") describe | FAIL |
| C22 | trigger failure -> 500, state unchanged | `TestAssignRoles_RollsBackTogether` PASS | `app/internal/features/rbac/assign_roles/assign_roles_test.go:237-238` - `StatusInternalServerError`, `require.Equal(t, was, f.state(t, user))` | PASS |
| C23 | 8 ops declare permission, 401, 403 without own, no writes | `TestRBAC_OperationAccess` PASS (8 subtests shown) | `app/internal/app/rbac_test.go:68` - `require.Len(t, declared, len(operations))`; `:81` declared permission; `:84` 401; `:94-98` 403, audit zero, roles/user_roles unchanged | PASS |
| C24 | ADMIN insert unique violation; Financeiro stored | `TestSchema_RoleNameUniqueIgnoringCase` PASS | (carried) `app/migrations/schema_test.go:63,67` | PASS |
| C25 | 8 ops each document exactly Surface statuses plus `500`; schema.d.ts regenerated | `TestOpenAPI_RBACStatuses` PASS; `task gen:openapi:check` exit 0; `npm --prefix web run gen:check` exit 0 | `app/internal/app/rbac_test.go:114-122` - `want` lists all 8 ops with the plan Surface statuses verbatim; `:127-128` - `documented := slices.Sorted(maps.Keys(operation.Responses))`, `require.Equal(t, slices.Sorted(slices.Values(append(statuses, "500"))), documented, key)` - set equality; plan `Surface` note records `+500` | PASS |
| C26 | archtest clean; web feature import rule with failing fixture | `TestImports_RepositoryIsClean`, `TestWebFeatures_DoNotImportEachOther` PASS | (carried) `app/archtest/imports_test.go:37`; `app/archtest/rbac_test.go:16-17,21` | PASS |
| C27 | lib/session.ts, lib/problems.ts exports; old files gone | `TestWebSharedSession_Moved` PASS | (carried) `app/archtest/rbac_test.go:27,31,35` | PASS |
| C28 | /roles table headers; `Todas`, `2` | `RolesList > shows the roles table` PASS | `web/src/features/rbac/RolesList.test.tsx:30,38-41` | PASS |
| C29 | role="status" while pending on /roles and /roles/$id | `RolesList > shows loading`, `RoleDetail > shows loading` PASS | `web/src/features/rbac/RolesList.test.tsx:47`; `web/src/features/rbac/RoleDetail.test.tsx:31` - `findByRole("status")` | PASS |
| C30 | 500 message and retry, both screens | `RolesList > shows error with retry`, `RoleDetail > shows error with retry` PASS | `web/src/features/rbac/RolesList.test.tsx:56,59`; `web/src/features/rbac/RoleDetail.test.tsx:41,44` | PASS |
| C31 | forbidden without rbac:read on 3 screens, no rbac call; API 403 on /roles, /roles/new, /roles/$id; menu link | `-t forbidden` on RolesList (2), RoleForm (2), RoleDetail (2); `UserMenu -t "roles link"` (2) PASS | `web/src/features/rbac/RolesList.test.tsx:65-66,75`; `web/src/features/rbac/RoleForm.test.tsx:88-89` (no rbac call); `:95` API 403 with `rbac:read`+`rbac:create` - `findByText("Você não tem permissão para acessar esta página.")`; `web/src/features/rbac/RoleDetail.test.tsx:50-51`; `:186` API 403 on `GET /roles/{id}` with `rbac:read` - same text; `web/src/features/users/UserMenu.test.tsx:35,42` | PASS |
| C32 | create sends body, navigates | `RoleForm > creates and navigates` PASS | `web/src/features/rbac/RoleForm.test.tsx:35-36` | PASS |
| C33 | grouped headings and checkboxes, both screens | `RoleForm > groups permissions`, `RoleDetail > groups permissions` PASS | `web/src/features/rbac/RoleForm.test.tsx:49`; `web/src/features/rbac/RoleDetail.test.tsx:71` | PASS |
| C34 | 409 message in `name-error`, referenced by the name input's `aria-describedby`, create and edit | `RoleForm > shows name conflict`, `RoleDetail > shows name conflict` PASS | `web/src/features/rbac/RoleForm.test.tsx:61-62` and `web/src/features/rbac/RoleDetail.test.tsx:107-108` - `expect(conflict.id).toBe("name-error")`, `expect(screen.getByLabelText("Nome")).toHaveAttribute("aria-describedby", "name-error")`; F2 killed | PASS |
| C35 | 422 messages under their fields | `RoleForm > shows field errors`, `RoleDetail > shows field errors` PASS | `web/src/features/rbac/RoleForm.test.tsx:81-82`; `web/src/features/rbac/RoleDetail.test.tsx:129-130` | PASS |
| C36 | PATCH only changed fields; success text | `RoleDetail > patches only changed fields` PASS | `web/src/features/rbac/RoleDetail.test.tsx:93,95` | PASS |
| C37 | admin locked | `RoleDetail > locks the admin role` PASS | `web/src/features/rbac/RoleDetail.test.tsx:137-144` | PASS |
| C38 | delete dialog, cancel, confirm | `RoleDetail > confirms deletion` PASS | `web/src/features/rbac/RoleDetail.test.tsx:156,159,166-167` | PASS |
| C39 | in-use message, disabled Excluir | `RoleDetail > blocks deletion in use` PASS | `web/src/features/rbac/RoleDetail.test.tsx:174,176` | PASS |
| C40 | 404 shows `Papel não encontrado.` | `RoleDetail > shows not found` PASS | `web/src/features/rbac/RoleDetail.test.tsx:61` | PASS |
| C41 | checkbox per role, checked for held | `UserRoles > checks held roles` PASS | `web/src/features/rbac/UserRoles.test.tsx:45-46` | PASS |
| C42 | read only without rbac:assign | `UserRoles > read only without assign` PASS | `web/src/features/rbac/UserRoles.test.tsx:53-54` | PASS |
| C43 | confirm dialog, cancel, PUT body, success | `UserRoles > confirms and saves` PASS | `web/src/features/rbac/UserRoles.test.tsx:67-68,72,76,78-79` | PASS |
| C44 | 409 last-admin message | `UserRoles > shows last admin conflict` PASS | `web/src/features/rbac/UserRoles.test.tsx:94` | PASS |
| C45 | hidden without rbac:read, no request | `UserRoles > hidden without read` PASS | `web/src/features/rbac/UserRoles.test.tsx:104-105` | PASS |
| C46 | 500 on either roles request shows load error in section | `UserRoles > shows load error`, `shows load error for held roles` PASS | `web/src/features/rbac/UserRoles.test.tsx:115` (`GET /rbac/roles` 500); `:119,123` - `GET ${rolesPath}` (`/api/v1/rbac/users/{id}/roles`, `:21`) 500, `within(element).findByText("Não foi possível carregar os papéis.")`; F4 killed | PASS |
| C47 | /users/$id renders user heading and Papéis | `/users/$id route > composes user and roles` PASS | (carried) `web/src/routes/userDetail.test.tsx:22-23` | PASS |
| C48 | assembled e2e | `task e2e -- rbac.spec.ts` 1 passed | (carried) `web/e2e/rbac.spec.ts:34,40,42` | PASS |
| C49 | audit actions exactly the 4 names | C5/C9/C12/C16 proofs PASS | `app/internal/features/rbac/create_role/create_role_test.go:58`; `app/internal/features/rbac/update_role/update_role_test.go:98`; `app/internal/features/rbac/delete_role/delete_role_test.go:44`; `app/internal/features/rbac/assign_roles/assign_roles_test.go:102` | PASS |
| C50 | /roles/new loading and load error with retry | `RoleForm > shows loading`, `RoleForm > shows error with retry` PASS | `web/src/features/rbac/RoleForm.test.tsx:101` - `findByRole("status")` with catalogue `pending`; `:110` message; `:111-113` click `Tentar novamente`, checkbox `users:read` appears, permissions requests `toHaveLength(2)` | PASS |
| C51 | `Novo papel` with rbac:create, absent without | `RolesList > new role link only with rbac:create`, `hides new role link without rbac:create` PASS | `web/src/features/rbac/RolesList.test.tsx:84` - `findByRole("link", { name: "Novo papel" })`; `:93-94` table rendered then `queryByRole("link", { name: "Novo papel" })).not.toBeInTheDocument()`; F5 killed | PASS |

## Coverage

Rows the fix touched recomputed at 7ad1473 from the plan's Surface (with the new `500` note), AC 9, AC 27, the Observable states and the web components' branches; other rows carried from fe13da6.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| route statuses, 8 routes (39 Surface + 8 `500` = 47) - verified at 7ad1473 | plan Surface + Surface note + `app/openapi.json` | contract: C25 asserts set equality per op (`app/internal/app/rbac_test.go:128`); behaviour: every Surface status C1-C23 as listed in checks.md; `500` behaviour PUT C22, others foundation (Swept `existing`) | - |
| access marker per RBAC operation (8) - carried from fe13da6 | `app/internal/features/rbac/register.go:22-29` | C23 | - |
| role validation inputs (9) - carried from fe13da6 | AC 8 | C8 | - |
| `PATCH` field combinations (3: name, permissions, both) - verified at 7ad1473 | AC 9 | name only C9 · permissions only C9 · both C9 (`TestUpdateRole_UpdatesBothFields`) | - |
| role guards (3) - carried from fe13da6 | AC 11, 13 | C11 · C13 · C14 | - |
| last-admin guard cases (5) - carried from fe13da6 | AC 19 | C20 | - |
| assignment outcomes (5) - carried from fe13da6 | AC 15-21 | C16 · C17 · C18 · C20/C21 · C22 | - |
| audit actions (4) - carried from fe13da6 | Landing door 3 | C5 · C9 · C12 · C16 | - |
| Relations edges (4) - carried from fe13da6 | plan Relations | C12 · C13 · C15/C16 · C16 | - |
| Landing doors (5) - verified at 7ad1473 | plan Landing | door 1 C6, C7, C24 · door 2 C20, C21 (F1 now killed 13/13; mechanism finding recorded under C21) · door 3 C23, C25, C49 · door 4 C26, C47 · door 5 C26, C27 | - |
| AC 27 forbidden by API 403 (3 screens) - verified at 7ad1473 | AC 27 | /roles C31 · /roles/new C31 (`RoleForm.test.tsx:95`, F3 killed) · /roles/$id C31 (`RoleDetail.test.tsx:186`) | - |
| screen `/roles` states (4) + link `Novo papel` (2) - verified at 7ad1473 | Observable + `RolesList.tsx:19-21,27` | C28 · C29 · C30 · C31 · `Novo papel` shown C51 · hidden C51 | - |
| screen `/roles/new` states (8) - verified at 7ad1473 | Observable + `RoleForm.tsx:37-41` + `roleErrors.ts:7-8` | C32 · C33 · C34 · C35 · C31 forbidden · C31 API 403 · C50 loading · C50 load error | - |
| screen `/roles/$id` states (12) - verified at 7ad1473 | Observable + `RoleDetail.tsx:81-84,97,134` | C29 · C30 · C31 · C31 API 403 · C33 · C34 · C35 · C36 · C37 · C38 · C39 · C40 | - |
| section `Papéis` states (7) - verified at 7ad1473 | Observable + `UserRoles.tsx:20,73,75` | C41 · C42 · C43 · C44 · C45 · C46 roles error · C46 held-roles error | - |
| menu link `Papéis` (2) - carried from fe13da6 | AC 27 | C31 both | - |
| startup assembly (2 places) - carried from fe13da6 | `app/internal/features/registry.go:17`, `app/internal/app/app.go:50` | C23, C25 · slice harness C1-C22 | - |
| auth-security rule 6 (1) - carried from fe13da6 | SKILL rule 6 | C16, C22 | - |

Round 1 unproven members, re-judged: `500` set mismatch -> proven (C25 equality); PATCH both -> proven (C9); /roles/new and /roles/$id API 403 -> proven (C31); 409 placement -> proven (C34); held-roles load error -> proven (C46); door 2 lock -> fault F1 killed, but see C21.

## Test policy rows

verified at 7ad1473 (both round 1 unmet rows re-judged; the other two rows classify no touched production file and are carried from fe13da6).

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `features/rbac/{assign_roles,create_role,update_role,delete_role}/endpoint.go`, `features/rbac/role/role.go` | boundary C23, C25 · own layer C5-C22 | yes - lock row of `assign_roles` now has a killing assertion (F1 killed 13/13), PATCH both-fields row asserted by C9; the mis-stated C21 precondition is recorded as a check finding |
| Decides, not reached across a boundary | `web/src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.tsx`, `roleErrors.ts`, `web/src/features/users/UserMenu.tsx` | own layer C28-C47, C50, C51 | yes - every row named in round 1 now asserted: `RoleForm.tsx:37` API 403 (C31), `:40-41` loading/load error (C50), `RoleDetail.tsx:81` API 403 (C31), `UserRoles.tsx:73` held-roles error (C46), `roleErrors.ts:7` 409 placement (C34), `RolesList.tsx:27` `Novo papel` (C51) |
| Entry point that decides nothing | `features/rbac/{list_permissions,list_roles,get_role,get_user_roles}/endpoint.go` | boundary C1-C4, C15, C19 | yes (carried from fe13da6) |
| Instrumentation, pass-throughs | `features/rbac/register.go`, `rbactest`, `web/src/features/rbac/{api,copy,States}.ts(x)`, routes | none of its own | yes (carried from fe13da6) |

Swept rows resolving to `existing` - carried from fe13da6: "dependency failure - database errors surface as `500` problem+json through foundation" holds (`app/internal/platform/httpx/problem.go:16`, `app/internal/platform/httpx/middleware.go:93`).

## Faults injected

verified at 7ad1473. Scratch worktree `git worktree add --detach <scratchpad>/wt2 HEAD`; baseline porcelain of the real tree `?? .claude/skills/auth-security/`, `?? .cursor/skills/auth-security/`; F1 regenerated sqlc with `go -C app run ./cmd/sqlcrun generate` in the scratch; web faults ran with a junction to the real `web/node_modules` (removed before worktree removal); each fault reverted with `git checkout` before the next; worktree removed with `git worktree remove --force`; real-tree porcelain afterwards identical to baseline. Surfaces chosen: the round 1 survivor plus four assertion surfaces the fix created.

| Mutation | Location | Killed |
| --- | --- | --- |
| F1 drop `FOR UPDATE` from `LockAdminRole` -> `TestAssignRoles_ConcurrentLastAdmin -count=3` | `app/internal/features/rbac/assign_roles/queries.sql:11` (`db/queries.sql.go`) | yes - 3/3 runs FAIL `expected []int{204, 409} actual []int{204, 204}`; repeated `-count=10`: 10/10 FAIL |
| F2 `roleErrors` maps 409 to `permissions` instead of `name` -> `-t "shows name conflict"` on RoleForm and RoleDetail | `web/src/features/rbac/roleErrors.ts:7` | yes - both FAIL, `Expected: "name-error" Received: "permissions-error"` |
| F3 `RoleForm` drops the API 403 branch (`if (!allowed)` only) -> `RoleForm > forbidden on API 403` | `web/src/features/rbac/RoleForm.tsx:37` | yes - forbidden text not found |
| F4 `UserRoles` error state ignores `held.isError` -> `UserRoles > shows load error for held roles` | `web/src/features/rbac/UserRoles.tsx:73` | yes - load error text not found |
| F5 `RolesList` shows `Novo papel` unconditionally -> `-t "new role link"` | `web/src/features/rbac/RolesList.tsx:27` | yes - `hides new role link without rbac:create` FAIL |

## Gate

verified at 7ad1473.

`go -C app test -count=1 <35 rbac proofs> -v` - 48 passed (35 tests + 13 subtests), 0 failed
`npm --prefix web run test -- <6 files>` - 40 passed, 0 failed
`task e2e -- rbac.spec.ts` - 1 passed, 0 failed
`task gen:openapi:check` - exit 0; `npm --prefix web run gen:check` - exit 0
Full `task check` not run by the Verifier.

Ranked gaps:

1. C21 - `app/internal/features/rbac/assign_roles/assign_roles_test.go:202,210-213`: the claim says both requests are held inside their transactions until each waits on a lock, but `LOCK TABLE sessions IN ACCESS EXCLUSIVE MODE` also blocks the authentication lookup (`app/internal/platform/auth/auth.go:99-101`), which runs before `WithTx` (`endpoint.go:68`); both counted waiters have `backend_xid` NULL. The kill of F1 (13/13) comes from a simultaneous release, not a forced interleave. Fix either the proof (a lock mode that blocks `DELETE FROM sessions` but not `SELECT`, e.g. `IN EXCLUSIVE MODE`, so the correct code parks one request on `sessions` and the other on the admin row, both in-transaction) or restate the claim.
