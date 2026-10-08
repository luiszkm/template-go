# RBAC checks

Profile: standard
Plan: `.specs/features/rbac/plan.md`

## Intent

51 checks in 8 slices · 5 one-way doors · 0 open

Comandos reais do repositório: `go -C app test <pkg> -run '<regex>'`, `npm --prefix web run test -- <file> -t "<name>"`
e `npm --prefix web run e2e -- <spec>`. Testes Go que tocam o banco usam `testkit.MigratedDB` (testcontainers,
Postgres 17, migrations reais); nenhum banco é simulado. Os slices de `rbac` não importam `features/users` nem nos
testes: para observar "a requisição seguinte" os testes registram uma operação própria com `Permission` ou
`Authenticated` no mesmo `api`.

## Checks

### S1 - Catálogo e leitura de papéis · ~8 files · ~25 KB · ~6k

**C1** - With `rbac` slices plus a test operation declaring `zeta:do` registered, `GET /api/v1/rbac/permissions` by a user holding `rbac:read` returns `200` with `items` sorted ascending, containing `rbac:read`, `rbac:assign` and `zeta:do`, and not containing `*` (RBAC-01, AC 1) `[done]`
Proof: `go -C app test ./internal/features/rbac/list_permissions -run '^TestListPermissions_ReturnsCatalogue$'`

**C2** - With roles `admin` (1 holder), `Zeta` (0 holders) and `Beta` (permissions `users:read`, `users:create`, 2 holders, one deactivated), `GET /api/v1/rbac/roles` returns `200` with `items` in the order `admin`, `Beta`, `Zeta`; `admin` has `permissions: ["*"]`; `Beta` has `permissions: ["users:create", "users:read"]` and `user_count: 2`; `Zeta` has `permissions: []` and `user_count: 0` (RBAC-01, AC 2) `[done]`
Proof: `go -C app test ./internal/features/rbac/list_roles -run '^TestListRoles_OrderedWithCounts$'`

**C3** - `GET /api/v1/rbac/roles/{id}` of `Beta` returns `200` with `id`, `name: "Beta"`, the sorted `permissions` and `user_count` (RBAC-01, AC 3) `[done]`
Proof: `go -C app test ./internal/features/rbac/get_role -run '^TestGetRole_ReturnsRole$'`

**C4** - A random UUID returns `404` and `abc` returns `422` on each of `GET`, `PATCH` and `DELETE` of `/api/v1/rbac/roles/{id}` (RBAC-01, AC 4) `[done]`
Proof: `go -C app test ./internal/features/rbac/get_role -run '^TestGetRole_404And422$'`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_404And422$'`
Proof: `go -C app test ./internal/features/rbac/delete_role -run '^TestDeleteRole_404And422$'`

### S2 - Criar, editar e excluir papéis · ~12 files · ~45 KB · ~11k

**C5** - `POST /api/v1/rbac/roles` with `name: "  Leitor "` and `permissions: ["users:read"]` returns `201` with a UUID `id`, `name: "Leitor"`, `permissions: ["users:read"]`, `user_count: 0`, and exactly one `audit_events` row with action `role.created`, `resource_type` `role`, `resource_id` = the new id, `actor_id` = the caller (RBAC-02, AC 5) `[done]`
Proof: `go -C app test ./internal/features/rbac/create_role -run '^TestCreateRole_Creates$'`

**C6** - With `Leitor` present, `POST` with `name: "LEITOR"` returns `409` and the `roles` count is unchanged; `PATCH` renaming another role to `" leitor "` returns `409` and that role keeps its name; neither writes an `audit_events` row (RBAC-02, AC 6, door 1) `[done]`
Proof: `go -C app test ./internal/features/rbac/create_role -run '^TestCreateRole_DuplicateNameIs409$'`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_DuplicateNameIs409$'`

**C7** - Two `POST /api/v1/rbac/roles` with `name: "Paralelo"` started together return one `201` and one `409`, and exactly one role `Paralelo` exists (RBAC-02, AC 7, door 1) `[done]`
Proof: `go -C app test ./internal/features/rbac/create_role -run '^TestCreateRole_ConcurrentSameName$'`

**C8** - `POST` returns `422` with an `errors` entry at `body.name` for name `""`, `"   "` and 51 characters, and at `body.permissions` for `["*"]`, `["nope:x"]` and `["users:read", "users:read"]`, creating no row; it accepts a 1-character name, a 50-character name and `permissions: []`; `PATCH` returns `422` at the same locations for the same 6 rejected inputs and changes nothing (RBAC-02, AC 8) `[done]`
Proof: `go -C app test ./internal/features/rbac/create_role -run '^TestCreateRole_Validation$'`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_Validation$'`

**C9** - On a role `Leitor` with `["users:read"]`: `PATCH {name: "Consulta"}` returns `200` with `permissions` still `["users:read"]`; `PATCH {permissions: ["users:create"]}` returns `200` with `name` still `Consulta` and `permissions` exactly `["users:create"]`; each writes one `audit_events` row `role.updated` whose `before` and `after` carry the name and permissions before and after; `PATCH {name: "Gestor", permissions: ["users:create", "users:read"]}` on `Leitor` changes both in one call and writes exactly one event (RBAC-02, AC 9) `[done]`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_UpdatesSentFields$'`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_UpdatesBothFields$'`

**C10** - A user whose only role grants `zeta:do` gets `200` on a test operation declaring `zeta:do`; after `PATCH` sets that role's `permissions` to `[]`, the same cookie gets `403` on that operation, and the user's `sessions` count is unchanged (RBAC-02, AC 10) `[done]`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_AppliesOnNextRequest$'`

**C11** - `PATCH` of `admin` (name or permissions) and `DELETE` of `admin` return `409`; `admin` still exists with exactly `*` and no `audit_events` row is written (RBAC-02, AC 11) `[done]`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_AdminIs409$'`
Proof: `go -C app test ./internal/features/rbac/delete_role -run '^TestDeleteRole_AdminIs409$'`

**C12** - `DELETE` of a role with 2 permissions and no holder returns `204`; the role and its `role_permissions` rows are gone and one `audit_events` row `role.deleted` has `resource_id` = the role id and `before` carrying its name and permissions (RBAC-02, AC 12) `[done]`
Proof: `go -C app test ./internal/features/rbac/delete_role -run '^TestDeleteRole_Deletes$'`

**C13** - `DELETE` of a role held by 1 user returns `409`; the role, its permissions and the `user_roles` row remain and no `audit_events` row is written (RBAC-02, AC 13) `[done]`
Proof: `go -C app test ./internal/features/rbac/delete_role -run '^TestDeleteRole_InUseIs409$'`

**C14** - While a test transaction that inserted a `user_roles` row for role `R` is still open, a `DELETE` of `R` does not answer until that transaction commits, and then returns `409` with the assignment intact (RBAC-02, AC 13, concurrency) `[done]`
Proof: `go -C app test ./internal/features/rbac/delete_role -run '^TestDeleteRole_WaitsForConcurrentAssignment$'`

### S3 - Atribuir papéis a um usuário · ~8 files · ~35 KB · ~9k

**C15** - `GET /api/v1/rbac/users/{id}/roles` for a user holding `Zeta` and `Beta` returns `200` with `items` `[{id, name: "Beta"}, {id, name: "Zeta"}]`; for a user with no role it returns `items: []` (RBAC-03, AC 14) `[done]`
Proof: `go -C app test ./internal/features/rbac/get_user_roles -run '^TestGetUserRoles_ReturnsRoles$'`

**C16** - For a user holding `Alfa` with 2 open sessions, `PUT /api/v1/rbac/users/{id}/roles` with the ids of `Gama` and `Beta` returns `204`; the user holds exactly `Beta` and `Gama`; `sessions` of that user is `0` and both cookies get `401` on a test `Authenticated` operation; exactly one `audit_events` row `user.roles_changed` has `resource_type` `user`, `resource_id` = the user id, `before` `{"roles": ["Alfa"]}` and `after` `{"roles": ["Beta", "Gama"]}` (RBAC-03, AC 15) `[done]`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_ReplacesAndRevokes$'`

**C17** - `PUT` with the user's current set, sent in a different order, returns `204`; the role set is unchanged, the user's `sessions` count is unchanged and no `audit_events` row is written (RBAC-03, AC 16) `[done]`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_SameSetIsNoop$'`

**C18** - `PUT` with a random UUID among `role_ids`, or with the same id twice, returns `422` with an `errors` entry at `body.role_ids`; the role set, the sessions and `audit_events` are unchanged (RBAC-03, AC 17) `[done]`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_InvalidRoleIds$'`

**C19** - A random UUID as `{id}` returns `404` and `abc` returns `422` on both `GET` and `PUT` of `/api/v1/rbac/users/{id}/roles` (RBAC-03, AC 18) `[done]`
Proof: `go -C app test ./internal/features/rbac/get_user_roles -run '^TestGetUserRoles_404And422$'`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_404And422$'`

**C20** - The last-admin guard, table-driven over 5 cases: removing `admin` from the only active admin returns `409`; removing it while the only other admin is deactivated returns `409`; removing it while another active admin exists returns `204`; adding a role to the only active admin while keeping `admin` returns `204`; removing `admin` from a deactivated admin while one active admin remains returns `204`; every `409` leaves roles, sessions and `audit_events` unchanged (RBAC-03, AC 19, door 2) `[done]`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_LastAdminGuard$'`

**C21** - With exactly two active admins `A` and `B`, a `PUT` removing `admin` from `A` and a `PUT` removing `admin` from `B` started together, both held inside their transactions until each is waiting on a lock (a test transaction holds `sessions` exclusively until `pg_stat_activity` shows two waiters), return one `204` and one `409`, and exactly one active user holds `admin` afterwards; repeated 3 times (RBAC-03, AC 20, door 2) `[done]`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_ConcurrentLastAdmin$'`

**C22** - With a test trigger that raises on `INSERT INTO user_roles`, a `PUT` that changes the set returns `500`, and the user's role set, the user's `sessions` count and the `audit_events` count are unchanged (RBAC-03, AC 21) `[done]`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_RollsBackTogether$'`

### S4 - Acesso às operações de RBAC · ~3 files · ~15 KB · ~4k

**C23** - Table-driven over the 8 RBAC operations of the assembled server: each declares exactly its permission (`list-permissions`, `list-roles`, `get-role` -> `rbac:read`; `create-role` -> `rbac:create`; `update-role` -> `rbac:update`; `delete-role` -> `rbac:delete`; `get-user-roles` -> `rbac:read`; `assign-roles` -> `rbac:assign`); each answers `401` without a cookie and `403` to a user holding every other `rbac:*` permission but its own, and the `403` writes no `audit_events` row and changes no `roles` or `user_roles` row (RBAC-04, AC 22, AC 23, door 3) `[done]`
Proof: `go -C app test ./internal/app -run '^TestRBAC_OperationAccess$'`

### S5 - Esquema, contrato e regras do repositório · ~6 files · ~25 KB · ~6k

**C24** - After `migrate up`, inserting a role `ADMIN` fails with a unique violation; a role `Financeiro` is stored as `Financeiro` (door 1) `[done]`
Proof: `go -C app test ./migrations -run '^TestSchema_RoleNameUniqueIgnoringCase$'`

**C25** - The committed `app/openapi.json` contains the 8 RBAC operations, each documenting exactly the statuses listed in the plan's `Surface` plus `500` (user decision, 2026-10-08, verification round 1), and `web/src/api/schema.d.ts` is regenerated from it (Surface, door 3) `[done]`
Proof: `go -C app test ./internal/app -run '^TestOpenAPI_RBACStatuses$'`
Proof: `task gen:openapi:check`
Proof: `npm --prefix web run gen:check`

**C26** - archtest reports no violation on the real module, where no package under `features/rbac` imports `features/users`; and no file under `web/src/features/<a>/` imports `@/features/<b>/` for `a != b`, which fails on a fixture where `features/rbac` imports `@/features/users/session` (AD-001, door 4, door 5) `[done]`
Proof: `go -C app test ./archtest -run '^TestImports_RepositoryIsClean$'`
Proof: `go -C app test ./archtest -run '^TestWebFeatures_DoNotImportEachOther$'`

**C27** - `web/src/lib/session.ts` exports `meQuery`, `useMe` and `can`, `web/src/lib/problems.ts` exports `fieldErrors` and `forbidden`, and `web/src/features/users/session.ts` and `problems.ts` no longer exist (door 5) `[done]`
Proof: `go -C app test ./archtest -run '^TestWebSharedSession_Moved$'`

### S6 - Web: telas de papéis · ~14 files · ~60 KB · ~15k

**C28** - `/roles` for a user with `rbac:read` shows a table with headers `Nome`, `Permissões` and `Usuários`; the `admin` row shows `Todas` and a role with 2 permissions shows `2` (RBAC-05, AC 24) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RolesList.test.tsx -t "shows the roles table"`

**C29** - While the roles request is pending an element with `role="status"` is visible, on `/roles` and on `/roles/$id` (RBAC-05, AC 25) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RolesList.test.tsx -t "shows loading"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "shows loading"`

**C30** - A `500` on the roles request shows `Não foi possível carregar os papéis.` and a button `Tentar novamente` that repeats the request, on `/roles` and on `/roles/$id` (RBAC-05, AC 26) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RolesList.test.tsx -t "shows error with retry"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "shows error with retry"`

**C31** - Without `rbac:read`, `/roles`, `/roles/new` and `/roles/$id` show `Você não tem permissão para acessar esta página.` and call no `/api/v1/rbac` route; with `rbac:read` and an API `403`, `/roles`, `/roles/new` and `/roles/$id` show the same text; the menu shows a link `Papéis` with `rbac:read` and none without it (RBAC-05, AC 27) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RolesList.test.tsx -t "forbidden"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "forbidden"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "forbidden"`
Proof: `npm --prefix web run test -- src/features/users/UserMenu.test.tsx -t "roles link"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "forbidden on API 403"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "forbidden on API 403"`

**C32** - Submitting `/roles/new` with name `Leitor` and `users:read` checked sends `POST /api/v1/rbac/roles` with `{name: "Leitor", permissions: ["users:read"]}` and navigates to `/roles/<returned id>` (RBAC-05, AC 28) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "creates and navigates"`

**C33** - With the catalogue `["rbac:read", "users:create", "users:read"]`, the role form shows a heading `rbac` followed by checkbox `rbac:read`, and a heading `users` followed by checkboxes `users:create` and `users:read`, on `/roles/new` and on `/roles/$id` (RBAC-05, AC 29) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "groups permissions"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "groups permissions"`

**C34** - A `409` on creation and a `409` on edition each show `Já existe um papel com este nome.` in the element `name-error`, which the name input references through `aria-describedby` (RBAC-05, AC 30) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "shows name conflict"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "shows name conflict"`

**C35** - A `422` with `errors` at `body.name` and `body.permissions` shows each message under its field, on creation and on edition (RBAC-05, AC 31) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "shows field errors"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "shows field errors"`

**C36** - On `/roles/$id`, changing only the name sends `PATCH` with `{name}` only; changing only the checked permissions sends `PATCH` with `{permissions}` only; each shows `Alterações salvas.` (RBAC-05, AC 32) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "patches only changed fields"`

**C37** - `/roles/$id` of `admin` shows `O papel admin tem todas as permissões e não pode ser alterado.`, the name input and every checkbox are disabled, and there is no `Salvar` and no `Excluir` button (RBAC-05, AC 33) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "locks the admin role"`

**C38** - Clicking `Excluir` opens a dialog reading `Excluir o papel Leitor?`; `Cancelar` closes it with no `DELETE` sent; `Excluir` in the dialog sends `DELETE /api/v1/rbac/roles/<id>` and navigates to `/roles` (RBAC-05, AC 34) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "confirms deletion"`

**C39** - A role with `user_count: 2` shows `Remova este papel dos 2 usuários antes de excluí-lo.` and a disabled `Excluir` (RBAC-05, AC 35) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "blocks deletion in use"`

**C40** - A `404` on the role request shows `Papel não encontrado.` (RBAC-05, AC 36) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "shows not found"`

### S7 - Web: papéis na tela do usuário · ~8 files · ~30 KB · ~8k

**C41** - With `rbac:read`, the section `Papéis` lists one checkbox per role from `GET /api/v1/rbac/roles`, checked exactly for the roles returned by `GET /api/v1/rbac/users/{id}/roles` (RBAC-06, AC 37) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "checks held roles"`

**C42** - With `rbac:read` and without `rbac:assign`, every checkbox is disabled and there is no `Salvar papéis` button (RBAC-06, AC 38) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "read only without assign"`

**C43** - With `rbac:assign`, checking `Leitor` and clicking `Salvar papéis` opens a dialog reading `Alterar os papéis de b@x.com? As sessões deste usuário serão encerradas.`; `Cancelar` sends no `PUT`; `Confirmar` sends `PUT /api/v1/rbac/users/<id>/roles` with the checked `role_ids` and shows `Papéis atualizados.` (RBAC-06, AC 39) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "confirms and saves"`

**C44** - A `409` on the `PUT` shows `Não é possível remover o último administrador.` (RBAC-06, AC 40) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "shows last admin conflict"`

**C45** - Without `rbac:read` the section renders nothing and no request to `/api/v1/rbac` is made (RBAC-06, AC 41) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "hidden without read"`

**C46** - A `500` on either roles request shows `Não foi possível carregar os papéis.` inside the section (Observable - seção `Papéis` error state) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "shows load error"`
Proof: `npm --prefix web run test -- src/features/rbac/UserRoles.test.tsx -t "shows load error for held roles"`

**C47** - The route `/users/$id` renders both the user detail (the email heading) and the section `Papéis` for an admin (door 4) `[done]`
Proof: `npm --prefix web run test -- src/routes/userDetail.test.tsx -t "composes user and roles"`

**C48** - Against the built binary, an admin creates role `Leitor` with `users:read` on `/roles/new`, assigns it on `/users/$id` of another user, and that user, after signing in, sees `/users` and gets the permission message on `/roles` (RBAC-03, RBAC-05, RBAC-06 assembled) `[done]`
Proof: `npm --prefix web run e2e -- rbac.spec.ts`

**C49** - The `audit_events` actions written by RBAC are exactly `role.created`, `role.updated`, `role.deleted` and `user.roles_changed`, each by its own check (door 3) `[done]`
Proof: `go -C app test ./internal/features/rbac/create_role -run '^TestCreateRole_Creates$'`
Proof: `go -C app test ./internal/features/rbac/update_role -run '^TestUpdateRole_UpdatesSentFields$'`
Proof: `go -C app test ./internal/features/rbac/delete_role -run '^TestDeleteRole_Deletes$'`
Proof: `go -C app test ./internal/features/rbac/assign_roles -run '^TestAssignRoles_ReplacesAndRevokes$'`

### S8 - Lacunas da verificação, rodada 1 · ~3 files · ~8 KB · ~2k

**C50** - On `/roles/new`, while the catalogue request is pending an element with `role="status"` is visible, and a `500` shows `Não foi possível carregar os papéis.` with `Tentar novamente` repeating the request (Test policy - `RoleForm.tsx` load states) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "shows loading"`
Proof: `npm --prefix web run test -- src/features/rbac/RoleForm.test.tsx -t "shows error with retry"`

**C51** - `/roles` shows the link `Novo papel` with `rbac:create` and does not show it without (Test policy - `RolesList.tsx` permission branch) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RolesList.test.tsx -t "new role link"`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| `GET /api/v1/rbac/permissions` statuses (3) | 200 C1 · 401 C23 · 403 C23 | - |
| `GET /api/v1/rbac/roles` statuses (3) | 200 C2 · 401 C23 · 403 C23 | - |
| `POST /api/v1/rbac/roles` statuses (5) | 201 C5 · 401 C23 · 403 C23 · 409 C6, C7 · 422 C8 | - |
| `GET /api/v1/rbac/roles/{id}` statuses (5) | 200 C3 · 401 C23 · 403 C23 · 404 C4 · 422 C4 | - |
| `PATCH /api/v1/rbac/roles/{id}` statuses (6) | 200 C9 · 401 C23 · 403 C23 · 404 C4 · 409 C6, C11 · 422 C4, C8 | - |
| `DELETE /api/v1/rbac/roles/{id}` statuses (6) | 204 C12 · 401 C23 · 403 C23 · 404 C4 · 409 C11, C13, C14 · 422 C4 | - |
| `GET /api/v1/rbac/users/{id}/roles` statuses (5) | 200 C15 · 401 C23 · 403 C23 · 404 C19 · 422 C19 | - |
| `PUT /api/v1/rbac/users/{id}/roles` statuses (6) | 204 C16, C17 · 401 C23 · 403 C23 · 404 C19 · 409 C20, C21 · 422 C18, C19 | - |
| access marker per RBAC operation (8) | C23, table-driven over all 8 | - |
| role validation inputs (9) | C8, table-driven over all 9 (6 rejected, 3 accepted) | - |
| `PATCH` field combinations (4) | name only C9 · permissions only C9 · both C9 · duplicate name C6 | - |
| role guards (3) | `admin` immutable C11 · in use C13 · in use under race C14 | - |
| last-admin guard cases (5) | C20, table-driven over all 5 | - |
| assignment outcomes (5) | changed C16 · same set C17 · invalid ids C18 · last admin C20, C21 · rollback C22 | - |
| session revocation triggers (2) | assignment change C16 · role permission edit leaves sessions C10 | - |
| audit actions (4) | `role.created` C5 · `role.updated` C9 · `role.deleted` C12 · `user.roles_changed` C16 | - |
| screen `/roles` states (4) | table C28 · loading C29 · error C30 · forbidden C31 | - |
| screen `/roles/new` states (8) | success C32 · grouped catalogue C33 · 409 C34 · 422 C35 · forbidden C31 · API 403 C31 · loading C50 · load error C50 | - |
| screen `/roles/$id` states (12) | loading C29 · error C30 · forbidden C31 · API 403 C31 · grouped catalogue C33 · 409 C34 · 422 C35 · patch C36 · admin locked C37 · delete confirm C38 · in use C39 · not found C40 | - |
| section `Papéis` states (7) | checked C41 · read only C42 · confirm and save C43 · 409 C44 · hidden C45 · roles load error C46 · held roles load error C46 | - |
| menu link `Papéis` (2) | shown with `rbac:read` C31 · hidden without C31 | - |
| link `Novo papel` (2) | shown with `rbac:create` C51 · hidden without C51 | - |
| startup assembly (2 places) | `features.Register` through `app.New` C23, C25 · test harness `testkit.NewAPI` C1-C22 | - |
| auth-security rule 6 (1) | role change revokes C16, C22 | - |
| Landing doors (5) | 1 C6, C7, C24 · 2 C20, C21 · 3 C23, C25, C49 · 4 C26, C47 · 5 C26, C27 | - |

- Claims naming a status code, route or response shape: C1-C23, C25 - each proof issues a real HTTP request against the slice's registered handler or the assembled server
- C23 and C48 prove the assembled path a second time; they do not stand in for the slice-level proofs C1-C22 nor for the component-level proofs C28-C46
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

- `features/rbac/assign_roles`: decides over user found, ids valid, duplicate, same set, last-admin guard - 5 branch points -> decides at the HTTP boundary of the slice; its HTTP test is its own layer (C16-C22)
- `features/rbac/{create_role,update_role}`: name and permission validation against the catalogue (6 rejected inputs), duplicate name, `admin` guard -> decides at the slice boundary (C5-C11)
- `features/rbac/delete_role`: found, `admin`, in use -> decides (C11-C14)
- `features/rbac/{list_permissions,list_roles,get_role,get_user_roles}`: a lookup and a lookup-or-404 -> entry point; accepted and rejected inputs proven (C1-C4, C15, C19)
- `web/src/features/rbac/{RolesList,RoleForm,RoleDetail,UserRoles}.tsx`: each maps a status or permission to a screen state (detail maps 403/404/error/409/422/admin/in use) -> decides, not reached across a boundary; one asserted case per row (C28-C46)
- `web/src/features/users/UserMenu.tsx`: gains one permission branch -> decides; both rows asserted (C31)
- closest analogue in the repo: `features/users/deactivate_user` (guard + revocation in one transaction, proven by its slice HTTP test) and `web/src/features/users/UserDetail.tsx` (status -> state mapping proven per row)

Cost: 22 slice-level Go proofs, 1 assembled table, 19 component proofs. Without them the 5-case last-admin guard would be proven only by the single path C48 crosses.

## Swept

- validation: C8, C18, C4, C19
- failure modes: C22 - assignment, revocation and audit roll back together; C11, C13 leave every row intact on refusal
- idempotency: C17 - re-sending the same set changes nothing; C6 - re-creating a name is a `409`, not a second row
- authorization: C23, C31, C42, C45; C11 and C20 guard the `admin` role against lock-out
- concurrency: C7 - unique index under concurrent creates; C14 - delete waits on an uncommitted assignment; C21 - last-admin guard under concurrent removals
- data lifecycle: C12, C13 - a role is deleted only when unused and its permissions go with it; `role_permissions` rows naming a permission later removed from code stay inert and disappear from the form on the next save (C33 lists only the catalogue)
- dependency failure: existing - database errors surface as `500` problem+json through foundation C3, C13; no new external dependency
- state transitions: C16, C17, C20 - the set of `admin` holders never reaches zero active users
- observability: C5, C9, C12, C16 - every mutation writes its audit event with actor and `request_id` (users C28); no new log line

## Handoff

Intended split, written before any code. Existing files the build reads as patterns (`features/users` slices,
`userstest`, `testkit`, `op`, `auth`, `audit`, `app`, `migrations`, `archtest`, web `features/users`, routes,
`test/render.tsx`, e2e) = ~120 KB = ~30k. Written per slice: S1 ~6k · S2 ~11k · S3 ~9k · S4 ~4k · S5 ~6k · S6 ~15k
· S7 ~8k = ~59k, plus the ~30k read = ~89k, under the 150k budget -> one builder. If the running total passes
150k, hand off after S5, where the surface changes from Go to web.
