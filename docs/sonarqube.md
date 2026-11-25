# Guia do SonarQube — my-screening-app

Este documento explica como usar o SonarQube que já está orquestrado no `docker-compose.yml`: acesso, criação de token, execução do scanner, onde analisar vulnerabilidades e qualidade de código Go, e os novos alvos do `Makefile`.

---

## 1. Visão geral da arquitetura

O `docker-compose.yml` sobe containers separados na rede `sonarnet`:

| Serviço | Porta | Propósito |
|---------|-------|-----------|
| `db` (Postgres) | 5433 (host) / 5432 (rede interna) | Banco da aplicação e banco dedicado do SonarQube |
| `sonarqube` | 9001 | Servidor de análise de código |
| `sonar-scanner` | profile `tools` | Cliente que envia a análise ao servidor |

**O scanner não sobe automaticamente** com `docker compose up`. Ele está no profile `tools` propositalmente: o SonRQube é pesado (Java + Elasticsearch) e o scanner consome muita memória. Você dispara a análise sob demanda, quando quiser.

---

## 2. Acessando o SonarQube

URL: `http://localhost:9001`

Login padrão: `admin` / `admin` (altere a senha no primeiro acesso).

---

## 3. Criando o token de autenticação

O scanner se autentica via **token**, não com usuário/senha.

1. No SonarQube, clique no seu usuário (canto superior direito) → **My Account**.
2. Vá na aba **Security**.
3. Em **Generate tokens**, dê um nome (ex: `my-screening-app-scanner`) e clique em **Generate**.
4. Copie o valor exibido — ele aparece apenas uma vez.

No seu `.env` local (não versionado, já está no `.gitignore`):

```env
SONAR_TOKEN=sua_token_copiada_aqui
```

---

## 4. É preciso criar um projeto?

**Sim, mas é automático.**

O arquivo `sonar-project.properties` já existe na raiz e define:

```properties
sonar.projectKey=my_screening_app
sonar.projectName=my_screening_app
sonar.projectVersion=1.0
```

Esse `projectKey` é o identificador único no SonarQube. Na **primeira execução do scanner**, o servidor cria o projeto automaticamente com essa chave. Você não precisa criar nada manualmente na UI antes de rodar o scanner.

---

## 5. Executando a análise (fluxo profissional)

### Passo a passo

```bash
# 1. Sobe banco de dados e SonarQube (scanner não sobe sozinho)
make sonar-up

# 2. Aguarda o SonarQube ficar saudável (pode demorar 2-3 minutos na primeira vez)
make sonar-wait

# 3. Verifica status manualmente (opcional)
curl http://localhost:9001/api/system/status

# 4. Roda a análise de código
make sonar-scan
```

Ou tudo de uma vez (testes + cobertura + scanner + link do dashboard):

```bash
make sonar
```

### O que acontece durante a execução

1. O container do scanner monta o repositório local em `/usr/src`.
2. Lê `sonar-project.properties`.
3. Conecta ao SonarQube via rede interna (`http://sonarqube:9001`).
4. Analisa arquivos Go, YAML, Docker, etc.
5. Envia o relatório para o servidor.
6. Exibe: `ANALYSIS SUCCESSFUL` e um link tipo:

```
http://sonarqube:9001/dashboard?id=my_screening_app
```

No seu navegador (host), acesse: `http://localhost:9001/dashboard?id=my_screening_app`

---

## 6. Onde analisar vulnerabilidades e qualidade

Depois da análise processada (alguns segundos após o `EXECUTION SUCCESS`):

| O que quer ver | Caminho no SonarQube |
|----------------|----------------------|
| Dashboard geral (bugs, vulnerabilidades, code smells, coverage) | `Dashboard` → projeto `my_screening_app` |
| Vulnerabilidades de segurança | `Security` → `Vulnerabilities` |
| Code smells / manutenibilidade | `Code` → `Issues` |
| Duplicação de código | `Code` → `Duplications` |
| Cobertura de testes | `Measures` → `Coverage` (requer `coverage.out`) |
| Quality Gate (pass/fail) | `Quality Gate` no topo do dashboard |

---

## 7. Logs do scanner e do servidor

```bash
# Logs do último run do scanner
docker compose --profile tools logs sonar-scanner

# Logs do servidor SonarQube (útil se o scanner falhar ao conectar)
docker logs my_screening_app-sonarqube-1 --tail=50
```

---

## 8. Linter local (independente do SonarQube)

O SonarQube faz análise estática, mas no ecossistema Go o padrão é rodar localmente antes de commitar:

```bash
# Vet nativo
go vet ./...

# Linter completo (mesmo do Makefile)
golangci-lint run ./...

# Via container (sem instalar nada no host)
docker run --rm -v "$(pwd):/app" -w /app golang:1.25-alpine \
  sh -c "apk add --no-cache git && GOTOOLCHAIN=auto go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./..."

# Relatório legível (HTML/JUnit)
make lint-report
# abre: http://localhost:8080/quality.html
```

---

## 9. Alvos do Makefile (SonarQube)

Os seguintes alvos foram adicionados/ajustados no `Makefile`:

| Alvo | Descrição |
|------|-----------|
| `make sonar-up` | Sobe apenas `db` e `sonarqube` |
| `make sonar-wait` | Aguarda o SonarQube reportar `status: UP` (máx. 60 tentativas de 5s) |
| `make sonar-scan` | Executa o scanner sob demanda após o servidor estar UP |
| `make sonar` | **Fluxo principal**: `go test ./... -coverprofile=coverage.out` + scanner + link do dashboard |
| `make sonar-open` | Abre o dashboard no navegador |

Exemplos de uso:

```bash
make sonar               # testes + análise + dashboard
make sonar-open          # apenas abre o navegador
make sonar SONAR_URL=http://localhost:9001  # sobrescreve URL, se necessário
```

---

## 10. Dicas profissionais para Go + SonarQube

### Código gerado pelo sqlc (não analisar)

`sonar-project.properties` já exclui:

```properties
sonar.exclusions=**/vendor/**,**/*.pb.go,reports/**,**/reports/**,internal/db/sqlc/**,coverage.out,.scannerwork/**,bin/**,**/.DS_Store,internal/web/templates/**,internal/web/static/**
```

Mantenha `internal/db/sqlc/**` excluído. Código gerado não deve ser analisado — gera falso-positivo e polui o Quality Gate.

### Cobertura de testes

Já configurado:

```properties
sonar.go.coverage.reportPaths=coverage.out
```

Gere antes de escanear, se quiser coverage no Sonar:

```bash
make cover
# ou
go test -race -coverprofile=coverage.out ./...
```

### Testes não devem ser analisados como código-fonte

Já configurado:

```properties
sonar.tests=internal
sonar.test.inclusions=**/*_test.go
```

Testes são medidos separadamente, não como código-fonte produtivo.

### Memória (Docker Desktop Windows/Mac)

Já mitigado no `docker-compose.yml`:

```yaml
SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true
```

E no `sonar-project.properties`:

```properties
sonar.javascript.node.maxspace=1024
```

Isso evita crash do scanner em ambientes com pouca RAM.

### Quality Gate

No SonarQube, vá em **Quality Gates** e garanta que o gate padrão esteja vinculado ao projeto `my_screening_app`. Recomendação profissional:

- Coverage ≥ 80% (ou o que seu time definir)
- Bugs = 0
- Vulnerabilities = 0
- Security Hotspots revisados = 100%

---

## 11. Resumo do fluxo diário

```
1. git clone -> cp .env.example .env -> preencha SONAR_TOKEN
2. make sonar-up            # sobe db + sonarqube (sem scanner)
3. make sonar-wait          # aguarda o SonarQube ficar UP
4. faça suas mudanças de código
5. git commit e push
6. make sonar               # testes + cobertura + análise no SonarQube
7. make sonar-open          # revise o dashboard no navegador
8. revise vulnerabilidades e code smells no SonarQube
9. rode golangci-lint localmente antes de abrir PR
```

---

## 12. Segurança

- O `SONAR_TOKEN` **não deve ser versionado**. Preencha apenas no seu `.env` local.
- O `.env` já está no `.gitignore`.
- Nunca commite tokens, senhas ou credenciais reais no repositório.
