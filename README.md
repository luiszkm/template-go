# Template

API em Go e interface em React, com PostgreSQL, sessão revogável, papéis e
auditoria. O repositório é um ponto de partida: `task rename` troca o nome do
projeto e `task new:slice` cria um caso de uso novo.

As regras de arquitetura e o fluxo de trabalho estão em `AGENTS.md`.

## Pré-requisitos

- Go 1.26.9
- Node.js 24 (`.nvmrc`; a mesma versão do CI e do Dockerfile)
- Task 3
- Docker com Compose v2

## Começar

Renomeie o módulo Go, o pacote web e o nome exibido (aba do navegador, home e
OpenAPI). O histórico em `.specs/` permanece.

```sh
task rename MODULE=github.com/acme/foo NAME=Foo
```

Suba o Postgres, aplique as migrações, a API em `:8080` e o Vite em `:5173`.
Abra http://localhost:5173.

```sh
task dev
```

Crie o primeiro usuário com o papel `admin` (permissão `*`). A senha vai pelo
stdin e tem de 12 a 128 caracteres. Rode em outro terminal, a partir de `app/`,
com o `task dev` no ar:

```sh
cd app
echo 'uma-senha-de-12' | DATABASE_URL=postgres://app:app@localhost:5432/app?sslmode=disable \
  go run ./cmd/api users create-admin --email admin@example.com --name Admin
```

O processo lê as variáveis do ambiente. Copie os valores de `.env.example` para
o ambiente do shell.

## Variáveis de ambiente

Só `DATABASE_URL` é obrigatória. As tarefas do `Taskfile.yml` usam
`postgres://app:app@localhost:5432/app?sslmode=disable` quando ela não está
definida. Os demais valores são o padrão do código.

| Variável | Padrão |
| --- | --- |
| `DATABASE_URL` | obrigatória |
| `HTTP_ADDR` | `:8080` |
| `SHUTDOWN_TIMEOUT` | `10s` |
| `READY_TIMEOUT` | `2s` |
| `LOG_LEVEL` | `info` |
| `HTTP_READ_TIMEOUT` | `30s` |
| `HTTP_WRITE_TIMEOUT` | `30s` |
| `HTTP_IDLE_TIMEOUT` | `120s` |
| `SESSION_TTL` | `12h` |
| `SESSION_SWEEP_INTERVAL` | `1h` |
| `COOKIE_SECURE` | `true` |
| `TRUSTED_PROXIES` | vazio |

`COOKIE_SECURE=true` marca o cookie de sessão como `Secure`, o que exige HTTPS.
`task dev` e o perfil `app` do Compose definem `false` para o HTTP local.
`TRUSTED_PROXIES` é uma lista de IPs ou CIDRs separados por vírgula. Com ela
vazia, o IP do cliente é o da conexão.

## Gate e fatias novas

`task check` é o gate: formatação, lint, código gerado, testes Go com Postgres,
typecheck, lint e testes da web. Ele precisa do Docker.

```sh
task new:slice FEATURE=orders NAME=create_order
```

## Imagem

O banco sozinho:

```sh
docker compose up -d --wait db
```

Banco, migrações e a imagem (API e interface em http://localhost:8080):

```sh
docker compose --profile app up --build
```

O serviço `migrate` roda `api migrate up` e termina. O `app` sobe em seguida.
Para criar o admin a partir da máquina, use `localhost` na `DATABASE_URL`, como
no comando acima. Dentro da rede do Compose o host do banco é `db`.

## Licença

Copyright (c) 2026 Luis Soares. Uso, cópia, modificação e distribuição só para
fins não comerciais, nos termos da PolyForm Noncommercial License 1.0.0. O
texto está em `LICENSE`. Uso comercial não está concedido. `task rename` não
altera o aviso de copyright nem esta licença.
