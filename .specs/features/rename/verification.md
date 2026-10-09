# Rename verification

**Verdict**: PASS
**Profile**: standard
**Diff range**: 738c1e4..b100ffe
**Round**: 2 - scoped
**Verifier**: independent sub-agent (author != verifier) - sub-agente independente, sem contexto do build nem da correção

As quatro lacunas da rodada 1 (FAIL em `be638d1`) estão fechadas em `b100ffe`: o mutante F7 agora morre em
`TestRename_ModuleAtEndOfFile`, o `\r` e a contagem em caracteres têm caso afirmado e matam seus mutantes, e cada
caminho de erro fora da validação que a rodada 1 enumerou tem prova pelo ponto de entrada `run`. A correção não toca
código de produção: `git diff be638d1..HEAD -- app web Taskfile.yml AGENTS.md` mostra só
`app/cmd/rename/rename_test.go` (+77 linhas, apenas acréscimos ao fim do arquivo), além de `checks.md` e do relatório
da rodada 1. As 7 falhas injetadas nesta rodada foram mortas. Restam observações que não bloqueiam (ver `Ranked
gaps`).

## Binding sources

carried from be638d1 - a correção não toca a interface nem o plano.

| Source | Opened | Contradiction | Uncovered |
| --- | --- | --- | --- |
| nenhuma fonte binding - o plano só cita conversa, `foundation/plan.md` e `cmd/newslice` como precedente; perfil `standard`, passo 1 não se aplica | n/a | - | - |

## Checks

verified at b100ffe (C1-C8, C11-C14) · C9/C10 carried from be638d1.

Provas unitárias executadas numa invocação: `go -C app test -count=1 ./cmd/rename -run '^(TestRename_|TestMain_|TestTaskfileAndDocs_)' -v`
- exit 0, `ok github.com/luiszkm/template-go/cmd/rename 4.049s`. Os 12 testes aparecem individualmente com
`--- PASS` (41 subtestes), entre eles `TestRename_ModuleAtEndOfFile`, os 3 subtestes de
`TestRename_NameRunesAndCarriageReturn` e os 5 de `TestRename_ErrorPaths`. Os três testes novos existem só em
`app/cmd/rename/rename_test.go:235`, `:243`, `:265`, todos acrescentados no diff `be638d1..b100ffe`; nenhuma prova
resolve para teste antigo.

C9/C10 carregados da rodada 1, como o procedimento escopado permite quando a correção não pode tê-los afetado: o
código de produção, `repo_test.go`, `Taskfile.yml` e `AGENTS.md` estão idênticos desde `be638d1`. O único efeito
possível da correção sobre C10 seria o `task check` da cópia reprovar o arquivo de teste novo; isso foi coberto à
parte em `b100ffe`: `gofmt -l app/cmd/rename` vazio, `go -C app vet ./cmd/rename` ok,
`go tool -modfile=tools/go.mod golangci-lint run ./cmd/rename/...` com `0 issues`, e o diff não tem comentário
(regra 8 do `AGENTS.md`).

As citações de C1-C11 continuam válidas: o diff só acrescenta linhas depois da linha 233 de `rename_test.go`, e
`repo_test.go` não mudou.

| Check | Claim | Proof run | Evidence | Result |
| --- | --- | --- | --- | --- |
| C1 | troca o módulo em `.go`, `go.mod`, `.golangci.yml` e arquivo sob `web/`; não troca `example.com/old/appx`; uma linha `wrote` por arquivo alterado | lote unitário em b100ffe, `--- PASS: TestRename_ReplacesModuleEverywhere` | `app/cmd/rename/rename_test.go:92` - `require.Equal(t, "package x\n\nimport _ \"github.com/acme/foo/internal/y\"\n", read(t, root, "app/internal/x/x.go"))`; `:95` - `require.Equal(t, fixtureFiles["other.txt"], read(t, root, "other.txt"))`; `:99` - `require.Equal(t, []string{"wrote app/.golangci.yml", ...}, lines)` | PASS |
| C2 | `app/tools/go.mod` declara `module github.com/acme/foo/tools` | lote unitário em b100ffe, `--- PASS: TestRename_ToolsModule` | `app/cmd/rename/rename_test.go:114` - `require.Equal(t, "module github.com/acme/foo/tools\n\ngo 1.26\n", read(t, root, "app/tools/go.mod"))` | PASS |
| C3 | `"Foo Bar API"`, `<title>Foo Bar</title>`, `>Foo Bar</h1>` | lote unitário em b100ffe, `--- PASS: TestRename_DisplayName` | `app/cmd/rename/rename_test.go:121` - `require.Contains(..., "\tAPITitle   = \"Foo Bar API\"\n")`; `:122` - `require.Contains(..., "<title>Foo Bar</title>")`; `:123` - `require.Equal(t, "<h1 className=\"text-2xl font-semibold\">Foo Bar</h1>\n", ...)` | PASS |
| C4 | `aigateway-web` e `foo-web` nos 3 campos | lote unitário em b100ffe, `--- PASS: TestRename_WebPackageName/aigateway-web` e `/foo-web` | `app/cmd/rename/rename_test.go:135` - `require.Equal(t, "{\n  \"name\": \""+c.pkg+"\",\n  \"private\": true\n}\n", ...)`; `:137` - `require.Equal(t, 2, strings.Count(lock, "\"name\": \""+c.pkg+"\""))` | PASS |
| C5 | 8 diretórios excluídos, binário com NUL, `text/template` e `Template` intactos | lote unitário em b100ffe, 9 subtestes `--- PASS: TestRename_LeavesExcludedFiles/...` | `app/cmd/rename/rename_test.go:155` - `require.Equal(t, "module "+oldModule+"\n", read(t, root, rel))`; `:165` - `require.Equal(t, blob, read(t, root, "app/blob.bin"))`; `:166` - `require.Equal(t, fixtureFiles["app/internal/render/render.go"], ...)` | PASS |
| C6 | 11 entradas recusadas com mensagem `module`/`name` e árvore intacta; 1 e 60 caracteres aceitos | lote unitário em b100ffe, 13 subtestes `--- PASS: TestRename_RejectsInvalidInput/...` | `app/cmd/rename/rename_test.go:189` - `require.ErrorContains(t, err, c.field)`; `:190` - `require.Equal(t, before, treeHash(t, root))`; `:199` - `require.Contains(t, read(t, root, "web/index.html"), "<title>"+name+"</title>")` | PASS |
| C7 | exit `1`, `usage:` no stderr e árvore intacta sem `--module`/`--name`; exit `1` com entrada inválida | lote unitário em b100ffe, 3 subtestes `--- PASS: TestMain_FailuresExit1/...` | `app/cmd/rename/rename_test.go:218` - `require.Equal(t, 1, code)`; `:219` - `require.Contains(t, errb, c.want)`; `:221` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C8 | mesmos valores: hash igual, sem `wrote`, sem erro | lote unitário em b100ffe, `--- PASS: TestRename_SameValuesChangesNothing` | `app/cmd/rename/rename_test.go:231` - `require.NotContains(t, out, "wrote")`; `:232` - `require.Equal(t, before, treeHash(t, root))` | PASS |
| C9 | na cópia, `task rename` sai 0, nenhum arquivo com `github.com/luiszkm/template-go`, `"title":"Foo Bar API"`, `"name": "foo-web"` | carried from be638d1 - `--- PASS: TestRenamed_RepoCarriesNewNames (63.18s)` | `app/cmd/rename/repo_test.go:36` - `require.NoError(t, renamedErr, renamedOut)`; `:94` - `require.Empty(t, stale)`; `:98` - `require.Contains(t, string(spec), `"title":"Foo Bar API"`)`; `:102` - `require.Contains(t, string(pkg), `"name": "foo-web"`)` | PASS |
| C10 | `task check` sai 0 na cópia renomeada | carried from be638d1 - `--- PASS: TestRenamed_TaskCheckPasses (561.74s)`; arquivo novo checado à parte em b100ffe (gofmt, vet, golangci-lint) | `app/cmd/rename/repo_test.go:108` - `require.NoError(t, err, out)` | PASS |
| C11 | linha `task rename MODULE=<m> NAME=<n>` no `AGENTS.md`; task exige `MODULE`/`NAME`, roda `go run ./cmd/rename` e depois `gen` | lote unitário em b100ffe, `--- PASS: TestTaskfileAndDocs_DeclareRename` | `app/cmd/rename/repo_test.go:117` - `require.Regexp` ancorado numa linha de tabela que começa com `task rename MODULE=<m> NAME=<n>`; `:124` - `require.Regexp(t, `vars: \[MODULE, NAME\]`, body)`; `:128` - `require.Greater(t, gen, run)` | PASS |
| C12 | arquivo terminando exatamente no caminho do módulo vira `see github.com/acme/foo` | lote unitário em b100ffe, `--- PASS: TestRename_ModuleAtEndOfFile` | `app/cmd/rename/rename_test.go:237` - `write(t, root, "web/tail.txt", "see "+oldModule)` (precondição nomeada na claim: sem quebra de linha); `:240` - `require.Equal(t, "see github.com/acme/foo", read(t, root, "web/tail.txt"))` | PASS |
| C13 | nome com `\r` recusado com `name` e árvore igual; 60 `ç` aceitos, 61 recusados | lote unitário em b100ffe, `--- PASS: TestRename_NameRunesAndCarriageReturn/rejects_carriage_return`, `/accepts_60_multibyte_characters`, `/rejects_61_multibyte_characters` | `app/cmd/rename/rename_test.go:248` - `require.ErrorContains(t, err, "name")` sobre `"Fo\ro"` (`:247`); `:249` - `require.Equal(t, before, treeHash(t, root))`; `:256` - `require.Contains(t, read(t, root, "web/index.html"), "<title>"+name+"</title>")` com `name := strings.Repeat("ç", 60)`; `:261` - `require.ErrorContains(t, err, "name")` sobre `strings.Repeat("ç", 61)` | PASS |
| C14 | exit `1`, arquivo citado no stderr, sem `wrote` e árvore igual para `app/go.mod` ausente/sem `module`, sem `<title>`, `package.json` inválido; falha de escrita de `web/package.json` sai `1` com `wrote` parciais e `app/go.mod` já reescrito | lote unitário em b100ffe, 5 subtestes `--- PASS: TestRename_ErrorPaths/...` | tabela (`:272-274`): `:281` - `require.Equal(t, 1, code)`; `:282` - `require.Contains(t, errb, c.want)`; `:283` - `require.Empty(t, out)`; `:284` - `require.Equal(t, before, treeHash(t, root))` · `go.mod` ausente: `:292` - `require.Equal(t, 1, code)`; `:293` - `require.Contains(t, errb, "go.mod")`; `:294` - `require.Empty(t, out)` · escrita: `:303` - `require.Equal(t, 1, code)`; `:304` - `require.Contains(t, out, "wrote app/go.mod\n")`; `:305` - `require.NotContains(t, out, "wrote web/package.json")`; `:307` - `require.Contains(t, errb, "package.json")`; `:308` - `require.Equal(t, "module github.com/acme/foo\n\ngo 1.26\n", read(t, root, "app/go.mod"))` | PASS |

## Coverage

verified at b100ffe para as linhas que a correção ou a rodada 1 deixaram em aberto; as demais carried from be638d1
(a autoridade delas, `app/cmd/rename/main.go`, não mudou). Membros recomputados de `main.go` em `b100ffe`.

| Set (size) | Recomputed from | Member -> proof | Unproven |
| --- | --- | --- | --- |
| fronteira do caminho do módulo (3) - verified at b100ffe | regex `QuoteMeta(old)` seguida de um caractere fora de `[A-Za-z0-9._~-]` ou do fim do texto, em `main.go:155` | seguido de caractere fora da classe C1 `rename_test.go:92-94`, C2 `:114` · seguido de caractere da classe (`appx`) C1 `:95` · fim de arquivo (`$`) C12 `:240` | - |
| caracteres proibidos no nome (8) - verified at b100ffe | `forbiddenNameChars = "\"<>\\{}\r\n"` em `main.go:36` | `"` `<` `>` `\` `{` `}` `\n` C6 `:177-183`, `:189` · `\r` C13 `:248` | - |
| limites do nome (6) - verified at b100ffe | `checkName` em `main.go:90` (`len([]rune(name))`, `maxNameLength = 60`) | vazio C6 `:175` · 61 ASCII C6 `:176` · 1 e 60 ASCII aceitos C6 `:194-199` · 60 multibyte aceito C13 `:256` · 61 multibyte recusado C13 `:261` | - |
| caminhos de erro fora da validação (6) - verified at b100ffe | `readCurrent` `main.go:110-138` e `applyEdits` `main.go:207-212` | `app/go.mod` ausente (`:111-113`) C14 `:292-294` · sem linha `module` (`:116-118`) C14 `:282-284` · sem `<title>` (`:124-126`) C14 `:282-284` · `package.json` inválido (`:134-136`) C14 `:282-284` · escrita falha após `wrote` parciais (`:211-213`) C14 `:303-308` · `os.Stat` falha (`:207-209`) - mesmo `return written, err` da escrita, coberto pelo mesmo contrato `:304-305` | - |
| saídas de `run` (5) - verified at b100ffe | `main.go:50-62` | `--module` ausente C7 · `--name` ausente C7 · erro de validação C7 · erro de leitura/escrita C14 `:281`, `:303` · sucesso C1 e C9 | - |
| lugares do módulo (5) | carried from be638d1 - `planEdits` `main.go:155`, `:186` | `.go` C1 `:92` · `app/go.mod` C1 `:91` · `app/tools/go.mod` C2 `:114` · `.golangci.yml` C1 `:93` · `web/notes.txt` C1 `:94` | - |
| lugares do nome exibido (3) | carried from be638d1 - `nameEdits` `main.go:157-159` | `APITitle` C3 `:121` · `<title>` C3 `:122` · `<h1>` C3 `:123` | - |
| campos do pacote web (3) | carried from be638d1 - `nameEdits` `main.go:160-161` | `package.json` C4 `:135` · lock raiz e `packages[""]` C4 `:137` | - |
| derivação do pacote web (2) | carried from be638d1 - `webPackage` `main.go:99-102` | segmento simples C4 · sufixo `/v2` C4 | - |
| diretórios excluídos (8) | carried from be638d1 - `skippedDirs` `main.go:32-35` | os 8, C5 `:144`/`:155`, um subteste cada | - |
| outros não alvos (3) | carried from be638d1 | binário C5 `:165` · `text/template` C5 `:166` · `Template` C5 `:167` | - |
| recusa de módulo (2) | carried from be638d1 - `module.CheckPath` `main.go:72` | `Not A Path` C6 `:173` · `.foo` C6 `:174` | - |
| Landing doors (2) | carried from be638d1 - `plan.md` `## Landing` | `CheckPath` C6 `:174` · task e flags C11 `repo_test.go:124-128` | - |

## Test policy rows

verified at b100ffe - re-julgadas as duas linhas não cumpridas na rodada 1; as demais carried from be638d1.

| Row | Files it classifies | Required proof | Expectation met |
| --- | --- | --- | --- |
| Decides, reached across a boundary | `app/cmd/rename/main.go` `Rename`/`planEdits`/`checkName`/`webPackage` | boundary C9, C10 · own layer C1-C8, C12, C13 | yes - as duas linhas da tabela de decisão sem caso na rodada 1 (fim de arquivo na fronteira, `\r`) agora têm caso afirmado e matam seus mutantes (F7, F10); a contagem em caracteres tem caso multibyte (F11) |
| Decides, not reached across a boundary | nenhum arquivo do diff | - | n/a - nada classificado |
| Entry point that decides nothing | `app/cmd/rename/main.go` `main`/`run` | C7 · C14 · C1 (sucesso) | yes - entrada aceita (C1), cada entrada recusada (C7, C6), e cada caminho de erro enumerado na rodada 1 provado pelo `run` com exit `1`, stderr citando o arquivo e o `wrote` parcial (C14) |
| Instrumentation, pass-throughs | `app/internal/platform/testkit/repocopy.go`, task `rename` do `Taskfile.yml` | carried from be638d1 - C9/C10, `newslice` `TestGenerated_TaskCheckPasses`, C11 | yes |

## Faults injected

verified at b100ffe. Isolamento: `git worktree add --detach <scratchpad>/wt-r2 HEAD`; cada mutação aplicada em
`app/cmd/rename/main.go` do worktree, o `git diff` do worktree conferido como não vazio antes de rodar a prova mais
estreita, arquivo restaurado com `git checkout`; worktree removido com `git worktree remove --force` e
`git worktree prune`. `git status --porcelain` da árvore real antes e depois: idêntico (só
`.claude/skills/auth-security/` e `.cursor/skills/auth-security/`). São 7 mutações, acima do teto de cinco do
`verify.md`, porque o brief pediu F7 reinjetada mais uma por superfície nova, e C14 tem quatro superfícies de
asserção distintas (stderr com o arquivo em três mensagens e o contrato de escrita parcial).

| Mutation | Location | Killed |
| --- | --- | --- |
| F7 - alternativa de fim do texto removida da fronteira (`([^A-Za-z0-9._~-])`) | `app/cmd/rename/main.go:155` | yes - `TestRename_ModuleAtEndOfFile` |
| F10 - `\r` removido de `forbiddenNameChars` | `app/cmd/rename/main.go:36` | yes - `TestRename_NameRunesAndCarriageReturn/rejects_carriage_return` |
| F11 - `len([]rune(name))` trocado por `len(name)` | `app/cmd/rename/main.go:90` | yes - `TestRename_NameRunesAndCarriageReturn/accepts_60_multibyte_characters` |
| F12 - `<title>` ausente ignorado (`title = []string{"", ""}` no lugar do erro) | `app/cmd/rename/main.go:124-126` | yes - `TestRename_ErrorPaths/index.html_without_title` |
| F13 - erro de escrita engolido (`continue` no lugar de `return written, err`) | `app/cmd/rename/main.go:212` | yes - `TestRename_ErrorPaths/write_fails_after_partial_writes` |
| F14 - mensagem sem o arquivo (`errors.New("no module line")`) | `app/cmd/rename/main.go:117` | yes - `TestRename_ErrorPaths/go.mod_without_module_line` |
| F15 - erro de JSON sem envolver o arquivo (`return names{}, err`) | `app/cmd/rename/main.go:135` | yes - `TestRename_ErrorPaths/invalid_package.json` |

## Swept existing

carried from be638d1 - nenhuma linha de `## Swept` resolve para restrição existente; a correção acrescentou C14 às
linhas de failure modes implicitamente (o `Swept` de `checks.md` ainda cita C6, C7), sem nada a reler no código.

## Gate

`go -C app test -count=1 ./cmd/rename -run '^(TestRename_|TestMain_|TestTaskfileAndDocs_)' -v` em b100ffe - 12 passed, 0 failed (41 subtestes).
C9/C10 carried from be638d1: `go -C app test -count=1 -timeout=40m ./cmd/rename -run '^(TestRenamed_RepoCarriesNewNames|TestRenamed_TaskCheckPasses)$' -v` - 2 passed, 0 failed (627.632s).

## Ranked gaps

Nenhum gap bloqueante. Observações:

1. Precisão (não bloqueia): o subteste `missing go.mod` de C14 não afirma `treeHash` igual, embora a claim diga
   "deixando a árvore igual" para os quatro erros de leitura; a propriedade fica afirmada só indiretamente por
   `require.Empty(t, out)` - C14 - `app/cmd/rename/rename_test.go:288-295`
2. Artefato (não bloqueia): a claim de C13 em `checks.md` contém um byte CR literal no lugar do texto `\r`, então
   renderiza como "Um nome com `` é recusado"; a linha de `Coverage` grafa `\r` corretamente - C13 -
   `.specs/features/rename/checks.md:43`
3. Portabilidade (não bloqueia): o subteste de escrita parcial depende de `os.Chmod(..., 0o444)` impedir a escrita;
   rodando como root no Linux a escrita passaria e o teste falharia. O CI usa `ubuntu-latest` sem container
   (usuário não root) - C14 - `app/cmd/rename/rename_test.go:300`
4. Carried from be638d1 (não bloqueia): na fronteira, C9 só afirma o título da OpenAPI e o `package.json`;
   `<title>` e `<h1>` do repositório real só são provados sobre o fixture (C3) - C9 -
   `app/cmd/rename/repo_test.go:96-102`
