# Proposal

## Why

O projeto já possui `Dockerfile` e um `docker-compose.yml` parcial, mas o compose não levanta a aplicação: os serviços `app` e `migrate` estão comentados, não há ordem de inicialização que garanta que o banco esteja pronto antes da API, e não há execução orquestrada de testes, linter, análise SonarQube nem logs nos containers. O objetivo da solicitação é fazer funcionar tudo em containers Docker: API, banco de dados, SonarQube, logs, testes e linter.

## What Changes

- Revisar e completar o `docker-compose.yml` com os serviços `app` e `migrate` ativos, com `depends_on` condicionado à saúde do banco de dados.
- Garantir que a API aguarde o banco e que as migrações sejam aplicadas antes da aplicação iniciar (via serviço `migrate` ou entrypoint da API).
- Expor/documentar os logs dos containers (API, banco, Sonar, scanner) via driver de log do Docker e `docker logs`.
- Adicionar ao Compose a execução de `make test`, `make vet` e `golangci-lint` (quality gate) e manter o serviço `sonar-scanner` já existente no profile `tools`.
- Ajustar variáveis de ambiente (`.env.example`) para funcionar com o Compose: `DATABASE_URL` apontando para o serviço `db` na rede interna, `PORT=8000`.
- Manter segredos por variáveis de ambiente; `.env` continua não versionado.
- Não alterar rotas, dependências Go nem comportamento de runtime da API.

## Capabilities

### New Capabilities

- `docker-infra`: infraestrutura de containers (docker-compose) que sobe API, banco Postgres, SonarQube e scanner, com execução de testes, linter e logs, tudo orquestrado pelo Compose.

### Modified Capabilities

- Nenhum. Esta change não altera requisitos de capacidades existentes (`project-constitution`, `agent-guidance`); apenas adiciona infraestrutura de execução.

## Impact

- Arquivos: `docker-compose.yml`, possíveis ajustes em `Dockerfile` e `.env.example`.
- Nenhuma mudança em rotas, endpoints, dependências Go ou comportamento em runtime da API.
- Segredos continuam por variáveis de ambiente (constituição, princípio 4).
- Prosa em português; identificadores e caminhos em inglês (constituição, princípio 5).
