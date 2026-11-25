# Design

## Context

Estado atual do repositório (observado no código, sem alterações de runtime):

- `Dockerfile` com três estágios: `builder` (roda `sqlc generate`, build da API e instala `migrate`), `gate` (executa `make gate`: generate-check, lint, vuln, cover) e runtime `distroless` expondo a porta 8000. O runtime só executa a API; não aplica migrações.
- `docker-compose.yml` com `db` (postgres:17-alpine, publicado em 5433:5432, healthcheck com `pg_isready`), `sonarqube` (9001:9001, healthcheck via `/api/system/status`) e `sonar-scanner` (profile `tools`). Os serviços `app` e `migrate` estão **comentados**.
- `db/init/01-sonarqube.sh` cria o usuário e banco do SonarQube no primeiro start do `db`.
- `cmd/api/main.go` falha imediatamente (`log.Fatalf`) se o banco não estiver pronto — não há retry — e não executa migrações na inicialização.
- `.env.example`: `PORT=8000`, `DATABASE_URL` apontando para `localhost:5433` (acesso do host), `POSTGRES_PASSWORD`, `SONAR_DB_PASSWORD`, `SONAR_TOKEN`.
- `Makefile`: `test -race`, `vet`, `fmt`, `generate`, `gate`, `gate-docker`, `sonar`, `migrate-up/down` (exigem `MIGRATE_URL`).
- Motivação e requisitos: `proposal.md` e `specs/docker-infra/spec.md`.

Restrições da constituição (fonte de verdade: `CONSTITUTION.md`; resumo em `openspec/config.yaml`): camadas `config → db → repository → service → handler/web`, segredos por variáveis de ambiente, não expor erros internos ao cliente, timeouts e desligamento graceful, prosa em português e identificadores/caminhos em inglês.

## Goals / Non-Goals

**Goals:**

- O `docker-compose.yml` levanta `app`, `db`, `migrate`, `sonarqube` e `sonar-scanner` com ordem de inicialização correta (`depends_on` condicionado a healthchecks).
- A API só inicia após o banco estar saudável e as migrações aplicadas.
- Testes (`go test -race ./...`), linter (`golangci-lint`) e `vet` executáveis em container.
- Logs de todos os serviços acessíveis via `docker logs`.
- Nenhuma alteração no comportamento de runtime, rotas ou endpoints da API.

**Non-Goals:**

- Alterar código Go da aplicação (handlers, service, repository, config).
- Traduzir o `README.md` existente.
- Configurar SonarQube para produção, reverse proxy ou autenticação externa.
- Corrigir divergências pré-existentes entre `README.md` e código (porta padrão, DSN) — registradas, não resolvidas aqui.

## Decisions

### 1. Serviços `app` e `migrate` ativos no Compose, com ordem por healthcheck

`migrate` declara `depends_on: db` com `condition: service_healthy` e executa `migrate up` contra `db:5432`. `app` declara `depends_on: db` (`service_healthy`) e `depends_on: migrate` (`condition: service_completed_successfully`).

- **Por quê:** garante que o esquema exista antes da API responder (requisito: "Migrações aplicadas antes da API") e que a API não rode sem banco (requisito: "API depende do banco estar saudável"), contornando a falta de retry em `main.go` sem alterar o código Go.
- **Alternativa considerada:** entrypoint da própria API rodando as migrações — descartado porque a imagem runtime é `distroless` e não inclui o binário `migrate`; adicionaria complexidade ao runtime.

### 2. Banco da aplicação acessível apenas na rede interna do Compose

O serviço `db` expõe a porta 5432 só na rede; o host continua acessando via `localhost:5433` (mapeamento já existente) para uso fora do Compose. Dentro da rede, `DATABASE_URL` aponta para `db:5432/my_screening_app`.

- **Por quê:** reduz área de exposta do banco e mantém compatibilidade com `.env.example` para desenvolvimento no host.

### 3. Segredos por variáveis de ambiente via `.env` (não versionado)

`POSTGRES_PASSWORD`, `SONAR_DB_PASSWORD` e `SONAR_TOKEN` seguem vindo do `.env`, preenchido a partir de `.env.example`. Nenhum valor real é版本izado.

- **Por quê:** constituição, princípio 4 (segredos por variáveis de ambiente).

### 4. Logs via driver padrão do Docker

Nenhum serviço de log dedicado (ex. ELK/Loki) é adicionado; stdout/stderr de cada container ficam disponíveis com `docker logs <serviço>`.

- **Por quê:** requisito de logs atendido com o mínimo de novos componentes; alternativa (stack de logging completa) é escopo de outra change.

### 5. `sonar-scanner` permanece no profile `tools`

Mantido como está; executado com `docker compose --profile tools run --rm sonar-scanner` quando SonarQube estiver saudável.

- **Por quê:** evita que o scanner rode sozinho no `docker compose up` (comportamento atual já desejado).

### 6. Sem alteração na `CONSTITUTION.md`

A change adiciona infraestrutura de execução, não altera arquitetura nem segurança da aplicação, então não há exigência de revisar a constituição (princípio 6).

## Risks / Trade-offs

- [Banco pode não estar pronto no momento em que o container da API inicia, apesar do `depends_on`] → `depends_on` com `service_healthy` cobre o healthcheck do Postgres; se a aplicação ainda falhar, avaliar retry de conexão no entrypoint do `app` sem alterar `main.go`.
- [SonarQube consome muita memória em ambientes com pouca RAM] → já mitigado com `SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true` no compose existente; manter.
- [Valores de placeholder em `.env.example` (senhas `troque-esta-senha`)] → documentar que `.env` real deve ser preenchido antes de subir o Compose; não versionar segredos.
- [Testes/linter como serviços separados vs. comandos documentados] → decisão de implementação a confirmar nas tarefas; não afeta os specs.

## Migration Plan

1. Editar `docker-compose.yml`: ativar e ajustar serviços `app` e `migrate` com `depends_on` condicionado a healthchecks; manter `db`, `sonarqube` e `sonar-scanner` (profile `tools`).
2. Ajustar `.env.example` se necessário (`DATABASE_URL` para rede interna; manter `localhost:5433` para host).
3. Validar: `docker compose config`, `docker compose up -d`, `GET /health`, `docker compose logs`.
4. Rollback: restaurar a versão anterior do `docker-compose.yml`; nenhum dado de aplicação a desfazer (volumes `pgdata` podem ser removidos se desejado).

## Open Questions

- Os serviços de testes e qualidade (`make test`, `vet`, `golangci-lint`) devem ser serviços dedicados no `docker-compose.yml` ou continuar como comandos documentados a serem rodados com `docker run` / `docker compose run`? (escopo de implementação, não muda os specs)
