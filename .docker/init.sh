#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  DROP DATABASE IF EXISTS sesamo_test;
  CREATE DATABASE sesamo_test;
EOSQL

for DB in $(psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -t -c "SELECT datname FROM pg_database WHERE datname NOT IN ('postgres', 'template0', 'template1')"); do
    echo $DB
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -d $DB -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
done

