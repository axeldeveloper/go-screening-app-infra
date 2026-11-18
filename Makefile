GOLANGCI_LINT_VERSION ?= v2.1.6
GOVULNCHECK_VERSION   ?= v1.1.4
SQLC_VERSION          ?= v1.27.0
COVER_MIN             ?= 30

MIGRATE_TAGS := postgres
# Defina MIGRATE_URL no ambiente (ver env.example). Não versione credenciais reais.
MIGRATE_URL ?= postgres://postgres:postgres@localhost:5432/my_screening_app?sslmode=disable

.PHONY: build run test vet fmt generate generate-check lint vuln cover gate migrate-up migrate-down gate-docker

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
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	@total=$$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Cobertura total: $$total% (minimo: $(COVER_MIN)%)"; \
	awk -v t="$$total" -v m="$(COVER_MIN)" 'BEGIN{ exit (t+0 < m+0) }' || { echo "Cobertura abaixo do minimo"; exit 1; }

gate: generate-check lint vuln cover

migrate-up:
	go run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database '$(MIGRATE_URL)' up

migrate-down:
	go run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database '$(MIGRATE_URL)' down

gate-docker:
	docker build --target gate --progress=plain .