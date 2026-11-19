#!/bin/sh
# Cria o banco dedicado do SonarQube no primeiro start do volume do Postgres.
# Executado apenas quando o diretório de dados está vazio (/docker-entrypoint-initdb.d).
set -eu

: "${SONAR_DB_PASSWORD:?SONAR_DB_PASSWORD is required}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE ROLE sonar LOGIN PASSWORD '${SONAR_DB_PASSWORD}';
    CREATE DATABASE sonarqube OWNER sonar;
EOSQL
