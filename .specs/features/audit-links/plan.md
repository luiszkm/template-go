# Audit links

Sources:

- conversation (2026-10-09) - candidato 1 do `STATE.md` Handoff: links "Ver auditoria" nas telas de usuário e de papel; o usuário escolheu dois links na tela de usuário (eventos sobre ele e ações dele)
- `.specs/features/audit/plan.md` `## Out of scope` ("a tela aceita os filtros pela URL, então o link entra depois sem mudar este contrato") e AC 25, AC 26 (filtros na URL chegam à API na primeira requisição e aparecem com `Limpar filtros`)
- `.specs/LESSONS.md` L-008 e L-010 (confirmadas) - estado aplicado a uma segunda tela é provado nela; componente web que ramifica tem todos os ramos provados

## Problem

O admin que está na tela de um usuário ou de um papel e quer saber o que aconteceu com ele precisa sair para
`/audit`, clicar num evento qualquer daquele recurso para filtrar, ou montar a URL à mão com o UUID copiado da
barra de endereço. Para saber o que um usuário fez, o caminho é o mesmo, achando uma linha em que ele aparece em
`Quem`. A pergunta "quem mexeu neste usuário?" custa vários passos e depende de o evento estar na primeira página.

A fonte não traz números: o template ainda não tem usuários reais.

Quando isto entra, quem tem `audit:read` vê na tela do usuário os links `Ver auditoria` e `Ver ações`, e na tela
do papel o link `Ver auditoria`, cada um abrindo `/audit` já filtrado.

## Out of scope

| Excluded | Why |
| --- | --- |
| Mudar a API ou a tela `/audit` | a tela já lê os filtros da URL (audit AC 26); este plano só cria links para ela |
| Mostrar eventos embutidos na tela do usuário ou do papel | duplicaria a lista de `/audit` com seus estados e paginação |
| Link na listagem `/users` ou `/roles` (por linha) | nenhuma fonte pediu; a tela de detalhe está a um clique |
| Link para eventos de sessão (`resource_type=session`) | o `resource_id` da sessão é o hash do token, que a tela do usuário não conhece; o login aparece em `Ver ações` (o ator do `session.created` é o usuário) |

## Assumptions

| Assumption | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Links na tela de usuário | dois: `Ver auditoria` -> `/audit?resource_type=user&resource_id=<id>` e `Ver ações` -> `/audit?actor_id=<id>` | respondem "quem mexeu neste usuário" e "o que este usuário fez"; a `/audit` já aceita os dois filtros | y |
| Link na tela de papel | um: `Ver auditoria` -> `/audit?resource_type=role&resource_id=<id>` | papel não age, só sofre ações | y |
| Quem vê os links | só quem tem `audit:read`; sem ela os links não aparecem | um link que leva a `Você não tem permissão` é ruído; mesma regra do menu `Auditoria` | y |
| Quando os links aparecem | só depois que o registro carregou; não aparecem em carregando, erro, `404` ou `403` | não há id válido para filtrar antes disso | y |
| Posição | abaixo do título, antes do formulário, como links sublinhados (`Link` do TanStack Router, como em `UsersList`) | é navegação, não ação; fica visível sem rolar | y |
| Papel `admin` (travado) | mostra `Ver auditoria` como qualquer papel | a trava impede editar, não consultar | y |
| Usuário vendo a si mesmo | mostra os dois links | a própria trilha não é segredo para quem tem `audit:read` | y |

**Open questions:** none - all resolved or logged above.

## Criteria

### S1: Web - links na tela do usuário (P1)

**Acceptance Criteria**

1. WHEN `/users/$id` has loaded the user and the viewer holds `audit:read` THEN the web app SHALL show a link `Ver auditoria` whose target is `/audit?resource_type=user&resource_id=<id>`
2. WHEN `/users/$id` has loaded the user and the viewer holds `audit:read` THEN the web app SHALL show a link `Ver ações` whose target is `/audit?actor_id=<id>`
3. IF the viewer does not hold `audit:read` THEN `/users/$id` SHALL show neither `Ver auditoria` nor `Ver ações`, and the rest of the screen SHALL be unchanged
4. WHILE the user request is pending, and IF it ends in `404`, `403` or another error, THEN `/users/$id` SHALL show neither link

**Independent test:** como admin, abrir um usuário e clicar `Ver auditoria`; `/audit` mostra só eventos daquele usuário, com `Limpar filtros`.

### S2: Web - link na tela do papel (P1)

**Acceptance Criteria**

5. WHEN `/roles/$id` has loaded the role and the viewer holds `audit:read` THEN the web app SHALL show a link `Ver auditoria` whose target is `/audit?resource_type=role&resource_id=<id>`
6. IF the viewer does not hold `audit:read` THEN `/roles/$id` SHALL NOT show `Ver auditoria`, and the rest of the screen SHALL be unchanged
7. WHILE the role request is pending, and IF it ends in `404`, `403` or another error, THEN `/roles/$id` SHALL NOT show `Ver auditoria`

**Independent test:** como admin, criar um papel, abrir e clicar `Ver auditoria`; `/audit` mostra o `role.created` dele.

### S3: Ponta a ponta (P2)

**Acceptance Criteria**

8. WHEN an admin creates a user, opens `/users/<id>` and follows `Ver auditoria` THEN `/audit` SHALL list that user's `user.created` event and no event whose `Recurso` is another resource
9. WHEN an admin follows `Ver ações` on their own `/users/<id>` THEN `/audit` SHALL list only events whose `Quem` is that admin's email

**Independent test:** `task e2e -- audit-links.spec.ts`.

## Traceability

| ID | Slice | Criteria | Status |
| --- | --- | --- | --- |
| AUDL-01 | S1 | 1, 2, 3, 4 | Implementing |
| AUDL-02 | S2 | 5, 6, 7 | Implementing |
| AUDL-03 | S3 | 8, 9 | Implementing |

## Observable

| Surface | Decision | Landing |
| --- | --- | --- |
| screen `/users/$id` | loading, error, not found states | AC 4 - os estados existentes não mudam; os links ficam fora deles |
| screen `/users/$id` | unauthorised state | AC 3 (sem `audit:read`), AC 4 (`403` de `users:read`, estado existente) |
| screen `/users/$id` | empty state | n/a - a tela mostra um registro; não há lista |
| screen `/users/$id` | density and ordering | Assumptions `Posição` - abaixo do título |
| screen `/users/$id` | destructive action confirms | n/a - link só navega |
| screen `/roles/$id` | loading, error, not found states | AC 7 |
| screen `/roles/$id` | unauthorised state | AC 6, AC 7 |
| screen `/roles/$id` | empty state | n/a - a tela mostra um registro; não há lista |
| screen `/roles/$id` | density and ordering | Assumptions `Posição` - abaixo do título |
| screen `/roles/$id` | destructive action confirms | n/a - link só navega |
| screen `/audit` (destino) | estado com filtro vindo da URL | existing - audit AC 25, AC 26; provado de ponta a ponta por AC 8, AC 9 |
| document `AGENTS.md` | what the reader does next | n/a - nenhum comando ou regra nova |

## Flow

Reaproveita a tela `/audit` e seu `validateSearch` (exists), que já leva os filtros da URL para a API, e o
`can(me, ...)` de `lib/session` (exists) que as telas já usam para decidir o que mostrar. Nenhuma chamada nova à
API: o id vem do registro que a tela já carregou.

1. `/users/$id` -> `features/users/UserDetail` (exists) - depois de carregar o usuário, com `audit:read`, renderiza os dois links
2. `/roles/$id` -> `features/rbac/RoleDetail` (exists) - depois de carregar o papel, com `audit:read`, renderiza o link
3. clique -> `/audit` (exists) com `resource_type`/`resource_id` ou `actor_id` na URL -> `GET /api/v1/audit/events` com os mesmos filtros

## Relations

None - no stored-data shape change

## Surface

None - nothing consumed outside; os parâmetros de `/audit` são internos da web e já existem

## Landing

| One-way door | Literal shape | Alternative rejected |
| --- | --- | --- |
| None - links de navegação na web; tirar ou mover um link é uma edição de JSX, e nenhum contrato, schema ou dependência muda | n/a | n/a |

- Nothing else in this change is hard to reverse

## Impact

| Front | What changes |
| --- | --- |
| domain | nothing - nenhum termo novo; `Ver ações` é só rótulo de link |
| stored data | nothing to migrate |
| web | `UserDetail` e `RoleDetail` passam a depender de `audit:read` para um ramo novo; os testes existentes dessas telas usam um `me` com permissões explícitas sem `audit:read`, então não veem os links e não mudam |
| web | o filtro por URL de `/audit` (`resource_type`, `resource_id`, `actor_id`) ganha consumidores fora de `features/audit`; renomear um desses parâmetros passa a quebrar os links |
