#!/bin/bash
set -e

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
wait-for-it.sh postgres:5432 -t 60

# Connect to default postgres database first to create our database and user
echo "Setting up initial database and user..."
PGPASSWORD=postgres psql -h postgres -U postgres -d postgres -f /app/migrations/000001_init_database.up.sql

# Now run the schema and data migrations
echo "Running schema migrations..."
migrate -path /app/migrations -database "postgres://xyz_user:xyz_password@postgres:5432/xyz_db?sslmode=disable" up

# Start the application
echo "Starting the application..."
exec ./main 