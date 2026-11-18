# syntax=docker/dockerfile:1

# ---- Builder ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0 generate
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/my_screening_app ./cmd/api
RUN CGO_ENABLED=0 GOBIN=/out go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

# ---- Quality gate: docker build --target gate --progress=plain . ----
FROM golang:1.25 AS gate
WORKDIR /app
ARG COVER_MIN=30
ARG GATE_TARGET=gate

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    git init -q . \
    && git add -A \
    && git -c user.email=gate@local -c user.name=gate commit -qm snapshot \
    && COVER_MIN=$COVER_MIN make $GATE_TARGET

# ---- Runtime (último estágio = alvo padrão do docker build) ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/my_screening_app /usr/local/bin/my_screening_app
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY db/migrations /migrations

EXPOSE 8000
ENTRYPOINT ["/usr/local/bin/my_screening_app"]