# Auth Scripts

This directory contains utility scripts for the Auth service.

## Scripts

### create_table.go
Creates the token blacklist table in the database.

### init_db.go
Initializes the Auth service database with required tables and creates an admin user.

### temp_hash.go
Utility script to generate password hashes.

### test_admin_password.go
Script to test and update the admin user password.

## Usage

All scripts use the `DATABASE_URL` environment variable or default to `postgres://postgres:password@localhost/taishanglaojun?sslmode=disable`.

```bash
# Run a script
cd scripts/auth
go run create_table.go
```

## Notes

- These scripts were moved from `services/auth/cmd` to this location as part of the project structure reorganization.
- The `create_table.go` script references the SQL file at `../../services/auth/migrations/create_blacklist_table.sql`.