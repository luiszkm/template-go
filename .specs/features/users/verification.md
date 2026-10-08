# Users verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 2c12067..fa85de1; fix under review afbc039..fa85de1
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier)

O veredito é FAIL por um único motivo: o gate `task check` saiu com código 201. O passo `test` falhou em
`TestLogin_UnknownEmailTimingMatches` (C13): `"275.8103ms" is not greater than or equal to "392.13095ms"`
(`unknown 275.8103ms, wrong 784.2619ms`, `login_test.go:224`). O mesmo teste passou em B1, no mesmo commit
fa85de1, com os pacotes do feature rodando sozinhos. A prova de C13 é, portanto, não determinística: compara a
mediana de 3 medições de relógio de parede entre duas séries, e sob a carga do gate (26 pacotes com `-p=4`, o
`cmd/newslice` rodando o gate aninhado, containers Postgres subindo em paralelo) a série de senha errada sofre um
pico e a razão cai abaixo de 1/2. O fix não tocou `login` nem esse teste; a falha é da prova, não do fix. C13 fica
como FAIL: uma prova que passa e falha no mesmo commit não prova a check.

As 89 checks têm prova rodada em fa85de1, com cada teste nomeado mostrado individualmente como passado, e uma
asserção localizada. Em B1 e B2 todas passam; no gate, a prova de tempo de C13 falhou (acima), então 88/89 ficam PASS. Os quatro membros sem prova da rodada 1 agora têm prova (C85-C89), as duas
linhas de Test policy não atendidas foram rejulgadas e estão atendidas, e os três precision gaps (C2, C50, C74)
estão fechados. Os 5 faults injetados nas superfícies que o fix criou foram mortos.

Escopo pelo diff `afbc039..fa85de1`, não pela mensagem do commit:

- **Testes novos ou alterados:** `auth_test.go` (C85), `LoginPage.test.tsx` (C88, C89), `UserDetail.test.tsx`
  (C86, C87), `change_password_test.go` (C50 passa a contar exatamente um evento; as linhas abaixo de `:69`
  deslocaram 2).
- **Helpers compartilhados (`testkit`):** `NewAPIWithoutDatabase` (novo, `http.go:45-51`) instala `auth.Install(api,
  nil, 12h)`, o mesmo valor nil que `app.New` passa quando `DB` é nil (`app/internal/app/app.go:64-69`).
  `pinDockerHost` (`db.go:58-65`, chamado em `db.go:69` e `postgres.go:23`) só define `DOCKER_HOST` no Windows e só
  quando está vazio. É ambiente, não comportamento: todos os testes de banco passaram com ele (B1) e no gate.
- **Config:** `Taskfile.yml` muda `GOFLAGS` para `-p=4 -timeout=30m` e `web/vite.config.ts` ganha
  `testTimeout: 30_000`. Os dois só alongam limites de tempo; nenhuma asserção, filtro ou inclusão de arquivo muda
  (`include` está intacto), então nenhum teste deixa de rodar. Efeito colateral aceito: um teste web travado agora
  leva 30 s para falhar em vez de 5 s.
- **Fonte de produção:** nenhuma. `git diff --stat afbc039..fa85de1` não toca `auth.go`, `LoginPage.tsx` nem
  `UserDetail.tsx`. O fix só acrescenta provas.

## Binding sources

carried from afbc039. O fix não tocou a interface (nenhum arquivo de produção mudou), então o passo 1 não é
refeito. O plano não marca nenhuma fonte como binding; as restrições citadas foram abertas na rodada 1.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| `.claude/skills/auth-security/SKILL.md` regras 1-8 | yes - arquivo local (rodada 1; relido nesta rodada) | none | - |
| `.specs/STATE.md` AD-005, AD-006, AD-007 | yes - arquivo local (rodada 1) | none | - |

## Checks

verified at fa85de1. Rodadas de prova, todas em fa85de1:

- **B1** - `GOFLAGS=-p=4 go -C app test ./archtest ./cmd/api ./cmd/newslice ./internal/app ./internal/features/users/{activate_user,bootstrap,change_password,create_user,deactivate_user,get_user,list_users,login,logout,me,password,update_user} ./internal/platform/{audit,auth,config,httpx,op} ./migrations -run '^(<74 nomes>)$' -v -count=1`:
  uma invocação, exit 0, 22 pacotes `ok`, 74/74 `--- PASS` individuais (os 74 nomes distintos dos `Proof:` Go,
  extraídos do `checks.md`; `comm` entre a lista e os `--- PASS` dá vazio). Inclui
  `--- PASS: TestMiddleware_NoDatabase503 (0.00s)`, `--- PASS: TestChangePassword_AuditsWithoutPassword (1.00s)`,
  `--- PASS: TestGeneratedSlice_RequiresSession (132.71s)`.
- **B2** - `npm --prefix web run test -- <7 arquivos> --reporter=verbose -t "<29 nomes>"`: uma invocação, exit 0,
  `Test Files 7 passed`, `Tests 31 passed | 1 skipped`. Os 31 nomes dos `Proof:` web aparecem com `✓`, inclusive
  `UserDetail > shows error and retries`, `UserDetail > maps 422 errors to fields on edit`,
  `LoginPage > shows a generic message for other failures`, `LoginPage > ignores redirects that leave the site`,
  `UsersList > hides pagination for 50 users`. `shows loading` e `shows error and retries` existem em dois arquivos
  e ambos rodaram. O skipped está fora do filtro.
- **B3** - `task gen:sqlc:check` exit 0, `task gen:openapi:check` exit 0, `npm --prefix web run gen:check` exit 0.
- **B4** - `task e2e -- -g "login and logout|api online"` (porta 8080 livre antes): exit 0, 2 passed:
  `e2e/users.spec.ts:4:1 › login and logout`, `e2e/status.spec.ts:5:1 › api online`.

Existência dos nomes novos por `rg -n`: `auth_test.go:123` `func TestMiddleware_NoDatabase503`;
`LoginPage.test.tsx:67` `it("shows a generic message for other failures"`, `:74` `it("ignores redirects that leave
the site"`; `UserDetail.test.tsx:101` `it("shows error and retries"`, `:113` `it("maps 422 errors to fields on
edit"`; `UsersList.test.tsx:83` `it("hides pagination for 50 users"`. Todos os arquivos de prova estão no diff
`2c12067..fa85de1`, exceto `app/archtest/imports_test.go` (C55), teste da foundation que roda sobre o módulo real.

Citações dos arquivos que o fix não tocou vêm da rodada 1 e foram conferidas: o diff `afbc039..fa85de1` só
acrescenta linhas no fim de `auth_test.go`, `LoginPage.test.tsx` e `UserDetail.test.tsx`, então as linhas
anteriores não se moveram. Em `change_password_test.go` as citações foram refeitas.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `create-admin` `  Ana@X.com `: exit 0, UUID no stdout, `ana@x.com` ativo, papel admin, evento `user.created` ator null | B1 `TestCreateAdmin_CreatesActiveAdmin` PASS | `app/cmd/api/users_test.go:27` - `require.Equal(t, 0, code, stderr)`; `:28-29` - `uuid.Parse(stdout)` NoError; `:35` - `require.Equal(t, "ana@x.com", email)`; `:36` - `require.True(t, active)`; `:42` - 1 vínculo `admin`; `:46-47` - 1 evento `user.created` `actor_id IS NULL AND resource_id = $1` | PASS |
| C2 | `ANA@x.com` com `ana@x.com` existente: exit 1, `already exists`, contagem igual | B1 `TestCreateAdmin_ExistingEmailExits1` PASS | `app/cmd/api/users_test.go:55` - segunda chamada com `"--email", "ANA@x.com"`; `:56` - `require.Equal(t, 1, code)`; `:57` - `require.Contains(t, stderr, "already exists")`; `:60` - `require.Equal(t, 1, users)`. O texto da check agora diz `ANA@x.com` (aprovado pelo usuário, `checks.md` Handoff); precision gap da rodada 1 fechado | PASS |
| C3 | exit 2 + `usage:` para 4 entradas, nenhuma linha; 12 e 128 aceitos | B1 `TestCreateAdmin_InvalidInputExits2`, `TestCreateAdmin_PasswordBounds` PASS | `users_test.go:69-72` - os 4 casos; `:77` - `require.Equal(t, 2, code)`; `:78` - `strings.HasPrefix(stderr, "usage:")`; `:83` - `require.Zero(t, users)`; `app/internal/features/users/bootstrap/bootstrap_test.go:14` - NoError para 12 e 128; `:17` - `ErrorIs(..., ErrInvalid)` para 11 e 129 | PASS |
| C4 | PHC fixado, salts distintos, `Verify` só na senha | B1 `TestHash_PinnedPHCFormat`, `TestVerify_MatchesOnlyThePassword` PASS | `app/internal/features/users/password/password_test.go:16` - regex fixada; `:23-24` - `require.Regexp(t, pinnedPHC, a/b)`; `:25` - `require.NotEqual(t, a, b)`; `:33` - `require.True(t, ok)`; `:37` - `require.False(t, ok, wrong)` | PASS |
| C5 | hash `m=19456,t=2,p=1` verifica, `NeedsRehash` true; atual false | B1 `TestVerify_ReadsParamsFromHash` PASS | `password_test.go:55` - `require.True(t, ok)`; `:56` - `require.True(t, password.NeedsRehash(old))`; `:60` - `require.False(t, password.NeedsRehash(current))` | PASS |
| C6 | login re-hash para `m=65536,t=3,p=4` e segue logando | B1 `TestLogin_RehashesOutdatedHash` PASS | `app/internal/features/users/login/login_test.go:93` - 204; `:96` - `strings.HasPrefix(stored, "$argon2id$v=19$m=65536,t=3,p=4$")`; `:97` - segundo login 204 | PASS |
| C7 | os 3 caminhos gravam o padrão de C4, nunca o texto | B1 `TestCreateAdmin_CreatesActiveAdmin`, `TestCreateUser_StoresArgon2idHash`, `TestChangePassword_StoresArgon2idHash` PASS | `users_test.go:37` - Regexp fixada; `app/internal/features/users/create_user/create_user_test.go:65` - Regexp; `:66` - `NotContains(t, hash, "doze-chars12")`; `app/internal/features/users/change_password/change_password_test.go:90` - `require.Regexp(t, ...m=65536,t=3,p=4..., hash)`; `:91` - `require.NotContains(t, hash, "senha-nova-1234")` | PASS |
| C8 | 204, um `Set-Cookie` 43 chars, `Path=/`, `Max-Age=43200`, `HttpOnly`, `SameSite=Lax`, `Secure` conforme config, `session.created` | B1 `TestLogin_SetsSessionCookie` PASS | `login_test.go:108` - 204; `:110` - `require.Len(t, headers, 1)`; `:112` - `^session=[A-Za-z0-9_-]{43}$`; `:113-114` - atributos; `:116` - `require.Equal(t, secure, slices.Contains(attrs, "Secure"))`; `:117-118` - 1 `session.created` com ator | PASS |
| C9 | `token_hash` = SHA-256 do cookie, nenhuma coluna contém o token | B1 `TestLogin_StoresOnlyTokenHash` PASS | `login_test.go:133` - `require.Equal(t, fmt.Sprintf("%x", sum), hashHex)`; `:135-136` - `NotEqual` / `NotContains(column, token)` | PASS |
| C10 | `NewToken` 43 chars, 32 bytes, 1000 distintos | B1 `TestNewToken_32RandomBytes` PASS | `app/internal/platform/auth/auth_test.go:50` - `require.Len(t, tok, 43)`; `:53` - `require.Len(t, raw, 32)`; `:54` - `require.False(t, seen[tok], "duplicate token")` | PASS |
| C11 | login com cookie: token novo, sessão antiga apagada e `401` em `/me` | B1 `TestLogin_RotatesExistingSession` PASS | `login_test.go:149` - `NotEqual(first.Value, second.Value)`; `:151` - 0 linhas com o hash antigo; `:152` - 401 | PASS |
| C12 | 3 falhas -> 401, `invalid email or password`, corpos iguais | B1 `TestLogin_FailuresAreIndistinguishable` PASS | `login_test.go:176` - 401; `:177` - `detail`; `:179-180` - `require.Equal(t, problemWithoutRequest(unknown), problemWithoutRequest(...))` | PASS |
| C13 | desconhecido verifica uma vez contra o fictício; mediana >= metade | B1 `TestLogin_UnknownEmailVerifiesDummyHash`, `TestLogin_UnknownEmailTimingMatches` PASS; gate `TestLogin_UnknownEmailTimingMatches` FAIL | `login_test.go:195` - `require.Equal(t, []string{password.DummyHash()}, hashes)` (parte determinística, PASS em B1 e no gate); `:224` - `require.GreaterOrEqual(t, unknown, wrong/2, ...)` PASS em B1, mas FAIL no gate em fa85de1: `"275.8103ms" is not greater than or equal to "392.13095ms"`. Prova de tempo não determinística (mediana de 3 amostras de relógio de parede, sensível à carga) | FAIL (flaky) |
| C14 | 422 para 4 corpos | B1 `TestLogin_InvalidBody422` PASS | `login_test.go:231-234` - os 4 corpos; `:239` - `require.Equal(t, http.StatusUnprocessableEntity, rec.Code)` | PASS |
| C15 | 5 falhas -> 6ª `429`, `Retry-After` 1..900, verificador não chamado; 4 falhas -> 204 | B1 `TestLogin_RateLimitPerEmail` PASS | `login_test.go:246-250` - 429 e `Retry-After` entre 1 e 900; `:260` - existente e inexistente; `:267` - `require.Equal(t, before, v.calls.Load())`; `:274` - 204 | PASS |
| C16 | 20 falhas por IP -> 429; outro IP 204; XFF variado segue limitado | B1 `TestLogin_RateLimitPerIP` PASS | `login_test.go:286` - `requireRetryAfter`; `:287` - 204 de outro IP; `:291` - XFF varia; `:294` - `requireRetryAfter` | PASS |
| C17 | 16 min não bloqueia; 14 min bloqueia | B1 `TestLogin_RateLimitWindowIs15Minutes` PASS | `login_test.go:314` - 204; `:315` - 429; `:318` - 204 (IP); `:320` - 429 (IP) | PASS |
| C18 | 4 falhas, sucesso, 4 falhas -> 204 | B1 `TestLogin_SuccessClearsEmailFailures` PASS | `login_test.go:334` - 204; `:336` - 204 após a segunda série | PASS |
| C19 | logout 204, cookie limpo, linha apagada, `session.deleted`, antigo 401; sem cookie 401 | B1 `TestLogout_DeletesSession`, `TestLogout_WithoutSession401` PASS | `app/internal/features/users/logout/logout_test.go:21` - 204; `:24-26` - `session=;`, `Path=/`, `Max-Age=0`; `:27` - 0 sessões; `:28-29` - evento; `:30-31` - 401; `:38` - 401 sem cookie | PASS |
| C20 | `/me` 200 com `permissions` `[]`, as duas, e todas para `*` sem `*` | B1 `TestMe_ReturnsPermissions` PASS | `app/internal/features/users/me/me_test.go:38-40` - os 3 casos; `:45` - 200; `:50` - `require.Equal(t, c.want, got.Permissions)`; `:51` - `NotContains(got.Permissions, "*")` | PASS |
| C21 | 401 problem+json sem handler nos 4 casos; TTL-1m passa | B1 `TestMiddleware_Rejects401`, `TestMe_401WithoutSession` PASS | `auth_test.go:64` - `ttl + 1s`; `:80` - 401; `:81` - problem+json; `:85` - `require.Zero(t, calls.Load())`; `:88-90` - TTL-1m 200; `me_test.go:61-62` | PASS |
| C22 | defaults 12h, true, vazio; 2 prefixos; `nope` cita `TRUSTED_PROXIES` | B1 `TestDefaults_UsersConfig`, `TestTrustedProxies_Parse` PASS | `app/internal/platform/config/config_test.go:29` - `12*time.Hour`; `:30` - `CookieSecure`; `:31` - `Empty(TrustedProxies)`; `:37` - 2 prefixos; `:40` - `ErrorContains(t, err, "TRUSTED_PROXIES")` | PASS |
| C23 | `api serve` `SESSION_TTL=1h`, `COOKIE_SECURE=false` -> `Max-Age=3600`, sem `Secure` | B1 `TestServe_AppliesSessionConfig` PASS | `users_test.go:97-98` - env; `:112` - 204; `:114` - `Contains(cookie, "Max-Age=3600")`; `:115` - `NotContains(cookie, "Secure")` | PASS |
| C24 | exatamente um marcador: 3 aceitos, 5 rejeitados citando o ID | B1 `TestRegister_ExactlyOneAccessMarker` PASS | `app/internal/platform/op/op_test.go:74-81` - 8 casos; `:89` - NoError; `:92` - Error; `:93` - `require.Contains(t, err.Error(), id)` | PASS |
| C25 | 403 problem+json sem handler: sem papel e `other:thing` | B1 `TestMiddleware_Forbids403` PASS | `auth_test.go:101` - `require.Equal(t, http.StatusForbidden, rec.Code, name)`; `:102` - problem+json; `:104` - `require.Zero(t, calls.Load())` | PASS |
| C26 | `*` e `x:y` exato executam | B1 `TestMiddleware_WildcardAndExactAllow` PASS | `auth_test.go:110-111` - 200 e 200; `:112` - `require.EqualValues(t, 2, calls.Load())` | PASS |
| C27 | `Authenticated` sem papel; `Public` sem cookie | B1 `TestMiddleware_AuthenticatedAndPublic` PASS | `auth_test.go:118` - 200; `:119` - 200; `:120` - 2 chamadas | PASS |
| C28 | `Record` commitado: 1 linha com os 8 campos | B1 `TestRecord_CommitsWithTx` PASS | `app/internal/platform/audit/audit_test.go:62` - count 1; `:63-70` - os 8 campos com valores literais | PASS |
| C29 | tx com erro: nenhum evento, nenhuma mutação | B1 `TestRecord_RolledBackWithTx` PASS | `audit_test.go:87` - `ErrorIs(err, boom)`; `:88-89` - `require.Zero` em eventos e na mutação | PASS |
| C30 | 4 chaves removidas em qualquer profundidade | B1 `TestRecord_RedactsPasswordKeys` PASS | `audit_test.go:99-104` - entradas; `:110-111` - `JSONEq` sem as 4 chaves | PASS |
| C31 | `UPDATE`/`DELETE` falham, linha intacta | B1 `TestAuditEvents_AppendOnly` PASS | `audit_test.go:121`, `:123` - `ErrorContains(err, "append-only")`; `:124` - 1 linha `action = 'a'` | PASS |
| C32 | par confiável: 4 casos de XFF | B1 `TestClientIP_TrustedPeerReadsForwardedFor` PASS | `app/internal/platform/httpx/clientip_test.go:32-35` - casos; `:39` - `require.Equal(t, netip.MustParseAddr(c.want), clientIP(...))` | PASS |
| C33 | par fora da lista ou lista vazia: XFF ignorado | B1 `TestClientIP_UntrustedPeerIgnoresForwardedFor` PASS | `clientip_test.go:47` - `198.51.100.4`; `:48` - `10.0.0.5` | PASS |
| C34 | `POST /users` 201, chaves exatas, `b@x.com`, ativo, `user.created` com ator | B1 `TestCreateUser_201` PASS | `create_user_test.go:44` - 201; `:52` - chaves exatas; `:53` - `"b@x.com"`; `:55` - `true`; `:56-57` - evento com ator | PASS |
| C35 | duplicado -> 409, contagem igual | B1 `TestCreateUser_DuplicateEmail409` PASS | `create_user_test.go:77` - 409; `:79` - contagem igual | PASS |
| C36 | 10 concorrentes -> 1x201, 9x409, 1 linha | B1 `TestCreateUser_ConcurrentSameEmail` PASS | `create_user_test.go:91` - `require.Equal(t, []int{201, 409 x9}, codes)`; `:92` - 1 linha | PASS |
| C37 | 422 com `location` para 5 corpos; 4 bordas aceitas | B1 `TestCreateUser_Validation` PASS | `create_user_test.go:101-105` - casos; `:110` - 422; `:119` - `Contains(locations, c.field)`; `:123-129` - 201 nas 4 bordas | PASS |
| C38 | 52 usuários: 50 ordenados, `total` 52; `offset=50` -> 2 | B1 `TestListUsers_PagesByEmail` PASS | `app/internal/features/users/list_users/list_users_test.go:42` - 52; `:43` - 50; `:44` - `IsSorted`; `:49` - 2 | PASS |
| C39 | 422 para 3 queries; 200 para 1 e 100 | B1 `TestListUsers_InvalidPage422` PASS | `list_users_test.go:59-61` - mapa; `:63` - `require.Equal(t, want, rec.Code, query)` | PASS |
| C40 | `GET /users/{id}` 200 com 6 campos | B1 `TestGetUser_200` PASS | `app/internal/features/users/get_user/get_user_test.go:26` - 200; `:29` - 6 chaves; `:34` - `Nil(deactivated_at)`; `:36` - `NotEmpty` | PASS |
| C41 | 404 e 422 nas 4 rotas | B1 os 4 `*_404And422` PASS | `get_user_test.go:45`, `:47`; `app/internal/features/users/update_user/update_user_test.go:115-116`; `app/internal/features/users/deactivate_user/deactivate_user_test.go:73-74`; `app/internal/features/users/activate_user/activate_user_test.go:56-57` | PASS |
| C42 | PATCH parcial; `user.updated` com diff só nos campos | B1 `TestUpdateUser_PartialFields` PASS | `update_user_test.go:71` - 200; `:75` - `ElementsMatch([name], changedKeys)`; `:80-81`; `:86-88`; `:89` - 3 eventos | PASS |
| C43 | 409 sem mudança; próprio 200; 3 corpos 422 | B1 `TestUpdateUser_ConflictAndValidation` PASS | `update_user_test.go:98` - 409; `:100` - inalterado; `:102` - 200; `:105-109` - 422 | PASS |
| C44 | desativar: 204, `deactivated_at`, 0 sessões, evento, 401 nos 2 cookies | B1 `TestDeactivateUser_RevokesSessions` PASS | `deactivate_user_test.go:37` - 204; `:38`; `:39` - `require.Equal(t, 0, ...)`; `:40-41` - evento; `:42-43` - 401 | PASS |
| C45 | desativar a si mesmo 409, nada muda, sessão 200 | B1 `TestDeactivateUser_SelfIs409` PASS | `deactivate_user_test.go:51` - 409; `:52` - 0 desativados; `:53` - 200 | PASS |
| C46 | no-op nos 2 sentidos: 204, igual, nenhum evento | B1 `TestDeactivateUser_AlreadyDeactivatedNoop`, `TestActivateUser_AlreadyActiveNoop` PASS | `deactivate_user_test.go:64-66`; `activate_user_test.go:47-49` | PASS |
| C47 | ativar: 204, limpa, `user.activated`, login volta | B1 `TestActivateUser_Reactivates` PASS | `activate_user_test.go:32` - 401 antes; `:34` - 204; `:35` - null; `:36-37` - evento; `:38` - 204 | PASS |
| C48 | servidor montado: 9 rotas 401; 6 rotas 403 | B1 `TestUsersRoutes_401WithoutSession`, `TestUsersRoutes_403WithoutPermission` PASS | `app/internal/app/users_test.go:27-40` - rotas; `:49` - `app.New(...)`; `:53-54` - 401 problem+json; `:65` - 403 | PASS |
| C49 | 10 operações com os marcadores certos | B1 `TestUsersOperations_AccessMarkers` PASS | `app/internal/app/users_test.go:75-86` - mapa esperado; `:104` - `require.Equal(t, want, got)` | PASS |
| C50 | troca de senha: 204, nova loga, antiga 401, outra sessão 401, atual 200, exatamente um `user.password_changed` sem `password` | B1 `TestChangePassword_RevokesOtherSessions`, `TestChangePassword_AuditsWithoutPassword` PASS | `change_password_test.go:54` - 204; `:55` - outra 401; `:56` - atual 200; `:57` - antiga 401; `:58` - nova 204; `:70-71` - `require.Equal(t, 1, userstest.Count(t, pool, "SELECT count(*) FROM audit_events WHERE action = 'user.password_changed' AND actor_id = $1", id))`; `:76` - `require.NotContains(t, after, "password")`. As duas provas estão nas linhas `Proof:` e o evento é contado; precision gap da rodada 1 fechado | PASS |
| C51 | `current_password` errada 422 em `body.current_password`, nada muda; `new_password` 11 -> 422 | B1 `TestChangePassword_Validation422` PASS | `change_password_test.go:112-113` - 422 e `[]string{"body.current_password"}`; `:114-116` - sessões 200, senha antiga 204; `:119-120` - 422 em `body.new_password` | PASS |
| C52 | constraints do schema | B1 `TestSchema_UsersConstraints` PASS | `app/migrations/schema_test.go:22`, `:24`, `:28`, `:34`, `:36`, `:37`, `:42` | PASS |
| C53 | `admin` semeado com exatamente `*` | B1 `TestSchema_AdminRoleSeeded` PASS | `schema_test.go:57` - `require.Equal(t, []string{"*"}, perms)` | PASS |
| C54 | `x/crypto` direto; radix dialog e label | B1 `TestDependencies_UsersFeature` PASS | `app/archtest/users_test.go:20-21`; `:31-32` | PASS |
| C55 | archtest limpo; `password` dentro da feature; `auth`/`audit` sem `features` | B1 `TestImports_RepositoryIsClean` PASS | `app/archtest/imports_test.go:41` - `require.Empty(t, v)` sobre o módulo real. Observação carregada da rodada 1: `password` tem 5 importadores não-teste contra os "4" do texto; o fix não tocou isso | PASS |
| C56 | slice gerado: 401, 501 problem+json, teste gerado passa | B1 `TestGeneratedSlice_RequiresSession` PASS (132.71s) | `app/cmd/newslice/repo_test.go:171-173` - teste gerado com 401, `SignIn(... "demo:get_thing")`, 501; `:176-177` - `--- PASS: TestEndpoint_NotImplemented` | PASS |
| C57 | `AGENTS.md` cita `Authenticated` e `api users create-admin` | B1 `TestAgentsDoc_MentionsUsersContract` PASS | `app/archtest/users_test.go:37-38` | PASS |
| C58 | login falho: 1 WARN com `request_id` e `ip`, sem o e-mail | B1 `TestLogin_FailureLogsWarnWithoutEmail` PASS | `login_test.go:356` - `Len(warns, 1)`; `:357-358`; `:359` - `NotContains(buf.String(), "secret.person")` | PASS |
| C59 | falha apaga tentativas com mais de 15 min | B1 `TestLogin_PrunesOldAttempts` PASS | `login_test.go:365`; `:368` - 0 linhas antigas | PASS |
| C60 | `openapi.json` com os status do `Surface`; `schema.d.ts` regenerado | B1 `TestOpenAPI_UsersStatuses` PASS; B3 `gen:openapi:check`, `gen:check` exit 0 | `app/internal/app/users_test.go:117-128` - 10 rotas e status; `:134` - `require.Contains(t, operation.Responses, status, key)` | PASS |
| C61 | `/users` sem sessão -> `/login?redirect=%2Fusers` | B2 `redirects to login without session` ✓ | `web/src/routes/authed.test.tsx:11` - `toBe("/login?redirect=%2Fusers")` | PASS |
| C62 | navega para `redirect` ou `/` | B2 `navigates after login` ✓ | `web/src/features/users/LoginPage.test.tsx:23` - `toBe("/users")`; `:32` - `toBe("/")` | PASS |
| C63 | 401 -> `E-mail ou senha inválidos.`, e-mail mantido | B2 `shows invalid credentials` ✓ | `LoginPage.test.tsx:41-42` | PASS |
| C64 | `Retry-After` 61 -> 2 min; 900 -> 15 | B2 `shows rate limit minutes` ✓ | `LoginPage.test.tsx:54`, `:56` | PASS |
| C65 | pendente: botão desabilitado `Entrando…` | B2 `disables submit while pending` ✓ | `LoginPage.test.tsx:63-64` - `toBeDisabled()` | PASS |
| C66 | `Sair` -> `DELETE` e `/login` | B2 `signs out` ✓ | `web/src/features/users/UserMenu.test.tsx:15-16` | PASS |
| C67 | 401 com sessão em `/users/abc` -> `/login?redirect=%2Fusers%2Fabc` | B2 `redirects on 401 while signed in` ✓ | `authed.test.tsx:22` | PASS |
| C68 | binário: login, `/users`, `Sair`, sem `session` em `document.cookie` | B4 `login and logout` | `web/e2e/users.spec.ts:8`, `:11`, `:12`, `:13` - `document.cookie` not toContain `session=`; `:16` - `/login$` | PASS |
| C69 | tabela com `E-mail`, `Nome`, `Status`, `Ativo`/`Desativado` | B2 `shows the users table` ✓ | `web/src/features/users/UsersList.test.tsx:25-27` | PASS |
| C70 | vazio -> `Nenhum usuário encontrado.` | B2 `shows empty state` ✓ | `UsersList.test.tsx:36` | PASS |
| C71 | `role="status"` pendente em `/users` e `/users/$id` | B2 `shows loading` x2 ✓ | `UsersList.test.tsx:42`; `web/src/features/users/UserDetail.test.tsx:21` - `findByRole("status")` | PASS |
| C72 | 500 -> mensagem + `Tentar novamente` refaz | B2 `UsersList > shows error and retries` ✓ | `UsersList.test.tsx:51-54` - texto, clique, `toHaveLength(2)` | PASS |
| C73 | sem `users:read` -> texto e menu sem `Usuários`; 403 no detalhe -> mesmo texto | B2 `shows forbidden`, `hides links without permission`, `shows forbidden on 403` ✓ | `UsersList.test.tsx:60-61`; `UserMenu.test.tsx:23`; `UserDetail.test.tsx:27` | PASS |
| C74 | `total` 120: botões e `offset=50`; `total` 50 sem botões | B2 `paginates by 50` ✓, `hides pagination for 50 users` ✓ | `UsersList.test.tsx:72-73`, `:76`, `:79`; `:90-91` - `queryByRole("button", { name: "Anterior" / "Próxima" })` `.not.toBeInTheDocument()` com `total: 50` (`:86`). As duas provas estão nas linhas `Proof:`; precision gap da rodada 1 fechado | PASS |
| C75 | `/users/new` posta e vai para `/users/<id>` | B2 `creates and navigates` ✓ | `web/src/features/users/UserForm.test.tsx:29-30` | PASS |
| C76 | 409 em criação e edição -> `Este e-mail já está em uso.` | B2 `shows email conflict` ✓ | `UserForm.test.tsx:46-47` (criação), `:56` (edição) | PASS |
| C77 | 422 em `body.name`/`body.password` sob cada campo | B2 `maps 422 errors to fields` ✓ | `UserForm.test.tsx:73` - `id` `name-error`; `:74` - `password-error` | PASS |
| C78 | só nome -> `PATCH {"name": ...}`, `Alterações salvas.` | B2 `patches only changed fields` ✓ | `UserDetail.test.tsx:48-49` | PASS |
| C79 | diálogo; `Cancelar` sem request; `Desativar` posta | B2 `confirms before deactivating` ✓ | `UserDetail.test.tsx:61`, `:65`, `:74` | PASS |
| C80 | próprio id: sem `Desativar` | B2 `hides deactivate for self` ✓ | `UserDetail.test.tsx:85` | PASS |
| C81 | `Ativar` -> `activate`, `Ativo` | B2 `activates` ✓ | `UserDetail.test.tsx:97-98` | PASS |
| C82 | 404 -> `Usuário não encontrado.` | B2 `shows not found` ✓ | `UserDetail.test.tsx:33` | PASS |
| C83 | senhas iguais -> `PUT`, `Senha alterada.`; 422 sob o campo | B2 `changes password` ✓ | `web/src/features/users/ChangePassword.test.tsx:26-27`, `:33` | PASS |
| C84 | senhas diferentes -> `As senhas não conferem.`, sem request | B2 `rejects mismatch` ✓ | `ChangePassword.test.tsx:40-41` | PASS |
| C85 | sem banco: `503` problem+json sem handler em `Authenticated` e `Permission`; `Public` 200 | B1 `TestMiddleware_NoDatabase503` PASS | `app/internal/platform/auth/auth_test.go:124` - `testkit.NewAPIWithoutDatabase(t)` (`auth.Install(api, nil, ...)`, `app/internal/platform/testkit/http.go:49`); `:131` - `/authn` e `/perm`; `:134` - `require.Equal(t, http.StatusServiceUnavailable, rec.Code, path)`; `:135` - `require.Equal(t, "application/problem+json", ...)`; `:137` - `require.Zero(t, calls.Load())`; `:138` - `require.Equal(t, http.StatusOK, get(t, h, "/public", nil))` | PASS |
| C86 | 500 em `/users/$id` -> `Não foi possível carregar os usuários.` e `Tentar novamente` refaz o `GET` | B2 `UserDetail > shows error and retries` ✓ | `web/src/features/users/UserDetail.test.tsx:104` - `[json(500), json(200, bia)]`; `:107` - `findByText("Não foi possível carregar os usuários.")`; `:108` - clique em `Tentar novamente`; `:109` - heading `b@x.com`; `:110` - `toHaveLength(2)` GETs | PASS |
| C87 | 422 no `PATCH` com erro em `body.name` -> mensagem sob o nome | B2 `maps 422 errors to fields on edit` ✓ | `UserDetail.test.tsx:117-120` - 422 com `location: "body.name"`; `:127` - `expect(await screen.findByText("expected length <= 100")).toHaveAttribute("id", "name-error")` | PASS |
| C88 | status fora de 204/401/429 -> `Não foi possível entrar. Tente novamente.` | B2 `shows a generic message for other failures` ✓ | `web/src/features/users/LoginPage.test.tsx:68` - `json(500, ...)`; `:71` - `findByText("Não foi possível entrar. Tente novamente.")` | PASS |
| C89 | `redirect` `//evil.example`, `https://evil.example`, `users` -> `/` | B2 `ignores redirects that leave the site` ✓ | `LoginPage.test.tsx:80` - os 3 valores codificados; `:83` - `expect(rendered.router.state.location.href).toBe("/")` | PASS |

## Coverage

Linhas cuja autoridade o fix tocou: verified at fa85de1. As demais: carried from afbc039. O fix não mudou nenhum
arquivo de produção, então nenhum conjunto ganhou membro novo. Ele só fecha membros sem prova da rodada 1.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| status das 10 rotas (47) - carried from afbc039 | plan `Surface` | session POST 204 C8 · 401 C12 · 422 C14 · 429 C15; DELETE 204/401 C19; `/me` 200 C20 · 401 C21; PUT password 204 C50 · 401 C48 · 422 C51; POST users 201 C34 · 409 C35 · 422 C37; GET users 200 C38 · 422 C39; `{id}` 200/204 C40, C42, C44, C47 · 404/422 C41 · 409 C43, C45; 401 e 403 C48; contrato C60 | - |
| middleware de auth (9 linhas) - verified at fa85de1 | `app/internal/platform/auth/auth.go:126-156` (`rg -n "huma.WriteErr"`: `:134` 503, `:139` 401, `:144` 401, `:148` 500, `:152` 403) | `Public` passa (`:129`) C27, C85 · `Querier` nil -> 503 (`:133-135`) C85 · cookie ausente 401 (`:138`) C21 · `ErrNoSession` 401 (`:143`) C21 · abaixo do TTL passa C21 · sem permissão 403 (`:151`) C25 · `*` / exata C26 · `Authenticated` sem papel C27 · erro de lookup 500 (`:147`) = repasse de erro, Swept `dependency failure` | - |
| `op.Spec` marcadores (8) + erros documentados (2) - carried from afbc039 | `op.go` | C24; C60 | - |
| login (`endpoint.go:77-145`) - carried from afbc039 | `login/endpoint.go` | C15 · C16 · C17 · C12 · C13 · C6 · C11 · C18 · C58 · C59 · C8 | - |
| saídas do `create-admin` (3), entradas rejeitadas (4) - verified at fa85de1 (texto de C2 mudou) | `app/cmd/api/main.go:137-176` | 0 C1 · 1 C2 (`ANA@x.com`, texto agora igual à prova) · 2 C3 | - |
| IP do cliente (6) - carried from afbc039 | `clientip.go:25-50` | C32, C33, `clientip_test.go:42` | - |
| redação da auditoria (4 chaves, mapa e lista) - carried from afbc039 | `audit.go:22`, `:66-81` | C30 | - |
| ações de auditoria (7) - verified at fa85de1 (C50 mudou) | `AuditAction` das 10 specs e `bootstrap.go` | `user.created` C1, C34 · `user.updated` C42 · `user.deactivated` C44 · `user.activated` C47 · `user.password_changed` C50 (agora contado: exatamente 1, `change_password_test.go:70-71`) · `session.created` C8 · `session.deleted` C19 | - |
| transições (4), gatilhos de revogação (4) - carried from afbc039 | slices | C44, C47, C46 x2 · C19, C11, C44, C50 | - |
| config (3), montagem (2) - verified at fa85de1 (testkit tocado) | `config.go`; `app.go:64-69`; `testkit/http.go:40`, `:45-51` | C22, C23 · `cmd/api serve` C23 · `app.New` C48. O helper novo `NewAPIWithoutDatabase` reproduz o caminho `DB == nil` de `app.New` (`sessions` nil passado a `auth.Install`, `app.go:64-69`); C85 é prova da própria camada e não substitui uma montagem | - |
| tela `/login` (AC 45-48 + mapeamento) - verified at fa85de1 | plan `Observable` + `LoginPage.tsx:21-31` | sucesso com/sem redirect C62 · 401 C63 · 429 C64 · pendente C65 · outro status (`:31`) C88 · `safeRedirect` (`:21-23`): `//host` C89, absoluto `https://` C89, sem `/` C89. `:26` (erro que não é `SignInRefused`) devolve o mesmo texto de `:31`; removê-lo leva ao mesmo retorno de `:31`, então não é linha distinta da tabela | - |
| tela `/users` (6) - carried from afbc039 | plan `Observable` + AC 51-56 | C69 · C70 · C71 · C72 · C73 · C74 (incl. `total` 50, agora em `Proof:`) | - |
| tela `/users/$id` (Observable: loading, error, not found + AC 55, 58-63) - verified at fa85de1 | plan `Observable` + AC 59 + `UserDetail.tsx:69-101` | loading C71 · erro e retry (`:94-101`) C86 · 403 C73 · 404 C82 · patch C78 · 409 (`:69`) C76 · 422 na edição (`:70`) C87 · confirmar C79 · self C80 · ativar C81 | - |
| telas `/users/new` (3), `/account/password` (3), redirects (2) - carried from afbc039 | AC 57-59, 64-65, 44, 50 | C75, C76, C77 · C83 x2, C84 · C61, C67 | - |
| regras auth-security (8) - carried from afbc039 | `.claude/skills/auth-security/SKILL.md` | 1 C9, C10 · 2 C11 · 3 C12, C13 · 4 C4-C6 · 5 C8, C23 · 6 C44 · 7 C15, C16 · 8 C32, C33 | - |
| doors do Landing (18) - verified at fa85de1 (tabela do `checks.md` corrigida para 18) | plan `Landing` | 1 C52 · 2 C35, C36, C52 · 3 C4, C5 · 4 C8-C10, C52 · 5 C24 · 6 C21, C25 · 7 C26, C53 · 8 C28, C31 · 9 C32, C33 · 10 C17, C52, C59 · 11 C55 · 12 C54 · 13 C1, C3 · 14 C49, C60 · 15 C22, C23 · 16 C61 · 17 C60 · 18 C48 | - |

Sweep refeito em fa85de1 (fonte de produção idêntica a afbc039):

- `rg -n "huma.WriteErr" app/internal/platform/auth/auth.go`: 5 escritas. A de `:134` agora tem C85; a de `:148` é
  o caminho de erro genérico.
- `rg -n "status === |isError|instanceof SignInRefused|safeRedirect" web/src/features/users/*.tsx web/src/routes/_authed.tsx`:
  `UserDetail.tsx:69`, `:70`, `:83`, `:86`, `:94`; `UsersList.tsx:28`, `:38`; `UserForm.tsx:27-28`;
  `ChangePassword.tsx:23`; `_authed.tsx:10`; `LoginPage.tsx:21`, `:26-28`, `:49`. Todos têm caso afirmado.
- O `Coverage` do `checks.md` agora lista os membros que a rodada 1 encontrou (middleware 10, `/login` 6,
  `/users/$id` 10, doors 18) e bate com o recálculo acima.

## Test policy rows

verified at fa85de1 para as duas linhas não atendidas na rodada 1 e para as linhas que classificam arquivos tocados
(`testkit/*`). As demais: carried from afbc039.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `internal/platform/auth/auth.go` (middleware) | own layer C21, C25, C26, C27, C85 · boundary C48 | yes - as 9 linhas da tabela de decisão têm caso afirmado, inclusive `Querier` nil -> 503 (C85, `auth_test.go:134`); boundary C48 (verified at fa85de1) |
| Decides, reached across a boundary | `internal/platform/op/op.go`; `features/users/login/endpoint.go` | own layer C24 · boundary C49, C60; login C6, C8-C18, C58, C59 · montado C23, C68 | yes (carried from afbc039) |
| Decides, not reached across a boundary | `httpx/clientip.go`, `audit/audit.go` (redação), `features/users/password/password.go`, `config.Prefixes` | C32, C33 · C30 · C4, C5 · C22 | yes (carried from afbc039) |
| Decides (slice HTTP é a borda e a própria camada) | `features/users/{create_user,update_user,deactivate_user,activate_user,change_password}`, `bootstrap` | C34-C37, C41-C47, C50, C51 · C1-C3 | yes (C50 agora conta o evento; verified at fa85de1) |
| Decides, not reached across a boundary (web) | `web/src/features/users/LoginPage.tsx`, `UserDetail.tsx`, `UsersList.tsx`, `UserForm.tsx`, `ChangePassword.tsx`, `routes/_authed.tsx`; agora classificados no `checks.md` | um caso por linha do mapeamento | yes - outro status C88, guarda de redirect C89, 422 na edição C87, erro do detalhe C86; as demais linhas C61-C84 (verified at fa85de1) |
| Entry point that decides nothing | `features/users/{list_users,get_user,me,logout}` | boundary C38-C41, C20, C21, C19 | yes (carried from afbc039) |
| Instrumentation, pass-throughs | `audit.Record` (insert), `testkit/*` (incl. `NewAPIWithoutDatabase`, `pinDockerHost`), `userstest`, `deps`, `email.Normalize`, `web/src/api/client.ts`, `session.ts` | none of its own | yes - `NewAPIWithoutDatabase` é coberto por C85 e `pinDockerHost` por todo teste de banco em B1 e no gate (verified at fa85de1) |

## Faults injected

verified at fa85de1. Um fault por superfície que o fix criou.

Isolamento: `git worktree add %TEMP%\users-verify-r2 HEAD` (fa85de1). Para o web, foi criada a junction
`web\node_modules` com `mklink /J`, removida com `rmdir` antes de `git worktree remove --force` e `git worktree
prune`. Cada fault foi aplicado sozinho e revertido com `git checkout -- .` no scratch antes do próximo; a porcelain
do scratch ficou vazia entre faults. A porcelain da árvore real antes e depois é idêntica (`diff` vazio): só os dois
diretórios `.claude/skills/auth-security/` e `.cursor/skills/auth-security/` não rastreados que já existiam.

| Mutation | Location | Killed |
| --- | --- | --- |
| ramo `Querier` nil abre em vez de responder 503 (`huma.WriteErr(... 503 ...)` -> `next(ctx)`) | `app/internal/platform/auth/auth.go:134` | yes - `TestMiddleware_NoDatabase503` FAIL: expected 503, actual 200 (C85) |
| `safeRedirect` sem a guarda `!redirect.startsWith("//")` | `web/src/features/users/LoginPage.tsx:22` | yes - `LoginPage > ignores redirects that leave the site` ×: expected `/evil.example` to be `/` (C89) |
| mensagem de "outro status" troca para `E-mail ou senha inválidos.` | `web/src/features/users/LoginPage.tsx:31` | yes - `LoginPage > shows a generic message for other failures` ×: unable to find `Não foi possível entrar. Tente novamente.` (C88) |
| mapeamento de `422` na edição removido | `web/src/features/users/UserDetail.tsx:70` | yes - `UserDetail > maps 422 errors to fields on edit` ×: unable to find `expected length <= 100` (C87) |
| `Tentar novamente` do detalhe deixa de chamar `user.refetch()` | `web/src/features/users/UserDetail.tsx:98` | yes - `UserDetail > shows error and retries` ×: unable to find heading `b@x.com` (C86) |

Os 5 faults da rodada 1 (curinga `*`, limite por e-mail, revogação na desativação, redação em listas, `Math.ceil`)
ficam carried from afbc039: o fix não tocou nenhuma dessas superfícies nem os testes que as mataram.

## Decisions and constraints

carried from afbc039, conferido no diff do fix. O fix não acrescenta comentários em código: `testkit/db.go`,
`testkit/http.go` e os testes novos não têm nenhum. O comentário de `testkit/postgres.go:20` já existia. A
observação da rodada 1 sobre AD-011 em `op.go:1-3`, `op.go:28` e `cmd/api/main.go:6` continua fora do veredito,
porque nenhuma check a afirma.

## Gate

verified at fa85de1. Rodado uma vez, sozinho, da raiz do repositório.

`task check` - **exit 201**, 708 s. FAIL.

- `fmt:check`, `lint`, `gen:check` (sqlc, openapi `TestOpenAPI_ServedMatchesCommitted` ok, web `gen:check`) e
  `archtest` (ok 93.0 s) passaram.
- `test` (`go test -count=1 ./...` com `GOFLAGS=-p=4 -timeout=30m`): 25 pacotes `ok`, inclusive `cmd/newslice`
  (561.1 s, gate aninhado) e todos os demais pacotes do feature; **1 pacote FAIL**:
  `internal/features/users/login` - `--- FAIL: TestLogin_UnknownEmailTimingMatches (3.73s)`,
  `login_test.go:224: "275.8103ms" is not greater than or equal to "392.13095ms"`, `unknown 275.8103ms, wrong
  784.2619ms`.
- `web:typecheck`, `web:lint` e `web:test` **não rodaram no gate**, porque `check` para no primeiro passo que falha.
  Rodados à parte depois do gate, só como informação (não substituem o gate): `task web:typecheck` exit 0,
  `task web:lint` exit 0 (biome, 57 arquivos), `task web:test` exit 0, 13 arquivos, 43 passed.

Pela instrução desta rodada o gate roda uma única vez; ele não foi repetido para fazer a falha sumir.

`task e2e -- -g "login and logout|api online"` - exit 0, 2 passed (B4).
