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
│   ├── chain/            # HTTP middleware chain builder
│   ├── dicer/            # Dice rolling logic
│   ├── statistics/       # Metrics collection and prometheus export
│   ├── statsd/           # StatsD client
│   ├── vectorstore/      # In-memory vector similarity search (cosine)
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

## Database Migration Examples

### Basic Table Creation

When adding a new table, follow this pattern:

**`NN_feature.up.sql`**
```sql
CREATE TABLE IF NOT EXISTS mytable
( id         VARCHAR(64) NOT NULL
, name       TEXT
, created_at INTEGER DEFAULT (strftime('%s', 'now'))
, PRIMARY KEY(id)
, FOREIGN KEY(gid) REFERENCES groups(gid)
);
```

**`NN_feature.down.sql`**
```sql
DROP TABLE IF EXISTS mytable;
```

### Foreign Key Relationships

Always reference the `groups` table for multi-tenant data:

```sql
CREATE TABLE IF NOT EXISTS mytable
( id  VARCHAR(64) NOT NULL
, gid VARCHAR(64) NOT NULL
, data TEXT
, PRIMARY KEY(id,gid)
, FOREIGN KEY(gid) REFERENCES groups(gid)
);
```

### Adding Columns

```sql
-- up
ALTER TABLE mytable ADD COLUMN new_field TEXT;

-- down
-- SQLite doesn't support DROP COLUMN directly, use a workaround or recreate table
```

### Full-Text Search (FTS5)

For search functionality, create FTS5 tables with triggers:

```sql
-- up
CREATE VIRTUAL TABLE mytable_fts USING
fts5( description
    , gid UNINDEXED
    , tokenize="trigram case_sensitive 0"
);

CREATE TRIGGER mytable_oninsert AFTER INSERT ON mytable BEGIN
    INSERT INTO mytable_fts(rowid, description, gid)
    VALUES (new.rowid, new.description, new.gid);
END;

CREATE TRIGGER mytable_onupdate AFTER UPDATE ON mytable BEGIN
    UPDATE mytable_fts
    SET description = new.description
    WHERE rowid = new.rowid;
END;

CREATE TRIGGER mytable_ondelete AFTER DELETE ON mytable BEGIN
    DELETE FROM mytable_fts WHERE rowid = old.rowid;
END;

-- down
DROP TRIGGER IF EXISTS mytable_oninsert;
DROP TRIGGER IF EXISTS mytable_onupdate;
DROP TRIGGER IF EXISTS mytable_ondelete;
DROP TABLE IF EXISTS mytable_fts;
```

**Note**: FTS5 requires the `fts5` build tag. Tests automatically skip FTS5 migrations if unavailable.

### Indexes

```sql
-- up
CREATE INDEX IF NOT EXISTS idx_mytable_gid ON mytable(gid);
CREATE INDEX IF NOT EXISTS idx_mytable_name ON mytable(name COLLATE NOCASE);

-- down
DROP INDEX IF EXISTS idx_mytable_gid;
DROP INDEX IF EXISTS idx_mytable_name;
```

## Detailed API Documentation

### Handler Types and Patterns

#### Available Handler Types

```go
const (
    UPD_STARTSWITH  // Command starts with exact text (e.g., "/command")
    UPD_CONTAINS    // Message contains text anywhere
    UPD_REGEXP      // Match using regular expression
    UPD_MEDIAPHOTO  // Photo attached to message
    UPD_MEDIAVIDEO  // Video attached to message
    UPD_MEDIADOCUMENT // Document attached to message
    UPD_CALLBACK    // Inline keyboard callback
    UPD_WILDCARD    // Catch-all handler
    UPD_MENTION     // Bot mentioned in message
)
```

#### Command Handler Pattern

```go
func MyCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
    // Parse command arguments
    command, rest := utils.SplitMessageWords(update.Message.Text)
    arg1, arg2 := utils.SplitMessageWords(rest)

    // Handle logic
    result := processData(arg1, arg2)

    // Return sendable or nil
    return &tg.Sendable{
        Method:  tg.MethodSendMessage,
        ChatID:  update.Message.Chat.ID,
        Text:    fmt.Sprintf("<b>%s</b>", result),
        ParseMode: "HTML",
    }, nil
}

// Return nil, nil for no response (e.g., just set a reaction)
return nil, nil
```

#### Media Handler Pattern

```go
func PhotoHandler(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
    // Access photo file ID
    photoID := update.Message.Photo[0].FileID

    // Get caption/reply text
    text := update.Message.Caption
    if update.Message.ReplyToMessage != nil {
        text = update.Message.ReplyToMessage.Text
    }

    // Set reaction to acknowledge
    return tg.SendableSetMessageReaction(update, "✅"), nil
}
```

#### Callback Handler Pattern

```go
func CallbackHandler(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
    // Parse callback data (hex-encoded)
    action, id := parseCallbackData(update.CallbackQuery.Data)

    switch action {
    case "show":
        // Handle show action
        return &tg.Sendable{
            Method: tg.MethodEditMessageText,
            MessageID: update.CallbackQuery.Message.MessageID,
            ChatID: update.CallbackQuery.Message.Chat.ID,
            Text: "Showing item " + id,
        }, nil
    }

    // Always answer callback query
    return &tg.Sendable{
        Method: tg.MethodAnswerCallbackQuery,
        CallbackQueryID: update.CallbackQuery.ID,
    }, nil
}
```

### Sendable Structures

All bot responses use `*tg.Sendable`:

```go
type Sendable struct {
    Method            string                 // API method name
    ChatID            int64                  // Target chat
    MessageID         int                    // For editing messages
    Text              string                 // Message text
    ParseMode         string                 // HTML/Markdown
    ReplyMarkup       *InlineKeyboardMarkup  // Inline keyboard
    ReplyToMessageID  int                    // Reply to message
    CallbackQueryID   string                 // For callback answers
    Reaction          []ReactionType         // Message reactions
    // ... other fields for photos, videos, etc.
}
```

### Inline Keyboards

```go
func createKeyboard(offset int) *tg.Sendable {
    keyboard := [][]tg.InlineKeyboardButton{
        {
            {Text: "Button 1", CallbackData: "action1"},
            {Text: "Button 2", CallbackData: "action2"},
        },
        {
            {Text: "Previous", CallbackData: formatCallbackData("prev", offset)},
            {Text: "Next", CallbackData: formatCallbackData("next", offset)},
        },
    }

    return &tg.Sendable{
        Method:      tg.MethodEditMessageReplyMarkup,
        ChatID:      chatID,
        MessageID:   messageID,
        ReplyMarkup: &tg.InlineKeyboardMarkup{InlineKeyboard: keyboard},
    }
}
```

### Telegram API Wrapper Methods

Key wrapper methods in `pkg/tg/methods.go`:

```go
// Send any sendable
b.SendSendable(ctx, sendable)

// Set bot commands
b.SetMyCommands(ctx, &tg.SetMyCommandsParams{
    Commands: []tg.BotCommand{
        {Command: "help", Description: "Show help"},
    },
    Scope:    &tg.BotCommandScope{Type: "default"},
    LanguageCode: "en",
})

// Set message reaction
b.SetMessageReaction(ctx, update, "👍", "❤️")

// Send raw API request
b.SendRaw(ctx, "methodName", params, result)
```

## LLM Integration

The bot integrates with Ollama for AI-powered responses using function calling.

### LLM Configuration

LLM models are defined in `internal/llm/client.go`:

```go
const (
    ModelGranite4      = `granite4:350m`
    ModelFunctiongemma = `functiongemma:270m`
    ModelLLama         = `llama:1b`
    ModelQwen          = `qwen3.5:0.5b`
)
```

### Adding Tools

Tools are registered in `internal/llm/tools.go`:

```go
func AllTools() []Tool {
    return []Tool{
        {
            Name:        "my_tool",
            Description: "Tool description",
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "param1": map[string]any{
                        "type":        "string",
                        "description": "Parameter description",
                    },
                },
                "required": []string{"param1"},
            },
            Handler: func(chatID int64, args api.ToolCallFunctionArguments) (string, error) {
                // Extract arguments
                param1, ok := args.Get("param1").(string)
                if !ok {
                    return "", fmt.Errorf("invalid param1")
                }

                // Process and return result
                result := doSomething(param1)
                return result, nil
            },
        },
    }
}
```

### System Prompt

The system prompt is defined in `internal/llm/client.go`:

```go
const SystemPrompt = `You are MarranoBot, an assistant in a Telegram group chat.
You have access to tools:
- roll_dice: Roll dice. Usage: describe what dice to roll (e.g., "2d6+3")
- search_media: Search for media files. Usage: search by description

Keep responses concise, use the adjective marrano to compliment the user.
Use HTML formatting when helpful.`
```

Modify this to change bot personality or available tool descriptions.

## Testing Patterns

### Handler Tests

Use `github.com/matryer/is` for assertions:

```go
func TestMyCommand(t *testing.T) {
    is := is.New(t)

    // Setup in-memory database
    is.NoErr(db.Open(":memory:"))
    is.NoErr(db.Migrate())
    defer db.Close()

    // Create test bot
    bot, err := tg.New("123456", "https://bot.marrani.lol")
    is.NoErr(err)

    // Create test update
    update := &tg.Update{
        UpdateID: -1,
        Message: &tg.Message{
            ID:   1,
            Chat: tg.Chat{ID: 123456},
            Text: "/mycommand arg1 arg2",
        },
    }

    // Execute handler
    ctx := context.TODO()
    sendable, err := MyCommand(ctx, bot, update)

    // Assert results
    is.NoErr(err)
    is.True(sendable != nil)
    is.Equal(sendable.Method, tg.MethodSendMessage)
    is.True(strings.Contains(sendable.Text, "expected text"))
}
```

### Database Tests

```go
func TestDatabaseOperation(t *testing.T) {
    is := is.New(t)

    is.NoErr(db.Open(":memory:"))
    is.NoErr(db.Migrate())
    defer db.Close()

    // Insert test data
    gid := int64(-123456)
    is.NoErr(db.InsertGroup(context.TODO(), gid, "test group"))

    // Query and verify
    var result MyStruct
    row := dbc.QueryRow(`SELECT * FROM mytable WHERE gid=?`, gid)
    is.NoErr(row.Scan(&result.ID, &result.GID, ...))

    is.Equal(result.GID, gid)
}
```

### Test Helpers

Common test helpers in `internal/core/handlers_test.go`:

```go
// Create a basic message update
func updMessage(text string) *tg.Update {
    return &tg.Update{
        UpdateID: -123456,
        Message: &tg.Message{
            ID:   1,
            Chat: tg.Chat{ID: 123456},
            Text: text,
        },
    }
}

// Create update with photo
func updMediaPhoto(text string, id string) *tg.Update {
    u := updMessage(text)
    u.Message.Photo = []*tg.PhotoSize{{FileID: id}}
    return u
}

// Pretty-print JSON for debugging
func pprint(t *testing.T, s any) {
    t.Helper()
    m, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        t.Log(err)
    }
    t.Log(string(m))
}
```

## Troubleshooting Guide

### Common Issues and Solutions

#### Test Failures: "no such table"

**Problem**: Tests fail with table not found errors.

**Cause**: Database not migrated or wrong migration order.

**Solution**:
```go
// Always migrate after opening
is.NoErr(db.Open(":memory:"))
is.NoErr(db.Migrate())  // Don't forget this!
defer db.Close()
```

#### FTS5 Errors: "no such module: fts5"

**Problem**: FTS5 migrations fail in tests.

**Cause**: FTS5 not available in SQLite build.

**Solution**: The codebase already handles this gracefully. Tests automatically skip FTS5 migrations and limit to version 10 when FTS5 is unavailable. No action needed.

#### Foreign Key Constraint Failed

**Problem**: `FOREIGN KEY constraint failed` error.

**Cause**: Trying to insert records with non-existent foreign keys.

**Solution**:
```go
// Always insert referenced rows first
is.NoErr(db.InsertGroup(ctx, gid, "group name"))
// Then insert records that reference it
is.NoErr(db.InsertMedia(ctx, &Media{GID: gid, ...}))
```

#### Handler Not Triggered

**Problem**: Command not responding.

**Cause**: Handler not registered or wrong handler type.

**Solution**:
```go
// Check registration in registerBotHandlers()
&tg.UpdateHandler{
    Type:    tg.UPD_STARTSWITH,  // Correct type?
    Param:   "/mycommand",       // Exact match?
    Aliases: []string{"/alias"}, // Aliases match?
    Fn:      MyCommand,          // Function exported?
}
```

#### Webhook Not Receiving Updates

**Problem**: Bot not receiving Telegram updates.

**Cause**: Webhook not set or wrong URL.

**Solution**:
```bash
# Check webhook status
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getWebhookInfo"

# Set webhook
./marrano-bot -W -c config.toml
```

Verify:
- Domain matches config
- API key in URL matches config
- HTTPS is working (TLS certificate valid)

#### Database Locked

**Problem**: "database is locked" errors.

**Cause**: Multiple connections or transactions not closed.

**Solution**:
```go
// Always close statements and rows
rows, err := q.QueryContext(ctx, args...)
defer rows.Close()  // Important!

// Use transactions for multi-statement operations
tx, err := dbc.BeginTxx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()  // Rollback if commit fails

// ... operations ...

err = tx.Commit()
```

#### Memory Leaks in Tests

**Problem**: Tests pass but memory grows.

**Cause**: Not closing resources.

**Solution**:
```go
// Always defer cleanup
is.NoErr(db.Open(":memory:"))
defer db.Close()  // Close database

rows, err := q.Query(...)
defer rows.Close()  // Close rows
```

#### LLM Not Responding

**Problem**: `/llm` command hangs or errors.

**Cause**: Ollama not running or model not available.

**Solution**:
```bash
# Check Ollama is running
curl http://localhost:11434/api/tags

# List available models
ollama list

# Pull model if needed
ollama pull granite4:350m
```

Check logs for:
- `LLMCommand: error getting conversation`
- Connection refused errors

#### CGO Errors During Build

**Problem**: Build fails with CGO errors.

**Cause**: SQLite driver requires CGO but no C compiler.

**Solution**:
```bash
# Enter nix shell (includes CGO toolchain)
nix develop

# Or ensure CC is set
export CC=gcc
go build -tags "sqlite_foreign_keys" ./cmd/marrano-bot
```

### Debugging Tips

1. **Enable debug logging**:
   ```go
   slog.SetLogLoggerLevel(slog.LevelDebug)
   ```

2. **Use pprint()** to inspect sendables:
   ```go
   pprint(t, sendable)
   ```

3. **Check database state**:
   ```go
   rows, _ := dbc.Query("SELECT * FROM media")
   defer rows.Close()
   for rows.Next() {
       // inspect rows
   }
   ```

4. **Test with curl**:
   ```bash
   # Test webhook endpoint
   curl -X POST http://localhost:6446/t/YOUR_API_KEY \
     -H "Content-Type: application/json" \
     -d '{"update_id": 1, "message": {"message_id": 1, "chat": {"id": 123}, "text": "test"}}'
   ```

5. **Check migration status**:
   ```sql
   SELECT * FROM schema_migrations ORDER BY version;
   ```

## Development Workflow

### Adding a New Feature

1. **Plan the feature**
   - Define what it does
   - Identify database changes needed
   - Plan handlers and API endpoints

2. **Create database migration** (if needed)
   ```bash
   # Create migration files
   internal/db/migrations/NN_feature.up.sql
   internal/db/migrations/NN_feature.down.sql
   ```

3. **Implement database operations** in `internal/db/`
   ```go
   func InsertMyData(ctx context.Context, data *MyData) error {
       q, err := prepareStmt(`INSERT INTO mytable (...) VALUES (...)`)
       if err != nil {
           return err
       }
       _, err = q.ExecContext(ctx, data.Field1, data.Field2)
       return err
   }
   ```

4. **Implement handler** in `internal/core/handlers.go`
   ```go
   func MyFeatureCommand(ctx context.Context, b *tg.Bot, update *tg.Update) (*tg.Sendable, error) {
       // Implementation
   }
   ```

5. **Register handler** in `registerBotHandlers()`
   ```go
   &tg.UpdateHandler{
       Type:  tg.UPD_STARTSWITH,
       Param: "/myfeature",
       Fn:    MyFeatureCommand,
   }
   ```

6. **Add to bot menu** in `registerCommands()`
   ```go
   {Command: "myfeature", Description: "Feature description"},
   ```

7. **Write tests** in `internal/core/handlers_test.go`
   ```go
   func TestMyFeatureCommand(t *testing.T) {
       // Test implementation
   }
   ```

8. **Run tests**
   ```bash
   make test
   ```

9. **Build and test locally**
   ```bash
   make marrano-bot
   ./marrano-bot -c config.toml
   ```

10. **Document changes**
    - Update AGENTS.md if adding patterns
    - Add comments to complex code
    - Consider adding examples to this file

### Code Review Checklist

Before committing, verify:
- [ ] Tests pass (`make test`)
- [ ] Code builds (`make marrano-bot`)
- [ ] No lint errors (`go vet ./...`)
- [ ] Foreign keys are properly used
- [ ] Errors are returned, not ignored
- [ ] Context is passed where needed
- [ ] Logging is appropriate (not too verbose, not silent)
- [ ] Handler returns nil,nil when no response needed
- [ ] Prepared statements use proper parameterization
- [ ] Database operations handle context cancellation
- [ ] New code follows existing patterns

### Performance Considerations

1. **Use prepared statements**: Always use `prepareStmt()` to cache SQL queries
2. **Limit query results**: Use `LIMIT` in SELECT queries
3. **Index appropriately**: Add indexes for frequently queried columns
4. **Use context properly**: Ensure long-running operations check context
5. **Batch operations**: Use transactions for multiple inserts/updates
6. **Avoid N+1 queries**: Use JOINs or batch queries instead of loops

### Security Best Practices

1. **Never log sensitive data**: Don't log API keys, tokens, or user messages
2. **Validate input**: Always validate user input before processing
3. **Use parameterized queries**: Never concatenate SQL with user input
4. **Check permissions**: Verify users have appropriate permissions before actions
5. **Rate limit**: Consider rate limiting expensive operations
6. **Sanitize HTML**: Use proper HTML parsing/escaping before rendering
