# Rename checks

Profile: standard
Plan: `.specs/features/rename/plan.md`

11 checks in 2 slices · 2 one-way doors · 0 open, of which 0 block

## Checks

Os testes unitários usam uma árvore temporária com o módulo fictício `example.com/old/app`, o nome `Old Name` e o
pacote `app-web`, nunca o módulo real: depois de um rename real, um literal do módulo real nos testes seria
reescrito e o teste viraria um no-op.

### S1 - comando `rename` · 4 files · 20 KB · ~5k

**C1** - Numa árvore com `app/go.mod` declarando `example.com/old/app`, `rename --module github.com/acme/foo --name Foo` troca o caminho em um `.go` (import), em `app/go.mod`, em `.golangci.yml` e num arquivo sob `web/`; não troca `example.com/old/appx` (outro caminho com o mesmo prefixo); imprime exatamente uma linha `wrote <path>` por arquivo alterado e nenhuma para os inalterados (REN-01, AC 1) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_ReplacesModuleEverywhere$' -v`

**C2** - Depois do rename, `app/tools/go.mod` declara `module github.com/acme/foo/tools` (REN-01, AC 2) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_ToolsModule$' -v`

**C3** - Com `--name "Foo Bar"`, `app/internal/platform/httpx/api.go` tem `"Foo Bar API"`, `web/index.html` tem `<title>Foo Bar</title>` e `web/src/routes/_authed/index.tsx` tem `>Foo Bar</h1>` (REN-01, AC 3) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_DisplayName$' -v`

**C4** - O `name` de `web/package.json` e os dois `name` de `web/package-lock.json` viram `aigateway-web` para `github.com/acme/AIGateway` e `foo-web` para `github.com/acme/foo/v2` - um caso afirmado por linha (REN-01, AC 4) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_WebPackageName$' -v`

**C5** - Arquivos com o módulo antigo sob `.specs`, `.git`, `node_modules`, `bin`, `dist`, `.task`, `test-results` e `playwright-report` ficam byte a byte iguais (um caso por diretório); um arquivo binário (com byte NUL) contendo o módulo fica igual; `text/template` e a palavra `Template` fora dos três lugares do C3 ficam iguais (REN-01, AC 5) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_LeavesExcludedFiles$' -v`

**C6** - Cada entrada inválida devolve erro cuja mensagem contém `module` ou `name` conforme o caso e deixa o hash da árvore igual: módulo `Not A Path`, módulo `github.com/acme/.foo`, nome vazio, nome de 61 caracteres, e nome contendo cada um de `"`, `<`, `>`, `\`, `{`, `}`, `\n`; nomes de 1 e de 60 caracteres são aceitos (REN-01, AC 6) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_RejectsInvalidInput$' -v`

**C7** - O binário sai com `1`, imprime `usage:` no stderr e não altera a árvore quando falta `--module` e quando falta `--name`; com entrada inválida também sai com `1` (REN-01, AC 6, AC 7) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestMain_FailuresExit1$' -v`

**C8** - Rodar com o módulo e o nome que a árvore já tem não altera nenhum arquivo (hash igual), não imprime `wrote` e retorna sem erro (REN-01, AC 8) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRename_SameValuesChangesNothing$' -v`

### S2 - `task rename` deixa a cópia verde · 3 files · 12 KB · ~3k

**C9** - Numa cópia do repositório, `task rename MODULE=github.com/acme/foo NAME="Foo Bar"` sai com `0`; nenhum arquivo copiado contém `github.com/luiszkm/template-go`; `app/openapi.json` contém `"title":"Foo Bar API"`; `web/package.json` tem `"name": "foo-web"` (REN-02, AC 9) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRenamed_RepoCarriesNewNames$' -v -timeout=30m`

**C10** - Na mesma cópia renomeada, `task check` sai com `0` (REN-02, AC 10) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestRenamed_TaskCheckPasses$' -v -timeout=30m`

**C11** - `AGENTS.md` tem uma linha da tabela de `## Workflow` começando com `` | `task rename MODULE=<m> NAME=<n>` | ``, e a task `rename` do `Taskfile.yml` exige `MODULE` e `NAME`, roda `go run ./cmd/rename` e depois a task `gen` (REN-02, AC 9, AC 11, door 2) `[done]`
Proof: `go -C app test -count=1 ./cmd/rename -run '^TestTaskfileAndDocs_DeclareRename$' -v`

## Coverage

| Set (size) | Member -> proof | Unproven |
| --- | --- | --- |
| lugares do módulo (5) | `.go` C1 · `app/go.mod` C1 · `app/tools/go.mod` C2 · `.golangci.yml` C1 · arquivo sob `web/` C1 | - |
| lugares do nome exibido (3) | `APITitle` C3 · `<title>` C3 · `<h1>` da home C3 | - |
| derivação do pacote web (2) | segmento simples C4 · sufixo de versão `/v2` C4 | - |
| arquivos do pacote web (3 campos) | `package.json` `name` C4 · lock raiz C4 · lock `packages[""]` C4 | - |
| diretórios excluídos (8) | `.specs` C5 · `.git` C5 · `node_modules` C5 · `bin` C5 · `dist` C5 · `.task` C5 · `test-results` C5 · `playwright-report` C5 | - |
| outros não alvos (3) | binário com NUL C5 · `text/template` C5 · `Template` fora dos três lugares C5 | - |
| fronteira do caminho do módulo (2) | caminho exato e com `/` seguinte C1, C2 · mesmo prefixo com mais caracteres C1 | - |
| entradas inválidas (11) | módulo `Not A Path` C6 · módulo `.foo` C6 · nome vazio C6 · 61 caracteres C6 · `"` C6 · `<` C6 · `>` C6 · `\` C6 · `{` C6 · `}` C6 · `\n` C6 | - |
| limites aceitos do nome (2) | 1 caractere C6 · 60 caracteres C6 | - |
| saídas do binário (4) | sucesso C1 (por `Rename`) e C9 (pela task) · falta `--module` C7 · falta `--name` C7 · entrada inválida C7 | - |
| Landing doors (2) | `golang.org/x/mod` `CheckPath` C6 (módulo `.foo` só é recusado pela regra do Go) · task e flags C11 | - |

- O plano não tem `Surface` nem `Relations`; nada mais a juntar
- C9 e C10 rodam o comando real numa cópia do repositório real; C1-C8 afirmam cada decisão na própria camada

## Test policy

O `AGENTS.md` já responde às duas perguntas (`## Test policy`); estas são as linhas dele, aplicadas.

| Code | Required proofs | Coverage expectation |
| --- | --- | --- |
| Decides, reached across a boundary | one at the boundary **and** one at its own layer | the contract at the boundary; one asserted case per row of the decision table at its own layer |
| Decides, not reached across a boundary | one at its own layer | one asserted case per row of the decision table |
| Entry point that decides nothing | one at the boundary | accepted input, each rejected input, each error path |
| Instrumentation, pass-throughs | none of its own | covered by its consumer's proof |

Evidence (planned code, by shape):

- `cmd/rename` `Rename`: validação (11 recusas, 2 limites), filtro de diretório (8), filtro de binário, fronteira do caminho, derivação do pacote (2), comparação antes de escrever -> decides, reached across a boundary (a task): own layer C1-C8, boundary C9-C10
- `cmd/rename` `main`/`run`: flags ausentes e código de saída -> entry point: C7
- closest analogue: `cmd/newslice`, mesmo shape (gerador chamado por task), provado na própria camada (`newslice_test.go`) e na fronteira por cópia do repo (`repo_test.go`, `TestGenerated_TaskCheckPasses`)

## Swept

- validation: C6
- failure modes: C6, C7 - validação antes da primeira escrita; falha não altera a árvore
- idempotency: C8
- authorization: n/a - comando local de desenvolvimento; não há usuário nem permissão
- concurrency: n/a - roda uma vez, manualmente, numa árvore de trabalho
- data lifecycle: n/a - nenhum dado persistido; só arquivos do repositório
- dependency failure: C10 - `task gen` e o gate rodam de verdade na cópia
- state transitions: n/a - o comando não tem estados; rodar de novo é C8
- observability: C1 - uma linha `wrote` por arquivo alterado

## Handoff

Intended split, with the arithmetic, written before any code:

- S1-S2 ≈ 8k de leitura (`newslice` como modelo, `Taskfile.yml`, `AGENTS.md`, os 3 arquivos do nome) -> um builder, sem handoff

O teste de cópia de `rename` pula quando `NEWSLICE_INNER_CHECK=1` ou `RENAME_INNER_CHECK=1`, e o `task check` interno roda com os dois, para que nenhuma cópia dispare outra cópia.
