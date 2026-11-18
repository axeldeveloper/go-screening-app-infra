# AGENTS.md — instruções para agentes

Este guia orienta agentes de IA e contribuidores a trabalhar neste repositório. Consulte também `CONSTITUTION.md`, que define os princípios do projeto.

## Visão geral

`my-screening-app` é um serviço Go (go 1.22, Gin, sqlc, PostgreSQL, golang-migrate) scaffoldado com TriStack, para manutenção de usuários e prestadores de serviços, auditorias e monitoramento. A aplicação expõe uma API HTTP e uma interface web com templates HTML embutidos.

## Estrutura de diretórios

```
cmd/api/             entrypoint: monta dependências, registra rotas, aplica timeouts e shutdown graceful
internal/config/     configuração por variáveis de ambiente (APP_NAME, PORT)
internal/db/         conexão com Postgres (pgx) e pool; EnsureSchema para o schema básico
internal/db/sqlc/    código GERADO pelo sqlc — não editar manualmente
internal/repository/ acesso a dados; implementação via sqlc
internal/service/    regras de negócio e validações
internal/model/      modelos de domínio
internal/handler/    handlers HTTP com Gin
internal/web/        templates e estáticos embedados, servidos por net/http
db/migrations/       migrações do golang-migrate
queries/             consultas SQL de entrada do sqlc
schema/              schema SQL de entrada do sqlc
openspec/            planejamento de mudanças (propostas, specs, design, tarefas)
Makefile             alvos de build, teste, formatação, geração e migração
```

## Arquitetura e fluxo de dados

A cadeia de dependências é `config → db → repository → service → handler/web`, com `cmd/api/main.go` como raiz de composição:

1. `config.Load()` lê `APP_NAME` e `PORT` do ambiente.
2. `db.Connect()` abre o PostgreSQL com pgx e configura o pool.
3. `repository.NewItemRepository(conn)` cria o repositório sobre o código sqlc.
4. `service.NewItemService(repo)` recebe a interface `repository.ItemRepository`.
5. `handler.NewHandler(items)` e `web.New(cfg.AppName, items)` recebem o serviço.
6. As rotas são registradas com Gin: `GET /health`, `GET /api/v1/items`, `POST /api/v1/items`, `/`, `/web/items` e `/static/*filepath`.

Regras de dependência: uma camada depende apenas da imediatamente inferior. Regras de negócio e validações ficam em `internal/service`; `internal/handler` apenas adapta requisições e respostas HTTP; `internal/repository` é o único lugar de acesso a dados. Ao adicionar casos de uso, siga o caminho modelo → query em `queries/` → sqlc → repositório → serviço → handler/rota, criando uma migração em `db/migrations/` se o esquema mudar.

## Convenções de código e qualidade

- Formate com `gofmt` e verifique com `go vet` e `golangci-lint` (configuração em `.golangci.yml`, linters padrão e formatador `gofmt`).
- Enriqueça erros com `%w` (`fmt.Errorf("contexto: %w", err)`); use erros sentinela como `service.ErrInvalidInput` e compare com `errors.Is`.
- Propague `context.Context` em chamadas de banco e serviços (`r.Context()` nos handlers).
- Identificadores de código, comandos e caminhos permanecem em inglês; prosa e documentação são escritas em português.
- Escreva testes table-driven com `t.Run` ao lado do código testado, como em `internal/service/item_service_test.go`.
- Toda entrada HTTP é limitada (`http.MaxBytesReader`) e validada no serviço antes de chegar ao domínio.

## Comandos

Todos os comandos abaixo estão definidos no `Makefile`:

| Comando | Finalidade |
| --- | --- |
| `make build` | compila o binário em `./bin/` (depende de `generate`) |
| `make run` | executa `go run ./cmd/api` |
| `make test` | `go test -race ./...` |
| `make vet` | `go vet ./...` |
| `make fmt` | `gofmt -w .` |
| `make generate` | gera `internal/db/sqlc/` a partir de `queries/` e `schema/` via sqlc |
| `make migrate-up` | aplica as migrações de `db/migrations/` |
| `make migrate-down` | reverte as migrações de `db/migrations/` |

Para migrações, a variável `MIGRATE_URL` sobrepõe a DSN padrão. Variáveis de ambiente são documentadas em `env.example` (o Go não carrega `.env` automaticamente).

## Fluxo de trabalho OpenSpec

- O planejamento vive em `openspec/changes/<nome>/` (proposta, specs, design e tarefas).
- Ciclo: `/opsx-propose` → `/opsx-apply` → `/opsx-archive`.
- Mudanças de comportamento observável DEVEM ser planejadas como uma change antes de editar o código da aplicação.
- Os princípios que norteiam as decisões estão em `CONSTITUTION.md`, com resumo em `openspec/constitution.md` e no campo `context` de `openspec/config.yaml`. Mudanças de arquitetura exigem atualizar a constituição na mesma change.

## Pendências conhecidas (confirmar antes de corrigir)

Estas divergências entre documentação e código foram observadas e **não** devem ser corrigidas como efeito colateral de outras tarefas:

- `README.md` informa `PORT` padrão `3000`, mas `internal/config/config.go` e `env.example` usam `8000`.
- A DSN de migração no `README.md` usa `postgres://postgres:pg766312@localhost:5432/my_screening_app?sslmode=disable`; `env.example` e `Makefile` usam `postgres://postgres:postgres@localhost:5432/my_tristack_app`, e `internal/db/db.go` usa essa mesma DSN sem `sslmode`.

Trate cada item como uma change separada de documentação/configuração, após confirmação.
