# AGENTS.md

Guide for AI agents working in the marrano-bot codebase.

## Project Overview

marrano-bot is a Telegram bot written in Go that provides media management, callouts, dice rolling, and other chat utilities. It receives updates via webhooks and stores data in SQLite.

## Essential Commands

```bash
make marrano-bot          # Build (requires -tags "sqlite_foreign_keys")
make watch                # Watch mode (requires fd and entr)
make test / go test -v ./...
make test-watch
nix build .#default       # Nix build (uses -tags "fts5")
nix develop               # Dev shell with CGO toolchain
./marrano-bot -I -c config.toml  # Init database
./marrano-bot -E -c config.toml  # Export to CSV
./marrano-bot -W -c config.toml  # Set webhook
```

## Code Organization

```
bot/
├── cmd/marrano-bot/        # Entrypoint (main.go, downloader.go)
├── internal/
│   ├── chain/              # HTTP middleware chain builder
│   ├── config/             # TOML config + env var overrides
│   ├── core/               # HTTP server (core.go) + handlers (handlers.go)
│   ├── db/                 # SQLite via sqlx (db.go, migrations/, domain ops)
│   ├── dicer/              # Dice rolling
│   ├── llm/                # Ollama integration (client.go, tools.go)
│   ├── statistics/         # Prometheus metrics
│   ├── statsd/             # StatsD client
│   ├── telegram/           # Telegram HTML/entity utilities
│   ├── utils/              # String helpers
│   └── vectorstore/        # In-memory cosine similarity search
└── pkg/tg/                 # Custom Telegram bot framework
    ├── bot.go              # Handler registration, middleware
    ├── client.go           # HTTP client for Telegram API
    ├── methods.go          # API method constants
    └── types.go            # Telegram API types
```

## Architecture

### Telegram Bot Framework (`pkg/tg/`)

Custom lightweight framework. Key types: `UPD_STARTSWITH`, `UPD_CONTAINS`, `UPD_REGEXP`, `UPD_CALLBACK`, `UPD_WILDCARD`, `UPD_MEDIAPHOTO`, `UPD_MEDIAVIDEO`, `UPD_MENTION`.

- Register: `b.RegisterHandlers(&tg.UpdateHandler{Type, Param, Aliases, Fn})`
- Middleware: `b.RegisterMiddlewares(fn)` — receives/returns `*Update`
- All responses: `*tg.Sendable{Method, ChatID, Text, ...}`
- No response: return `nil, nil`
- Reactions: `tg.SendableSetMessageReaction(update, emoji)`

### Database Layer (`internal/db/`)

- `sqlx` + SQLite (mattn/go-sqlite3, CGO required)
- Prepared statements cached via `prepareStmt()` in `stmts` map
- Migrations: `//go:embed migrations/*.sql`, auto-applied via `db.Migrate()`
- Naming: `NN_description.up.sql` / `NN_description.down.sql`
- Connection: `file:<path>?cache=private&mode=rw&_txlock=immediate&_journal_mode=WAL`
- Multi-tenant: always reference `groups(gid)` for foreign keys

### HTTP Server (`internal/core/core.go`)

- Chi router with httplog middleware
- Webhook: `POST /t/{apikey}` (apikey from config)
- Health: `/ping`, `/health`, `/stats`, `/stats.json`

### LLM Integration (`internal/llm/`)

Ollama-based AI responses with function calling. Models in `client.go`, tools in `tools.go`, system prompt in `client.go`. To add a tool: append to `AllTools()` with `Name`, `Description`, `Parameters` (JSON schema), and `Handler`.

## Adding a New Command

1. Handler in `internal/core/handlers.go`: `func MyCommand(ctx, b, update) (*tg.Sendable, error)`
2. Register in `registerBotHandlers()`: `&tg.UpdateHandler{Type: tg.UPD_STARTSWITH, Param: "/mycommand", Fn: MyCommand}`
3. Add to `registerCommands()` for bot menu
4. Add tests in `handlers_test.go`

## Adding a Database Migration

Create `internal/db/migrations/NN_name.up.sql` and `NN_name.down.sql`. FTS5 tables need `fts5` build tag — tests auto-skip FTS5 migrations when unavailable. See existing migrations for patterns (FK references, indexes, triggers).

## Committing

Always use the git skill (commit workflow) when asked to commit changes.

## Code Style

- **Indentation**: Tabs for Go, spaces for Nix/JSON/SQL
- **Logging**: `log/slog` with structured logging
- **Testing**: `github.com/matryer/is` for assertions, table-driven tests
- **Errors**: Return up the stack, log at handler level
- **Context**: First parameter where needed

## Configuration

TOML file with env var overrides:

| Config key    | Env var          | Notes                   |
|---------------|------------------|-------------------------|
| database      | DATABASE         | SQLite file path        |
| port          | PORT             | HTTP server port        |
| telegram.token| TELEGRAM_TOKEN   | Bot token               |
| telegram.apikey| TELEGRAM_KEY    | Webhook validation key  |
| —             | LOG_LEVEL        | debug/info/warn/error   |

## Gotchas

- **CGO**: sqlite3 requires CGO; use `nix develop` or set `CC=gcc`
- **Build tags**: `sqlite_foreign_keys` for `go build`, `fts5` for nix
- **Inline keyboards**: Use `formatCallbackData()` with hex encoding
- **Callback data**: Always answer with `MethodAnswerCallbackQuery`
- **FK constraints**: Insert referenced rows (groups) before child records
- **DB locked**: Always `defer rows.Close()`, use transactions for multi-statement ops
- **LLM**: Requires Ollama running (`curl localhost:11434/api/tags`); pull models with `ollama pull`
- **Webhook debug**: `curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo`
- **Test webhook**: `curl -X POST http://localhost:6446/t/<APIKEY> -H "Content-Type: application/json" -d '{"update_id":1,"message":{"message_id":1,"chat":{"id":123},"text":"test"}}'`
- **Migration status**: `SELECT * FROM schema_migrations ORDER BY version;`

## Key Dependencies

`go-chi/chi/v5` (HTTP) · `jmoiron/sqlx` (SQL) · `mattn/go-sqlite3` (CGO) · `golang-migrate/migrate/v4` (migrations) · `pelletier/go-toml` (config) · `lmittmann/tint` (logging) · `matryer/is` (test assertions)
