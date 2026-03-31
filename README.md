# marrano-bot

A Telegram bot written in Go that provides media management, callouts, dice rolling, AI-powered responses, and other chat utilities. It receives updates via webhooks and stores data in SQLite.

## Features

- **Media management** — remember, forget, search, and rate photos and videos with descriptions
- **Callouts** — create `!callout` triggers that reply with randomized responses
- **Abraxas** — keyword triggers that reply with media attachments
- **Dice rolling** — standard dice notation (`2d6+3`, `d20`, etc.)
- **LLM integration** — AI responses via Ollama when the bot is @mentioned
- **Statistics** — Prometheus metrics and channel stats tracking
- **Aliases** — short aliases for frequently used commands
- **Media export** — sync media files to a local folder

## Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `/remember` | `/r` | Remember a photo or video with a description |
| `/forget` | `/f` | Forget a saved photo or video |
| `/search` | `/s`, `/pupy` | Search saved media by description |
| `/top` | `/t`, `/top10` | Show top-rated media |
| `/tte` | `/bottom` | Show bottom-rated media |
| `/abraxas` | `/ab` | Manage keyword triggers (`add\|rm <word> [photo\|video]`) |
| `/callout` | `/c`, `/oh` | Create or delete a `!callout` |
| `/dice` | `/d` | Roll dice using standard notation |
| `/alias` | `/a` | Manage command aliases |
| `/grumpyness` | `/g`, `/grumpy` | Show channel statistics |
| `!callout` | | Trigger a callout response |

## Usage

```
./marrano-bot -c config.toml
```

```
Flags:
  -c, --config string       bot configuration path (default "./marrano-bot.toml")
  -D, --dump                dump configuration object
  -E, --export              export database data as csv
      --export-dir string   folder to write exported csv files
  -I, --init                initialize the database
  -M, --export-media string sync media files to a folder
  -v, --verbose             set verbose output
  -h, --help                show help message
```

## Configuration

TOML config file:

```toml
database = "marrano-bot.sqlite"
port = 6446

[telegram]
name = "marrano-bot"
token = "123456789-bot"
domain = "bot.marrani.lol"
apikey = "webhook-secret-key"
```

### Config options

| Option | Description |
|--------|-------------|
| `database` | SQLite database file path |
| `port` | HTTP server listen port (default `6446`) |
| `telegram.name` | Bot display name |
| `telegram.token` | Telegram bot token |
| `telegram.domain` | Public domain for webhook registration |
| `telegram.apikey` | Secret key for webhook URL validation |

### Environment variables

Environment variables override config file values:

| Variable | Overrides |
|----------|-----------|
| `DATABASE` | `database` |
| `PORT` | `port` |
| `TELEGRAM_TOKEN` | `telegram.token` |
| `TELEGRAM_KEY` | `telegram.apikey` |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, `warn`, `error`) |

## Building

### Go

```bash
go build -tags "sqlite_foreign_keys" -v ./cmd/marrano-bot
```

Or with make:

```bash
make
```

### Nix

```bash
nix build .#default
```

### Development shell

```bash
nix develop
```

## NixOS module

The flake provides a NixOS module with a systemd service, Caddy and nginx reverse proxy support:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
    marrano-bot.url = "github:moolite/bot";
  };

  outputs = { nixpkgs, marrano-bot, ... }@inputs: {
    nixosConfigurations.my-machine = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        marrano-bot.nixosModules.default
        {
          services.marrano-bot = {
            enable = true;
            hostName = "bot.example.com";
            logLevel = "info";
          };
        }
      ];
    };
  };
}
```

### Module options

| Option | Default | Description |
|--------|---------|-------------|
| `services.marrano-bot.enable` | `false` | Enable the service |
| `services.marrano-bot.port` | `64041` | HTTP listen port |
| `services.marrano-bot.hostName` | `bot.marrani.lol` | Public hostname for webhooks |
| `services.marrano-bot.token` | `""` | Telegram bot token |
| `services.marrano-bot.dataDir` | `/var/lib/marrano-bot` | Data directory |
| `services.marrano-bot.logLevel` | `info` | Log level |
| `services.marrano-bot.openPort` | `false` | Open firewall port |

## Architecture

```
cmd/marrano-bot/       # Entrypoint, CLI flags, media sync
internal/
  config/              # TOML config + env var overrides
  core/                # HTTP server (chi), all Telegram handlers
  db/                  # SQLite layer (sqlx) with Client struct
  dicer/               # Dice rolling logic
  llm/                 # Ollama integration with function calling
  statistics/          # Prometheus metrics
  telegram/            # Telegram file download utilities
  utils/               # Shared string helpers
pkg/tg/                # Custom Telegram bot framework
```

The bot uses a custom lightweight Telegram framework (`pkg/tg/`) rather than an external library. Updates arrive via webhooks on `POST /t/{apikey}`. The database layer uses a `db.Client` struct with `sync.RWMutex`-protected prepared statement caching.

## Testing

```bash
go test -v ./...
```

## License

Copyright &copy; 2023-2026 Lorenzo Giuliani. Released under [MPL-2.0](LICENSE).
