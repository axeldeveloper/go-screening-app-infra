# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/my_tristack_app ./cmd/api
RUN CGO_ENABLED=0 GOBIN=/out go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/my_tristack_app /usr/local/bin/my_tristack_app
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY db/migrations /migrations

EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/my_tristack_app"]
