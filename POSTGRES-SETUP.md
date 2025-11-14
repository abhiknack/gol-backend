# PostgreSQL Temporary Database Setup

## Overview

A temporary PostgreSQL database has been added to your development environment for testing and development purposes.

## Connection Details

| Parameter | Value |
|-----------|-------|
| **Host** | `localhost` (from host machine) or `postgres` (from containers) |
| **Port** | `5432` |
| **Database** | `middleware_db` |
| **Username** | `postgres` |
| **Password** | `postgres` |

## Connection Strings

### From Host Machine
```
postgresql://postgres:postgres@localhost:5432/middleware_db
```

### From Docker Containers
```
postgresql://postgres:postgres@postgres:5432/middleware_db
```

### Environment Variables
```env
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=middleware_db
```

## Quick Start

### 1. PostgreSQL is Already Running

The container `supabase-postgres-temp` is running on port 5432.

### 2. Access PostgreSQL CLI

```bash
# Connect to PostgreSQL
docker exec -it supabase-postgres-temp psql -U postgres -d middleware_db
```

### 3. Basic Commands

```sql
-- List all databases
\l

-- Connect to middleware_db
\c middleware_db

-- List all tables
\dt

-- Create a sample table
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO products (name, category, price) VALUES
    ('Milk', 'dairy', 3.99),
    ('Bread', 'bakery', 2.49),
    ('Eggs', 'dairy', 4.99);

-- Query data
SELECT * FROM products;

-- Exit psql
\q
```

## Using with Go Application

### Install PostgreSQL Driver

```bash
go get github.com/lib/pq
# or for pgx (recommended)
go get github.com/jackc/pgx/v5
```

### Example Connection Code

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
)

func main() {
    // Build connection string from environment variables
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("POSTGRES_HOST"),
        os.Getenv("POSTGRES_PORT"),
        os.Getenv("POSTGRES_USER"),
        os.Getenv("POSTGRES_PASSWORD"),
        os.Getenv("POSTGRES_DB"),
    )

    // Connect to database
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Test connection
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Successfully connected to PostgreSQL!")
}
```

## Database Management

### Create a New Database

```bash
docker exec -it supabase-postgres-temp psql -U postgres -c "CREATE DATABASE myapp_db;"
```

### Backup Database

```bash
# Backup to file
docker exec supabase-postgres-temp pg_dump -U postgres middleware_db > backup.sql

# Restore from file
docker exec -i supabase-postgres-temp psql -U postgres middleware_db < backup.sql
```

### View Database Size

```bash
docker exec supabase-postgres-temp psql -U postgres -c "SELECT pg_size_pretty(pg_database_size('middleware_db'));"
```

## Using Database GUI Tools

### pgAdmin
- **Host**: `localhost`
- **Port**: `5432`
- **Username**: `postgres`
- **Password**: `postgres`
- **Database**: `middleware_db`

### DBeaver
- **Connection Type**: PostgreSQL
- **Host**: `localhost`
- **Port**: `5432`
- **Database**: `middleware_db`
- **Username**: `postgres`
- **Password**: `postgres`

### VS Code Extensions
- **PostgreSQL** by Chris Kolkman
- **SQLTools** with PostgreSQL driver

## Container Management

### Start PostgreSQL
```bash
docker-compose up -d postgres
```

### Stop PostgreSQL
```bash
docker-compose stop postgres
```

### Restart PostgreSQL
```bash
docker-compose restart postgres
```

### View Logs
```bash
docker logs supabase-postgres-temp -f
```

### Check Health
```bash
docker exec supabase-postgres-temp pg_isready -U postgres
```

## Data Persistence

PostgreSQL data is stored in a Docker volume: `gol-backend_postgres-data`

### View Volume
```bash
docker volume ls | findstr postgres
```

### Remove Volume (Delete All Data)
```bash
# WARNING: This will delete all data!
docker-compose down -v
```

### Backup Volume
```bash
# Create backup
docker run --rm -v gol-backend_postgres-data:/data -v ${PWD}:/backup alpine tar czf /backup/postgres-backup.tar.gz -C /data .

# Restore backup
docker run --rm -v gol-backend_postgres-data:/data -v ${PWD}:/backup alpine tar xzf /backup/postgres-backup.tar.gz -C /data
```

## Migration Tools

### golang-migrate

```bash
# Install
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create migration
migrate create -ext sql -dir migrations -seq create_products_table

# Run migrations
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable" up

# Rollback
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/middleware_db?sslmode=disable" down
```

### Example Migration Files

**migrations/000001_create_products_table.up.sql:**
```sql
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10, 2),
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
```

**migrations/000001_create_products_table.down.sql:**
```sql
DROP TABLE IF EXISTS products;
```

## Sample Schema for Middleware

```sql
-- Supermarket Products
CREATE TABLE supermarket_products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10, 2),
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Movies
CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    genre VARCHAR(100),
    duration INTEGER,
    rating DECIMAL(3, 1),
    release_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Movie Showtimes
CREATE TABLE showtimes (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id),
    theater VARCHAR(255),
    showtime TIMESTAMP,
    available_seats INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Pharmacy Medicines
CREATE TABLE medicines (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10, 2),
    prescription_required BOOLEAN DEFAULT false,
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO supermarket_products (name, category, price, stock) VALUES
    ('Milk', 'dairy', 3.99, 50),
    ('Bread', 'bakery', 2.49, 100),
    ('Eggs', 'dairy', 4.99, 75);

INSERT INTO movies (title, genre, duration, rating, release_date) VALUES
    ('Action Hero', 'action', 120, 8.5, '2024-01-15'),
    ('Comedy Night', 'comedy', 95, 7.8, '2024-02-01');

INSERT INTO medicines (name, category, price, prescription_required, stock) VALUES
    ('Aspirin', 'pain-relief', 5.99, false, 200),
    ('Amoxicillin', 'antibiotic', 12.99, true, 50);
```

## Troubleshooting

### Connection Refused

**Check if container is running:**
```bash
docker ps | findstr postgres
```

**Check logs:**
```bash
docker logs supabase-postgres-temp
```

**Restart container:**
```bash
docker-compose restart postgres
```

### Port Already in Use

If port 5432 is already in use, change it in docker-compose.yml:
```yaml
postgres:
  ports:
    - "5433:5432"  # Use port 5433 on host
```

Then update your connection string to use port 5433.

### Permission Denied

```bash
# Fix permissions
docker exec -u root supabase-postgres-temp chown -R postgres:postgres /var/lib/postgresql/data
```

### Database Not Found

```bash
# Create the database
docker exec supabase-postgres-temp psql -U postgres -c "CREATE DATABASE middleware_db;"
```

## Security Notes

⚠️ **This is a TEMPORARY database for development only!**

- Default credentials are used (postgres/postgres)
- No SSL/TLS encryption
- Data is stored in a local Docker volume
- **DO NOT use in production**
- **DO NOT store sensitive data**

For production, use:
- Strong passwords
- SSL/TLS connections
- Proper user permissions
- Regular backups
- Supabase or managed PostgreSQL service

## Quick Reference Commands

```bash
# Connect to database
docker exec -it supabase-postgres-temp psql -U postgres -d middleware_db

# Run SQL file
docker exec -i supabase-postgres-temp psql -U postgres -d middleware_db < schema.sql

# Export data
docker exec supabase-postgres-temp pg_dump -U postgres middleware_db > dump.sql

# Check connection
docker exec supabase-postgres-temp pg_isready -U postgres

# View running queries
docker exec supabase-postgres-temp psql -U postgres -c "SELECT * FROM pg_stat_activity;"

# Kill all connections
docker exec supabase-postgres-temp psql -U postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='middleware_db';"
```

## Next Steps

1. ✅ PostgreSQL is running on port 5432
2. Connect using your preferred database tool
3. Create your schema and tables
4. Update your Go application to use PostgreSQL
5. Test your application with the local database

---

**PostgreSQL is ready to use!** 🐘

Access it at: `postgresql://postgres:postgres@localhost:5432/middleware_db`
