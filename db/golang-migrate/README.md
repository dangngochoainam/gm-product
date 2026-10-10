# Database Migration Guide with Atlas

This guide explains how to use Atlas for database migrations in this project.

## Prerequisites

- [Atlas CLI](https://atlasgo.io/getting-started#installation) installed on your machine or inside the Docker container
- Go toolchain installed (only needed to build the loader binary once)
- PostgreSQL database running
- Environment variables configured in `export-env.local.sh`

## Setup

### 1. Configure Environment Variables

Edit `export-env.local.sh` with your database credentials:

```bash
USER_NAME=root
PASSWORD=Abc12345
HOST=localhost
PORT=5434
DB_NAME=ecommerce
SCHEMA_NAME=migrate_user
```

### 2. Build the Schema Loader

The schema loader is a standalone binary that reads GORM entity definitions and outputs the schema to Atlas. Build it once (and rebuild whenever entities in `db/entities/` change):

```bash
./shmake build
```

This produces an `atlas-loader` binary in the current directory. The binary is not committed to git.

> After adding a new entity to `db/entities/`, register it in `loader/main.go` and run `./shmake build` again.

## Available Commands

The `shmake` script provides wrapper functions for all operations:

| Command | Description |
|---------|-------------|
| `./shmake build` | Build the `atlas-loader` binary from GORM entities |
| `./shmake diff <name>` | Generate a new migration by diffing entities vs database |
| `./shmake apply` | Apply all pending migrations to the database |

## Workflow

### Making Schema Changes

1. Modify GORM entity structs in `db/entities/`
2. If a new entity was added, register it in `loader/main.go`
3. Rebuild the loader: `./shmake build`
4. Generate the migration: `./shmake diff <migration_name>`
5. Review the generated files in `migrations/`
6. Apply: `./shmake apply`

### Example

```bash
# 1. Edit db/entities/product.go to add a new field
# 2. Rebuild the loader
./shmake build

# 3. Generate migration
./shmake diff add_product_description

# 4. Review the generated migration
cat migrations/<timestamp>_add_product_description.up.sql

# 5. Apply
./shmake apply
```

## File Structure

```
db/golang-migrate/
├── README.md                # This guide
├── atlas.hcl                # Atlas configuration
├── export-env.local.sh      # Database credentials (not committed)
├── shmake                   # Helper script
├── atlas-loader             # Compiled schema loader binary (not committed)
├── loader/                  # Loader source (separate Go module)
│   ├── go.mod
│   ├── go.sum
│   └── main.go
└── migrations/              # Generated migration files
    ├── <timestamp>_init.up.sql
    └── <timestamp>_init.down.sql
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `USER_NAME` | Database username |
| `PASSWORD` | Database password |
| `HOST` | Database host |
| `PORT` | Database port |
| `DB_NAME` | Database name |
| `SCHEMA_NAME` | Schema name |
| `DB_CONN_STRING` | Full connection string (auto-generated) |

## Troubleshooting

### `atlas-loader: no such file or directory`

The binary has not been built yet. Run `./shmake build`.

### Connection Issues

- Verify your database is running
- Check credentials in `export-env.local.sh`
- Ensure the database and schema exist

### Migration Issues

- Review generated migration files before applying
- Check Atlas logs for detailed error messages

### New entity not reflected in migration

Ensure the entity is registered in `loader/main.go` inside the `Load(...)` call, then run `./shmake build` before `./shmake diff`.

### Reset Migrations

```bash
# 1. Drop and recreate your database schema
# 2. Remove all files from migrations/
# 3. Regenerate
./shmake diff initial
./shmake apply
```

## Resources

- [Atlas Documentation](https://atlasgo.io/)
- [Golang Migrate Documentation](https://github.com/golang-migrate/migrate)
- [Atlas GORM Provider](https://atlasgo.io/guides/orms/gorm)
