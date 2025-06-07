# Database Migration Documentation

## Overview
This document describes the database migration process for PT XYZ Multifinance's backend service. We use `golang-migrate` for handling database migrations.

## Migration Files Structure
```
migrations/
├── 000001_init_database.up.sql   # Create database, user, and extensions
├── 000001_init_database.down.sql # Revert database initialization
├── 000002_init_schema.up.sql     # Create tables and constraints
├── 000002_init_schema.down.sql   # Drop all tables
├── 000003_insert_dummy_data.up.sql   # Insert initial dummy data
└── 000003_insert_dummy_data.down.sql # Remove dummy data
```

## Migration Steps

### 1. Database Initialization (000001)
- Creates database user `xyz_user` with password
- Grants necessary privileges
- Creates required PostgreSQL extensions:
  * uuid-ossp
  * pgcrypto

### 2. Schema Creation (000002)
Creates the following tables with their constraints:

#### Tables
1. `customers`
   - Primary customer information
   - Unique NIK constraint
   - Soft delete support (deleted_at)

2. `credit_limits`
   - Customer credit limits
   - Tenor constraints (1-4 months)
   - Unique customer_id and tenor combination

3. `transactions`
   - Customer transactions
   - Source validation (e-commerce, website, dealer)
   - Status tracking (pending, approved, rejected, cancelled)
   - Soft delete support

4. `installments`
   - Transaction installments
   - Status tracking (paid, unpaid, overdue)
   - Payment tracking

#### Indexes
- customers: idx_customers_nik
- credit_limits: idx_credit_limits_customer_id
- transactions: idx_transactions_contract_number, idx_transactions_customer_id
- installments: idx_installments_transaction_id, idx_installments_due_date

### 3. Initial Data (000003)
Inserts test data including:
- Sample customers
- Credit limits
- Test transactions
- Installment records

## Running Migrations

### Using Docker

1. Build migration image:
```bash
docker build -t xyz-migrate -f Dockerfile.migrate .
```

2. Run migrations:
```bash
docker run --network golangstudikasusptxyz_default xyz-migrate \
  -path=/migrations/ \
  -database "postgres://xyz_user:xyz_password@postgres:5432/xyz_db?sslmode=disable" \
  up
```

3. Rollback migrations:
```bash
docker run --network golangstudikasusptxyz_default xyz-migrate \
  -path=/migrations/ \
  -database "postgres://xyz_user:xyz_password@postgres:5432/xyz_db?sslmode=disable" \
  down
```

### Local Development

1. Install golang-migrate:
```bash
# MacOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

2. Run migrations:
```bash
migrate -path ./migrations \
  -database "postgres://xyz_user:xyz_password@localhost:5432/xyz_db?sslmode=disable" \
  up
```

3. Rollback migrations:
```bash
migrate -path ./migrations \
  -database "postgres://xyz_user:xyz_password@localhost:5432/xyz_db?sslmode=disable" \
  down
```

## Troubleshooting

### Dirty Database State
If you encounter "dirty database" error:

1. Force a specific version:
```bash
migrate -path ./migrations \
  -database "postgres://xyz_user:xyz_password@localhost:5432/xyz_db?sslmode=disable" \
  force VERSION
```

2. Then run migrations again:
```bash
migrate -path ./migrations \
  -database "postgres://xyz_user:xyz_password@localhost:5432/xyz_db?sslmode=disable" \
  up
```

### Database In Use
If you can't drop the database because it's in use:

1. Stop the application:
```bash
docker-compose stop app
```

2. Drop and recreate the database:
```bash
docker-compose exec postgres psql -U xyz_user -d postgres -c "DROP DATABASE IF EXISTS xyz_db;"
docker-compose exec postgres psql -U xyz_user -d postgres -c "CREATE DATABASE xyz_db;"
```

3. Restart the application:
```bash
docker-compose start app
```

## Database Credentials

### Docker Environment
- Host: postgres
- Port: 5432
- Database: xyz_db
- Username: xyz_user
- Password: xyz_password

### Local Environment
- Host: localhost
- Port: 5432
- Database: xyz_db
- Username: xyz_user
- Password: xyz_password 