# my-screening-app

Project for the maintenance of users and service providers, audits, and monitoring

A Go project scaffolded with [TriStack](https://tristack.space).


## Stack

- **Framework:** gin
- **ORM / DB layer:** sqlc
- **Database:** postgres
- **Migrations:** golang-migrate
- **Package manager:** go

## Getting Started

```bash
# 1. Install dependencies

go mod tidy

# 2. Generate query code (sqlc)

go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate

# 3. Load .env and run database migrations (host: 5433)

set -a && . ./.env && set +a
export MIGRATE_URL="postgres://postgres:${POSTGRES_PASSWORD}@localhost:5433/my_screening_app?sslmode=disable"
make migrate-up

# 4. Start the dev server (same shell, .env already loaded)

go run ./cmd/api
```

## Common commands

```bash
go test ./...                          # run all tests
go vet ./...                           # static analysis
gofmt -l .                             # check formatting
go build -o ./bin/my_screening_app ./cmd/api  # compile the binary
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate  # regenerate query code
make sonar                             # testes + análise no SonarQube (requer .env)
```

## API

- `GET /health` — liveness probe.
- `GET /api/v1/items` / `POST /api/v1/items` — example resource; replace with your own models, repositories, and handlers.

## Project layout

```
cmd/api/             entrypoint: config -> db -> repository -> service -> handler
internal/config/     environment configuration (PORT, APP_NAME)
internal/handler/    HTTP handlers (framework-specific)
internal/service/    business logic
internal/repository/ data access per ORM
internal/model/      domain models
internal/db/         connection + connection pool
Makefile             convenience targets (build, run, test, vet, fmt, generate, migrate-up/down)
```

## Environment

Defaults are baked in, so the app runs with no setup. Override them with environment variables:

- `PORT` — HTTP port (default `3000`)
- `APP_NAME` — application name
- `DATABASE_URL` — database DSN; no host use `localhost:5433` (o Compose publica `5433:5432`); veja `.env.example`

O `.env` é lido pelo Docker Compose e pode ser carregado no shell para o app
(Go não lê `.env` sozinho):

```bash
# macOS / Linux
set -a && . ./.env && set +a

# Windows PowerShell
Get-Content ./.env | Where-Object { $_ -match '=' -and $_ -notmatch '^\s*#' } | ForEach-Object { $n, $v = $_ -split '=', 2; Set-Item -Path "env:$n" -Value $v }
```


## SonarQube + sonar-scanner

A análise estática roda em um SonarQube local via Docker Compose. O Sonar usa
um banco dedicado (`sonarqube`, usuário `sonar`), separado do banco da
aplicação (`my_screening_app`).

1. Crie o `.env` local (ignorado pelo git) e preencha as senhas:

   ```bash
   cp .env.example .env
   # edite POSTGRES_PASSWORD, SONAR_DB_PASSWORD e SONAR_TOKEN
   ```

   O `POSTGRES_PASSWORD` deve ser o mesmo usado em `DATABASE_URL` no `.env`.

2. Suba o banco e o SonarQube:

   ```bash
   docker compose up -d db sonarqube
   ```

   O host publica a porta **5433**; dentro da rede do Compose o Postgres
   continua em **5432** (é a porta que o SonarQube usa). Não altere o
   `SONAR_JDBC_URL` para 5433.

3. No primeiro acesso ao dashboard (`http://localhost:9001`), entre com
   `admin`/`admin`, troque a senha e gere um token em
   **My Account > Security**. Coloque o token em `SONAR_TOKEN` no `.env`
   (ou no secret do CI).

4. Aplique as migrações no banco da aplicação (o banco do Sonar não é usado
   pelo app):

   ```bash
   set -a && . ./.env && set +a
   export MIGRATE_URL="postgres://postgres:${POSTGRES_PASSWORD}@localhost:5433/my_screening_app?sslmode=disable"
   make migrate-up
   ```

5. Rode os testes + análise:

   ```bash
   make sonar
   ```

O `make sonar` gera `coverage.out`, sobe as dependências se necessário, espera
o healthcheck do SonarQube e roda o `sonar-scanner` na rede do Compose. O
dashboard fica em `http://localhost:9001/dashboard?id=my_screening_app`.

No Azure Pipelines, o estágio `SonarAnalysis` é opcional: defina
`SONAR_HOST_URL` e `SONAR_TOKEN` (secret) nas variáveis do pipeline. Ele roda
com `-Dsonar.qualitygate.wait=true`, ou seja, reprova o build se o quality
gate falhar. A URL precisa ser acessível pelo agente do pipeline.

> O gate padrão do Sonar exige 80% de cobertura no código novo. Com a
> cobertura atual (~0,9%), ajuste as condições do gate ou suba os testes
> antes de tornar o estágio obrigatório.

Observações:

- `reports/**` e `internal/db/sqlc/**` ficam fora da análise (artefatos e
  código gerado).
- No Apple Silicon a imagem do scanner é `linux/amd64` e roda em emulação; se
  o parser Go falhar em algum arquivo, rode o estágio do CI (amd64) ou
  instale o `sonar-scanner` nativo.
- Em um volume Postgres já existente, o script `db/init/01-sonarqube.sh` não
  roda. Crie o banco/usuário manualmente ou recrie o volume (`docker compose
  down -v`).


# Gate Usage

docker build --target gate --build-arg GATE_TARGET=generate-check --progress=plain --no-cache-filter gate .


docker build --target gate --build-arg GATE_TARGET=lint --progress=plain --no-cache .

# lista dos achados
docker build --target gate --progress=plain . 2>&1 | tee lint.log

# contagem por linter
grep -oE '\([a-z]+\)$' lint.log | sort | uniq -c | sort -rn

# lista dos achados
grep -E '\((revive|gosec|errcheck|staticcheck|gocritic|errorlint|govet)\)$' lint.log


axel@NB1395:/mnt/c/Dev/go/go-screening-app-infra$ make sonar
docker compose up -d db sonarqube
[+] Running 2/2
 ✔ Container my_screening_app-db-1         Healthy                                                                                         0.6s
 ✔ Container my_screening_app-sonarqube-1  Running                                                                                         0.0s
Aguardando SonarQube ficar saudável em http://localhost:9001...
SonarQube está UP.
go test ./... -coverprofile=coverage.out
make: go: No such file or directory