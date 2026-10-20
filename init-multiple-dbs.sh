#!/bin/bash
set -euo pipefail

# Инициализация выполняется PostgreSQL только при создании нового тома.
IFS=',' read -ra databases <<< "${POSTGRES_MULTIPLE_DATABASES:-}"
for database in "${databases[@]}"; do
    [[ "$database" =~ ^[a-z][a-z0-9_]*$ ]] || { echo "Invalid database name" >&2; exit 1; }
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres --set "database=$database" <<'SQL'
CREATE DATABASE :"database";
SQL
    for migration in "/migrations/$database/"*.up.sql; do
        [ -f "$migration" ] || continue
        psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$database" --file "$migration"
    done
done
