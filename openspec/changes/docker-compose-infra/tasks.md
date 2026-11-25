# Tasks

Assunção registrada (decisão de implementação, não muda os specs): testes (`make test`), `vet` e `golangci-lint` são executados em container via `docker compose run` sobre a imagem de build/gate do `Dockerfile`, sem adicionar serviços permanentes ao `docker-compose.yml`. Se o time preferir serviços dedicados (`test`, `quality`), ajuste este plano e os tasks do grupo 5.

## 1. Levantamento e ambiente

- [x] 1.1 Revisar `docker-compose.yml`, `Dockerfile`, `.env.example` e `db/init/01-sonarqube.sh` e confirmar o estado atual (serviços `app` e `migrate` comentados; healthchecks de `db` e `sonarqube` existentes) — verifique lendo os arquivos e executando `docker compose config`.
- [x] 1.2 Confirmar que Docker e Docker Compose estão instalados e funcionam no ambiente — verifique com `docker version` e `docker compose version`.

## 2. Banco de dados e migrações

- [x] 2.1 Reescrever/ativar o serviço `db` no `docker-compose.yml` com rede interna, healthcheck (`pg_isready -U postgres -d my_screening_app -h 127.0.0.1`) e volumes já existentes (`pgdata`, `db/init`) — verifique com `docker compose up -d db` e `docker compose ps`.
- [x] 2.2 Adicionar o serviço `migrate` que depende de `db:service_healthy` e roda `migrate up` contra `db:5432/my_screening_app` com `MIGRATE_URL` do `.env` — verifique executando `docker compose up migrate` e confirmando migração aplicada com sucesso.
- [x] 2.3 Fazer o serviço `app` depender de `db:service_healthy` e `migrate:service_completed_successfully` — verifique a ordem de inicialização com `docker compose up -d`.

## 3. Serviço da aplicação

- [x] 3.1 Ativar o serviço `app` no `docker-compose.yml`, construído a partir do `Dockerfile`, expondo a porta 8000 e lendo `DATABASE_URL` e `PORT` do `.env` — verifique com `docker compose up -d app`.
- [x] 3.2 Confirmar que a API responde apenas após banco e migrações estarem prontos — verifique com `curl http://localhost:8000/health` retornando `{"status":"ok"}`.

## 4. SonarQube e scanner

- [x] 4.1 Revisar serviço `sonarqube` e healthcheck já existentes (porta 9001, banco dedicado via `db`) — verifique `curl http://localhost:9001/api/system/status` retornando `{"status":"UP"}`.
- [x] 4.2 Confirmar que `sonar-scanner` roda sob demanda pelo profile `tools` — verifique `docker compose --profile tools run --rm -T sonar-scanner` concluindo a análise do projeto.

## 5. Testes e qualidade em container

- [x] 5.1 Rodar os testes da aplicação em container — verifique a execução de `go test -race ./...` via `docker compose run` sobre a imagem gate/builder com resultado de saída zero.
- [x] 5.2 Rodar `go vet ./...` em container — verifique conclusão sem falhas.
- [x] 5.3 Rodar `golangci-lint run ./...` em container — verifique conclusão sem falhas de lint.
- [x] 5.4 Rodar o quality gate completo (`make gate`: generate-check, lint, vuln, cover) em container — verifique `gate-docker` ou equivalente passando.

## 6. Logs e validação final

- [x] 6.1 Confirmar que os logs de cada serviço estão acessíveis — verifique `docker logs <serviço>` para `app`, `db`, `sonarqube` e `sonar-scanner`.
- [x] 6.2 Validação de ponta a ponta: `docker compose up -d` sobe todos os serviços, `/health` responde, SonarQube fica UP e os logs são consultáveis — verifique com `docker compose ps` e `curl`.

## Workflow follow-up

- Revisar todos os artifacts da change com o solicitante antes de arquivar.
- Arquivar a change quando a revisão estiver concluída (workflow de arquivamento do OpenSpec).
