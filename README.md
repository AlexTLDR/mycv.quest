# MyCV Quest

A CV builder website built with Go, Templ, and PostgreSQL.

## Prerequisites

- Go 1.24+
- [Task](https://taskfile.dev/installation/) (install as `go-task`)
- [Templ](https://templ.guide/quick-start/installation)
- PostgreSQL (for production) or Docker (for development)

## Development Setup

### 1. Start PostgreSQL Test Instance

```bash
docker run --name mycv-postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:latest
```

### 2. Set Environment Variables

```bash
export DB_DSN="postgres:password@localhost:5432/postgres?sslmode=disable"
```

### 3. Available Commands

View all available tasks:
```bash
go-task --list
```

**Development:**
- `go-task build` - Build the application (automatically generates templ files)
- `go-task run` - Build and run the application
- `go-task dev` - Run with templ watching and live reload
- `go-task test` - Run all tests
- `go-task clean` - Clean build artifacts and generated templ files

**Templ:**
- `go-task templ:generate` - Generate templ files
- `go-task templ:watch` - Watch and regenerate templ files on changes

**Database Migrations:**
- `go-task migrations:new NAME=migration_name` - Create new migration
- `go-task migrations:up` - Apply all migrations
- `go-task migrations:down` - Rollback all migrations
- `go-task migrations:version` - Show current migration version

**Quality Control:**
- `go-task audit` - Run quality control checks
- `go-task test:cover` - Run tests with coverage
- `go-task tidy` - Format code and tidy modules

## Project Structure

```
├── assets/
│   ├── emails/templ/          # Email templates (templ)
│   ├── static/                # Static assets
│   └── templates/templ/       # Web templates (templ)
├── cmd/web/                   # Main application
├── internal/
│   ├── response/              # HTTP response helpers
│   ├── smtp/                  # Email functionality
│   └── version/               # Version information
└── Taskfile.yml              # Task runner configuration
```

## Technology Stack

- **Backend:** Go with standard library HTTP router
- **Templates:** [Templ](https://templ.guide/) for type-safe templates
- **Database:** PostgreSQL with migrations
- **Email:** SMTP with templ-based email templates
- **Build:** Task runner for development workflow
