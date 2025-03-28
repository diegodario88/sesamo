#!/bin/bash
set -e

echo "Initializing databases..."

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  DROP DATABASE IF EXISTS sesamo_test;
  CREATE DATABASE sesamo_test;
  CREATE EXTENSION IF NOT EXISTS "pg_cron";
EOSQL

echo "Listing databases..."
DATABASES=$(psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -t -c "SELECT datname FROM pg_database WHERE datname NOT IN ('postgres', 'template0', 'template1')")

for DB in $DATABASES; do
    echo "Processing database: $DB"

    echo "Checking pgx_ulid extension availability..."
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" -c '\dx pgx_ulid'

    echo "Creating extensions in $DB..."
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" -c '
        CREATE EXTENSION IF NOT EXISTS "ulid";
        CREATE EXTENSION IF NOT EXISTS "citext";
        CREATE EXTENSION IF NOT EXISTS "hstore";
    '

    echo "Verifying extensions in $DB..."
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" -c '\dx'
done

echo "Database initialization complete."
