# Audit links checks

Profile: standard
Plan: `.specs/features/audit-links/plan.md`

11 checks in 3 slices · 0 one-way doors · 0 open, of which 0 block

## Checks

### S1 - links na tela do usuário · 2 files · 13 KB · ~4k

**C1** - Com o usuário carregado e `audit:read`, `/users/$id` mostra o link `Ver auditoria` com `href` exatamente `/audit?resource_type=user&resource_id=<id>` (AUDL-01, AC 1) `[done]`
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "links to the audit trail of the user"`

**C2** - Com o usuário carregado e `audit:read`, `/users/$id` mostra o link `Ver ações` com `href` exatamente `/audit?actor_id=<id>` (AUDL-01, AC 2) `[done]`
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "links to the actions of the user"`

**C3** - Sem `audit:read`, `/users/$id` carregado não mostra `Ver auditoria` nem `Ver ações`, e ainda mostra o e-mail no título e o botão `Salvar` (AUDL-01, AC 3) `[done]`
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "hides audit links without audit:read"`

**C4** - Com `audit:read`, `/users/$id` não mostra nenhum dos dois links enquanto o pedido está pendente, nem depois de `404`, `403` ou `500` - um caso afirmado por estado, cada um esperando o texto do próprio estado antes de afirmar a ausência (AUDL-01, AC 4) `[done]`
Proof: `npm --prefix web run test -- src/features/users/UserDetail.test.tsx -t "hides audit links while"`

### S2 - link na tela do papel · 2 files · 15 KB · ~4k

**C5** - Com o papel carregado e `audit:read`, `/roles/$id` mostra o link `Ver auditoria` com `href` exatamente `/audit?resource_type=role&resource_id=<id>` (AUDL-02, AC 5) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "links to the audit trail of the role"`

**C6** - Sem `audit:read`, `/roles/$id` carregado não mostra `Ver auditoria` e ainda mostra o nome no título e o botão `Salvar` (AUDL-02, AC 6) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "hides audit link without audit:read"`

**C7** - Com `audit:read`, `/roles/$id` não mostra `Ver auditoria` enquanto o pedido está pendente, nem depois de `404`, `403` ou `500` - um caso afirmado por estado, cada um esperando o texto do próprio estado (AUDL-02, AC 7) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "hides audit link while"`

**C8** - O papel `admin` travado, com `audit:read`, mostra `Ver auditoria` com o próprio id (AUDL-02, AC 5; Assumptions `Papel admin`) `[done]`
Proof: `npm --prefix web run test -- src/features/rbac/RoleDetail.test.tsx -t "links to the audit trail of the admin role"`

### S3 - ponta a ponta · 1 file · 2 KB · ~1k

**C9** - Um admin cria um usuário, abre `/users/<id>`, clica `Ver auditoria`; a URL tem `resource_type=user` e `resource_id=<id>`, a tabela tem exatamente uma linha `user.created` com `Recurso` `user <id>`, toda linha tem `Recurso` `user <id>`, e `Limpar filtros` aparece (AUDL-03, AC 8) `[done]`
Proof: `task e2e -- audit-links.spec.ts -g "follows Ver auditoria from a user"`

**C10** - Um admin abre o próprio `/users/<id>` e clica `Ver ações`; a URL tem `actor_id=<id>`, a tabela tem pelo menos uma linha e toda linha tem `Quem` igual ao e-mail do admin (AUDL-03, AC 9) `[done]`
Proof: `task e2e -- audit-links.spec.ts -g "follows Ver ações on self"`

**C11** - Um admin cria um papel, abre `/roles/<id>` e clica `Ver auditoria`; a URL tem `resource_type=role` e `resource_id=<id>`, a tabela tem exatamente uma linha `role.created`, toda linha tem `Recurso` `role <id>`, e `Limpar filtros` aparece (AUDL-02, AC 5; Test policy - prova na fronteira de `RoleDetail`, lacuna da verificação rodada 1) `[done]`
Proof: `task e2e -- audit-links.spec.ts -g "follows Ver auditoria from a role"`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| links da tela de usuário (2) | `Ver auditoria` C1 · `Ver ações` C2 | - |
| ramos de visibilidade em `UserDetail` (6) | carregado com `audit:read` C1, C2 · sem `audit:read` C3 · pendente C4 · `404` C4 · `403` C4 · `500` C4 | - |
| ramos de visibilidade em `RoleDetail` (7) | carregado com `audit:read` C5 · `admin` travado C8 · sem `audit:read` C6 · pendente C7 · `404` C7 · `403` C7 · `500` C7 | - |
| filtros de `/audit` usados por link (3) | `resource_type`+`resource_id` de usuário C1, C9 · `resource_type`+`resource_id` de papel C5, C11 · `actor_id` C2, C10 | - |

- O plano não tem `Surface`, `Relations` nem `Landing` com linhas; nada mais a juntar
- Claims naming a URL: C1, C2, C5, C8 afirmam o `href` exato na própria camada; C9, C10 cruzam até a API real

## Test policy

O `AGENTS.md` já responde às duas perguntas (`## Test policy`); estas são as linhas dele, aplicadas.

| Code | Required proofs | Coverage expectation |
| --- | --- | --- |
| Decides, reached across a boundary | one at the boundary **and** one at its own layer | the contract at the boundary; one asserted case per row of the decision table at its own layer |
| Decides, not reached across a boundary | one at its own layer | one asserted case per row of the decision table |
| Entry point that decides nothing | one at the boundary | accepted input, each rejected input, each error path |
| Instrumentation, pass-throughs | none of its own | covered by its consumer's proof |

Evidence (planned code, by shape):

- `features/users/UserDetail`: o ramo novo decide mostrar ou não os links (permissão `audit:read` × 5 estados do pedido, 6 casos) -> decides, reached across a boundary: own layer C1-C4 (Vitest), boundary C9-C10 (Playwright contra a API real)
- `features/rbac/RoleDetail`: mesmo ramo, mais o `admin` travado (7 casos) -> decides, reached across a boundary: own layer C5-C8, boundary C11 (Playwright contra a API real)
- closest analogue: o link `Auditoria` do `UserMenu`, gated por `audit:read` e provado em `UserMenu.test.tsx`

## Swept

- validation: n/a - não há entrada do usuário; o id vem do registro que a API devolveu
- failure modes: C4, C7 - pedido em erro não mostra link
- idempotency: n/a - links só navegam, nada é escrito
- authorization: C3, C6; a API segue exigindo `audit:read` em `/audit/events` (existing - a check de permissão do plano `audit`, AUD-04)
- concurrency: n/a - nenhuma escrita, nenhum estado compartilhado
- data lifecycle: n/a - nada é armazenado
- dependency failure: C4, C7 - falha da API do registro
- state transitions: n/a - os links não mudam estado; usuário desativado continua com trilha consultável
- observability: n/a - navegação na web não gera log; a leitura da auditoria não é auditada (audit Out of scope)

## Handoff

Intended split, with the arithmetic, written before any code:

- S1-S3 ≈ 9k, bem abaixo do budget de 150k -> um builder, sem handoff
