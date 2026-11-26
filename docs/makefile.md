# Guia dos comandos do Makefile

Este documento explica cada alvo (`target`) do `Makefile` do projeto `my-screening-app`, o que faz, quando usar e pré-requisitos.

O projeto segue o padrão de variáveis com `?=`, permitindo sobrescrever valores via ambiente ou linha de comando, por exemplo:

```bash
make sonar SONAR_URL=http://localhost:9001
```

---

## Variáveis globais

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `GOLANGCI_LINT_VERSION` | `v2.15.0` | Versão fixada do golangci-lint usada quando o binário não está no PATH |
| `GOLANGCI_LINT` | binário do PATH ou `go run ...@v2.15.0` | Comando do linter |
| `SQLC_VERSION` | `v27.0` | Versão do sqlc usada para geração de código |
| `COVER_MIN` | `30` | Cobertura mínima exigida pelo gate |
| `MIGRATE_TAGS` | `postgres` | Tags de build para o migrate |
| `MIGRATE_URL` | vazio | DSN do banco para migrações (obrigatório para `migrate-up/down`) |
| `SONAR_URL` | `http://localhost:9001` | URL do servidor SonarQube |
| `SONAR_DASHBOARD` | `http://localhost:9001/dashboard?id=my_screening_app` | Link do dashboard do projeto |
| `REPORTS_PORT` | `8080` | Porta para servir relatórios de qualidade |

---

## 1. Build e execução da aplicação

### `make build`

**O que faz:** Gera o código com sqlc (`generate`) e compila o binário da API em `./bin/my_screening_app`.

```bash
make build
```

**Quando usar:** Antes de rodar a aplicação localmente ou para criar o binário de implantação.

**Pré-requisitos:** Go instalado no host, conexão com internet para baixar dependências.

---

### `make run`

**O que faz:** Gera o código com sqlc e executa a API diretamente com `go run ./cmd/api`.

```bash
make run
```

**Quando usar:** Desenvolvimento local rápido, sem precisar gerar binário.

**Pré-requisitos:** Go instalado no host, banco de dados disponível.

---

## 2. Testes e qualidade de código

### `make test`

**O que faz:** Executa todos os testes do projeto com o detector de race conditions (`-race`).

```bash
make test
```

**Quando usar:** Desenvolvimento local, antes de commitar alterações.

**Pré-requisitos:** Go instalado no host.

---

### `make vet`

**O que faz:** Executa `go vet ./...` para análise estática básica do código Go.

```bash
make vet
```

**Quando usar:** Verificação rápida de problemas evidentes no código.

**Pré-requisitos:** Go instalado no host.

---

### `make fmt`

**O que faz:** Formata todos os arquivos Go com `gofmt -w .`.

```bash
make fmt
```

**Quando usar:** Padronizar formatação antes de commitar.

**Pré-requisitos:** Go instalado no host.

---

### `make generate`

**O que faz:** Executa `sqlc generate` para gerar código Go a partir dos arquivos SQL em `queries/` e `schema/`. O código gerado fica em `internal/db/sqlc/`.

```bash
make generate
```

**Quando usar:** Sempre que alterar queries SQL ou o schema do banco.

**Pré-requisitos:** Conexão com internet para baixar o sqlc.

---

### `make generate-check`

**O que faz:** Executa `generate` e verifica com `git diff --exit-code` se houve alterações no código gerado (`internal/db/sqlc`). Falha se o código gerado não estiver commitado.

```bash
make generate-check
```

**Quando usar:** CI/CD para garantir que o código gerado está atualizado.

**Pré-requisitos:** Git, conexão com internet.

---

### `make lint`

**O que faz:** Executa `golangci-lint run ./...` com a versão fixada (ou binário local se disponível).

```bash
make lint
```

**Quando usar:** Análise estática completa (estilo, boas práticas, possíveis bugs).

**Pré-requisitos:** Go ou `golangci-lint` instalado.

---

### `make vuln`

**O que faz:** Executa `govulncheck` para verificar vulnerabilidades conhecidas nas dependências Go.

```bash
make vuln
```

**Quando usar:** Verificação de segurança de dependências.

**Pré-requisitos:** Go instalado, conexão com internet.

---

### `make cover`

**O que faz:** Executa testes com cobertura (`-coverprofile=coverage.out`), exibe o percentual total e falha se estiver abaixo de `COVER_MIN` (padrão 30%).

```bash
make cover
```

**Quando usar:** Verificar se a cobertura de testes atende ao mínimo exigido.

**Pré-requisitos:** Go instalado no host.

---

### `make gate`

**O que faz:** Executa o quality gate completo: `generate-check` + `lint` + `vuln` + `cover`. É o mesmo comando usado no CI e no estágio `gate` do Dockerfile.

```bash
make gate
```

**Quando usar:** Antes de abrir um Pull Request, no CI local ou no estágio de qualidade do Docker.

**Pré-requisitos:** Go, conexão com internet, git limpo (para `generate-check`).

---

### `make gate-docker`

**O que faz:** Constrói apenas o estágio `gate` do Dockerfile (`docker build --target gate`), que roda o quality gate completo dentro do container.

```bash
make gate-docker
```

**Quando usar:** Validar o gate em ambiente isolado (Docker), sem instalar ferramentas no host.

**Pré-requisitos:** Docker instalado e em execução.

---

## 3. Relatórios de qualidade

### `make lint-report`

**O que faz:** Executa o golangci-lint e gera relatórios em `reports/` nos formatos JSON, JUnit XML e Checkstyle. Depois gera um relatório legível HTML (`reports/quality.html`) usando a ferramenta `./tools/lintreport`.

```bash
make lint-report
```

**Quando usar:** Gerar relatórios para auditoria ou visualização local da qualidade de código.

**Pré-requisitos:** golangci-lint, Go, diretório `reports/` gravável.

---

### `make quality-report`

**O que faz:** Alias para `lint-report`.

```bash
make quality-report
```

---

### `make reports-serve`

**O que faz:** Após gerar o relatório (`lint-report`), serve os arquivos de `reports/` em um servidor HTTP local na porta `REPORTS_PORT` (padrão 8080).

```bash
make reports-serve
```

**Quando usar:** Visualizar o relatório HTML no navegador.

**Acesse:** `http://localhost:8080/quality.html`

**Pré-requisitos:** Python 3 (para `http.server`).

---

### `make reports-open`

**O que faz:** Serve o relatório e tenta abrir automaticamente no navegador padrão (`open` no macOS, `xdg-open` no Linux).

```bash
make reports-open
```

**Quando usar:** Abrir o relatório de qualidade rapidamente.

---

## 4. SonarQube + scanner

Esses alvos integram o projeto com o SonarQube orquestrado pelo `docker-compose.yml`.

### `make sonar-up`

**O que faz:** Sobe apenas os serviços `db` (Postgres) e `sonarqube` via Docker Compose. O scanner não sobe sozinho (fica no profile `tools`).

```bash
make sonar-up
```

**Quando usar:** Iniciar a infraestrutura do SonarQube sem rodar a análise ainda.

**Pré-requisitos:** Docker e Docker Compose, arquivo `.env` com `POSTGRES_PASSWORD` e `SONAR_DB_PASSWORD`.

---

### `make sonar-wait`

**O que faz:** Depende de `sonar-up`. Aguarda até 60 tentativas (5s cada) que o SonarQube responda `status: UP` em `http://localhost:9001/api/system/status`.

```bash
make sonar-wait
```

**Quando usar:** Garantir que o SonarQube esteja pronto antes de rodar o scanner.

**Pré-requisitos:** `curl` instalado, SonarQube subindo.

---

### `make sonar-scan`

**O que faz:** Depende de `sonar-wait`. Executa o `sonar-scanner` sob demanda via `docker compose --profile tools run --rm -T sonar-scanner`. O scanner analisa o código e envia o relatório para o SonarQube.

```bash
make sonar-scan
```

**Quando usar:** Disparar uma análise de código sob demanda, após o SonarQube estar saudável.

**Pré-requisitos:** Docker Compose, `.env` com `SONAR_TOKEN`, SonarQube UP.

---

### `make sonar` (alvo principal)

**O que faz:** Fluxo completo:
1. `sonar-wait` (sobe db + sonarqube e aguarda ficar UP)
2. Roda `go test ./... -coverprofile=coverage.out` **dentro de um container Docker** (`golang:1.25-alpine`), com suporte a CGO
3. Executa o `sonar-scanner` via `docker compose --profile tools run`
4. Exibe o link do dashboard

```bash
make sonar
```

**Quando usar:** Análise completa de qualidade e segurança em um único comando, especialmente quando não tem Go instalado no host (WSL, por exemplo).

**Pré-requisitos:** Docker, `.env` com `SONAR_TOKEN`.

---

### `make sonar-open`

**O que faz:** Tenta abrir o dashboard do SonarQube no navegador padrão.

```bash
make sonar-open
```

**Quando usar:** Abrir rapidamente os resultados da análise no navegador.

**URL:** `http://localhost:9001/dashboard?id=my_screening_app`

---

## 5. Migrações do banco de dados

### `make migrate-up`

**O que faz:** Executa `migrate up` para aplicar todas as migrações pendentes em `db/migrations/` no banco definido por `MIGRATE_URL`.

```bash
export MIGRATE_URL='postgres://usuario:senha@localhost:5433/my_screening_app?sslmode=disable'
make migrate-up
```

**Quando usar:** Aplicar mudanças de schema no banco de desenvolvimento ou produção.

**Pré-requisitos:** `MIGRATE_URL` definida, conexão com o banco.

---

### `make migrate-down`

**O que faz:** Executa `migrate down` para reverter a última migração aplicada.

```bash
export MIGRATE_URL='postgres://usuario:senha@localhost:5433/my_screening_app?sslmode=disable'
make migrate-down
```

**Quando usar:** Desfazer a última migração durante desenvolvimento.

**Pré-requisitos:** `MIGRATE_URL` definida, conexão com o banco.

---

## Fluxos recomendados

### Desenvolvimento local (com Go instalado)

```bash
# 1. Formatar código
make fmt

# 2. Gerar código do banco (se alterou SQL)
make generate

# 3. Rodar testes
make test

# 4. Verificar qualidade
make vet
make lint

# 5. Build
make build
```

### Análise completa com SonarQube (sem Go no host, ex: WSL)

```bash
# 1. Preencha o .env com SONAR_TOKEN, POSTGRES_PASSWORD, etc.
cp .env.example .env

# 2. Rode a análise completa (testes em container + scanner)
make sonar

# 3. Abra o dashboard
make sonar-open
```

### CI / qualidade antes de PR

```bash
make gate
```

Ou via Docker (sem instalar ferramentas no host):

```bash
make gate-docker
```

---

## Arquivos relacionados

- `Dockerfile` — estágios `builder`, `gate` e runtime
- `docker-compose.yml` — orquestração de `app`, `db`, `migrate`, `sonarqube`, `sonar-scanner`
- `.env.example` — variáveis de ambiente de exemplo
- `sonar-project.properties` — configuração do projeto no SonarQube
- `docs/sonarqube.md` — guia detalhado de uso do SonarQube
