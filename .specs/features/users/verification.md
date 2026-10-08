# Users verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 2c12067..afbc039
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier)

As 84 checks têm prova rodada em afbc039 e uma asserção localizada; todas passam. Os 5 faults injetados foram
mortos. `task gen:*:check`, a web `gen:check` e o e2e (`login and logout`, `api online`) passam.

O veredito é FAIL por quatro membros sem prova, encontrados ao recalcular a Coverage a partir do código e do plano.
Nenhum aparece no `Coverage` do `checks.md`:

1. **Middleware de auth, linha "sem banco"** (`app/internal/platform/auth/auth.go:133-135`). Com `Querier` nil,
   toda operação não `Public` responde `503 database not configured`. Nenhum teste chega a esse ramo. A tabela de
   decisão do `checks.md` ("auth middleware decision (9)") não o lista. Por isso a linha de Test policy *Decides,
   reached across a boundary* não é atendida para `platform/auth`.
2. **Estado de erro de `/users/$id`** (`web/src/features/users/UserDetail.tsx:94-101`). O `Observable` do plano diz
   "screen `/users/$id` | loading, error, not found | AC 53, AC 54 aplicados à tela de detalhe". Loading (C71) e not
   found (C82) têm prova; o erro (`Não foi possível carregar os usuários.` + `Tentar novamente`) não tem nenhuma. O
   autor viu o membro no próprio plano e a tabela o escondeu.
3. **AC 59 na edição** (`UserDetail.tsx:70`). AC 59 diz "IF creation **or edition** returns `422`". C77 prova só a
   criação (`UserForm.test.tsx`); o mapeamento de `422` no `PATCH` de `/users/$id` não tem prova.
4. **Linhas de decisão da tela de login** (`web/src/features/users/LoginPage.tsx`):
   - `refusalMessage`, linha "outro erro" (`:26`, `:31`): `Não foi possível entrar. Tente novamente.` não tem prova.
   - Guarda de open-redirect `safeRedirect` (`:21-23`): os casos `//host` e caminho sem `/` caem em `/`, e nenhum tem
     prova. C62 prova só `/users` e "ausente".

   O `Test policy` do `checks.md` não classifica nenhum arquivo web, e esses dois são decisão pelo formato.

Também há três precision gaps (C2, C50, C74). Neles a asserção existe, mas o texto ou a lista de provas da check não
bate com o que foi provado. Fora do gate, há uma não conformidade com AD-011 em arquivos tocados pelo diff (seção
`Decisions and constraints`).

## Binding sources

O plano não marca nenhuma fonte como binding, então o passo 1 (`ui`) não se aplica. O plano cita como restrições
`.claude/skills/auth-security/SKILL.md` e AD-005/006/007. Abri os três e comparei com as checks; nenhuma check os
contradiz.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| `.claude/skills/auth-security/SKILL.md` regras 1-8 | yes - arquivo local | none | - |
| `.specs/STATE.md` AD-005, AD-006, AD-007 | yes - arquivo local | none | - |

Cada regra contra as checks:

- **Regra 1** (32 B de `crypto/rand`, só SHA-256 persistido): C9, C10.
- **Regra 2** (token novo a cada login): C11. Tokens distintos: C10.
- **Regra 3** (mesmo corpo e status, hash fictício): C12, C13.
- **Regra 4** (parâmetros fixados e codificados no hash): C4, C5, C6.
- **Regra 5** (`Secure` vindo da config): C8, C23.
- **Regra 6** (desativação apaga sessões na mesma transação): C44. A troca de papéis é da feature `rbac`, fora do
  escopo, e o plano registra isso em `Out of scope`.
- **Regra 7** (rate limit por conta e por IP): C15, C16.
- **Regra 8** (IP só via proxy confiável): C32, C33. O mesmo IP alimenta o limite por IP, e C16 prova que o XFF
  forjado atrás de um par não confiável não escapa do limite.

Contra as decisões:

- **AD-005** (cookie httpOnly, SameSite=Lax, revogável): C8, C19, C44.
- **AD-006** ("seeded role admin holds all permissions"): realizado pelo curinga `*` da door 7, que o plano decidiu
  explicitamente. C20 e C53 provam o efeito. Não é contradição.
- **AD-007** (mesma transação, campos do evento): C28, C29, C31.

## Checks

Rodadas de prova, todas em afbc039:

- **B1** - `GOFLAGS=-p=4 go -C app test ./cmd/api ./internal/features/users/... ./internal/platform/{auth,config,op,audit,httpx} ./internal/app ./migrations ./archtest -run '^(<71 nomes>)$' -v -count=1`:
  exit 0, 71/71 `--- PASS` individuais.
- **B2** - `GOFLAGS=-p=4 go -C app test ./cmd/newslice -run '^(TestGeneratedSlice_RequiresSession)$' -v -count=1`:
  exit 0, `--- PASS (46.24s)`.
- **B3** - `npm --prefix web run test -- <7 arquivos> --reporter=verbose -t "<25 nomes>"`: exit 0, 26 passed.
  `shows loading` existe em dois arquivos. Os 2 skipped estão fora do filtro.
- **B4** - testes fora das listas `Proof:` que carregam cláusulas de checks:
  - `go -C app test ./internal/features/users/change_password -run '^(TestChangePassword_AuditsWithoutPassword)$' -v`: PASS.
  - vitest `-t "hides pagination for 50 users|shows the users link with users:read"`: 2 passed.
- **B5** - `task gen:sqlc:check` exit 0, `task gen:openapi:check` exit 0, `npm --prefix web run gen:check` exit 0.
- **B6** - `task e2e -- -g "login and logout|api online"`: exit 0, 2 passed:
  `e2e\users.spec.ts:4:1 › login and logout`, `e2e\status.spec.ts:5:1 › api online`.

Existência: cada nome aparece individualmente na saída `-v` / `verbose` de B1-B4. Foi checado também com `rg -n` ou
lendo o arquivo de teste, de onde vêm as linhas citadas abaixo. Todos os arquivos de prova estão no diff
`2c12067..afbc039`, exceto `app/archtest/imports_test.go` (C55). Esse é um teste da foundation que roda sobre o
módulo real, que o diff mudou.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `create-admin` `  Ana@X.com `: exit 0, UUID no stdout, `ana@x.com` ativo, papel admin, evento `user.created` ator null | B1 `TestCreateAdmin_CreatesActiveAdmin` PASS | `app/cmd/api/users_test.go:27` - `require.Equal(t, 0, code, stderr)`; `:28-29` - `uuid.Parse(stdout)` NoError; `:35` - `require.Equal(t, "ana@x.com", email)`; `:36` - `require.True(t, active)`; `:42` - 1 vínculo `admin`; `:46-47` - 1 evento `user.created` `actor_id IS NULL AND resource_id = $1` (id impresso) | PASS |
| C2 | e-mail existente: exit 1, `already exists`, contagem igual | B1 `TestCreateAdmin_ExistingEmailExits1` PASS | `users_test.go:55` - segunda chamada com `ANA@x.com`; `:56` - `require.Equal(t, 1, code)`; `:57` - `require.Contains(t, stderr, "already exists")`; `:60` - `require.Equal(t, 1, users)`. **Precision gap**: o texto da check usa `A@x.com`, que normaliza para `a@x.com` e não é `ana@x.com`. Lida ao pé da letra, uma implementação correta sairia com 0. A prova usa a variante de caixa `ANA@x.com` e prova AC 2 com a normalização de AC 1. O texto da check precisa de correção. | PASS |
| C3 | exit 2 + `usage:` para 4 entradas, nenhuma linha; 12 e 128 aceitos | B1 `TestCreateAdmin_InvalidInputExits2` (4 subtests), `TestCreateAdmin_PasswordBounds` PASS | `users_test.go:69-72` - os 4 casos; `:77` - `require.Equal(t, 2, code)`; `:78` - `strings.HasPrefix(stderr, "usage:")`; `:83` - `require.Zero(t, users)`; `app/internal/features/users/bootstrap/bootstrap_test.go:14` - `require.NoError(... Validate(..., Repeat("p", n)))` para 12 e 128; `:17` - `ErrorIs(..., ErrInvalid)` para 11 e 129 | PASS |
| C4 | PHC fixado, salts distintos, `Verify` só na senha | B1 `TestHash_PinnedPHCFormat`, `TestVerify_MatchesOnlyThePassword` PASS | `app/internal/features/users/password/password_test.go:16` - regex `^\$argon2id\$v=19\$m=65536,t=3,p=4\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`; `:23-24` - `require.Regexp(t, pinnedPHC, a/b)`; `:25` - `require.NotEqual(t, a, b)`; `:33` - `require.True(t, ok)`; `:37` - `require.False(t, ok, wrong)` | PASS |
| C5 | hash `m=19456,t=2,p=1` verifica, `NeedsRehash` true; atual false | B1 `TestVerify_ReadsParamsFromHash` PASS | `password_test.go:55` - `require.True(t, ok)`; `:56` - `require.True(t, password.NeedsRehash(old))`; `:60` - `require.False(t, password.NeedsRehash(current))` | PASS |
| C6 | login re-hash para `m=65536,t=3,p=4` e segue logando | B1 `TestLogin_RehashesOutdatedHash` PASS | `app/internal/features/users/login/login_test.go:93` - 204; `:96` - `require.True(t, strings.HasPrefix(stored, "$argon2id$v=19$m=65536,t=3,p=4$"))`; `:97` - segundo login 204 | PASS |
| C7 | os 3 caminhos gravam o padrão de C4, nunca o texto | B1 `TestCreateAdmin_CreatesActiveAdmin`, `TestCreateUser_StoresArgon2idHash`, `TestChangePassword_StoresArgon2idHash` PASS | `users_test.go:37` - `require.Regexp(... pinned ...)`; `app/internal/features/users/create_user/create_user_test.go:65` - Regexp; `:66` - `require.NotContains(t, hash, "doze-chars12")`; `app/internal/features/users/change_password/change_password_test.go:88` - Regexp; `:89` - `require.NotContains(t, hash, "senha-nova-1234")` | PASS |
| C8 | 204, um `Set-Cookie` 43 chars, `Path=/`, `Max-Age=43200`, `HttpOnly`, `SameSite=Lax`, `Secure` conforme config, evento `session.created` | B1 `TestLogin_SetsSessionCookie` (`secure=true`, `secure=false`) PASS | `login_test.go:108` - 204; `:110` - `require.Len(t, headers, 1)`; `:112` - `require.Regexp(t, "^session=[A-Za-z0-9_-]{43}$", attrs[0])`; `:113-114` - `Path=/`, `Max-Age=43200`, `HttpOnly`, `SameSite=Lax`; `:116` - `require.Equal(t, secure, slices.Contains(attrs, "Secure"))`; `:117-118` - 1 `session.created` com `actor_id` = usuário | PASS |
| C9 | `token_hash` = SHA-256 do cookie, nenhuma coluna contém o token | B1 `TestLogin_StoresOnlyTokenHash` PASS | `login_test.go:133` - `require.Equal(t, fmt.Sprintf("%x", sum), hashHex)`; `:135-136` - `NotEqual` / `NotContains(column, token)` para as 3 colunas | PASS |
| C10 | `NewToken` 43 chars, 32 bytes, 1000 distintos | B1 `TestNewToken_32RandomBytes` PASS | `app/internal/platform/auth/auth_test.go:50` - `require.Len(t, tok, 43)`; `:53` - `require.Len(t, raw, 32)`; `:54` - `require.False(t, seen[tok], "duplicate token")` | PASS |
| C11 | login com cookie: token novo, sessão antiga apagada e `401` em `/me` | B1 `TestLogin_RotatesExistingSession` PASS | `login_test.go:149` - `require.NotEqual(t, first.Value, second.Value)`; `:151` - 0 linhas com o hash antigo; `:152` - `require.Equal(t, http.StatusUnauthorized, meStatus(t, h, first))` | PASS |
| C12 | 3 falhas -> 401, `invalid email or password`, corpos iguais sem `request_id`/`instance` | B1 `TestLogin_FailuresAreIndistinguishable` PASS | `login_test.go:176` - 401; `:177` - `require.Equal(t, "invalid email or password", ...["detail"])`; `:179-180` - `require.Equal(t, problemWithoutRequest(unknown), problemWithoutRequest(wrong/deactivated))` | PASS |
| C13 | e-mail desconhecido verifica uma vez contra o hash fictício; mediana >= metade | B1 `TestLogin_UnknownEmailVerifiesDummyHash`, `TestLogin_UnknownEmailTimingMatches` PASS | `login_test.go:195` - `require.Equal(t, []string{password.DummyHash()}, hashes)`; `:224` - `require.GreaterOrEqual(t, unknown, wrong/2)` | PASS |
| C14 | 422 para 4 corpos | B1 `TestLogin_InvalidBody422` (4 subtests) PASS | `login_test.go:231-234` - os 4 corpos; `:239` - `require.Equal(t, http.StatusUnprocessableEntity, rec.Code)` | PASS |
| C15 | 5 falhas por e-mail (existente ou não) -> 6ª `429`, `Retry-After` 1..900, verificador não chamado; 4 falhas -> 204 | B1 `TestLogin_RateLimitPerEmail` PASS | `login_test.go:260` - `x@y.com` e `ghost@y.com`; `:246-250` - 429, `Retry-After` inteiro `>= 1`, `<= 900`; `:267` - `require.Equal(t, before, v.calls.Load())`; `:274` - 204 após 4 falhas | PASS |
| C16 | 20 falhas por IP -> 21ª `429`; outro IP 204; XFF variado atrás de par não confiável segue limitado | B1 `TestLogin_RateLimitPerIP` PASS | `login_test.go:286` - `requireRetryAfter` no mesmo IP; `:287` - 204 de `203.0.113.2`; `:291` - `xff` varia por tentativa; `:294` - `requireRetryAfter` | PASS |
| C17 | 5 registros de 16 min não bloqueiam; de 14 min bloqueiam | B1 `TestLogin_RateLimitWindowIs15Minutes` PASS | `login_test.go:314` - 204 (16 min, e-mail); `:315` - 429 (14 min); `:318` - 204 (16 min, IP); `:320` - 429 (14 min, IP) | PASS |
| C18 | 4 falhas, sucesso, 4 falhas -> senha correta 204 | B1 `TestLogin_SuccessClearsEmailFailures` PASS | `login_test.go:334` - 204; `:336` - `require.Equal(t, http.StatusNoContent, ...)` após a segunda série | PASS |
| C19 | logout 204, cookie limpo, linha apagada, `session.deleted`, cookie antigo 401; sem cookie 401 | B1 `TestLogout_DeletesSession`, `TestLogout_WithoutSession401` PASS | `app/internal/features/users/logout/logout_test.go:21` - 204; `:23` - um `Set-Cookie`; `:24-26` - `session=;`, `Path=/`, `Max-Age=0`; `:27` - 0 sessões; `:28-29` - 1 `session.deleted`; `:30-31` - 401 em `/me`; `:38` - 401 sem cookie | PASS |
| C20 | `/me` 200 com `permissions` `[]`, as duas ordenadas, e todas as do `op` para `*`, sem `*` | B1 `TestMe_ReturnsPermissions` (3 subtests) PASS | `app/internal/features/users/me/me_test.go:38-40` - casos `[]`, `["users:create","users:read"]`, `["users:read","zeta:do"]`; `:45` - 200; `:47-49` - `id`, `email`, `name`; `:50` - `require.Equal(t, c.want, got.Permissions)`; `:51` - `NotContains(got.Permissions, "*")` | PASS |
| C21 | 401 problem+json sem rodar o handler: sem cookie, token desconhecido, TTL+1s, desativado; TTL-1m passa | B1 `TestMiddleware_Rejects401` (4 subtests), `TestMe_401WithoutSession` PASS | `auth_test.go:64` - `created_at = now() - (ttl + 1s)`; `:80` - 401; `:81` - `application/problem+json`; `:85` - `require.Zero(t, calls.Load())`; `:88-90` - TTL-1m -> 200; `me_test.go:61-62` - 401 problem+json | PASS |
| C22 | defaults 12h, true, vazio; 2 prefixos; `nope` falha citando `TRUSTED_PROXIES` | B1 `TestDefaults_UsersConfig`, `TestTrustedProxies_Parse` PASS | `app/internal/platform/config/config_test.go:29` - `12*time.Hour`; `:30` - `require.True(t, cfg.CookieSecure)`; `:31` - `require.Empty(t, cfg.TrustedProxies)`; `:37` - 2 prefixos; `:40` - `require.ErrorContains(t, err, "TRUSTED_PROXIES")` | PASS |
| C23 | `api serve` com `SESSION_TTL=1h`, `COOKIE_SECURE=false` -> `Max-Age=3600`, sem `Secure` | B1 `TestServe_AppliesSessionConfig` PASS | `users_test.go:97-98` - `run(ctx, []string{"serve"}, ... "SESSION_TTL": "1h", "COOKIE_SECURE": "false")`; `:112` - 204; `:114` - `require.Contains(t, cookie, "Max-Age=3600")`; `:115` - `require.NotContains(t, cookie, "Secure")` | PASS |
| C24 | exatamente um marcador: 3 aceitos, 5 rejeitados citando o ID | B1 `TestRegister_ExactlyOneAccessMarker` (8 subtests) PASS | `app/internal/platform/op/op_test.go:74-81` - os 8 casos; `:89` - `require.NoError(t, err)`; `:92` - `require.Error(t, err)`; `:93` - `require.Contains(t, err.Error(), id)` | PASS |
| C25 | 403 problem+json sem handler: sem papel e só `other:thing` | B1 `TestMiddleware_Forbids403` PASS | `auth_test.go:97-98` - os 2 usuários; `:101` - `require.Equal(t, http.StatusForbidden, rec.Code, name)`; `:102` - problem+json; `:104` - `require.Zero(t, calls.Load())` | PASS |
| C26 | `*` e `x:y` exato executam `x:y` | B1 `TestMiddleware_WildcardAndExactAllow` PASS | `auth_test.go:110` - 200 com `auth.Wildcard`; `:111` - 200 com `x:y`; `:112` - `require.EqualValues(t, 2, calls.Load())` | PASS |
| C27 | `Authenticated` executa sem papel; `Public` sem cookie | B1 `TestMiddleware_AuthenticatedAndPublic` PASS | `auth_test.go:118` - `/authn` 200 sem papel; `:119` - `/public` 200 sem cookie; `:120` - 2 chamadas | PASS |
| C28 | `Record` em tx commitada: 1 linha com os 8 campos | B1 `TestRecord_CommitsWithTx` PASS | `app/internal/platform/audit/audit_test.go:62` - count 1; `:63-70` - `actor`, `"thing.changed"`, `"thing"`, `"t-1"`, `before`, `after`, `ip` `203.0.113.7`, `"req-28"` | PASS |
| C29 | tx com erro: nenhum evento, nenhuma mutação | B1 `TestRecord_RolledBackWithTx` PASS | `audit_test.go:87` - `require.ErrorIs(t, err, boom)`; `:88` - `require.Zero(... audit_events)`; `:89` - `require.Zero(... roles 'rolled-back')` | PASS |
| C30 | as 4 chaves removidas em qualquer profundidade | B1 `TestRecord_RedactsPasswordKeys` PASS | `audit_test.go:99-104` - `password_hash` em struct, `password`/`current_password`/`new_password` no topo, aninhado em mapa e em lista; `:110` - `JSONEq({"email":"a@x.com"}, b)`; `:111` - `JSONEq({"email":...,"nested":{"keep":1},"list":[{"keep":2}]}, a)` | PASS |
| C31 | `UPDATE`/`DELETE` em `audit_events` falham, linha intacta | B1 `TestAuditEvents_AppendOnly` PASS | `audit_test.go:121` - `require.ErrorContains(t, err, "append-only")` (UPDATE); `:123` - idem (DELETE); `:124` - 1 linha `action = 'a'` | PASS |
| C32 | par confiável: 4 casos de XFF | B1 `TestClientIP_TrustedPeerReadsForwardedFor` (4 subtests) PASS | `app/internal/platform/httpx/clientip_test.go:32-35` - os 4 casos com o esperado; `:39` - `require.Equal(t, netip.MustParseAddr(c.want), clientIP(...))` | PASS |
| C33 | par fora da lista ou lista vazia: XFF ignorado | B1 `TestClientIP_UntrustedPeerIgnoresForwardedFor` PASS | `clientip_test.go:47` - `198.51.100.4`; `:48` - lista nil -> `10.0.0.5` | PASS |
| C34 | `POST /users` 201, chaves exatas, `b@x.com`, ativo, `user.created` com ator admin | B1 `TestCreateUser_201` PASS | `create_user_test.go:44` - 201; `:52` - `require.Equal(t, []string{"active","created_at","email","id","name"}, keys)`; `:53` - `"b@x.com"`; `:55` - `true`; `:56-57` - 1 `user.created` com `actor_id` = admin | PASS |
| C35 | ` b@x.com` após `B@x.com` -> 409, contagem igual | B1 `TestCreateUser_DuplicateEmail409` PASS | `create_user_test.go:77` - `require.Equal(t, http.StatusConflict, rec.Code)`; `:79` - contagem igual | PASS |
| C36 | 10 concorrentes -> 1x201, 9x409, 1 linha | B1 `TestCreateUser_ConcurrentSameEmail` PASS | `create_user_test.go:91` - `require.Equal(t, []int{201, 409 x9}, codes)`; `:92` - 1 linha | PASS |
| C37 | 422 com `location` para 5 corpos; 4 bordas aceitas | B1 `TestCreateUser_Validation` (5 subtests) PASS | `create_user_test.go:101-105` - casos e campos; `:110` - 422; `:119` - `require.Contains(t, locations, c.field)`; `:123-126` + `:129` - 1, 100, 12, 128 -> 201 | PASS |
| C38 | 52 usuários: 50 ordenados, `total` 52; `offset=50` -> 2 | B1 `TestListUsers_PagesByEmail` PASS | `app/internal/features/users/list_users/list_users_test.go:42` - `require.Equal(t, 52, first.Total)`; `:43` - `Len(..., 50)`; `:44` - `slices.IsSorted`; `:49` - `Len(last.Items, 2)`; `:51-52` - ordenado e 52 distintos | PASS |
| C39 | 422 para `limit=0`, `limit=101`, `offset=-1`; 200 para 1 e 100 | B1 `TestListUsers_InvalidPage422` PASS | `list_users_test.go:59-61` - mapa query -> status; `:63` - `require.Equal(t, want, rec.Code, query)` | PASS |
| C40 | `GET /users/{id}` 200 com 6 campos; `deactivated_at` null / timestamp | B1 `TestGetUser_200` PASS | `app/internal/features/users/get_user/get_user_test.go:26` - 200; `:29` - `require.Contains(t, body, key)` sobre os 6; `:34` - `require.Nil(t, body["deactivated_at"])`; `:36` - `require.NotEmpty(...)` | PASS |
| C41 | UUID aleatório 404, `abc` 422 nas 4 rotas | B1 `TestGetUser_404And422`, `TestUpdateUser_404And422`, `TestDeactivateUser_404And422`, `TestActivateUser_404And422` PASS | `get_user_test.go:45` - 404; `:47` - 422; `app/internal/features/users/update_user/update_user_test.go:115-116`; `app/internal/features/users/deactivate_user/deactivate_user_test.go:73-74`; `app/internal/features/users/activate_user/activate_user_test.go:56-57` | PASS |
| C42 | PATCH parcial: só `name`, só `email`, ambos; 200; `user.updated` com diff só nos campos | B1 `TestUpdateUser_PartialFields` PASS | `update_user_test.go:71` - 200; `:74` - `[ana@x.com, Ana Maria]`; `:75` - `ElementsMatch([name], changedKeys)`; `:80-81` - `email` sozinho; `:86-88` - ambos; `:89` - 3 `user.updated` com ator | PASS |
| C43 | e-mail de outro -> 409 sem mudança; o próprio -> 200; inválido, nome vazio, `{}` -> 422 | B1 `TestUpdateUser_ConflictAndValidation` PASS | `update_user_test.go:98` - 409; `:100` - e-mail inalterado; `:102` - 200 para o próprio; `:105-109` - 3 corpos -> 422 | PASS |
| C44 | desativar com 2 sessões: 204, `deactivated_at`, 0 sessões, `user.deactivated`, ambos os cookies 401 | B1 `TestDeactivateUser_RevokesSessions` PASS | `deactivate_user_test.go:37` - 204; `:38` - `deactivated_at IS NOT NULL`; `:39` - `require.Equal(t, 0, ... sessions WHERE user_id)`; `:40-41` - evento; `:42-43` - 401 nos dois cookies | PASS |
| C45 | desativar a si mesmo 409, nada muda, sessão segue 200 | B1 `TestDeactivateUser_SelfIs409` PASS | `deactivate_user_test.go:51` - 409; `:52` - 0 linhas desativadas; `:53` - `/me` 200 | PASS |
| C46 | no-op em ambos os sentidos: 204, `deactivated_at` igual, nenhum evento | B1 `TestDeactivateUser_AlreadyDeactivatedNoop`, `TestActivateUser_AlreadyActiveNoop` PASS | `deactivate_user_test.go:64` - 204; `:65` - `deactivated_at = '2026-01-01T00:00:00Z'`; `:66` - 0 eventos; `activate_user_test.go:47-49` - 204, ainda null, 0 eventos | PASS |
| C47 | ativar: 204, limpa, `user.activated`, login volta a funcionar | B1 `TestActivateUser_Reactivates` PASS | `activate_user_test.go:32` - login 401 antes; `:34` - 204; `:35` - `deactivated_at IS NULL`; `:36-37` - evento; `:38` - login 204 | PASS |
| C48 | servidor montado: 9 rotas -> 401 sem cookie; 6 -> 403 sem papel | B1 `TestUsersRoutes_401WithoutSession`, `TestUsersRoutes_403WithoutPermission` PASS | `app/internal/app/users_test.go:27-40` - 3 + 6 rotas; `:49` - `app.New(...)`; `:53-54` - 401 problem+json; `:65` - 403 | PASS |
| C49 | 10 operações com os marcadores certos | B1 `TestUsersOperations_AccessMarkers` PASS | `app/internal/app/users_test.go:75-86` - mapa esperado com as 10; `:104` - `require.Equal(t, want, got)` (igualdade exata) | PASS |
| C50 | troca de senha: 204, nova loga, antiga 401, outra sessão 401, atual 200, um `user.password_changed` sem `password` | B1 `TestChangePassword_RevokesOtherSessions` PASS; B4 `TestChangePassword_AuditsWithoutPassword` PASS | `change_password_test.go:54` - 204; `:55` - outra 401; `:56` - atual 200; `:57` - senha antiga 401; `:58` - nova 204; `:72-74` - `SELECT ... WHERE action = 'user.password_changed'` + `require.NotContains(t, after, "password")`. **Precision gap**: a cláusula de auditoria está num teste que não aparece em `Proof:`, e "um" evento não é contado. `QueryRow().Scan` só prova que existe ao menos um. | PASS |
| C51 | `current_password` errada -> 422 em `body.current_password`, nada muda; `new_password` de 11 -> 422 em `body.new_password` | B1 `TestChangePassword_Validation422` PASS | `change_password_test.go:110-111` - 422 e `[]string{"body.current_password"}`; `:112-114` - as duas sessões 200, senha antiga ainda loga; `:117-118` - 422 em `body.new_password` | PASS |
| C52 | constraints do schema | B1 `TestSchema_UsersConstraints` PASS | `app/migrations/schema_test.go:22` - `A@x.com` falha; `:24` - `a@x.com` duplicado falha; `:28` - `id` não nil; `:34` - `(user, role)` duplicado; `:36` - `(role, permission)` duplicado; `:37` - `kind 'other'`; `:42` - cascade: 0 sessões | PASS |
| C53 | `admin` semeado com exatamente `*` | B1 `TestSchema_AdminRoleSeeded` PASS | `schema_test.go:57` - `require.Equal(t, []string{"*"}, perms)` | PASS |
| C54 | `golang.org/x/crypto` direto; radix dialog e label | B1 `TestDependencies_UsersFeature` PASS | `app/archtest/users_test.go:20` - `require.NotEmpty(t, crypto)`; `:21` - `require.NotContains(t, crypto, "// indirect")`; `:31-32` - `@radix-ui/react-dialog`, `@radix-ui/react-label` | PASS |
| C55 | archtest limpo; `password` importado dentro da feature; `auth`/`audit` sem `features` | B1 `TestImports_RepositoryIsClean` PASS | `app/archtest/imports_test.go:41` - `require.Empty(t, v)` sobre o módulo real. Por `rg`, `features/users/password` tem 5 importadores não-teste: `login`, `create_user`, `change_password`, `bootstrap` e o helper `userstest`, contra os "4" do texto. `rg -n features` em `auth.go`/`audit.go` não dá nenhum hit. | PASS |
| C56 | slice gerado: 401 sem sessão, 501 problem+json com a permissão, teste gerado passa | B2 `TestGeneratedSlice_RequiresSession` PASS | `app/cmd/newslice/repo_test.go:171-173` - teste gerado contém `http.StatusUnauthorized`, `testkit.SignIn(t, pool, "demo:get_thing")`, `http.StatusNotImplemented`; `:176-177` - `go test` do slice gerado: `--- PASS: TestEndpoint_NotImplemented`. O template (`app/cmd/newslice/main.go:265`, `:269-270`) afirma 401, 501 e `application/problem+json` | PASS |
| C57 | `AGENTS.md` cita `Authenticated` e `api users create-admin` | B1 `TestAgentsDoc_MentionsUsersContract` PASS | `app/archtest/users_test.go:37` - `require.Contains(t, agents, "Authenticated")`; `:38` - `"api users create-admin"` | PASS |
| C58 | login falho: 1 linha WARN com `request_id` e `ip`, sem o e-mail | B1 `TestLogin_FailureLogsWarnWithoutEmail` PASS | `login_test.go:356` - `require.Len(t, warns, 1)`; `:357` - `request_id` = header; `:358` - `"192.0.2.77"`; `:359` - `require.NotContains(t, buf.String(), "secret.person")` | PASS |
| C59 | falha registrada apaga tentativas com mais de 15 min | B1 `TestLogin_PrunesOldAttempts` PASS | `login_test.go:365` - 3 linhas de 20 min; `:368` - `require.Equal(t, 0, ... at <= now() - interval '15 minutes')` | PASS |
| C60 | `openapi.json` com os status do `Surface`; `schema.d.ts` regenerado | B1 `TestOpenAPI_UsersStatuses` PASS; B5 `gen:openapi:check`, `gen:check` exit 0 | `app/internal/app/users_test.go:117-128` - as 10 rotas com os status do `Surface`; `:134` - `require.Contains(t, operation.Responses, status, key)` | PASS |
| C61 | `/users` sem sessão -> `/login?redirect=%2Fusers` | B3 `redirects to login without session` ✓ | `web/src/routes/authed.test.tsx:10` - pathname `/login`; `:11` - `expect(router.state.location.href).toBe("/login?redirect=%2Fusers")` | PASS |
| C62 | navega para `redirect` ou `/` | B3 `navigates after login` ✓ | `web/src/features/users/LoginPage.test.tsx:23` - `toBe("/users")`; `:32` - `toBe("/")` | PASS |
| C63 | 401 -> `E-mail ou senha inválidos.`, e-mail mantido | B3 `shows invalid credentials` ✓ | `LoginPage.test.tsx:41` - `findByText("E-mail ou senha inválidos.")`; `:42` - `toHaveValue("ana@x.com")` | PASS |
| C64 | `Retry-After: 61` -> 2 minutos; 900 -> 15 | B3 `shows rate limit minutes` ✓ | `LoginPage.test.tsx:54` - `"Muitas tentativas. Tente novamente em 2 minutos."`; `:56` - `"... em 15 minutos."` | PASS |
| C65 | pendente: botão desabilitado `Entrando…` | B3 `disables submit while pending` ✓ | `LoginPage.test.tsx:63` - `findByRole("button", { name: "Entrando…" })`; `:64` - `toBeDisabled()` | PASS |
| C66 | `Sair` -> `DELETE` e `/login` | B3 `signs out` ✓ | `web/src/features/users/UserMenu.test.tsx:15` - pathname `/login`; `:16` - `requests.some(DELETE /api/v1/users/session)` true | PASS |
| C67 | 401 com sessão em `/users/abc` -> `/login?redirect=%2Fusers%2Fabc` | B3 `redirects on 401 while signed in` ✓ | `authed.test.tsx:22` - `toBe("/login?redirect=%2Fusers%2Fabc")` | PASS |
| C68 | binário: login pelo `/login` com admin da CLI chega em `/users`, `Sair` volta, `document.cookie` sem `session` | B6 `login and logout` ✓ | `web/e2e/users.spec.ts:8` - `toHaveURL(/\/login\?redirect=%2Fusers$/)`; `:11` - `/users$`; `:12` - tabela contém o e-mail; `:13` - `document.cookie` not toContain `session=`; `:16` - `/login$`; admin via `create-admin` em `web/e2e/session.ts:33` | PASS |
| C69 | tabela com `E-mail`, `Nome`, `Status`, `Ativo`/`Desativado` | B3 `shows the users table` ✓ | `web/src/features/users/UsersList.test.tsx:25` - `toEqual(["E-mail", "Nome", "Status"])`; `:26-27` - `Ativo`, `Desativado` | PASS |
| C70 | vazio -> `Nenhum usuário encontrado.` | B3 `shows empty state` ✓ | `UsersList.test.tsx:36` | PASS |
| C71 | `role="status"` pendente em `/users` e `/users/$id` | B3 `shows loading` x2 ✓ | `UsersList.test.tsx:42` - `findByRole("status")`; `web/src/features/users/UserDetail.test.tsx:21` - `findByRole("status")` | PASS |
| C72 | 500 -> mensagem + `Tentar novamente` refaz | B3 `shows error and retries` ✓ | `UsersList.test.tsx:51` - mensagem; `:52` - clique; `:54` - `toHaveLength(2)` | PASS |
| C73 | sem `users:read` -> texto de permissão e menu sem `Usuários`; 403 no detalhe -> mesmo texto | B3 `shows forbidden`, `hides links without permission`, `shows forbidden on 403` ✓ | `UsersList.test.tsx:60` - texto; `:61` - nenhuma chamada a `/api/v1/users`; `UserMenu.test.tsx:23` - `queryByRole("link", { name: "Usuários" })` not in document; `UserDetail.test.tsx:27` - texto | PASS |
| C74 | `total` 120: anterior desabilitado, próxima `offset=50`, pág. 3 próxima desabilitada; `total` 50 sem botões | B3 `paginates by 50` ✓; B4 `hides pagination for 50 users` ✓ | `UsersList.test.tsx:72-73` - disabled / enabled; `:76` - `?limit=50&offset=50`; `:79` - próxima disabled; `:90-91` - sem `Anterior`/`Próxima` com `total: 50`. **Precision gap**: a cláusula "`total` 50" está num teste fora da lista `Proof:` | PASS |
| C75 | `/users/new` posta e vai para `/users/<id>` | B3 `creates and navigates` ✓ | `web/src/features/users/UserForm.test.tsx:29` - pathname `/users/${created.id}`; `:30` - corpo postado | PASS |
| C76 | 409 em criação e edição -> `Este e-mail já está em uso.` | B3 `shows email conflict` ✓ | `UserForm.test.tsx:46` - texto (criação); `:47` - `aria-describedby="email-error"`; `:56` - texto (edição) | PASS |
| C77 | 422 em `body.name`/`body.password` sob cada campo | B3 `maps 422 errors to fields` ✓ | `UserForm.test.tsx:73` - `toHaveAttribute("id", "name-error")`; `:74` - `"password-error"`. Prova só a criação; a edição de AC 59 está em `## Coverage` | PASS |
| C78 | só nome alterado -> `PATCH {"name": ...}`, `Alterações salvas.` | B3 `patches only changed fields` ✓ | `UserDetail.test.tsx:48` - `Alterações salvas.`; `:49` - `toEqual({ name: "Bia Souza" })` | PASS |
| C79 | diálogo com o texto; `Cancelar` sem request; `Desativar` posta | B3 `confirms before deactivating` ✓ | `UserDetail.test.tsx:61` - `toHaveTextContent("Desativar b@x.com? As sessões deste usuário serão encerradas.")`; `:65` - 0 POST após cancelar; `:74` - 1 POST em `/deactivate` | PASS |
| C80 | próprio id: sem `Desativar` | B3 `hides deactivate for self` ✓ | `UserDetail.test.tsx:85` - `queryByRole("button", { name: "Desativar" })` not in document | PASS |
| C81 | `Ativar` -> `activate` e status `Ativo` | B3 `activates` ✓ | `UserDetail.test.tsx:97` - `findByText("Ativo")`; `:98` - 1 POST em `/activate` | PASS |
| C82 | 404 -> `Usuário não encontrado.` | B3 `shows not found` ✓ | `UserDetail.test.tsx:33` | PASS |
| C83 | senhas iguais -> `PUT`, `Senha alterada.`; 422 em `body.current_password` sob o campo | B3 `changes password` ✓ | `web/src/features/users/ChangePassword.test.tsx:26` - `Senha alterada.`; `:27` - corpo do PUT; `:33` - `toHaveAttribute("id", "current_password-error")` | PASS |
| C84 | senhas diferentes -> `As senhas não conferem.`, sem request | B3 `rejects mismatch` ✓ | `ChangePassword.test.tsx:40` - texto; `:41` - 0 PUT | PASS |

## Coverage

Recalculado da autoridade de cada conjunto, não lido da tabela do autor:

- **Rotas e status:** do `Surface` do plano.
- **Decisões:** do código.
- **Telas:** do `Observable` e das ACs do plano.
- **Doors:** do `Landing`, que tem 18 doors.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| status das 10 rotas (47) | plan `Surface` | session POST 204 C8 · 401 C12 · 422 C14 · 429 C15; DELETE 204/401 C19; `/me` 200 C20 · 401 C21; PUT password 204 C50 · 401 C48 (`users_test.go:30`) · 422 C51; POST users 201 C34 · 409 C35 · 422 C37; GET users 200 C38 · 422 C39; GET/PATCH/deactivate/activate `{id}` 200/204 C40, C42, C44, C47 · 404/422 C41 · 409 C43, C45; 401 e 403 das 6 rotas com permissão C48. Documentadas no contrato: C60 | - |
| middleware de auth (8 linhas) | `app/internal/platform/auth/auth.go:127-156` | `Public` passa (`:129`) C27 · cookie ausente ou vazio 401 (`:138`) C21 · `ErrNoSession` 401 (`:143`: desconhecido, expirado, desativado) C21 · abaixo do TTL passa C21 · sem permissão nem `*` 403 (`:151`) C25 · `*` / exata passam (`:31-36`, `:155`) C26 · `Authenticated` passa sem papel C27 · erro de lookup 500 (`:147`) = repasse de erro, Swept `dependency failure` | `Querier` nil -> `503 database not configured` (`auth.go:133-135`): nenhum teste (`rg -n "database not configured" --glob '*_test.go'` sem hits; `TestUsersOperations_AccessMarkers` monta `app.New` sem DB mas não faz request) |
| `op.Spec` marcadores (8) + erros documentados (2) | `app/internal/platform/op/op.go` `validate`, `documentedErrors` | C24 tabela com as 8 combinações; 401 em não `Public`, 403 em `Permission`: C60 (`users_test.go:117-134`) | - |
| login (`endpoint.go:77-145`) | `app/internal/features/users/login/endpoint.go` | rate limit e-mail C15 · IP C16 · janela C17 · encontrado/senha/desativado C12 · hash fictício C13 · re-hash C6 · cookie existente C11 · limpa falhas C18 · log WARN C58 · prune C59 · sucesso C8 | - |
| saídas do `create-admin` (3) e entradas rejeitadas (4) | `app/cmd/api/main.go:137-176` | 0 C1 · 1 (já existe) C2 · 2 (flags, 11, 129) C3; o exit 1 por config/DB é repasse de erro | - |
| IP do cliente (6) | `app/internal/platform/httpx/clientip.go:25-50` | par não confiável C33 · lista vazia C33 · sem XFF -> par (`clientip_test.go:42`) · inválido -> par C32 · direita não confiável C32 · todos confiáveis -> mais à esquerda C32. O fallback de `peerAddr` sem porta (`:56-58`) não é alcançável por `net/http` sobre TCP | - |
| redação da auditoria (4 chaves, mapa e lista) | `app/internal/platform/audit/audit.go:22`, `:66-81` | C30 cobre as 4 chaves, mapa aninhado e lista | - |
| ações de auditoria (7) | `AuditAction` das 10 specs e `bootstrap.go` | `user.created` C1, C34 · `user.updated` C42 · `user.deactivated` C44 · `user.activated` C47 · `user.password_changed` C50 (B4, existência) · `session.created` C8 · `session.deleted` C19 | - |
| transições de estado (4), gatilhos de revogação (4) | slices `deactivate_user`, `activate_user`, `login`, `logout`, `change_password` | C44, C47, C46 x2 · C19, C11, C44, C50 | - |
| config (3), montagem (2) | `config.go`; `app/internal/app/app.go:69`, `testkit/http.go:40` | C22, C23 · `cmd/api serve` C23 · `app.New` C48 | - |
| tela `/login` (AC 45-48 + mapeamento) | plan `Observable` + `LoginPage.tsx:21-31` | sucesso com redirect / sem redirect C62 · 401 C63 · 429 C64 · pendente C65 | `refusalMessage` "outro erro" (`LoginPage.tsx:26`, `:31`) sem prova; `safeRedirect` rejeitando `//host` ou caminho sem `/` (`LoginPage.tsx:21-23`) sem prova |
| tela `/users` (6) | plan `Observable` + AC 51-56 | tabela C69 · vazio C70 · loading C71 · erro C72 · proibido C73 · paginação C74 (incl. `total` 50, B4) | - |
| tela `/users/$id` (Observable: loading, error, not found + AC 55, 58-63) | plan `Observable` "AC 53, AC 54 aplicados à tela de detalhe" + AC 59 | loading C71 · 403 C73 · 404 C82 · patch C78 · 409 C76 · confirmar C79 · self C80 · ativar C81 | estado de erro (`UserDetail.tsx:94-101`, AC 54 aplicado ao detalhe) sem prova; `422` na edição (`UserDetail.tsx:70`, AC 59 "or edition") sem prova |
| telas `/users/new` (3), `/account/password` (3), redirects (2) | AC 57-59, 64-65, 44, 50 | C75, C76, C77 · C83 x2, C84 · C61, C67 | - |
| regras auth-security (8) | `.claude/skills/auth-security/SKILL.md` | 1 C9, C10 · 2 C11 · 3 C12, C13 · 4 C4-C6 · 5 C8, C23 · 6 C44 · 7 C15, C16 · 8 C32, C33 | - |
| doors do Landing (18) | plan `Landing` | 1 C52 · 2 C35, C36, C52 · 3 C4, C5 · 4 C8-C10, C52 · 5 C24 · 6 C21, C25 · 7 C26, C53 · 8 C28, C31 · 9 C32, C33 · 10 C17, C52, C59 · 11 C55 · 12 C54 · 13 C1, C3 · 14 C49, C60 · 15 C22, C23 · 16 C61 · **17** C60 · **18** C48 (`app.New` instala o middleware: 401/403 no servidor montado). O `checks.md` diz "Landing doors (16)" e omite 17 e 18; ambas têm prova | - |

Sweep de conjuntos que nenhuma tabela trouxe:

- **Ramos do middleware.** `rg -n "huma.WriteErr" app/internal/platform/auth/auth.go` dá 5 escritas: `:134` 503,
  `:139` 401, `:144` 401, `:148` 500, `:152` 403. Só a de `:134` é uma decisão própria sem prova. A de `:148` é o
  caminho de erro genérico.
- **Ramos das telas.** `rg -n "status === |isError" web/src/features/users/*.tsx`:
  - `UserDetail.tsx:69-70` (409/422), `:83`, `:86`, `:94`.
  - `UsersList.tsx:28`, `:38`.
  - `UserForm.tsx:27-28`.
  - `ChangePassword.tsx:23`.

  `UserDetail.tsx:70` e `:94` são os únicos sem prova.
- **Swept "existing" relido.** "database errors surface as `500`/`503` problem+json" está de fato no código:
  `app/internal/platform/httpx/problem.go:19` sobrescreve `huma.NewError`, e o middleware usa `huma.WriteErr`, que
  passa por esse modelo. A restrição citada existe. Só C3/C13 não exercitam o caminho do middleware, e por isso o
  `503` de `auth.go:133` continua sem prova acima.

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `internal/platform/auth/auth.go` (middleware) | own layer C21, C25, C26, C27 · boundary C48 | no - 7 de 8 linhas da tabela de decisão têm caso afirmado; `Querier` nil -> 503 (`auth.go:133-135`) não tem |
| Decides, reached across a boundary | `internal/platform/op/op.go`; `features/users/login/endpoint.go` | own layer C24 · boundary C49, C60; login HTTP é a própria camada: C6, C8-C18, C58, C59 · servidor montado C23, C68 | yes |
| Decides, not reached across a boundary | `internal/platform/httpx/clientip.go`, `internal/platform/audit/audit.go` (redação), `features/users/password/password.go`, `config.Prefixes` | C32, C33 · C30 · C4, C5 · C22 | yes |
| Decides (slice HTTP é a borda e a própria camada) | `features/users/{create_user,update_user,deactivate_user,activate_user,change_password}`, `bootstrap` | C34-C37, C41-C47, C50, C51 · C1-C3 | yes |
| Decides, not reached across a boundary (web) | `web/src/features/users/LoginPage.tsx`, `UserDetail.tsx`, `UsersList.tsx`, `UserForm.tsx`, `ChangePassword.tsx`, `routes/_authed.tsx`; nenhuma linha do `checks.md` os classifica | um caso por linha do mapeamento | no - sem caso: `LoginPage.tsx:26`/`:31` (outro erro), `LoginPage.tsx:21-23` (guarda de redirect), `UserDetail.tsx:70` (422 na edição), `UserDetail.tsx:94-101` (erro). As demais linhas têm caso (C61-C84) |
| Entry point that decides nothing | `features/users/{list_users,get_user,me,logout}` | boundary C38-C41, C20, C21, C19 | yes - aceito, cada rejeitado e cada caminho de erro do `Surface` |
| Instrumentation, pass-throughs | `audit.Record` (insert), `testkit/*`, `userstest`, `deps`, `email.Normalize`, `web/src/api/client.ts`, `session.ts` | none of its own | yes - coberto por C28 e pelas asserções de auditoria de cada slice; pelos testes de tela |

## Faults injected

Isolamento: `git worktree add %TEMP%\users-verify-wt HEAD`. Para o web, foi criada uma junction `web\node_modules`;
ela foi removida com `rmdir` antes de `git worktree remove --force`. Cada fault foi aplicado sozinho e revertido com
`git checkout --` no scratch antes do próximo. A porcelain da árvore real antes e depois é idêntica (`diff` vazio):
só os dois diretórios `auth-security/` não rastreados que já existiam.

| Mutation | Location | Killed |
| --- | --- | --- |
| `Principal.Can` deixa de honrar o curinga `*` (removido o `if _, ok := p.Permissions[Wildcard]`) | `app/internal/platform/auth/auth.go:32-34` | yes - `TestMiddleware_WildcardAndExactAllow` FAIL: expected 200, actual 403 (C26) |
| limite por e-mail 5 -> 6 | `app/internal/features/users/login/endpoint.go:26` | yes - `TestLogin_RateLimitPerEmail` FAIL: expected 429, actual 204 (C15) |
| desativar sem `DeleteUserSessions` | `app/internal/features/users/deactivate_user/endpoint.go:60-62` | yes - `TestDeactivateUser_RevokesSessions` FAIL em `deactivate_user_test.go:39`: expected 0, actual 2 (C44) |
| redação deixa de descer em listas (`case []any` removido) | `app/internal/platform/audit/audit.go:76-79` | yes - `TestRecord_RedactsPasswordKeys` FAIL em `audit_test.go:111` (C30) |
| minutos de `Retry-After` com `Math.floor` em vez de `Math.ceil` | `web/src/features/users/LoginPage.tsx:29` | yes - `LoginPage > shows rate limit minutes` × (C64) |

## Decisions and constraints

Nenhuma check cobre AD-011 / regra 8 do `AGENTS.md` (sem comentários em código). Ainda assim, o diff edita e
acrescenta comentários em arquivos que tocou:

- `app/internal/platform/op/op.go:1-3` (doc do pacote reescrita).
- `op.go:28` (`// Exactly one of Permission, Public or Authenticated must be set.`).
- `app/cmd/api/main.go:6` (linha de uso de `create-admin` no doc do comando).

São blocos que já existiam na foundation e foram estendidos em vez de removidos. Isso não entra no veredito, porque
nenhuma check o afirma. Fica registrado para o usuário decidir.

Fixtures da foundation alteradas (declaradas pelo autor e conferidas):

- `web/src/test/render.tsx` passou a responder `GET /api/v1/users/me` como admin em `stubFetch`, e o contador de
  respostas ignora `/me`, então a sequência dos testes da foundation se mantém.
- `web/e2e/status.spec.ts` agora faz login antes de `API: online`.

Nenhum arquivo de teste da foundation mudou (`git diff --stat 2c12067..afbc039 -- 'web/src/**/*.test.ts*'` lista só
arquivos de `users` e `routes/authed.test.tsx`). As asserções de C49 da foundation estão intactas e o e2e `api online`
passou (B6).

## Gate

`task check` - GATE_RESULT
