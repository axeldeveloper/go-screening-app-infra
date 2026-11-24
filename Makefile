GOLANGCI_LINT_VERSION ?= v2.15.0
# Usa o binário do asdf/PATH se existir; senão compila a versão fixada (Docker/CI sem o binário)
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || echo go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))
LINTREPORT_FLAGS ?=
GOVULNCHECK_VERSION   ?= v1.1.4
SQLC_VERSION          ?= v1.27.0
COVER_MIN             ?= 30

MIGRATE_TAGS := postgres
# Defina MIGRATE_URL no ambiente (ver env.example). Não versione credenciais reais.
MIGRATE_URL ?=

.PHONY: build run test vet fmt generate generate-check lint vuln cover gate gate-docker \
	lint-report quality-report sonar check-migrate-url migrate-up migrate-down

build: generate
	go build -o ./bin/my_screening_app ./cmd/api

run: generate
	go run ./cmd/api

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

generate:
	CGO_ENABLED=0 go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

# ---------------------------------------------------------------
# Quality gate (determinístico) — o mesmo comando roda local e no CI
# ---------------------------------------------------------------
generate-check: generate
	git diff --exit-code -- internal/db/sqlc

lint:
	$(GOLANGCI_LINT) run ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Cobertura total: $$total% (minimo: $(COVER_MIN)%)"; \
	awk -v t="$$total" -v m="$(COVER_MIN)" 'BEGIN{ exit (t+0 < m+0) }' || { echo "Cobertura abaixo do minimo"; exit 1; }

gate: generate-check lint vuln cover

# Roda o gate dentro do Docker (estágio "gate" do Dockerfile)
gate-docker:
	docker build --target gate --progress=plain .

# ---------------------------------------------------------------
# Relatório legível (reports/quality.html) com comparação entre execuções
# ---------------------------------------------------------------
lint-report:
	@mkdir -p reports
	@rm -f reports/lint.json
	@$(GOLANGCI_LINT) run ./... \
	  --output.json.path=reports/lint.json \
	  --output.junit-xml.path=reports/lint-junit.xml \
	  --output.checkstyle.path=reports/lint-checkstyle.xml \
	  >/dev/null; code=$$?; \
	if [ $$code -ne 0 ] && [ $$code -ne 1 ]; then \
	  echo "golangci-lint falhou (exit $$code): verifique configuracao e versao do Go"; exit $$code; \
	fi
	@go run ./tools/lintreport -in reports/lint.json -out reports $(LINTREPORT_FLAGS)

quality-report: lint-report

# ---------------------------------------------------------------
# Servir o relatório de qualidade localmente
# ---------------------------------------------------------------
# Depois de rodar `make lint-report`, use:
#   make reports-serve        # serve http://localhost:8080/reports/
#   make reports-open         # abre no navegador (macOS)
# ---------------------------------------------------------------
REPORTS_PORT ?= 8080

reports-serve: lint-report
	@echo "Servindo relatorios em http://localhost:$(REPORTS_PORT)/quality.html"
	@cd reports && python3 -m http.server $(REPORTS_PORT)

reports-open: reports-serve
	@open http://localhost:$(REPORTS_PORT)/quality.html 2>/dev/null || \
		xdg-open http://localhost:$(REPORTS_PORT)/quality.html 2>/dev/null || \
		echo "Abra manualmente: http://localhost:$(REPORTS_PORT)/quality.html"

# ---------------------------------------------------------------
# SonarQube + sonar-scanner
# ---------------------------------------------------------------
	go test ./... -coverprofile=coverage.out
	docker compose --profile tools run --rm -T sonar-scanner


# ---------------------------------------------------------------
# Migrações
# ---------------------------------------------------------------
check-migrate-url:
	@test -n "$(MIGRATE_URL)" || { echo "Defina MIGRATE_URL, ex.: export MIGRATE_URL='postgres://usuario:senha@localhost:5432/my_screening_app?sslmode=disable'"; exit 1; }

migrate-up: check-migrate-url
	go run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database '$(MIGRATE_URL)' up

migrate-down: check-migrate-url
	go run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database '$(MIGRATE_URL)' down
