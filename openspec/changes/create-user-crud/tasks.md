# Tasks

## 1. Banco de dados, esquema e consultas

- [ ] 1.1 Adicionar DDL da tabela `users` a `schema/schema.sql` (`id VARCHAR(36) PRIMARY KEY`, `name TEXT NOT NULL`, `email TEXT NOT NULL UNIQUE`, `created_at TIMESTAMP NOT NULL`); verifique que a tabela foi adicionada e que o arquivo continua definindo apenas `items` e `users`.
- [ ] 1.2 Criar `queries/users.sql` com consultas sqlc `CreateUser`, `ListUsers`, `GetUser`, `UpdateUser` e `DeleteUser`; verifique que o arquivo existe e que cada consulta tem a anotação `-- name:` e o verbo correto (`:many`, `:one`, `:exec`).
- [ ] 1.3 Criar migrações `db/migrations/0002_create_users.up.sql` e `db/migrations/0002_create_users.down.sql`; verifique que a up cria a tabela `users` e a down a remove.
- [ ] 1.4 Executar `make migrate-up` com `MIGRATE_URL` definido (banco PostgreSQL em execução) e confirmar que a tabela `users` existe; verifique com `make migrate-down` seguido de `make migrate-up` novamente.
- [ ] 1.5 Executar `make generate` e confirmar que `internal/db/sqlc/` foi regenerado com os métodos de `users`; verifique que `internal/db/sqlc/items.sql.go` não foi alterado.

## 2. Modelo e repositório de usuários

- [ ] 2.1 Criar `internal/model/user.go` com a struct `User` (`ID`, `Name`, `Email`, `CreatedAt` com tags `json`); verifique que o pacote compila com `go build ./internal/model/`.
- [ ] 2.2 Criar `internal/repository/user.go` com a interface `UserRepository` (`Create`, `Get`, `List`, `Update`, `Delete`) e implementação sobre sqlc; verifique que o pacote compila e que `go vet ./internal/repository/` passa sem erros.

## 3. Serviço e regras de negócio

- [ ] 3.1 Criar `internal/service/user_service.go` com `UserService`, validação de nome (não vazio, no máximo 200 caracteres) e de e-mail (formato válido), erro sentinela `ErrInvalidInput`, geração de ID (UUID) e `created_at` em UTC; verifique com `go vet ./internal/service/`.
- [ ] 3.2 Criar `internal/service/user_service_test.go` com testes table-driven (`t.Run`), repositório mock implementando `UserRepository` e comparações com `errors.Is` contra `ErrInvalidInput`; verifique que `go test -race ./internal/service/` passa.

## 4. Handler HTTP e rotas da API

- [ ] 4.1 Criar `internal/handler/user_handler.go` com o handler Gin para `GET/POST/PUT/DELETE /api/v1/users`, usando `http.MaxBytesReader` (1 MiB), `ShouldBindJSON`, `c.Request.Context()` e respostas de erro genéricas ao cliente; verifique com `go vet ./internal/handler/`.
- [ ] 4.2 Criar `internal/handler/user_handler_test.go` com testes pelo contexto de teste do Gin e repositório mock; verifique status 200/201, 400 (entrada inválida) e 500 (erro inesperado).
- [ ] 4.3 Registrar as rotas de usuário e a composição (repositório, serviço, handler) em `cmd/api/main.go`; verifique que `make build` produz o binário sem erros.

## 5. Interface web

- [ ] 5.1 Criar templates `internal/web/templates/users.html` e `user-form.html` (HTMX, espelhando o padrão de `items.html`); verifique que `web.New` carrega os templates sem erro.
- [ ] 5.2 Adicionar as rotas de usuários ao `net/http.ServeMux` de `internal/web/web.go`; verifique que `go test -race ./internal/web/` passa.
- [ ] 5.3 Acrescentar casos de teste das rotas de usuário a `internal/web/web_test.go`; verifique que `make test` passa para o pacote web.

## 6. Verificação final de consistência

- [ ] 6.1 Executar `gofmt -w .`, `go vet ./...` e `golangci-lint run ./...`; verifique que não há erros de formatação ou lint.
- [ ] 6.2 Executar `make test` (`go test -race ./...`); verifique que todos os testes passam.
- [ ] 6.3 Executar `git diff --stat` e confirmar que apenas os arquivos do CRUD de usuários foram criados ou alterados, sem mudanças em `CONSTITUTION.md` nem na funcionalidade existente de itens.

## Workflow follow-up

- Revisar e aprovar os artefatos de planejamento antes de aplicar a change.
- Aplicar a change com `/opsx-apply` quando os artefatos estiverem revisados.
- Arquivar a change após a implementação e a revisão estarem concluídas.
