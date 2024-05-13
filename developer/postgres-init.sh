#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER yhs WITH PASSWORD 'yhs';
    CREATE DATABASE yhs;
    ALTER DATABASE yhs OWNER TO yhs;
    GRANT ALL PRIVILEGES ON DATABASE yhs TO yhs;
EOSQL
