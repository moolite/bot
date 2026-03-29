# AGENTS.md

Guide for AI agents working in the marrano-bot codebase.

## Project Overview

marrano-bot is a Telegram bot written in Go that provides media management, callouts, dice rolling, and other chat utilities. It receives updates via webhooks and stores data in SQLite.

## Essential Commands

### Build

```bash
# Build the binary (requires sqlite_foreign_keys build tag)
go build -tags "sqlite_foreign_keys" -v ./cmd/marrano-bot

# Or use make
make marrano-bot

# Watch mode (requires fd and entr)
make watch
```

### Test

```bash
# Run all tests
go test -v ./...

# Watch mode
make test-watch
```

### Nix

```bash
# Build with nix
nix build .#default

# Enter dev shell
nix develop
```

### Database

```bash
# Initialize database
./marrano-bot -I -c config.toml

# Export database to CSV
./marrano-bot -E -c config.toml
```

## Code Organization

```
bot/
├── cmd/marrano-bot/      # Main application entrypoint
│   ├── main.go           # CLI flags, config loading, startup
│   └── downloader.go     # Media sync functionality
├── internal/
│   ├── config/           # Configuration loading (TOML + env vars)
│   ├── core/             # Bot core logic, HTTP server, handlers
│   │   ├── core.go       # HTTP server setup with chi router
│   │   └── handlers.go   # All Telegram command/message handlers
│   ├── db/               # Database layer (SQLite with sqlx)
│   │   ├── db.go         # Connection management, prepared statements
│   │   ├── migrations/   # SQL migration files (embedded)
│   │   └── *.go          # Domain-specific DB operations
│   ├── dicer/            # Dice rolling logic
│   ├── statistics/       # Metrics collection and prometheus export
│   ├── statsd/           # StatsD client
│   ├── telegram/         # Telegram-specific utilities (HTML, entities)
│   └── utils/            # Shared utilities (string helpers)
└── pkg/tg/               # Custom Telegram bot framework
    ├── bot.go            # Bot core, handler registration, middleware
    ├── client.go         # HTTP client for Telegram API
    ├── methods.go        # API method constants
    └── types.go          # Telegram API types
```

## Architecture

### Telegram Bot Framework (`pkg/tg/`)

Custom lightweight framework instead of using external libraries:

- **Handler Types**: `UPD_STARTSWITH`, `UPD_CONTAINS`, `UPD_REGEXP`, `UPD_CALLBACK`, `UPD_WILDCARD`, etc.
- **Handler Registration**: `b.RegisterHandlers(&tg.UpdateHandler{Type, Param, Aliases, Fn})`
- **Middleware**: `b.RegisterMiddlewares(fn)` - middleware receives/returns `*Update`
- **Sendable**: All responses use `*tg.Sendable` struct with `Method` field

### Database Layer (`internal/db/`)

- Uses `sqlx` with SQLite (mattn/go-sqlite3, CGO required)
- Prepared statements cached in `stmts` map
- Migrations embedded with `//go:embed migrations/*.sql`
- Connection string: `file:<path>?cache=private&mode=rw&_txlock=immediate&_journal_mode=WAL`

### HTTP Server (`internal/core/core.go`)

- Chi router with httplog middleware
- Webhook endpoint: `POST /t/{apikey}` where apikey must match config
- Health endpoints: `/ping`, `/health`, `/stats`, `/stats.json`

## Code Style

- **Indentation**: Tabs for Go files (4-space tabs), spaces for Nix/JSON/SQL
- **Logging**: Use `log/slog` with structured logging
- **Testing**: Use `github.com/matryer/is` for assertions
- **Errors**: Return errors up the stack, log at the handler level
- **Context**: Pass `context.Context` as first parameter to functions that need it

## Configuration

Config file is TOML format:

```toml
database = "marrano-bot.sqlite"
port = 6446

[telegram]
name = "marrano-bot"
token = "123456789-bot"
domain = "bot.marrani.lol"
apikey = "secret-key-for-webhook"
```

Environment variables override config:
- `DATABASE` - database file path
- `PORT` - HTTP server port
- `TELEGRAM_TOKEN` - bot token
- `TELEGRAM_KEY` - API key for webhook validation
- `LOG_LEVEL` - debug/info/warn/error

## Adding a New Command

1. Add handler function in `internal/core/handlers.go`:
   ```go
   func MyCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
       // ...
   }
   ```

2. Register in `registerBotHandlers()`:
   ```go
   &tg.UpdateHandler{
       Type:    tg.UPD_STARTSWITH,
       Param:   "/mycommand",
       Aliases: []string{"/alias1", "/alias2"},
       Fn:      MyCommand,
   }
   ```

3. Add to `registerCommands()` for bot command menu

4. Add tests in `handlers_test.go`

## Database Migrations

Migrations are numbered SQL files in `internal/db/migrations/`:
- `NN_description.up.sql` - Apply migration
- `NN_description.down.sql` - Rollback migration

Migrations are auto-applied on startup via `db.Migrate()`.

## Key Dependencies

- `github.com/go-chi/chi/v5` - HTTP routing
- `github.com/jmoiron/sqlx` - SQL with named params
- `github.com/mattn/go-sqlite3` - SQLite driver (CGO)
- `github.com/golang-migrate/migrate/v4` - DB migrations
- `github.com/pelletier/go-toml` - Config parsing
- `github.com/lmittmann/tint` - Colored terminal logging
- `github.com/matryer/is` - Test assertions

## Gotchas

- **CGO Required**: sqlite3 driver requires CGO; ensure CC is set in nix shell
- **Build Tags**: Must use `-tags "sqlite_foreign_keys"` when building
- **Nix Build Tags**: Flake uses `-tags "fts5"` for FTS5 support
- **Inline Keyboard Callbacks**: Use `formatCallbackData()` with hex encoding
- **Handler Return**: Return `nil, nil` to indicate "no response needed"
- **Message Reactions**: Use `tg.SendableSetMessageReaction(update, emoji)` for reactions
