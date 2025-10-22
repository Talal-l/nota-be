#!/bin/bash
set -e

echo "=========================================="
echo "PostgreSQL User Initialization"
echo "=========================================="
echo "Admin User: $POSTGRES_USER (superuser for migrations)"
echo "App User: $DB_USER (limited user for normal operations)"
echo "Database: $POSTGRES_DB"
echo "=========================================="

# Create the application user with limited privileges
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create the application user if it doesn't exist
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '$DB_USER') THEN
            CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
            RAISE NOTICE 'User $DB_USER created successfully';
        ELSE
            RAISE NOTICE 'User $DB_USER already exists';
        END IF;
    END
    \$\$;

    -- Grant connect privilege to the database
    GRANT CONNECT ON DATABASE $POSTGRES_DB TO $DB_USER;

    -- Grant usage on the public schema
    GRANT USAGE ON SCHEMA public TO $DB_USER;

    -- Grant SELECT, INSERT, UPDATE, DELETE on all existing tables
    -- (excludes DROP, TRUNCATE, and REFERENCES)
    GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO $DB_USER;

    -- Grant usage and SELECT on all sequences (for auto-increment columns)
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;

    -- Grant execute on all functions
    GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO $DB_USER;

    -- Set default privileges for future tables created by the admin
    -- This ensures new tables will automatically grant privileges to the app user
    ALTER DEFAULT PRIVILEGES FOR ROLE $POSTGRES_USER IN SCHEMA public
        GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO $DB_USER;

    ALTER DEFAULT PRIVILEGES FOR ROLE $POSTGRES_USER IN SCHEMA public
        GRANT USAGE, SELECT ON SEQUENCES TO $DB_USER;

    ALTER DEFAULT PRIVILEGES FOR ROLE $POSTGRES_USER IN SCHEMA public
        GRANT EXECUTE ON FUNCTIONS TO $DB_USER;

    -- Create commonly used extensions
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
    CREATE EXTENSION IF NOT EXISTS "pg_trgm";

    -- Display summary
    SELECT
        'User Permissions Summary' as info,
        usename as username,
        usesuper as is_superuser,
        usecreatedb as can_create_db
    FROM pg_user
    WHERE usename IN ('$POSTGRES_USER', '$DB_USER')
    ORDER BY usesuper DESC;

EOSQL

echo "=========================================="
echo "✓ User setup completed successfully!"
echo "=========================================="
echo "Admin user ($POSTGRES_USER):"
echo "  - Full superuser privileges"
echo "  - Use for migrations and schema changes"
echo ""
echo "App user ($DB_USER):"
echo "  - SELECT, INSERT, UPDATE, DELETE on tables"
echo "  - USAGE, SELECT on sequences"
echo "  - EXECUTE on functions"
echo "  - NO DROP, TRUNCATE, or GRANT privileges"
echo "  - Use for normal application operations"
echo "=========================================="
