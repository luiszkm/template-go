# Users checks

Profile: standard
Plan: `.specs/features/users/plan.md`

## Intent

84 checks in 8 slices · 16 one-way doors · 0 open

Comandos reais do repositório: `go -C app test <pkg> -run '<regex>'`, `npm --prefix web run test -- <file> -t "<name>"`
e `npm --prefix web run e2e -- <spec>` (Playwright contra o binário, via `task e2e:serve`). Testes Go que tocam o
banco usam `testkit.StartPostgres` (testcontainers, Postgres 17) e aplicam as migrations reais; nenhum banco é
simulado. Um verificador de senha contável (`password.Verifier`) pode ser injetado no slice `login` para provar
que a verificação rodou - é a única dupla, e não substitui o banco.

## Checks

### S1 - Primeiro admin pela CLI e formato de senha · ~8 files · ~45 KB · ~11k

**C1** - `api users create-admin --email "  Ana@X.com " --name Ana` with `senha-longa-123` on stdin exits `0`, prints the new user id (a UUID) on stdout, and leaves one active user with email `ana@x.com` holding role `admin` and one `audit_events` row with action `user.created`, `actor_id` null, `resource_id` equal to the printed id (USR-01, AC 1)
Proof: `go -C app test ./cmd/api -run '^TestCreateAdmin_CreatesActiveAdmin$'`

**C2** - With `ana@x.com` already present, `create-admin --email A@x.com` exits `1`, stderr contains `already exists`, and the `users` count is unchanged (USR-01, AC 2)
Proof: `go -C app test ./cmd/api -run '^TestCreateAdmin_ExistingEmailExits1$'`

**C3** - `create-admin` exits `2` with a line starting `usage:` on stderr and creates no row for each of 4 inputs: no `--email`, no `--name`, an 11-character password, a 129-character password; passwords of exactly 12 and 128 characters are accepted (USR-01, AC 3)
Proof: `go -C app test ./cmd/api -run '^TestCreateAdmin_InvalidInputExits2$'`
Proof: `go -C app test ./internal/features/users/bootstrap -run '^TestCreateAdmin_PasswordBounds$'`

**C4** - `password.Hash` returns a string matching `^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$` (16-byte salt, 32-byte key); two hashes of the same password differ; `Verify` is true for the password and false for any other (USR-01, AC 4, door 3)
Proof: `go -C app test ./internal/features/users/password -run '^TestHash_PinnedPHCFormat$'`
Proof: `go -C app test ./internal/features/users/password -run '^TestVerify_MatchesOnlyThePassword$'`

**C5** - A PHC hash built with `m=19456,t=2,p=1` verifies with its own parameters and `NeedsRehash` reports `true`; a hash from `password.Hash` reports `false` (USR-01, AC 5, door 3)
Proof: `go -C app test ./internal/features/users/password -run '^TestVerify_ReadsParamsFromHash$'`

**C6** - After a successful login for a user stored with an `m=19456,t=2,p=1` hash, `users.password_hash` starts with `$argon2id$v=19$m=65536,t=3,p=4$` and the same password still logs in (USR-01, AC 5)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_RehashesOutdatedHash$'`

**C7** - Every path that writes a password - `create-admin`, `POST /api/v1/users`, `PUT /api/v1/users/me/password` - stores a value matching the C4 pattern and never the plaintext (USR-01, AC 4)
Proof: `go -C app test ./cmd/api -run '^TestCreateAdmin_CreatesActiveAdmin$'`
Proof: `go -C app test ./internal/features/users/create_user -run '^TestCreateUser_StoresArgon2idHash$'`
Proof: `go -C app test ./internal/features/users/change_password -run '^TestChangePassword_StoresArgon2idHash$'`

### S2 - Sessão: login, logout, quem sou eu · ~14 files · ~70 KB · ~18k

**C8** - A correct login returns `204` with exactly one `Set-Cookie` whose value matches `^[A-Za-z0-9_-]{43}$` and whose attributes are `Path=/`, `Max-Age=43200`, `HttpOnly`, `SameSite=Lax`; `Secure` is present with `COOKIE_SECURE=true` and absent with `false`; one `audit_events` row `session.created` has `actor_id` = the user (USR-02, AC 6, door 4)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_SetsSessionCookie$'`

**C9** - After login, `sessions` holds one row whose `token_hash` equals SHA-256 of the cookie value, and no column of that row, cast to text or hex, equals the cookie value (USR-02, AC 7, door 4)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_StoresOnlyTokenHash$'`

**C10** - `auth.NewToken` returns a 43-character base64url string that decodes to 32 bytes, and 1000 calls produce 1000 distinct tokens (USR-02, AC 7, door 4)
Proof: `go -C app test ./internal/platform/auth -run '^TestNewToken_32RandomBytes$'`

**C11** - A login sent with a valid `session` cookie returns a different token; the session row of the sent token no longer exists and that token gets `401` on `GET /api/v1/users/me` (USR-02, AC 8)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_RotatesExistingSession$'`

**C12** - Unknown email, wrong password and deactivated user each return `401` with `detail` `invalid email or password`, and the three bodies are equal once `request_id` and `instance` are removed (USR-02, AC 9)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_FailuresAreIndistinguishable$'`

**C13** - A login for an unknown email calls the injected verifier exactly once, against the fixed dummy hash; over 3 runs the median latency of unknown-email logins is at least half the median of wrong-password logins (USR-02, AC 10)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_UnknownEmailVerifiesDummyHash$'`
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_UnknownEmailTimingMatches$'`

**C14** - Login returns `422` for each of 4 bodies: `email` `nope`, `email` absent, `password` empty, `password` of 129 characters (USR-02, AC 11)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_InvalidBody422$'`

**C15** - After 5 failed logins for `x@y.com` - existing user or not - the 6th login, even with the correct password, returns `429` with a `Retry-After` integer between `1` and `900`, and the injected verifier is not called; after only 4 failures the correct password returns `204` (USR-02, AC 12)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_RateLimitPerEmail$'`

**C16** - After 20 failed logins from one client IP across 20 different emails, the 21st from that IP returns `429` with `Retry-After`, while a login from another IP returns `204`; a client behind an untrusted peer that varies `X-Forwarded-For` on each attempt is still limited (USR-02, AC 13, door 9)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_RateLimitPerIP$'`

**C17** - 5 failure records dated 16 minutes ago do not block a login; 5 dated 14 minutes ago do (USR-02, AC 12, AC 13, door 10)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_RateLimitWindowIs15Minutes$'`

**C18** - 4 failures, a success, then 4 more failures for the same email still let the correct password return `204` (USR-02, AC 14)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_SuccessClearsEmailFailures$'`

**C19** - `DELETE /api/v1/users/session` with a valid cookie returns `204` with `Set-Cookie: session=; Path=/; Max-Age=0`, deletes the row, records `session.deleted`, and the old cookie then gets `401`; without a cookie it returns `401` (USR-02, AC 15)
Proof: `go -C app test ./internal/features/users/logout -run '^TestLogout_DeletesSession$'`
Proof: `go -C app test ./internal/features/users/logout -run '^TestLogout_WithoutSession401$'`

**C20** - `GET /api/v1/users/me` returns `200` with `id`, `email`, `name`, `permissions`: `[]` for a user with no role, `["users:create","users:read"]` for a role granting those two, and for `admin` the sorted list of every permission registered in `op`, without `*` (USR-02, AC 16)
Proof: `go -C app test ./internal/features/users/me -run '^TestMe_ReturnsPermissions$'`

**C21** - The auth middleware answers `401` problem+json without running the handler for each of 4 cases on a non-public operation: no cookie, unknown token, session created `SESSION_TTL + 1s` ago, deactivated user; a session created `SESSION_TTL - 1m` ago passes (USR-02, AC 17)
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_Rejects401$'`
Proof: `go -C app test ./internal/features/users/me -run '^TestMe_401WithoutSession$'`

**C22** - Config defaults are `SESSION_TTL=12h`, `COOKIE_SECURE=true`, `TRUSTED_PROXIES` empty; `TRUSTED_PROXIES="10.0.0.0/8, 192.168.1.1/32"` parses into 2 prefixes; `TRUSTED_PROXIES=nope` fails `config.Load` with a message containing `TRUSTED_PROXIES` (USR-02, door 15)
Proof: `go -C app test ./internal/platform/config -run '^TestDefaults_UsersConfig$'`
Proof: `go -C app test ./internal/platform/config -run '^TestTrustedProxies_Parse$'`

**C23** - `api serve` with `SESSION_TTL=1h` and `COOKIE_SECURE=false` issues a login cookie with `Max-Age=3600` and no `Secure` (USR-02, door 15 - assembly in `cmd/api`)
Proof: `go -C app test ./cmd/api -run '^TestServe_AppliesSessionConfig$'`

### S3 - Autorização e auditoria · ~12 files · ~55 KB · ~14k

**C24** - `op.Register` accepts each spec with exactly one of `Permission`, `Public`, `Authenticated` (3 cases) and rejects, with an error containing the operation ID, the 5 others: none, and each of the 4 combinations of two or three markers (USR-03, AC 18, door 5)
Proof: `go -C app test ./internal/platform/op -run '^TestRegister_ExactlyOneAccessMarker$'`

**C25** - The middleware answers `403` problem+json without running the handler for a user with no role and for a user whose role grants only `other:thing`, on an operation requiring `x:y` (USR-03, AC 19)
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_Forbids403$'`

**C26** - A user whose role grants `*` runs an operation requiring `x:y`; a user whose role grants exactly `x:y` also runs it (USR-03, AC 20, door 7)
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_WildcardAndExactAllow$'`

**C27** - An `Authenticated` operation runs for a valid session with no role; a `Public` operation runs with no cookie at all (USR-03, AC 21, door 5)
Proof: `go -C app test ./internal/platform/auth -run '^TestMiddleware_AuthenticatedAndPublic$'`

**C28** - `audit.Record` inside a committed `db.WithTx` leaves exactly one `audit_events` row carrying `actor_id` from the principal, `action`, `resource_type`, `resource_id`, `before`, `after`, `ip` from `httpx.ClientIP` and `request_id` from the context (USR-03, AC 22, door 8)
Proof: `go -C app test ./internal/platform/audit -run '^TestRecord_CommitsWithTx$'`

**C29** - `audit.Record` inside a `db.WithTx` whose function returns an error leaves no `audit_events` row and no row of the mutation (USR-03, AC 23)
Proof: `go -C app test ./internal/platform/audit -run '^TestRecord_RolledBackWithTx$'`

**C30** - `audit.Record` given `before`/`after` values that contain the keys `password_hash`, `password`, `current_password` or `new_password`, at any depth, stores JSON without any of those 4 keys (USR-03, AC 24)
Proof: `go -C app test ./internal/platform/audit -run '^TestRecord_RedactsPasswordKeys$'`

**C31** - `UPDATE audit_events SET action = 'x'` and `DELETE FROM audit_events` each raise an error and leave the row unchanged (USR-03, AC 25, door 8)
Proof: `go -C app test ./internal/platform/audit -run '^TestAuditEvents_AppendOnly$'`

**C32** - With `TRUSTED_PROXIES=10.0.0.0/8` and peer `10.0.0.5`: `X-Forwarded-For: 203.0.113.9, 10.0.0.7` gives `203.0.113.9`; `1.1.1.1, 203.0.113.9` gives `203.0.113.9`; `10.0.0.8, 10.0.0.7` gives `10.0.0.8`; an unparsable entry gives the peer `10.0.0.5` (USR-03, AC 26, door 9)
Proof: `go -C app test ./internal/platform/httpx -run '^TestClientIP_TrustedPeerReadsForwardedFor$'`

**C33** - With peer `198.51.100.4` outside `TRUSTED_PROXIES`, `X-Forwarded-For: 1.2.3.4` is ignored and the client IP is `198.51.100.4`; with `TRUSTED_PROXIES` empty, the same holds for any peer (USR-03, AC 27, door 9)
Proof: `go -C app test ./internal/platform/httpx -run '^TestClientIP_UntrustedPeerIgnoresForwardedFor$'`

### S4 - Gestão de usuários pela API · ~22 files · ~95 KB · ~24k

**C34** - `POST /api/v1/users` as admin with `email` `B@X.com `, `name` `Bia`, a 12-character password returns `201` whose JSON keys are exactly `id`, `email`, `name`, `active`, `created_at`, with `email` `b@x.com` and `active` `true`; one `user.created` audit row has the admin as actor (USR-04, AC 28)
Proof: `go -C app test ./internal/features/users/create_user -run '^TestCreateUser_201$'`

**C35** - Creating ` b@x.com` after `B@x.com` exists returns `409` and the `users` count is unchanged (USR-04, AC 29, door 2)
Proof: `go -C app test ./internal/features/users/create_user -run '^TestCreateUser_DuplicateEmail409$'`

**C36** - 10 concurrent `POST /api/v1/users` with the same email produce exactly one `201`, nine `409`, and one row (USR-04, AC 30, door 2)
Proof: `go -C app test ./internal/features/users/create_user -run '^TestCreateUser_ConcurrentSameEmail$'`

**C37** - `POST /api/v1/users` returns `422` with an `errors[].location` of `body.email`, `body.name` or `body.password` for each of 5 bodies: email `nope`, name empty, name of 101 characters, password of 11, password of 129; names of 1 and 100 and passwords of 12 and 128 characters return `201` (USR-04, AC 31)
Proof: `go -C app test ./internal/features/users/create_user -run '^TestCreateUser_Validation$'`

**C38** - With 52 users, `GET /api/v1/users` returns `200`, 50 items ordered by email ascending and `total` `52`; `offset=50` returns the last 2 (USR-04, AC 32)
Proof: `go -C app test ./internal/features/users/list_users -run '^TestListUsers_PagesByEmail$'`

**C39** - `GET /api/v1/users` returns `422` for `limit=0`, `limit=101` and `offset=-1`, and `200` for `limit=1` and `limit=100` (USR-04, AC 33)
Proof: `go -C app test ./internal/features/users/list_users -run '^TestListUsers_InvalidPage422$'`

**C40** - `GET /api/v1/users/{id}` returns `200` with `id`, `email`, `name`, `active`, `created_at`, `deactivated_at`: `null` for an active user, a timestamp for a deactivated one (USR-04, AC 34)
Proof: `go -C app test ./internal/features/users/get_user -run '^TestGetUser_200$'`

**C41** - A random UUID returns `404` and `abc` returns `422` on each of the 4 routes taking `{id}`: `GET`, `PATCH`, `deactivate`, `activate` (USR-04, AC 35)
Proof: `go -C app test ./internal/features/users/get_user -run '^TestGetUser_404And422$'`
Proof: `go -C app test ./internal/features/users/update_user -run '^TestUpdateUser_404And422$'`
Proof: `go -C app test ./internal/features/users/deactivate_user -run '^TestDeactivateUser_404And422$'`
Proof: `go -C app test ./internal/features/users/activate_user -run '^TestActivateUser_404And422$'`

**C42** - `PATCH` with only `name` leaves `email` unchanged, with only `email` leaves `name` unchanged, with both changes both; each returns `200` with the updated user and records one `user.updated` audit row whose `before` and `after` differ only in the fields sent (USR-04, AC 36)
Proof: `go -C app test ./internal/features/users/update_user -run '^TestUpdateUser_PartialFields$'`

**C43** - `PATCH` setting another user's email returns `409` and changes nothing; setting the user's own current email returns `200`; an invalid email, an empty name or an empty body `{}` returns `422` (USR-04, AC 37, AC 36)
Proof: `go -C app test ./internal/features/users/update_user -run '^TestUpdateUser_ConflictAndValidation$'`

**C44** - Deactivating a user with 2 open sessions returns `204`, sets `deactivated_at`, leaves 0 sessions for that user, records `user.deactivated`, and both of that user's cookies get `401` on the very next `GET /api/v1/users/me` (USR-04, AC 38, AC 17)
Proof: `go -C app test ./internal/features/users/deactivate_user -run '^TestDeactivateUser_RevokesSessions$'`

**C45** - A caller deactivating their own id gets `409`; `deactivated_at` stays null and the caller's session still answers `200` on `/me` (USR-04, AC 39)
Proof: `go -C app test ./internal/features/users/deactivate_user -run '^TestDeactivateUser_SelfIs409$'`

**C46** - Deactivating an already deactivated user, and activating an already active user, each return `204`, leave `deactivated_at` unchanged and add no `audit_events` row (USR-04, AC 40)
Proof: `go -C app test ./internal/features/users/deactivate_user -run '^TestDeactivateUser_AlreadyDeactivatedNoop$'`
Proof: `go -C app test ./internal/features/users/activate_user -run '^TestActivateUser_AlreadyActiveNoop$'`

**C47** - Activating a deactivated user returns `204`, clears `deactivated_at`, records `user.activated`, and that user can log in again (USR-04, AC 41)
Proof: `go -C app test ./internal/features/users/activate_user -run '^TestActivateUser_Reactivates$'`

**C48** - On the assembled server, each of the 9 non-public operations answers `401` with no cookie, and each of the 6 operations requiring a `users:*` permission answers `403` to a signed-in user with no role (USR-04, USR-03, AC 17, AC 19)
Proof: `go -C app test ./internal/app -run '^TestUsersRoutes_401WithoutSession$'`
Proof: `go -C app test ./internal/app -run '^TestUsersRoutes_403WithoutPermission$'`

**C49** - The assembled server registers the 10 user operations with these markers: login `Public`; logout, me, change password `Authenticated`; list and get `users:read`; create `users:create`; update `users:update`; deactivate `users:deactivate`; activate `users:activate` (USR-04, AC 18)
Proof: `go -C app test ./internal/app -run '^TestUsersOperations_AccessMarkers$'`

### S5 - Trocar a própria senha · ~4 files · ~15 KB · ~4k

**C50** - `PUT /api/v1/users/me/password` with the right current password and a 12-character new one returns `204`; the new password logs in and the old one gets `401`; of the caller's 2 sessions the other gets `401` and the current still gets `200` on `/me`; one `user.password_changed` audit row exists whose JSON contains no `password` key (USR-05, AC 42)
Proof: `go -C app test ./internal/features/users/change_password -run '^TestChangePassword_RevokesOtherSessions$'`

**C51** - A wrong `current_password` returns `422` with `errors[].location` `body.current_password`, leaves the hash and both sessions intact; an 11-character `new_password` returns `422` at `body.new_password` (USR-05, AC 43)
Proof: `go -C app test ./internal/features/users/change_password -run '^TestChangePassword_Validation422$'`

### S6 - Esquema, contrato e regras do repositório · ~8 files · ~40 KB · ~10k

**C52** - The applied schema enforces: `INSERT` of email `A@x.com` fails the lowercase check; a second `a@x.com` fails uniqueness; a new `users` row gets a UUID `id` by default; deleting a user cascades its `sessions`; a duplicate `(user, role)` and a duplicate `(role, permission)` fail; `login_attempts.kind` rejects `other` (door 1, door 2, door 4, door 10)
Proof: `go -C app test ./migrations -run '^TestSchema_UsersConstraints$'`

**C53** - After `api migrate up`, `roles` has a row `admin` and `role_permissions` grants it exactly `*` (door 7)
Proof: `go -C app test ./migrations -run '^TestSchema_AdminRoleSeeded$'`

**C54** - `app/go.mod` requires `golang.org/x/crypto` without `// indirect`; `web/package.json` lists `@radix-ui/react-dialog` and `@radix-ui/react-label` (door 12)
Proof: `go -C app test ./archtest -run '^TestDependencies_UsersFeature$'`

**C55** - archtest reports no violation on the real module with `features/users/password` imported by 4 packages of the same feature; `internal/platform/auth` and `internal/platform/audit` import nothing under `features` (door 11, AD-001)
Proof: `go -C app test ./archtest -run '^TestImports_RepositoryIsClean$'`

**C56** - A slice generated by `newslice` answers `401` without a session and `501` problem+json to a session whose role grants its permission, and its generated test passes (Impact - generator template)
Proof: `go -C app test ./cmd/newslice -run '^TestGeneratedSlice_RequiresSession$'`

**C57** - `AGENTS.md` contains the strings `Authenticated` and `api users create-admin` (Observable - AGENTS.md)
Proof: `go -C app test ./archtest -run '^TestAgentsDoc_MentionsUsersContract$'`

**C58** - A failed login writes one log line at level `WARN` carrying `request_id` and `ip`, and no attribute of that line contains the submitted email (Assumption - login failure observability)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_FailureLogsWarnWithoutEmail$'`

**C59** - Recording a failed login deletes every `login_attempts` row older than 15 minutes: 3 rows dated 20 minutes ago are gone after one new failure (door 10)
Proof: `go -C app test ./internal/features/users/login -run '^TestLogin_PrunesOldAttempts$'`

**C60** - The committed `app/openapi.json` contains the 10 user operations with the statuses listed in the plan's `Surface`, and `web/src/api/schema.d.ts` is regenerated from it (Surface, foundation C28-C29)
Proof: `go -C app test ./internal/app -run '^TestOpenAPI_UsersStatuses$'`
Proof: `task gen:openapi:check`
Proof: `npm --prefix web run gen:check`

### S7 - Web: entrar, sair, rotas protegidas · ~10 files · ~35 KB · ~9k

**C61** - With `GET /api/v1/users/me` answering `401`, opening `/users` lands on `/login?redirect=%2Fusers` (USR-06, AC 44, door 16)
Proof: `npm --prefix web run test -- src/routes/authed.test.tsx -t "redirects to login without session"`

**C62** - Submitting valid credentials navigates to the `redirect` path when present and to `/` when absent (USR-06, AC 45)
Proof: `npm --prefix web run test -- src/features/users/LoginPage.test.tsx -t "navigates after login"`

**C63** - A `401` login shows `E-mail ou senha inválidos.` and the email field keeps its value (USR-06, AC 46)
Proof: `npm --prefix web run test -- src/features/users/LoginPage.test.tsx -t "shows invalid credentials"`

**C64** - A `429` login with `Retry-After: 61` shows `Muitas tentativas. Tente novamente em 2 minutos.` and with `Retry-After: 900` shows `... em 15 minutos.` (USR-06, AC 47)
Proof: `npm --prefix web run test -- src/features/users/LoginPage.test.tsx -t "shows rate limit minutes"`

**C65** - While the login request is pending, the submit button is disabled and reads `Entrando…` (USR-06, AC 48)
Proof: `npm --prefix web run test -- src/features/users/LoginPage.test.tsx -t "disables submit while pending"`

**C66** - Clicking `Sair` sends `DELETE /api/v1/users/session` and lands on `/login` (USR-06, AC 49)
Proof: `npm --prefix web run test -- src/features/users/UserMenu.test.tsx -t "signs out"`

**C67** - When a signed-in screen at `/users/abc` receives `401` from any API call, the app lands on `/login?redirect=%2Fusers%2Fabc` (USR-06, AC 50)
Proof: `npm --prefix web run test -- src/routes/authed.test.tsx -t "redirects on 401 while signed in"`

**C68** - In the built binary, logging in through `/login` with the admin created by `create-admin` reaches `/users`, and `Sair` returns to `/login`; `document.cookie` does not expose `session` (USR-06, AC 44, AC 45, AC 49, door 4)
Proof: `npm --prefix web run e2e -- e2e/users.spec.ts -g "login and logout"`

### S8 - Web: telas de usuários e da própria senha · ~16 files · ~70 KB · ~18k

**C69** - `/users` for a user with `users:read` shows a table with headers `E-mail`, `Nome`, `Status` and rows labelled `Ativo` and `Desativado` (USR-07, AC 51)
Proof: `npm --prefix web run test -- src/features/users/UsersList.test.tsx -t "shows the users table"`

**C70** - An empty `items` shows `Nenhum usuário encontrado.` (USR-07, AC 52)
Proof: `npm --prefix web run test -- src/features/users/UsersList.test.tsx -t "shows empty state"`

**C71** - While `GET /api/v1/users` is pending an element with `role="status"` is visible, on `/users` and on `/users/$id` (USR-07, AC 53)
Proof: `npm --prefix web run test -- src/features/users/UsersList.test.tsx -t "shows loading"`
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "shows loading"`

**C72** - A `500` shows `Não foi possível carregar os usuários.` and a `Tentar novamente` button that issues the request again (USR-07, AC 54)
Proof: `npm --prefix web run test -- src/features/users/UsersList.test.tsx -t "shows error and retries"`

**C73** - A user whose `permissions` lack `users:read` sees `Você não tem permissão para acessar esta página.` on `/users` and no `Usuários` link in the menu; an API `403` on `/users/$id` shows the same text (USR-07, AC 55)
Proof: `npm --prefix web run test -- src/features/users/UsersList.test.tsx -t "shows forbidden"`
Proof: `npm --prefix web run test -- src/features/users/UserMenu.test.tsx -t "hides links without permission"`
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "shows forbidden on 403"`

**C74** - With `total` `120`: page 1 has `Anterior` disabled and `Próxima` enabled; `Próxima` requests `offset=50`; page 3 has `Próxima` disabled; with `total` `50` neither button is shown (USR-07, AC 56)
Proof: `npm --prefix web run test -- src/features/users/UsersList.test.tsx -t "paginates by 50"`

**C75** - Submitting `/users/new` with valid data posts the body and lands on `/users/<id from the 201>` (USR-07, AC 57)
Proof: `npm --prefix web run test -- src/features/users/UserForm.test.tsx -t "creates and navigates"`

**C76** - A `409` on create and on edit shows `Este e-mail já está em uso.` under the email field (USR-07, AC 58)
Proof: `npm --prefix web run test -- src/features/users/UserForm.test.tsx -t "shows email conflict"`

**C77** - A `422` with `errors` at `body.name` and `body.password` shows each message under its field (USR-07, AC 59)
Proof: `npm --prefix web run test -- src/features/users/UserForm.test.tsx -t "maps 422 errors to fields"`

**C78** - Changing only the name on `/users/$id` sends `PATCH` with body exactly `{"name": ...}` and shows `Alterações salvas.` (USR-07, AC 60)
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "patches only changed fields"`

**C79** - `Desativar` opens a dialog reading `Desativar b@x.com? As sessões deste usuário serão encerradas.`; `Cancelar` closes it with no request; the dialog's `Desativar` sends `POST .../deactivate` (USR-07, AC 61)
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "confirms before deactivating"`

**C80** - `/users/$id` for the signed-in user's own id shows no `Desativar` button (USR-07, AC 62)
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "hides deactivate for self"`

**C81** - `Ativar` on a deactivated user sends `POST .../activate` and the status then reads `Ativo` (USR-07, AC 63)
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "activates"`

**C82** - A `404` on `/users/$id` shows `Usuário não encontrado.` (Observable - screen `/users/$id`)
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "shows not found"`

**C83** - `/account/password` with matching new passwords sends `PUT /api/v1/users/me/password` and shows `Senha alterada.`; a `422` at `body.current_password` shows its message under that field (USR-07, AC 64, AC 43)
Proof: `npm --prefix web run test -- src/features/users/ChangePassword.test.tsx -t "changes password"`

**C84** - Different values in the two new-password fields show `As senhas não conferem.` and send no request (USR-07, AC 65)
Proof: `npm --prefix web run test -- src/features/users/ChangePassword.test.tsx -t "rejects mismatch"`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| `POST /api/v1/users/session` statuses (4) | 204 C8 · 401 C12 · 422 C14 · 429 C15 | - |
| `DELETE /api/v1/users/session` statuses (2) | 204 C19 · 401 C19 | - |
| `GET /api/v1/users/me` statuses (2) | 200 C20 · 401 C21 | - |
| `PUT /api/v1/users/me/password` statuses (3) | 204 C50 · 401 C48 · 422 C51 | - |
| `POST /api/v1/users` statuses (5) | 201 C34 · 401 C48 · 403 C48 · 409 C35 · 422 C37 | - |
| `GET /api/v1/users` statuses (4) | 200 C38 · 401 C48 · 403 C48 · 422 C39 | - |
| `GET /api/v1/users/{id}` statuses (5) | 200 C40 · 401 C48 · 403 C48 · 404 C41 · 422 C41 | - |
| `PATCH /api/v1/users/{id}` statuses (6) | 200 C42 · 401 C48 · 403 C48 · 404 C41 · 409 C43 · 422 C43 | - |
| `POST /api/v1/users/{id}/deactivate` statuses (6) | 204 C44 · 401 C48 · 403 C48 · 404 C41 · 409 C45 · 422 C41 | - |
| `POST /api/v1/users/{id}/activate` statuses (5) | 204 C47 · 401 C48 · 403 C48 · 404 C41 · 422 C41 | - |
| `create-admin` exits (3) | 0 C1 · 1 C2 · 2 C3 | - |
| `create-admin` rejected inputs (4) | C3, table-driven over all 4 | - |
| auth middleware decision (9) | public passes C27 · authenticated passes C27 · no cookie C21 · unknown token C21 · expired C21 · deactivated C21 · under TTL passes C21 · missing permission C25 · wildcard and exact allow C26 | - |
| `op.Spec` access markers (8 combinations) | C24, table-driven over all 8 | - |
| access marker per user operation (10) | C49, table-driven over all 10 | - |
| login outcomes (7) | success C8 · rotate C11 · unknown C12, C13 · wrong password C12 · deactivated C12 · invalid body C14 · rate limited C15, C16 | - |
| login invalid bodies (4) | C14, table-driven over all 4 | - |
| rate limit keys (2) | email C15 · ip C16 | - |
| rate limit window edges (2) | 14 min counted C17 · 16 min ignored C17 | - |
| session cookie attributes (6) | value format C8 · `Path` C8 · `Max-Age` C8, C23 · `HttpOnly` C8 · `SameSite=Lax` C8 · `Secure` both ways C8, C23 | - |
| audit actions (7) | `user.created` C1, C34 · `user.updated` C42 · `user.deactivated` C44 · `user.activated` C47 · `user.password_changed` C50 · `session.created` C8 · `session.deleted` C19 | - |
| audit fields (8) | `actor_id` C28 · `action` C28 · `resource_type` C28 · `resource_id` C28 · `before` C28 · `after` C28 · `ip` C28 · `request_id` C28 | - |
| redacted keys (4) | C30, table-driven over all 4 | - |
| append-only statements (2) | `UPDATE` C31 · `DELETE` C31 | - |
| client IP cases (6) | trusted + untrusted rightmost C32 · trusted skips trusted hops C32 · all trusted C32 · unparsable C32 · untrusted peer C33 · empty list C33 | - |
| password write paths (3) | CLI C7 · create C7 · change C7 | - |
| create validation edges (9) | C37, table-driven over all 9 (5 rejected, 4 accepted) | - |
| list page bounds (5) | `limit=0` C39 · `limit=101` C39 · `offset=-1` C39 · `limit=1` C39 · `limit=100` C39 | - |
| `{id}` routes for 404/422 (4) | get C41 · patch C41 · deactivate C41 · activate C41 | - |
| user state transitions (4) | active -> deactivated C44 · deactivated -> active C47 · deactivated -> deactivated C46 · active -> active C46 | - |
| session revocation triggers (4) | logout C19 · login rotation C11 · deactivation C44 · password change C50 | - |
| config fields (3) | `SESSION_TTL` C22, C23 · `COOKIE_SECURE` C22, C23 · `TRUSTED_PROXIES` C22 | - |
| startup assembly (2 places) | `cmd/api serve` C23 · test harness via `app.New` C48 | - |
| screen `/login` states (4) | success C62 · 401 C63 · 429 C64 · pending C65 | - |
| screen `/users` states (6) | table C69 · empty C70 · loading C71 · error C72 · forbidden C73 · pagination C74 | - |
| screen `/users/$id` states (8) | loading C71 · forbidden C73 · not found C82 · patch C78 · conflict C76 · deactivate confirm C79 · self hides C80 · activate C81 | - |
| screen `/users/new` states (3) | success C75 · 409 C76 · 422 C77 | - |
| screen `/account/password` states (3) | success C83 · 422 C83 · mismatch C84 | - |
| unauthorised redirects (2) | no session C61 · 401 while signed in C67 | - |
| auth-security rules (8) | 1 C9, C10 · 2 C11 · 3 C12, C13 · 4 C4, C5, C6 · 5 C8, C23 · 6 C44 · 7 C15, C16 · 8 C32, C33 | - |
| Landing doors (16) | 1 C52 · 2 C35, C36, C52 · 3 C4, C5 · 4 C8, C9, C10, C52 · 5 C24 · 6 C21, C25 · 7 C26, C53 · 8 C28, C31 · 9 C32, C33 · 10 C17, C52, C59 · 11 C55 · 12 C54 · 13 C1, C3 · 14 C49, C60 · 15 C22, C23 · 16 C61 | - |

- Claims naming a status code, route or response shape: C8, C11, C12, C14-C16, C19-C21, C34-C51, C60 - each proof issues a real HTTP request against the slice's registered handler or the assembled server
- C48 and C68 prove the assembled path a second time; they do not stand in for C21, C25-C27 (middleware decision at its own layer) or C61-C67 (each state at component level)
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

- `internal/platform/auth` middleware: decides over public / authenticated / permission, cookie present, session found, expired, deactivated, permission or `*` held - 7 branch points -> decides, reached across a boundary (C21, C25-C27 own layer; C48 assembled)
- `internal/platform/op`: decides over 3 access markers, 8 combinations -> decides, reached across a boundary (C24 own layer; C49 assembled)
- `internal/platform/httpx` client IP: decides over trusted peer, XFF walk, parse failure - 4 branch points -> decides (C32, C33 own layer; C16 through login)
- `internal/platform/audit`: redaction walks 4 keys at any depth -> decides (C30); `Record` itself is one insert -> instrumentation, proven through C28 and every slice's audit assertion
- `features/users/login`: decides over rate limit (2 keys), user found, active, password match, rehash needed, existing cookie - 7 branch points -> decides, reached across a boundary; its HTTP test is its own layer (C6, C8-C18, C58, C59)
- `features/users/password`: decides over parameter parse and rehash -> decides (C4, C5)
- `features/users/{create,update,deactivate,activate,change_password}`: each a validation plus a guard (`409`, self, no-op) -> decides at the HTTP boundary of the slice (C34-C47, C50-C51)
- `features/users/{list_users,get_user,me,logout}`: list bounds and a lookup -> entry point; accepted, rejected and error inputs proven (C19, C20, C38-C41)
- closest analogue in the repo: `internal/platform/op` (`TestRegister_*`), a guard table proven at its own layer and again through `app.New` (foundation C19-C22)

Cost: 19 own-layer proofs across 6 platform/shared packages beyond the slice HTTP tests. Without them the middleware's 9-row decision table would be proven only by the 2 rows C48 happens to cross.

## Swept

- validation: C3, C14, C37, C39, C41, C43, C51, C84
- failure modes: C29, C2, C12
- idempotency: C46, C2, C35
- authorization: C21, C24, C25, C26, C27, C48, C49, C73
- concurrency: C36 - two creates race on the unique index; C44 - revocation is in the deactivation transaction, so a request after commit sees no session
- data lifecycle: C59 - old login attempts pruned; expired sessions are rejected (C21) but not deleted - rows are inert and are removed on deactivation (C44) or logout (C19); a purge job is out of scope for this feature
- dependency failure: existing - database errors surface as `500`/`503` problem+json through foundation C3, C13; no new external dependency
- state transitions: C44, C46, C47
- observability: C58, C28 (`request_id` and `ip` on every audit event)

## Handoff

Intended split, written before any code. Existing files the build reads (`op`, `httpx`, `app`, `config`, `cmd/api`,
`cmd/newslice`, `registry.go`, web routes/client/status, `Taskfile.yml`, `archtest`) = 70 KB = ~18k. Written per
slice: S1 ~11k · S2 ~18k · S3 ~14k · S4 ~24k · S5 ~4k · S6 ~10k · S7 ~9k · S8 ~18k = ~108k, plus the ~18k read
= ~126k, under the 150k budget -> one builder. If the running total passes 150k, hand off after S6, where the
surface changes from Go to web.
