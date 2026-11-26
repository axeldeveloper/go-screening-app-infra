# Design

## Context

O projeto já implementa um CRUD de itens (`items`) com a cadeia `config → db → repository → service → handler/web`, e a raiz de composição é `cmd/api/main.go`. Não há domínio de usuários ainda: não existem modelo, consultas SQL, repositório, serviço, handler, migração nem template para usuários. Esta change adiciona o domínio `user-management` seguindo o mesmo padrão do CRUD de itens existente. Veja a motivação em `proposal.md` e os requisitos em `specs/user-management/spec.md` — não são reescritos aqui.

## Goals / Non-Goals

**Goals:**
- Implementar cadastro, listagem, consulta, atualização e remoção de usuários via API JSON e interface web.
- Reutilizar os padrões estabelecidos: sqlc para acesso a dados, interface de repositório no serviço, validação de entrada no service, `context.Context` propagado, erros com `%w` e testes table-driven.
- Adicionar migração `db/migrations/0002_create_users` para a tabela `users`.
- Manter `internal/db/sqlc/` como código gerado (não editar manualmente).

**Non-Goals:**
- Autenticação, senhas, controle de sessão ou autorização (coberto por uma change separada, se aplicável).
- Alteração no domínio de itens existente.
- Mudança na arquitetura em camadas ou nas fronteiras de responsabilidade.

## Decisions

### Seguir o padrão do CRUD de itens
**Decisão:** implementar usuários com a mesma estrutura: `internal/model/user.go`, `queries/users.sql`, `schema/schema.sql` (DDL da tabela `users`), repositório `internal/repository/user.go` (interface + implementação sqlc), serviço `internal/service/user_service.go`, handler `internal/handler/user_handler.go`, rotas em `cmd/api/main.go` e templates web em `internal/web/templates/`.
**Racional:** consistência com o código existente, menor curva de aprendizado e reaproveitamento de padrões de teste já validados.
**Alternativa considerada:** criar um pacote dedicado `internal/user/` agregando camadas. Rejeitado porque quebra a convenção de "uma responsabilidade por pacote em internal/" e separa camadas que o projeto mantém divididas.

### Modelo de dados
**Decisão:** tabela `users` com colunas `id VARCHAR(36) PRIMARY KEY`, `name TEXT NOT NULL`, `email TEXT NOT NULL UNIQUE`, `created_at TIMESTAMP NOT NULL`.
**Racional:** `id` gerado pelo serviço (UUID string) espelha `model.Item`; `email` com restrição `UNIQUE` garante unicidade no banco como backstop da validação de aplicação.
**Alternativa considerada:** gerar `id` no banco. Rejeitado porque o serviço já gera UUID para itens e o repositório recebe o modelo completo.

### Unicidade de e-mail
**Decisão:** validar unicidade no serviço antes de gravar e confiar na restrição `UNIQUE` do banco para tratar condições de corrida; converter violação de restrição em erro de domínio (e-mail duplicado).
**Racional:** a validação de aplicação dá feedback rápido; a restrição do banco é a fonte de verdade sob concorrência.

### Rotas e interface web
**Decisão:** API JSON `GET/POST/PUT/DELETE /api/v1/users` no Gin (espelha `/api/v1/items`) e templates `users.html`/`user-form.html` com HTMX no `net/http.ServeMux` (espelha `/web/items`).
**Racional:** mantém a separação já existente entre API (Gin) e interface web (mux + HTMX).

### Testes
**Decisão:** testes table-driven com `t.Run`, repositório mock implementando `repository.UserRepository`, comparações com `errors.Is` contra `service.ErrInvalidInput`.
**Racional:** segue o padrão de `item_service_test.go` e `handler_test.go`.

## Risks / Trade-offs

- [Condição de corrida na criação de e-mail duplicado] → duas requisições simultâneas com o mesmo e-mail podem passar na validação de aplicação; mitigado pela restrição `UNIQUE` no banco e tratamento do erro sql como erro de domínio.
- [sqlc deve ser regenerado] → `make generate` é obrigatório após alterar `queries/users.sql` ou `schema/schema.sql`; o build falha se o código gerado não estiver atualizado.
- [Esquema gerenciado por migrações] → `db.EnsureSchema` existe em `internal/db/db.go`, mas não é chamado em `main.go`; a tabela `users` será criada via `make migrate-up` com `MIGRATE_URL` definido (banco PostgreSQL em execução necessário).
- [Divergências conhecidas de DSN/PORT] → documentadas no `AGENTS.md` (pendências de documentação/configuração); não são corrigidas nesta change para evitar efeitos colaterais.
- [Sem autenticação] → endpoints de usuários estarão expostos sem controle de acesso; decision consciente de escopo, a ser revisado em uma change de autenticação futura.

## Migration Plan

1. Adicionar DDL da tabela `users` a `schema/schema.sql`.
2. Adicionar consultas em `queries/users.sql` e arquivos de migração `db/migrations/0002_create_users.up.sql` / `.down.sql`.
3. Executar `make migrate-up` com `MIGRATE_URL` definido (requer Postgres rodando).
4. Executar `make generate` para regenerar `internal/db/sqlc/`.
5. Implementar modelo, repositório, serviço, handler e templates web.
6. Registrar rotas e composição em `cmd/api/main.go`.
7. Executar `make test`, `make vet` e `make lint`.
Rollback: `make migrate-down` reverte a migração; remover os arquivos de código adicionados.

## Open Questions

- Nenhuma que altere os requisitos, a abordagem ou a divisão de tarefas. Autenticação e senhas são deliberadamente deixados para uma change futura (`user-auth`); se o escopo mudar, os requisitos em `specs/user-management/spec.md` precisam ser revisados antes de continuar.
