-- Create user if not exists
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_user WHERE usename = 'xyz_user') THEN
    CREATE USER xyz_user WITH PASSWORD 'xyz_password';
  END IF;
END
$$;

-- Grant privileges
ALTER USER xyz_user WITH LOGIN;
GRANT ALL PRIVILEGES ON DATABASE xyz_db TO xyz_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO xyz_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO xyz_user;

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; 