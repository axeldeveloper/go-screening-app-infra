# Proposal

## Why

O projeto já implementa um CRUD de itens (`items`) seguindo a arquitetura em camadas `config → db → repository → service → handler/web`, mas não existe nenhum domínio de usuários: não há modelo, repositório, serviço, handler, consulta SQL, migração nem template para usuários. O README declara o projeto como voltado à manutenção de usuários e prestadores de serviços, mas essa funcionalidade ainda não foi implementada. Esta change introduz o CRUD de usuários, cobrindo criação, listagem, atualização e remoção, seguindo os mesmos padrões do CRUD de itens existente.

## What Changes

- Modelo de domínio `User` em `internal/model/user.go` (campos `id`, `name`, `email`, `created_at`).
- Consultas SQL em `queries/users.sql`: `CreateUser`, `ListUsers`, `GetUser`, `UpdateUser`, `DeleteUser`.
- DDL da tabela `users` adicionado a `schema/schema.sql` (usado pelo sqlc para geração de código).
- Nova migração `db/migrations/0002_create_users.up.sql` e `0002_create_users.down.sql`.
- Código sqlc regenerado via `make generate` em `internal/db/sqlc/` (não editar manualmente).
- Repositório `internal/repository/user.go`: interface `UserRepository` e implementação sobre sqlc.
- Serviço `internal/service/user_service.go`: regras de negócio, validação de entrada e erro sentinela `ErrInvalidInput`.
- Handler `internal/handler/user_handler.go`: API JSON `GET/POST/PUT/DELETE /api/v1/users`, com `http.MaxBytesReader`, `context.Context` e erros genéricos ao cliente.
- Rotas registradas em `cmd/api/main.go` e composição do repositório/serviço/handler.
- Templates web `internal/web/templates/users.html` e `user-form.html` (HTMX), com rotas no `net/http.ServeMux` de `internal/web/web.go`.
- Testes table-driven com `t.Run` para serviço e handler, usando repositório mock e `errors.Is`.

## Capabilities

### New Capabilities

- `user-management`: criar, listar, atualizar e remover usuários via API JSON e interface web, respeitando a arquitetura em camadas do projeto (regras no service, acesso a dados no repository, handler apenas adapta HTTP).

### Modified Capabilities

- (nenhum)

## Impact

- `cmd/api/main.go` (composição de dependências e registro de rotas).
- `internal/model/`, `internal/repository/`, `internal/service/`, `internal/handler/`, `internal/web/` (novos arquivos e templates).
- `queries/`, `schema/`, `db/migrations/`, `internal/db/sqlc/` (gerado pelo sqlc).
- Comandos `make generate`, `make migrate-up`, `make test`, `make vet`, `make lint`.
- Nenhuma mudança em `CONSTITUTION.md`: não altera arquitetura nem fronteiras de segurança.
