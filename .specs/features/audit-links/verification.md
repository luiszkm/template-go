# Audit links verification

**Verdict**: PASS
**Profile**: standard
**Diff range**: 5d54186..46c42a6 (fix desta rodada: b3cd52f..46c42a6)
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier) - sub-agente independente, sem contexto do build nem do fix

A única lacuna da rodada 1 (linha de `Test policy` "Decides, reached across a boundary" sem prova na fronteira para
`RoleDetail.tsx`) está fechada pela nova check C11, um teste Playwright que cria um papel, segue `Ver auditoria`
a partir de `/roles/<id>` e afirma a URL e a tabela de `/audit` contra a API real. As 11 provas passam em `46c42a6`
e as duas falhas injetadas sobre a superfície nova de C11 foram mortas.

Escopo desta rodada: `git diff --stat b3cd52f..HEAD` toca apenas `.specs/features/audit-links/checks.md`,
`.specs/features/audit-links/verification.md` e `web/e2e/audit-links.spec.ts` (+21 linhas, só o teste novo).
Nenhum código de produção mudou, então `UserDetail.tsx`, `RoleDetail.tsx` e seus testes Vitest são os mesmos de
`b3cd52f`. O que não foi tocado pelo fix vem da rodada 1 e está marcado `carried from b3cd52f`; todas as provas
foram reexecutadas em `46c42a6`.

## Binding sources

carried from b3cd52f - o fix não tocou a interface.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| nenhuma fonte binding - o plano não marca design ou contrato binding; perfil `standard`, passo 1 não se aplica | n/a | - | - |

## Checks

verified at 46c42a6 - provas reexecutadas em lote:

- Vitest: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx src/features/rbac/RoleDetail.test.tsx --reporter=verbose` - exit 0, `Test Files 2 passed (2)`, `Tests 35 passed (35)`; cada teste nomeado abaixo aparece individualmente com `✓`.
- Playwright: `task e2e -- audit-links.spec.ts --reporter=list` - exit 0, `3 passed`; linhas `✓ 1 ... audit-links.spec.ts:15:1 › admin follows Ver auditoria from a user`, `✓ 2 ... :37:1 › admin follows Ver ações on self`, `✓ 3 ... :51:1 › admin follows Ver auditoria from a role`. Porta 8080 livre antes da execução (`netstat` sem `LISTEN`), então o binário foi construído e servido pelo próprio `webServer`.

Citações de C1-C10: os arquivos de teste não mudaram no fix (`web/e2e/audit-links.spec.ts` só recebeu linhas após a `:49`), então os `file:line` da rodada 1 continuam válidos e foram reconferidos na leitura do arquivo em `HEAD`.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `Ver auditoria` com `href` `/audit?resource_type=user&resource_id=<id>` | Vitest lote, `✓ links to the audit trail of the user` | `web/src/features/users/UserDetail.test.tsx:134` - `expect(link).toHaveAttribute("href", \`/audit?resource_type=user&resource_id=${id}\`)` | PASS |
| C2 | `Ver ações` com `href` `/audit?actor_id=<id>` | Vitest lote, `✓ links to the actions of the user` | `web/src/features/users/UserDetail.test.tsx:141` - `expect(link).toHaveAttribute("href", \`/audit?actor_id=${id}\`)` | PASS |
| C3 | sem `audit:read`: nenhum link, título e `Salvar` presentes | Vitest lote, `✓ hides audit links without audit:read` | `web/src/features/users/UserDetail.test.tsx:147-150` - `findByRole("heading", { name: "b@x.com" })`, `getByRole("button", { name: "Salvar" })`, `queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument()`, idem `Ver ações` | PASS |
| C4 | pendente, `404`, `403`, `500`: nenhum link, após o texto do estado | Vitest lote, `✓ hides audit links while pending` / `while not found` / `while forbidden` / `while failed` | `web/src/features/users/UserDetail.test.tsx:161-163` - `expect(await screen.findByText(text))` seguido de `queryByRole("link", ...)).not.toBeInTheDocument()` para os dois links; casos em `:153-157` | PASS |
| C5 | `Ver auditoria` com `href` `/audit?resource_type=role&resource_id=<id>` | Vitest lote, `✓ links to the audit trail of the role` | `web/src/features/rbac/RoleDetail.test.tsx:193` - `expect(link).toHaveAttribute("href", \`/audit?resource_type=role&resource_id=${id}\`)` | PASS |
| C6 | sem `audit:read`: sem link, nome no título e `Salvar` presentes | Vitest lote, `✓ hides audit link without audit:read` | `web/src/features/rbac/RoleDetail.test.tsx:211-213` - `findByRole("heading", { name: "Leitor" })`, `getByRole("button", { name: "Salvar" })`, `queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument()` | PASS |
| C7 | pendente, `404`, `403`, `500`: sem link, após o texto do estado | Vitest lote, `✓ hides audit link while pending` / `while not found` / `while forbidden` / `while failed` | `web/src/features/rbac/RoleDetail.test.tsx:228-229` - `expect(await screen.findByText(text))` seguido de `queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument()`; casos em `:216-220` | PASS |
| C8 | papel `admin` travado mostra `Ver auditoria` com o próprio id | Vitest lote, `✓ links to the audit trail of the admin role` | `web/src/features/rbac/RoleDetail.test.tsx:205` - `expect(link).toHaveAttribute("href", \`/audit?resource_type=role&resource_id=${id}\`)` (papel `name: "admin"` em `:199`) | PASS |
| C9 | e2e: URL com `resource_type=user`/`resource_id`, exatamente um `user.created`, toda linha `Recurso` `user <id>`, `Limpar filtros` | Playwright lote, `✓ admin follows Ver auditoria from a user` | `web/e2e/audit-links.spec.ts:29-34` - `search.get("resource_type")).toBe("user")`, `search.get("resource_id")).toBe(userId)`, `getByRole("button", { name: "Limpar filtros" })).toBeVisible()`, `filter({ hasText: "user.created" })).toHaveCount(1)`, `row.locator("td").nth(3)).toHaveText(\`user ${userId}\`)` | PASS |
| C10 | e2e: URL com `actor_id=<id>`, pelo menos uma linha, toda linha `Quem` = e-mail do admin | Playwright lote, `✓ admin follows Ver ações on self` | `web/e2e/audit-links.spec.ts:45-48` - `searchParams.get("actor_id")).toBe(me.id)`, `eventRows(page).first()).toBeVisible()`, `row.locator("td").nth(2)).toHaveText(admin.email)` | PASS |
| C11 | e2e: admin cria papel, segue `Ver auditoria` de `/roles/<id>`; URL com `resource_type=role` e `resource_id=<id>`, exatamente um `role.created`, toda linha `Recurso` `role <id>`, `Limpar filtros` | Playwright lote, `✓ admin follows Ver auditoria from a role` (`audit-links.spec.ts:51:1`) | `web/e2e/audit-links.spec.ts:64` - `expect(search.get("resource_type")).toBe("role")`; `:65` - `expect(search.get("resource_id")).toBe(roleId)`; `:66` - `getByRole("button", { name: "Limpar filtros" })).toBeVisible()`; `:67` - `eventRows(page).filter({ hasText: "role.created" })).toHaveCount(1)`; `:69` - `row.locator("td").nth(3)).toHaveText(\`role ${roleId}\`)`. Pré-condição nomeada pela claim: papel criado pela UI em `:53-58` e o clique em `Ver auditoria` a partir de `/roles/<id>` em `:60` | PASS |

Notas de leitura (verified at 46c42a6):

- C11 segue o clique real a partir da tela do papel (`:57` exige `/roles/<uuid>` antes do clique em `:60`), então prova o link renderizado por `RoleDetail.tsx:110-117`, não uma navegação direta para `/audit`.
- O laço de `:68-69` sobre "toda linha" não é vácuo: `:67` exige pelo menos (exatamente) uma linha `role.created` antes dele.
- `nth(3)` é a coluna `Recurso` - carried from b3cd52f (`web/src/features/audit/AuditList.tsx:120-123`, arquivo fora do diff).
- Regra sem comentários (`AGENTS.md` regra 8): as 21 linhas adicionadas em `web/e2e/audit-links.spec.ts` não contêm `//` nem `/*`.
- A prova de C11 toca apenas arquivo alterado pelo fix (`web/e2e/audit-links.spec.ts`).

## Coverage

Linha "filtros de `/audit` usados por link" verified at 46c42a6 (o fix acrescentou C11 a ela); demais linhas carried from b3cd52f - nenhum código de produção mudou, então as autoridades desses conjuntos não mudaram.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| links da tela de usuário (2) | carried from b3cd52f - `web/src/features/users/UserDetail.tsx:124-133` | `Ver auditoria` C1 (`UserDetail.test.tsx:134`) · `Ver ações` C2 (`:141`) | - |
| ramos de visibilidade em `UserDetail` (6) | carried from b3cd52f - AC 1-4 + retornos em `UserDetail.tsx:85-104` | carregado com `audit:read` C1, C2 · sem `audit:read` C3 · pendente, `404`, `403`, `500` C4 (`:153-157`) | - |
| ramos de visibilidade em `RoleDetail` (7) | carried from b3cd52f - AC 5-7 + Assumptions `Papel admin` + retornos em `RoleDetail.tsx:80-92` | carregado C5 · `admin` travado C8 · sem `audit:read` C6 · pendente, `404`, `403`, `500` C7 (`:216-220`) | - |
| filtros de `/audit` usados por link (3) | verified at 46c42a6 - `web/src/routes/_authed/audit/index.tsx:6-12` (`validateSearch`) e os `search` em `UserDetail.tsx:126,129` e `RoleDetail.tsx:113` | `resource_type`+`resource_id` de usuário C1, C9 · `resource_type`+`resource_id` de papel C5 (camada própria), C11 (fronteira, `audit-links.spec.ts:64-69`) · `actor_id` C2, C10 | - |

Varredura de conjuntos sem linha: carried from b3cd52f (`Surface`, `Relations` e `Landing` vazios; ramos de catálogo
e `users:read` compartilham os retornos provados). O ponto que a rodada 1 tratou à parte - o filtro de papel sem
travessia até a API real - agora tem prova na fronteira (C11).

## Test policy rows

verified at 46c42a6 - linha não atendida da rodada 1 rejulgada; a linha de `UserDetail` carried from b3cd52f (nenhum arquivo classificado por ela foi tocado).

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `web/src/features/users/UserDetail.tsx` | carried from b3cd52f - boundary C9, C10 (`web/e2e/audit-links.spec.ts:29-48`) · own layer C1-C4 (6 casos, um por linha da tabela de decisão) | yes |
| Decides, reached across a boundary | `web/src/features/rbac/RoleDetail.tsx` | boundary C11 (`web/e2e/audit-links.spec.ts:51-70`, contrato `/audit?resource_type=role&resource_id=<id>` afirmado contra a API real em `:64-69`) · own layer C5-C8 (7 casos, um por linha: carregado, `admin` travado, sem `audit:read`, pendente, `404`, `403`, `500`) | yes - lacuna da rodada 1 fechada; F5 e F6 mostram que a prova da fronteira falha sob mutação |
| Decides, not reached across a boundary | nenhum arquivo do diff | - | n/a |
| Entry point that decides nothing | nenhum arquivo do diff | - | n/a |
| Instrumentation, pass-throughs | `validateSearch` de `/audit` (existente, não alterado) | coberto pelos consumidores C9, C10, C11 | yes |

## Faults injected

F1-F4 carried from b3cd52f (superfícies não tocadas pelo fix; os testes Vitest e as linhas `:15-49` do spec não
mudaram). F5-F6 verified at 46c42a6, sobre a superfície criada pelo fix (C11).

Isolamento de F5-F6: `git worktree add <scratchpad>/wt HEAD` (`46c42a6`) com junção para `web/node_modules`; a
árvore real nunca foi editada. Baseline `git status --porcelain` gravado antes (só `?? .claude/skills/auth-security/`
e `?? .cursor/skills/auth-security/`). Cada mutante rodou `task e2e -- audit-links.spec.ts` no worktree com
`COMPOSE_PROJECT_NAME=aigateway` e a porta 8080 livre, então o binário foi construído do código mutado. Depois:
junção removida, `git worktree remove --force` e `git worktree prune`; `git worktree list` mostra só a árvore real
e o porcelain da árvore real é idêntico ao baseline (`diff` vazio).

| Mutation | Location | Killed |
| --- | --- | --- |
| F1 - troca os `search` dos dois links (carried from b3cd52f) | `web/src/features/users/UserDetail.tsx:126,129` | yes - Vitest C1, C2; Playwright C9, C10 |
| F2 - links fora da trava `audit:read` e antes do carregamento (carried from b3cd52f) | `web/src/features/users/UserDetail.tsx:30,124` | yes - C3, C4 |
| F3 - `resource_type: "role"` -> `"user"` (carried from b3cd52f) | `web/src/features/rbac/RoleDetail.tsx:113` | yes - Vitest C5, C8 |
| F4 - link fora da trava `audit:read` e antes do carregamento (carried from b3cd52f) | `web/src/features/rbac/RoleDetail.tsx:34,110` | yes - C6, C7 |
| F5 - `resource_type: "role"` -> `"user"` no link do papel, binário reconstruído | `web/src/features/rbac/RoleDetail.tsx:113` | yes - Playwright exit 1: `✘ 3 ... audit-links.spec.ts:51:1 › admin follows Ver auditoria from a role`, `Expected: "role"`, `Received: "user"` em `audit-links.spec.ts:64` (C11); C9 e C10 seguiram verdes |
| F6 - filtro `resource_id` ignorado pela API (`pgtype.Text{..., Valid: true}` -> `Valid: false`), URL do link intacta | `app/internal/features/audit/list_events/endpoint.go:95` | yes - Playwright exit 1: `✘ ... :51:1 › admin follows Ver auditoria from a role`, `toHaveCount` `Expected: 1`, `Received: 9` em `audit-links.spec.ts:67` (C11); também `✘ ... :15:1` em `:32` (C9, `Received: 46`) |

F5 mata a asserção de URL de C11 e F6 mata as asserções da tabela de C11 com a URL correta - as duas superfícies de
asserção da prova nova. F6 muta código pré-existente (fora do diff da feature) de propósito: é a única forma de
deixar a URL certa e o conteúdo errado, que é exatamente o contrato na fronteira que a linha de `Test policy` exige.

## Swept existing

carried from b3cd52f - authorization: "a API segue exigindo `audit:read` em `/audit/events`" confirmado em
`app/internal/features/audit/list_events/endpoint.go:18` (`const permission op.Permission = "audit:read"`) e `:48`
(`Permission: permission`). Demais linhas `n/a` são política aprovada.

## Gate

verified at 46c42a6:

`npm --prefix web run test -- src/features/users/UserDetail.test.tsx src/features/rbac/RoleDetail.test.tsx --reporter=verbose` - 35 passed, 0 failed
`task e2e -- audit-links.spec.ts --reporter=list` - 3 passed, 0 failed

## Ranked gaps

Nenhuma. A lacuna única da rodada 1 (prova na fronteira de `RoleDetail.tsx`) está fechada por C11.
