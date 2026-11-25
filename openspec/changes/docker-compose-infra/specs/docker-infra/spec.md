# Spec Delta

## Purpose

Define o comportamento observável da infraestrutura de containers (docker-compose) do projeto: sobe API, banco Postgres, SonarQube e scanner, e orquestra testes, linter e logs, sem alterar o comportamento de runtime da API.

## ADDED Requirements

### Requirement: API em container expõe a API
O sistema SHALL levantar o serviço `app` pelo `docker-compose.yml` expondo a API da aplicação na porta 8000 do host.

#### Scenario: Health da API responde via Compose
- **WHEN** `docker compose up -d app` é executado com o banco saudável
- **THEN** `GET http://localhost:8000/health` retorna `{"status":"ok"}`

### Requirement: API depende do banco estar saudável
O sistema SHALL iniciar o serviço `app` apenas depois de o serviço `db` reportar saudável no healthcheck do Compose.

#### Scenario: App aguarda o banco antes de iniciar
- **WHEN** o serviço `db` ainda não passou no healthcheck
- **THEN** o serviço `app` não inicia e aguarda a condição `service_healthy` do banco

### Requirement: Migrações aplicadas antes da API
O sistema SHALL aplicar as migrações do banco (`db/migrations`) antes de iniciar a aplicação, por meio de um serviço dedicado ou do entrypoint.

#### Scenario: Migração roda antes da aplicação ficar disponível
- **WHEN** o Compose inicia os serviços `db` e `migrate`
- **THEN** `migrate up` é executado com sucesso antes do serviço `app` responder requisições

### Requirement: Banco Postgres com healthcheck
O serviço `db` SHALL expor o Postgres na rede interna e declarar um healthcheck que confirme prontidão.

#### Scenario: Healthcheck do banco passa
- **WHEN** o serviço `db` inicia
- **THEN** `pg_isready -U postgres -d my_screening_app` retorna sucesso dentro do período configurado

### Requirement: SonarQube sobe e fica saudável
O serviço `sonarqube` SHALL subir com banco dedicado e reportar `status: UP` em seu healthcheck.

#### Scenario: SonarQube reporta status UP
- **WHEN** o serviço `sonarqube` inicia e depende do `db` saudável
- **THEN** `GET http://localhost:9001/api/system/status` retorna `{"status":"UP"}`

### Requirement: Sonar scanner executável sob demanda
O serviço `sonar-scanner` SHALL continuar disponível apenas pelo profile `tools` e executar a análise do código quando executado.

#### Scenario: Scanner roda sob demanda via profile tools
- **WHEN** `docker compose --profile tools run --rm sonar-scanner` é executado com SonarQube saudável
- **THEN** o scanner se conecta a `http://sonarqube:9001` e executa a análise do projeto

### Requirement: Testes da aplicação em container
O sistema SHALL oferecer um serviço ou tarefa de container que rode `go test -race ./...` e relate o resultado.

#### Scenario: Suite de testes executa no container
- **WHEN** o serviço de testes é executado no container da aplicação
- **THEN** `go test -race ./...` é executado e encerra com código de saída zero ou exibe falhas de teste

### Requirement: Qualidade de código (linter e vet) em container
O sistema SHALL oferecer um serviço ou tarefa de container que rode `go vet ./...` e `golangci-lint run ./...`.

#### Scenario: Gate de qualidade executa no container
- **WHEN** o serviço de qualidade é executado no container da aplicação
- **THEN** `go vet ./...` e `golangci-lint run ./...` são executados e relatam o resultado

### Requirement: Logs dos serviços acessíveis
Todos os serviços de container SHALL enviar stdout/stderr para os drivers de log do Docker, consultáveis com `docker logs`.

#### Scenario: Logs de um serviço são consultáveis
- **WHEN** um serviço (app, db, sonarqube ou sonar-scanner) está em execução
- **THEN** `docker logs <nome-do-serviço>` exibe a saída do serviço
