.PHONY: build run test vet fmt generate migrate-up migrate-down

build: generate
run: generate

build:
	go build -o ./bin/my_tristack_app ./cmd/api

run:
	go run ./cmd/api

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .


generate:
	CGO_ENABLED=0 go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate


MIGRATE_TAGS := postgres
MIGRATE_URL ?= postgres://postgres:postgres@localhost:5432/my_tristack_app?sslmode=disable

migrate-up:
	go run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database '$(MIGRATE_URL)' up

migrate-down:
	go run -tags '$(MIGRATE_TAGS)' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1 -path db/migrations -database '$(MIGRATE_URL)' down
