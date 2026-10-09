# Audit links verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 5d54186..b3cd52f
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier) - sub-agente independente, sem contexto do build

Todas as 10 checks têm prova verde em `HEAD` com asserção localizada, as 4 linhas de `Coverage` recompostas não
deixam membro sem prova e as 5 falhas injetadas foram mortas. O FAIL vem de uma única linha de `Test policy` não
atendida: `RoleDetail` foi classificado como "Decides, reached across a boundary", mas a prova na fronteira é
declarada "por analogia ao C9" e não existe - nenhum teste segue `Ver auditoria` a partir de `/roles/<id>`. É
exatamente a lição confirmada L-008 (estado aplicado a uma segunda tela tem de ser provado nela).

## Binding sources

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| nenhuma fonte binding - o plano não marca design ou contrato binding; perfil `standard`, passo 1 não se aplica | n/a | - | - |

## Checks

Provas executadas em `b3cd52f`, em lote:

- Vitest: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx src/features/rbac/RoleDetail.test.tsx --reporter=verbose` - exit 0, 35 passed (2 arquivos); cada teste nomeado abaixo aparece individualmente com `✓`.
- Playwright: `task e2e -- audit-links.spec.ts` - exit 0, `2 passed (3.2m)`; binário construído do zero (porta 8080 livre antes da execução, sem reaproveitar servidor).

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | `Ver auditoria` com `href` `/audit?resource_type=user&resource_id=<id>` | Vitest lote, `✓ links to the audit trail of the user` | `web/src/features/users/UserDetail.test.tsx:134` - `expect(link).toHaveAttribute("href", \`/audit?resource_type=user&resource_id=${id}\`)` | PASS |
| C2 | `Ver ações` com `href` `/audit?actor_id=<id>` | Vitest lote, `✓ links to the actions of the user` | `web/src/features/users/UserDetail.test.tsx:141` - `expect(link).toHaveAttribute("href", \`/audit?actor_id=${id}\`)` | PASS |
| C3 | sem `audit:read`: nenhum link, título e `Salvar` presentes | Vitest lote, `✓ hides audit links without audit:read` | `web/src/features/users/UserDetail.test.tsx:147-150` - `findByRole("heading", { name: "b@x.com" })`, `getByRole("button", { name: "Salvar" })`, `queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument()`, idem `Ver ações` | PASS |
| C4 | pendente, `404`, `403`, `500`: nenhum link, após o texto do estado | Vitest lote, `✓ hides audit links while pending` / `while not found` / `while forbidden` / `while failed` | `web/src/features/users/UserDetail.test.tsx:161-163` - `expect(await screen.findByText(text))` seguido de `queryByRole("link", ...)).not.toBeInTheDocument()` para os dois links; tabela de casos em `:153-157` | PASS |
| C5 | `Ver auditoria` com `href` `/audit?resource_type=role&resource_id=<id>` | Vitest lote, `✓ links to the audit trail of the role` | `web/src/features/rbac/RoleDetail.test.tsx:193` - `expect(link).toHaveAttribute("href", \`/audit?resource_type=role&resource_id=${id}\`)` | PASS |
| C6 | sem `audit:read`: sem link, nome no título e `Salvar` presentes | Vitest lote, `✓ hides audit link without audit:read` | `web/src/features/rbac/RoleDetail.test.tsx:211-213` - `findByRole("heading", { name: "Leitor" })`, `getByRole("button", { name: "Salvar" })`, `queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument()` | PASS |
| C7 | pendente, `404`, `403`, `500`: sem link, após o texto do estado | Vitest lote, `✓ hides audit link while pending` / `while not found` / `while forbidden` / `while failed` | `web/src/features/rbac/RoleDetail.test.tsx:228-229` - `expect(await screen.findByText(text))` seguido de `queryByRole("link", { name: "Ver auditoria" })).not.toBeInTheDocument()`; casos em `:216-220` | PASS |
| C8 | papel `admin` travado mostra `Ver auditoria` com o próprio id | Vitest lote, `✓ links to the audit trail of the admin role` | `web/src/features/rbac/RoleDetail.test.tsx:205` - `expect(link).toHaveAttribute("href", \`/audit?resource_type=role&resource_id=${id}\`)` (papel `name: "admin"` em `:199`) | PASS |
| C9 | e2e: URL com `resource_type=user`/`resource_id`, exatamente um `user.created`, toda linha `Recurso` `user <id>`, `Limpar filtros` | Playwright lote, `✓ admin follows Ver auditoria from a user (7.6s)` | `web/e2e/audit-links.spec.ts:29-34` - `search.get("resource_type")).toBe("user")`, `search.get("resource_id")).toBe(userId)`, `getByRole("button", { name: "Limpar filtros" })).toBeVisible()`, `filter({ hasText: "user.created" })).toHaveCount(1)`, `row.locator("td").nth(3)).toHaveText(\`user ${userId}\`)` | PASS |
| C10 | e2e: URL com `actor_id=<id>`, pelo menos uma linha, toda linha `Quem` = e-mail do admin | Playwright lote, `✓ admin follows Ver ações on self (3.7s)` | `web/e2e/audit-links.spec.ts:45-48` - `searchParams.get("actor_id")).toBe(me.id)`, `eventRows(page).first()).toBeVisible()`, `row.locator("td").nth(2)).toHaveText(admin.email)` | PASS |

Notas de leitura:

- Colunas da tabela de `/audit` conferidas em `web/src/features/audit/AuditList.tsx:120-123` (`Quando`, `Ação`, `Quem`, `Recurso`): `nth(2)` é `Quem` e `nth(3)` é `Recurso`, como as claims exigem.
- As provas tocam só arquivos do diff (`UserDetail.test.tsx`, `RoleDetail.test.tsx`, `e2e/audit-links.spec.ts`), todos alterados em `b3cd52f`.
- Regra sem comentários (`AGENTS.md` regra 8): nenhuma linha adicionada no diff de `web/` contém `//` ou `/*` de comentário.

## Coverage

Recomposto a partir do código (`UserDetail.tsx:85-133`, `RoleDetail.tsx:80-118`) e da autoridade do conjunto (as AC do plano), não da tabela do `checks.md`.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| links da tela de usuário (2) | `web/src/features/users/UserDetail.tsx:124-133` - dois `Link` | `Ver auditoria` C1 (`:134`) · `Ver ações` C2 (`:141`) | - |
| ramos de visibilidade em `UserDetail` (6) | AC 1-4 + retornos antecipados em `UserDetail.tsx:85-104` | carregado com `audit:read` C1, C2 · sem `audit:read` C3 · pendente, `404`, `403`, `500` C4 (um caso cada, `:153-157`) | - |
| ramos de visibilidade em `RoleDetail` (7) | AC 5-7 + Assumptions `Papel admin` + retornos em `RoleDetail.tsx:80-92` | carregado C5 · `admin` travado C8 · sem `audit:read` C6 · pendente, `404`, `403`, `500` C7 (`:216-220`) | - |
| filtros de `/audit` usados por link (3) | `web/src/routes/_authed/audit/index.tsx:6-12` (`validateSearch`) | `resource_type`+`resource_id` de usuário C1, C9 · `resource_type`+`resource_id` de papel C5 · `actor_id` C2, C10 | - |

Varredura de conjuntos sem linha:

- `Surface`, `Relations` e `Landing` do plano estão vazios (`None`); nenhuma rota ou entidade deve linha.
- Ramo `!can(me, "users:read")` de `UserDetail.tsx:85` e estados do catálogo (`catalogue.isPending`/`isError`, `failure = role.error ?? catalogue.error`) de `RoleDetail.tsx:80-92`: compartilham o mesmo `return` dos casos provados (`403`, pendente, erro), e as falhas F2/F4 mostraram que os links não podem escapar desses retornos sem matar C3/C4/C6/C7. Não são membros das AC (que nomeiam o pedido do registro); sem lacuna.
- O membro "filtro de papel" é provado na própria camada (C5, `href` exato) mas não atravessa até a API real - tratado em `Test policy rows`, não como membro sem prova, porque a claim C5 é sobre o `href`.

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `web/src/features/users/UserDetail.tsx` | boundary C9, C10 (`web/e2e/audit-links.spec.ts:29-48`) · own layer C1-C4 (6 casos da tabela de decisão, um por linha) | yes |
| Decides, reached across a boundary | `web/src/features/rbac/RoleDetail.tsx` | own layer C5-C8 presente (7 casos) · boundary **ausente**: o `checks.md` declara "boundary por analogia ao C9"; `rg -n "roles" web/e2e/audit-links.spec.ts` não acha nada e nenhum spec em `web/e2e/` segue `Ver auditoria` de `/roles/<id>` nem prova `/audit?resource_type=role` contra a API (filtros por URL em `AuditList.test.tsx:180-201` e `list_events_test.go:190-191` só usam `user`). O "Independent test" do S2 no plano fica sem prova automatizada. Lição L-008 recorrente | no - gap |
| Decides, not reached across a boundary | nenhum arquivo do diff | - | n/a |
| Entry point that decides nothing | nenhum arquivo do diff | - | n/a |
| Instrumentation, pass-throughs | `validateSearch` de `/audit` (existente, não alterado) | coberto pelo consumidor C9, C10 | yes |

## Faults injected

Isolamento: `git worktree add <scratchpad>/wt HEAD` com junção para `web/node_modules`; a árvore real nunca foi
editada. Baseline `git status --porcelain` gravado antes; após remover a junção e o worktree, o porcelain da árvore
real é idêntico ao baseline (`diff` vazio - só os `??` pré-existentes de `.claude/skills/auth-security/` e
`.cursor/skills/auth-security/`). O e2e do mutante rodou no worktree com `COMPOSE_PROJECT_NAME=aigateway` e porta
8080 livre, então o binário foi construído do código mutado.

| Mutation | Location | Killed |
| --- | --- | --- |
| F1 - troca os `search` dos dois links (`Ver auditoria` -> `actor_id`, `Ver ações` -> `resource_type`/`resource_id`) | `web/src/features/users/UserDetail.tsx:126,129` | yes - Vitest: `× links to the audit trail of the user`, `× links to the actions of the user` (C1, C2) |
| F1 no e2e - mesmo mutante, binário reconstruído | `web/src/features/users/UserDetail.tsx:126,129` | yes - Playwright exit 1: `✘ admin follows Ver auditoria from a user` (`audit-links.spec.ts:29`, Expected "user", Received null), `✘ admin follows Ver ações on self` (`:45`, Received null) (C9, C10) |
| F2 - links renderizados antes do carregamento e sem a trava `audit:read` (wrapper que mostra os links com o `id` da prop acima do corpo) | `web/src/features/users/UserDetail.tsx:30,124` | yes - `× hides audit links without audit:read`, `× hides audit links while pending/not found/forbidden/failed` (C3, C4) |
| F3 - `resource_type: "role"` -> `"user"` | `web/src/features/rbac/RoleDetail.tsx:113` | yes - `× links to the audit trail of the role`, `× links to the audit trail of the admin role` (C5, C8) |
| F4 - link renderizado antes do carregamento e sem a trava `audit:read` (wrapper acima do corpo) | `web/src/features/rbac/RoleDetail.tsx:34,110` | yes - `× hides audit link without audit:read`, `× hides audit link while pending/not found/forbidden/failed` (C6, C7) |

Toda prova que carrega uma check (C1-C10) falhou ao menos uma vez. Observação lateral: na rodada F1 do Vitest, os
testes pré-existentes `shows loading` (UserDetail e RoleDetail) estouraram o timeout de 5 s enquanto o build do e2e
disputava CPU; RoleDetail não estava mutado, e em `HEAD` sem concorrência os dois passam. Não é achado desta feature.

## Swept existing

- authorization: "a API segue exigindo `audit:read` em `/audit/events`" - confirmado em `app/internal/features/audit/list_events/endpoint.go:18` (`const permission op.Permission = "audit:read"`) e `:48` (`Permission: permission`).
- demais linhas `n/a` são política aprovada; nada a conferir no código.

## Gate

`npm --prefix web run test -- src/features/users/UserDetail.test.tsx src/features/rbac/RoleDetail.test.tsx --reporter=verbose` - 35 passed, 0 failed
`task e2e -- audit-links.spec.ts` - 2 passed, 0 failed

## Ranked gaps

1. Linha de `Test policy` "Decides, reached across a boundary" não atendida para `web/src/features/rbac/RoleDetail.tsx`: falta a prova na fronteira (Playwright seguindo `Ver auditoria` de `/roles/<id>` até `/audit` com `resource_type=role`/`resource_id=<id>`, listando o `role.created` do papel). O "Independent test" do S2 do plano descreve exatamente esse teste. C5-C8 - no evidence at the boundary (`web/e2e/audit-links.spec.ts` só cobre usuário). Recorrência de L-008.
