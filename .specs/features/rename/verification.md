# Rename verification

**Verdict**: FAIL
**Profile**: standard
**Diff range**: 738c1e4..be638d1
**Round**: 1 - full
**Verifier**: independent sub-agent (author != verifier) - sub-agente independente, sem contexto do build

As 11 checks passam em `be638d1` com evidência localizada, e o refactor de `cmd/newslice` é um movimento literal
para `testkit` que não toca nenhuma asserção. O veredito é FAIL por três achados do recompute: um mutante não
equivalente sobrevive na fronteira do caminho do módulo (o ramo `$` da regex - caminho no fim do arquivo), o
caractere `\r` que o código recusa no nome não tem prova, e a linha `Test policy` do ponto de entrada não tem
prova para nenhum caminho de erro que não seja de validação (falha ao ler `app/go.mod`, `<title>` ou
`package.json`, falha de escrita no meio).

## Binding sources

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| nenhuma fonte binding - o plano só cita conversa, `foundation/plan.md` e `cmd/newslice` como precedente; perfil `standard`, passo 1 não se aplica | n/a | - | - |

## Checks

verified at be638d1 - provas executadas em duas invocações:

- unitárias (C1-C8, C11): `go -C app test -count=1 ./cmd/rename -run '^(TestRename_ReplacesModuleEverywhere|TestRename_ToolsModule|TestRename_DisplayName|TestRename_WebPackageName|TestRename_LeavesExcludedFiles|TestRename_RejectsInvalidInput|TestMain_FailuresExit1|TestRename_SameValuesChangesNothing|TestTaskfileAndDocs_DeclareRename)$' -v` - exit 0, `ok github.com/luiszkm/template-go/cmd/rename 7.738s`; os 9 testes aparecem individualmente com `--- PASS`, incluindo os 8 subtestes de diretório de C5, os 11 de recusa e os 2 de aceite de C6 e os 3 de C7.
- cópia do repositório (C9, C10): `go -C app test -count=1 -timeout=40m ./cmd/rename -run '^(TestRenamed_RepoCarriesNewNames|TestRenamed_TaskCheckPasses)$' -v` - exit 0, `ok github.com/luiszkm/template-go/cmd/rename 627.632s`; `--- PASS: TestRenamed_RepoCarriesNewNames (63.18s)` e `--- PASS: TestRenamed_TaskCheckPasses (561.74s)` individualmente, Docker disponível.

Os 11 nomes existem em `app/cmd/rename/rename_test.go` e `app/cmd/rename/repo_test.go`, ambos criados no diff
`738c1e4..HEAD`; nenhuma prova resolve para um teste antigo.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | troca o módulo em `.go`, `go.mod`, `.golangci.yml` e arquivo sob `web/`; não troca `example.com/old/appx`; uma linha `wrote` por arquivo alterado | lote unitário, `--- PASS: TestRename_ReplacesModuleEverywhere` | `app/cmd/rename/rename_test.go:92` - `require.Equal(t, "package x\n\nimport _ \"github.com/acme/foo/internal/y\"\n", read(t, root, "app/internal/x/x.go"))`; `:95` - `require.Equal(t, fixtureFiles["other.txt"], read(t, root, "other.txt"))`; `:99` - `require.Equal(t, []string{"wrote app/.golangci.yml", ... "wrote web/package.json"}, lines)` | PASS |
| C2 | `app/tools/go.mod` declara `module github.com/acme/foo/tools` | lote unitário, `--- PASS: TestRename_ToolsModule` | `app/cmd/rename/rename_test.go:114` - `require.Equal(t, "module github.com/acme/foo/tools\n\ngo 1.26\n", read(t, root, "app/tools/go.mod"))` | PASS |
| C3 | `"Foo Bar API"`, `<title>Foo Bar</title>`, `>Foo Bar</h1>` | lote unitário, `--- PASS: TestRename_DisplayName` | `app/cmd/rename/rename_test.go:121` - `require.Contains(t, read(t, root, "app/internal/platform/httpx/api.go"), "\tAPITitle   = \"Foo Bar API\"\n")`; `:122` - `require.Contains(..., "<title>Foo Bar</title>")`; `:123` - `require.Equal(t, "<h1 className=\"text-2xl font-semibold\">Foo Bar</h1>\n", ...)` | PASS |
| C4 | `aigateway-web` para `github.com/acme/AIGateway`, `foo-web` para `.../foo/v2`, nos 3 campos | lote unitário, `--- PASS: TestRename_WebPackageName/aigateway-web` e `/foo-web` | `app/cmd/rename/rename_test.go:135` - `require.Equal(t, "{\n  \"name\": \""+c.pkg+"\",\n  \"private\": true\n}\n", ...)`; `:137` - `require.Equal(t, 2, strings.Count(lock, "\"name\": \""+c.pkg+"\""))` | PASS |
| C5 | 8 diretórios excluídos, binário com NUL, `text/template` e `Template` intactos | lote unitário, 9 subtestes `--- PASS: TestRename_LeavesExcludedFiles/...` | `app/cmd/rename/rename_test.go:155` - `require.Equal(t, "module "+oldModule+"\n", read(t, root, rel))` (um subteste por diretório, lista em `:144`); `:165` - `require.Equal(t, blob, read(t, root, "app/blob.bin"))`; `:166` - `require.Equal(t, fixtureFiles["app/internal/render/render.go"], ...)` | PASS |
| C6 | 11 entradas recusadas com mensagem `module`/`name` e árvore intacta; 1 e 60 caracteres aceitos | lote unitário, 13 subtestes `--- PASS: TestRename_RejectsInvalidInput/...` | `app/cmd/rename/rename_test.go:189` - `require.ErrorContains(t, err, c.field)`; `:190` - `require.Equal(t, before, treeHash(t, root))`; `:199` - `require.Contains(t, read(t, root, "web/index.html"), "<title>"+name+"</title>")` | PASS |
| C7 | exit `1`, `usage:` no stderr e árvore intacta sem `--module`/`--name`; exit `1` com entrada inválida | lote unitário, 3 subtestes `--- PASS: TestMain_FailuresExit1/...` | `app/cmd/rename/rename_test.go:218` - `require.Equal(t, 1, code)`; `:219` - `require.Contains(t, errb, c.want)` com `"usage:"` em `:210-211`; `:221` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C8 | mesmos valores: hash igual, sem `wrote`, sem erro | lote unitário, `--- PASS: TestRename_SameValuesChangesNothing` | `app/cmd/rename/rename_test.go:231` - `require.NotContains(t, out, "wrote")`; `:232` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C9 | na cópia, `task rename` sai 0, nenhum arquivo com `github.com/luiszkm/template-go`, `"title":"Foo Bar API"`, `"name": "foo-web"` | lote cópia, `--- PASS: TestRenamed_RepoCarriesNewNames (63.18s)` | `app/cmd/rename/repo_test.go:36` - `require.NoError(t, renamedErr, renamedOut)`; `:94` - `require.Empty(t, stale)`; `:98` - `require.Contains(t, string(spec), `"title":"Foo Bar API"`)`; `:102` - `require.Contains(t, string(pkg), `"name": "foo-web"`)` | PASS |
| C10 | `task check` sai 0 na cópia renomeada | lote cópia, `--- PASS: TestRenamed_TaskCheckPasses (561.74s)` | `app/cmd/rename/repo_test.go:108` - `require.NoError(t, err, out)` sobre `testkit.RunIn(root, ..., "task", "check")` em `:107` | PASS |
| C11 | linha `task rename MODULE=<m> NAME=<n>` no `AGENTS.md`; task exige `MODULE`/`NAME`, roda `go run ./cmd/rename` e depois `gen` | lote unitário, `--- PASS: TestTaskfileAndDocs_DeclareRename` | `app/cmd/rename/repo_test.go:117` - `require.Regexp(t, "(?m)^\\| `task rename MODULE=<m> NAME=<n>` \\| .+ \\|$", ...)`; `:124` - `require.Regexp(t, `vars: \[MODULE, NAME\]`, body)`; `:128` - `require.Greater(t, gen, run)` | PASS |

Refactor de `testkit`: `git diff 738c1e4..HEAD -- app/cmd/newslice` só remove `skipDirs`, `copyRepo`, `linkDir` e
`runIn` e troca as chamadas por `testkit.CopyRepo`/`testkit.RunIn`; o corpo movido para
`app/internal/platform/testkit/repocopy.go:13-65` é idêntico (só `skipDirs` virou `repoCopySkipDirs`), e nenhum
`require` de `app/cmd/newslice/repo_test.go` foi alterado ou removido. A extração cumpre a regra 3 do `AGENTS.md`
(segundo consumidor: `cmd/rename`).

## Coverage

Membros recomputados de `app/cmd/rename/main.go` em `be638d1`.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| lugares do módulo (5) | `planEdits` aplica `modulePattern` a todo arquivo de texto (`main.go:155`, `:186`) | `.go` C1 `rename_test.go:92` · `app/go.mod` C1 `:91` · `app/tools/go.mod` C2 `:114` · `.golangci.yml` C1 `:93` · `web/notes.txt` C1 `:94` | - |
| fronteira do caminho do módulo (3, o artefato declara 2) | regex `QuoteMeta(old) + ([^A-Za-z0-9._~-]\|$)` em `main.go:155` | seguido de caractere fora da classe (`/`, `\n`, espaço) C1 `:92-94`, C2 `:114` · seguido de caractere da classe (`appx`) C1 `:95` | seguido de fim de arquivo (`$`) - nenhum fixture termina no caminho do módulo; mutante F7 sobreviveu |
| lugares do nome exibido (3) | `nameEdits` em `main.go:157-159` | `APITitle` C3 `:121` · `<title>` C3 `:122` · `<h1>` C3 `:123` | - |
| campos do pacote web (3) | `nameEdits` em `main.go:160-161` | `package.json` `name` C4 `:135` · lock raiz e `packages[""]` C4 `:137` (`Count == 2`) | - |
| derivação do pacote web (2) | `webPackage` em `main.go:99-102` (`SplitPathVersion` + `ToLower(Base)`) | segmento simples com maiúsculas C4 `aigateway-web` · sufixo `/v2` C4 `foo-web` | - |
| diretórios excluídos (8) | `skippedDirs` em `main.go:32-35` | `.specs` · `.git` · `node_modules` · `bin` · `dist` · `.task` · `test-results` · `playwright-report` - C5 `:144`/`:155`, um subteste cada | - |
| outros não alvos (3) | NUL em `main.go:178`; palavra `Template` fora de `nameEdits` | binário C5 `:165` · `text/template` C5 `:166` · `Template` em `other.txt` C5 `:167` | - |
| caracteres proibidos no nome (8, o artefato declara 7) | `forbiddenNameChars = "\"<>\\{}\r\n"` em `main.go:36` | `"` · `<` · `>` · `\` · `{` · `}` · `\n` - C6 `:177-183`, `:189` | `\r` - nenhum teste o usa (`rg '\\r' app/cmd/rename` acha só `main.go:36`) |
| limites do nome (4) | `checkName` em `main.go:90` (contagem em runas) | vazio C6 `:175` · 61 C6 `:176` · 1 e 60 aceitos C6 `:194-199` | - |
| recusa de módulo (2) | `module.CheckPath` em `main.go:72` | `Not A Path` C6 `:173` · `.foo` C6 `:174` | - |
| saídas de `run` (5) | `main.go:50-62` | `--module` ausente C7 `:210` · `--name` ausente C7 `:211` · `Rename` com erro de validação C7 `:212` · sucesso exit 0 C1 `:89` e C9 `repo_test.go:36` · sem alterações exit 0 C8 `:230` | - |
| caminhos de erro sem validação (6) | `readCurrent` `main.go:111-137` (sem `app/go.mod`, sem linha `module`, sem `<title>`, `package.json` inválido) e `applyEdits` `main.go:207-212` (stat/escrita falham após `wrote` parciais) | - | os 6 - nenhum teste os exercita; o plano (`Observable`, "what it prints when it fails halfway") nomeia o último |
| Landing doors (2) | `plan.md` `## Landing` | `golang.org/x/mod` `CheckPath` C6 `:174` (`.foo`) e `app/go.mod` sem `// indirect` no diff · task e flags C11 `repo_test.go:124-128` | - |

## Test policy rows

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `app/cmd/rename/main.go` `Rename`/`planEdits`/`checkName`/`webPackage` | boundary C9, C10 (`task rename` + `task check` na cópia) · own layer C1-C8 | no - fronteira presente, mas a tabela de decisão da própria camada tem linhas sem caso afirmado: fim de arquivo na fronteira do módulo (F7 sobreviveu) e `\r` em `forbiddenNameChars` |
| Decides, not reached across a boundary | nenhum arquivo do diff | - | n/a - nada classificado |
| Entry point that decides nothing | `app/cmd/rename/main.go` `main`/`run` | C7 · C1 (sucesso via `run`) | no - entrada aceita e cada entrada recusada provadas; nenhum caminho de erro não-validação (`readCurrent`, `applyEdits`) tem prova, e `AGENTS.md` pede "each error path" |
| Instrumentation, pass-throughs | `app/internal/platform/testkit/repocopy.go`, task `rename` do `Taskfile.yml` | coberto pelo consumidor: C9/C10 e `newslice` `TestGenerated_TaskCheckPasses` · C11 | yes |

## Faults injected

Isolamento: `git worktree add --detach <scratchpad>/wt HEAD`; cada mutação aplicada com `perl`, prova mais estreita
rodada no worktree, arquivo restaurado com `git checkout`; worktree removido com `git worktree remove --force` e
`git worktree prune`. `git status --porcelain` da árvore real antes e depois: idêntico (só
`.claude/skills/auth-security/` e `.cursor/skills/auth-security/`). São 9 mutações, acima do teto de cinco do
`verify.md`, porque o brief do orquestrador pediu uma por superfície listada; F7 foi acrescentada para testar o
membro de fronteira que o recompute achou sem prova. C9/C10 não receberam falhas (custo de 12-18 min por rodada);
F9 cobre a superfície da task pelo C11.

| Mutation | Location | Killed |
| --- | --- | --- |
| F1 - `dist` removido de `skippedDirs` | `app/cmd/rename/main.go:33` | yes - `TestRename_LeavesExcludedFiles/dist` |
| F2 - grupo de fronteira trocado por `()` (troca por prefixo puro) | `app/cmd/rename/main.go:155` | yes - `TestRename_ReplacesModuleEverywhere` |
| F3 - `{` removido de `forbiddenNameChars` | `app/cmd/rename/main.go:36` | yes - `TestRename_RejectsInvalidInput/name_with_open_brace` |
| F4 - `webPackage` usa `path.Base(modulePath)` sem tirar `/v2` | `app/cmd/rename/main.go:101` | yes - `TestRename_WebPackageName/foo-web` |
| F5 - entrada de `apiTitleFile` removida de `nameEdits` | `app/cmd/rename/main.go:157` | yes - `TestRename_DisplayName` |
| F6 - `readCurrent`+`planEdits`+`applyEdits` antes de `CheckPath`/`checkName` | `app/cmd/rename/main.go:71` | yes - `TestRename_RejectsInvalidInput` (4 subtestes) |
| F7 - `\|$` removido da fronteira (`([^A-Za-z0-9._~-])`) | `app/cmd/rename/main.go:155` | no - survived `TestRename_ReplacesModuleEverywhere` e `TestRename_ToolsModule`; não equivalente: uma sonda temporária com `app/eof.txt = "see "+oldModule` passa no original e falha no mutante |
| F8 - `if text != string(raw)` trocado por `if true` | `app/cmd/rename/main.go:190` | yes - `TestRename_SameValuesChangesNothing` |
| F9 - `task: gen` antes de `go run ./cmd/rename` na task | `Taskfile.yml:145-146` | yes - `TestTaskfileAndDocs_DeclareRename` |

## Swept existing

Nenhuma linha de `## Swept` resolve para uma restrição existente: validation, failure modes, idempotency,
dependency failure e observability apontam para checks novas (C6, C7, C8, C10, C1), e as demais são `n/a`
aprovadas. Nada a reler no código.

## Gate

`go -C app test -count=1 ./cmd/rename -run '<9 provas unitárias>' -v` - 9 passed, 0 failed (34 subtestes).
`go -C app test -count=1 -timeout=40m ./cmd/rename -run '^(TestRenamed_RepoCarriesNewNames|TestRenamed_TaskCheckPasses)$' -v` - 2 passed, 0 failed (627.632s).

## Ranked gaps

1. Mutante sobrevivente F7: a fronteira do caminho do módulo no fim do arquivo (`|$`) não tem caso afirmado; um arquivo terminando no caminho do módulo sem quebra de linha deixaria de ser renomeado sem nenhuma prova falhar - C1/C2 - `app/cmd/rename/main.go:155`
2. Membro sem prova: `\r` em `forbiddenNameChars`; a check C6 lista 7 caracteres e o código recusa 8 - C6 - `app/cmd/rename/main.go:36`
3. Linha `Entry point` do `Test policy` não cumprida: nenhum caminho de erro não-validação tem prova (falta `app/go.mod`, falta linha `module`, falta `<title>`, `package.json` inválido, falha de escrita após `wrote` parciais - este último nomeado no `Observable` do plano) - C7 - `app/cmd/rename/main.go:111-137`, `:207-212`
4. Gap de precisão (não bloqueia sozinho): C6 afirma "61 caracteres" só com ASCII, então a contagem em runas de `main.go:90` (`len([]rune(name))`) passaria também com `len(name)`; a check não fixa se o limite é em caracteres ou bytes - C6 - `app/cmd/rename/rename_test.go:176`
5. Observação (não bloqueia): na fronteira, C9 só afirma o título da OpenAPI e o `package.json`; `<title>` e `<h1>` do repositório real só são provados sobre o fixture (C3), que hoje reproduz o formato real (`web/src/routes/_authed/index.tsx:11`, `web/index.html:6`) - C9 - `app/cmd/rename/repo_test.go:96-102`
