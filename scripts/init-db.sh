#!/bin/bash
# Initialize PostgreSQL database with sample data

echo "Initializing PostgreSQL database..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until docker exec supabase-postgres-temp pg_isready -U postgres > /dev/null 2>&1; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is ready!"

# Run the initialization SQL script
echo "Running initialization script..."
docker exec -i supabase-postgres-temp psql -U postgres -d middleware_db < init-postgres.sql

if [ $? -eq 0 ]; then
    echo "✓ Database initialized successfully!"
    echo ""
    echo "Connection details:"
    echo "  Host: localhost"
    echo "  Port: 5432"
    echo "  Database: middleware_db"
    echo "  Username: postgres"
    echo "  Password: postgres"
    echo ""
    echo "DATABASE_URL=postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable"
else
    echo "✗ Failed to initialize database"
    exit 1
fi
