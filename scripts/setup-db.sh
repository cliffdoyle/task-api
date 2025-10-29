#!/bin/bash

# This script assumes psql client is available or can connect directly
# to a local or containerized PostgreSQL instance.

# Environment variables for database connection (match docker-compose.yml)
DB_USER=${POSTGRES_USER:-taskapi}
DB_PASSWORD=${POSTGRES_PASSWORD:-taskapi123}
DB_NAME=${POSTGRES_DB:-taskapi}
DB_HOST=${DB_HOST:-localhost} # Use localhost if running psql from host, or 'postgres' if from another container

echo "Attempting to connect to PostgreSQL at $DB_HOST for database setup..."

# Wait for PostgreSQL to be ready
# We'll try to connect a few times, useful if running immediately after `docker-compose up`
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 2
done

>&2 echo "Postgres is up - executing schema setup"

PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" << EOF
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
EOF

echo "Database setup complete!"