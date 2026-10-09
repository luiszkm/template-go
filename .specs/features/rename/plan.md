# Rename

Sources:

- conversation (2026-10-09) - candidato restante do `STATE.md` Handoff; o usuário escolheu trocar módulo Go + nome exibido, e não aplicar o rename neste repositório
- `.specs/features/foundation/plan.md` `## Out of scope` - "Renomear o template (`task rename`) - só faz sentido depois que o esqueleto existir"
- `app/cmd/newslice` - precedente de gerador em Go chamado por uma task, com prova que copia o repositório e roda o gate na cópia

## Problem

Quem começa um projeto a partir deste template herda o módulo `github.com/luiszkm/template-go`, o pacote web
`template-go-web` e o nome `Template` na aba do navegador, no título da home e na OpenAPI. Trocar isso à mão
significa editar 84 arquivos Go, `go.mod`, `tools/go.mod`, `.golangci.yml` (que usa o módulo para ordenar
imports), `archtest` (que fixa o caminho do módulo), `package.json`, `package-lock.json` e três textos, e depois
regenerar a OpenAPI e o cliente TS. Um arquivo esquecido só aparece como erro de compilação, de lint ou de drift
no gate.

A fonte não traz números: o template ainda não foi usado para iniciar outro projeto.

Quando isto entra, `task rename MODULE=github.com/acme/foo NAME="Foo"` deixa a cópia com o módulo, o pacote web
e o nome novos, e `task check` passa nela sem nenhuma edição manual.

## Out of scope

| Excluded | Why |
| --- | --- |
| Aplicar o rename neste repositório | decisão do usuário: este repo continua sendo o template |
| Renomear banco/usuário Postgres `app`, binário `api`, serviço e volume do compose | nomes genéricos que não identificam o projeto; mexer neles toca `task dev`, e2e e CI sem ganho |
| Reescrever `.specs/` | é histórico das decisões do template; reescrever apagaria o que o texto citava |
| Trocar o remoto git, autor ou licença | operações do git e do repositório remoto, fora dos arquivos |
| Desfazer um rename | rodar de novo com os valores antigos refaz a troca, porque o comando lê os valores atuais |

## Assumptions

| Assumption | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Entradas | `task rename MODULE=<caminho do módulo> NAME=<nome exibido>`, as duas obrigatórias | o usuário escolheu módulo + nome exibido | y |
| De onde vêm os valores atuais | módulo da linha `module` de `app/go.mod`; nome do `<title>` de `web/index.html` | o comando roda de novo sobre um repo já renomeado sem precisar saber o nome original | y |
| Onde troca o módulo | todo arquivo de texto do repositório que contém o caminho atual, exceto `.git`, `node_modules`, `bin`, `dist`, `.task` e `.specs`; `tools/go.mod` vira `<módulo>/tools` pela mesma troca de prefixo | os 84 `.go`, os dois `go.mod`, `.golangci.yml` e `archtest` têm o caminho literal; uma troca literal por prefixo cobre todos | y |
| Onde troca o nome exibido | exatamente 3 lugares: `APITitle` (`"<NAME> API"`), `<title>` de `web/index.html` e o `<h1>` da home | são os únicos que mostram o nome; trocar a palavra `Template` no repo inteiro pegaria `text/template` e `TemplateURL` | y |
| Nome do pacote web | último segmento do módulo, em minúsculas, mais `-web` (`github.com/acme/AIGateway` -> `aigateway-web`), em `package.json` e `package-lock.json` | segue o padrão atual `template-go-web`; npm exige minúsculas | y |
| Validação | `MODULE` passa em `module.CheckPath` (golang.org/x/mod); `NAME` tem de 1 a 60 caracteres, sem `"`, `<`, `>`, `\`, `{`, `}` nem quebra de linha | o nome entra numa string Go, em HTML e em JSX; um caractere desses quebra a compilação ou injeta marcação | y |
| Falha | valida tudo e calcula todas as edições antes de escrever o primeiro arquivo; entrada inválida sai com 1 e não toca nada | uma troca pela metade deixa o repo sem compilar | y |
| Depois da troca | a task roda `task gen` para regenerar `app/openapi.json` e `web/src/api/schema.d.ts` | o título da OpenAPI muda e o gate falha por drift sem regenerar | y |
| Saída | uma linha `wrote <arquivo>` por arquivo alterado, como `newslice` | mesmo formato do outro gerador | y |
| Mesmos valores | rodar com o módulo e o nome atuais não altera nenhum arquivo e sai com 0 | repetir o comando é seguro | y |

**Open questions:** none - all resolved or logged above.

## Criteria

### S1: Comando `rename` (P1)

**Acceptance Criteria**

1. WHEN `rename --module <m> --name <n>` runs over a tree whose `app/go.mod` declares module `<old>` THEN the system SHALL replace every occurrence of `<old>` with `<m>` in every text file outside `.git`, `node_modules`, `bin`, `dist`, `.task` and `.specs`, and print `wrote <path>` for each file changed
2. WHEN the rename runs THEN `app/tools/go.mod` SHALL declare module `<m>/tools`
3. WHEN the rename runs THEN `APITitle` SHALL be `"<n> API"`, `web/index.html` `<title>` SHALL be `<n>`, and the home `<h1>` SHALL be `<n>`
4. WHEN the rename runs THEN the `name` in `web/package.json` and both `name` fields of `web/package-lock.json` SHALL be the last segment of `<m>` lowercased plus `-web`
5. The system SHALL NOT change any file under `.specs`, and SHALL NOT change the word `Template` anywhere other than the three places of AC 3
6. IF `<m>` fails `module.CheckPath` or `<n>` is empty, longer than 60 characters, or contains `"`, `<`, `>`, `\`, `{`, `}` or a line break THEN the system SHALL exit `1` with a message naming the invalid input and SHALL change no file
7. IF `--module` or `--name` is missing THEN the system SHALL exit `1` with the usage and SHALL change no file
8. WHEN the rename runs with the module and name already in place THEN the system SHALL change no file and exit `0`

**Independent test:** numa cópia do repo, `go run ./cmd/rename --module github.com/acme/foo --name Foo` e `git grep luiszkm/template-go -- ':!.specs'` não acha nada.

### S2: `task rename` deixa a cópia verde (P1)

**Acceptance Criteria**

9. WHEN `task rename MODULE=<m> NAME=<n>` runs THEN it SHALL run the rename and then `task gen`, and `app/openapi.json` SHALL carry `"title":"<n> API"`
10. WHEN `task rename MODULE=github.com/acme/foo NAME="Foo Bar"` runs on a copy of the repository THEN `task check` SHALL exit `0` in that copy
11. The `AGENTS.md` `## Workflow` table SHALL list `task rename MODULE=<m> NAME=<n>` with what it does

**Independent test:** a prova de S2 copia o repo, roda `task rename` e `task check` na cópia.

## Traceability

| ID | Slice | Criteria | Status |
| --- | --- | --- | --- |
| REN-01 | S1 | 1, 2, 3, 4, 5, 6, 7, 8 | Verified |
| REN-02 | S2 | 9, 10, 11 | Verified |

## Observable

| Surface | Decision | Landing |
| --- | --- | --- |
| command `task rename` / `cmd/rename` | output format and verbosity | AC 1 - `wrote <path>` por arquivo |
| command `task rename` / `cmd/rename` | every flag and its default | AC 7 - `--module` e `--name` obrigatórios, sem default; `--root` com default `.` como em `newslice` |
| command `task rename` / `cmd/rename` | exit codes | AC 6, AC 7, AC 8 - `0` sucesso ou nada a fazer, `1` entrada inválida ou falha |
| command `task rename` / `cmd/rename` | what it prints when it fails halfway | AC 6 - validação e cálculo antes da primeira escrita; uma falha de escrita no meio imprime o erro e os `wrote` já impressos dizem o que mudou |
| document `AGENTS.md` | what the reader does next | AC 11 |

## Flow

Reaproveita o padrão de `cmd/newslice` (flags, `--root`, saída `wrote`, chamado por uma task) e o `task gen`
existente para regenerar o que deriva do título. Nada no runtime da aplicação muda.

1. `task rename MODULE NAME` -> `app/cmd/rename` (new, door 1) - lê o módulo de `app/go.mod` e o nome de `web/index.html`, valida, calcula as edições, escreve
2. `task gen` (exists) - regenera `app/openapi.json` (título) e `web/src/api/schema.d.ts`
3. out: repo com módulo, pacote web e nome novos; `task check` (exists) passa

## Relations

None - no stored-data shape change

## Surface

None - nothing consumed outside; o comando é local ao repositório e a OpenAPI só muda o `info.title`

## Landing

| One-way door | Literal shape | Alternative rejected |
| --- | --- | --- |
| 1. dependência direta `golang.org/x/mod` | `golang.org/x/mod v0.41.0` passa de `// indirect` para direta em `app/go.mod`; `module.CheckPath(m)` valida `MODULE` | regex própria para caminho de módulo - reimplementa uma regra que o Go já define e erra nos casos de borda (maiúsculas, `.` inicial, elementos reservados) |
| 2. task e flags | `task rename MODULE=<m> NAME=<n>` -> `go run ./cmd/rename --module "<m>" --name "<n>"` e depois `task gen` | script shell/PowerShell - o repo roda em Windows e Unix, e os outros geradores já são Go |

- Nothing else in this change is hard to reverse

## Impact

| Front | What changes |
| --- | --- |
| domain | nothing - nenhum termo novo |
| stored data | nothing to migrate |
| tooling | novo `app/cmd/rename` e task `rename`; `app/go.mod` promove `golang.org/x/mod` a direta |
| tests | `copyRepo`/`runIn` saem de `cmd/newslice/repo_test.go` para `internal/platform/testkit/repocopy.go` (`CopyRepo`, `RunIn`), porque `cmd/rename` é o segundo consumidor (AGENTS regra 3); o `newslice` passa a usá-los sem mudar o que testa |
| tooling | `.golangci.yml` exclui `gosec` de `cmd/rename/` (como dos outros geradores) e de `internal/platform/testkit/repocopy.go` (copia arquivos e roda processos só em teste) |
| tests | a prova de AC 10 copia o repo e roda `task check` na cópia, como `TestGenerated_TaskCheckPasses` do `newslice` - soma alguns minutos ao `go test` do gate |
| docs | `AGENTS.md` ganha a linha `task rename` na tabela de comandos |
