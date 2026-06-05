# gumlet

CLI for the [Gumlet](https://www.gumlet.com/) Image CDN API. Manage sources, purge caches, query analytics, and build transform URLs.

## Install

```bash
go install github.com/nicolasacchi/gumlet/cmd/gumlet@latest
```

Or build from source:

```bash
git clone https://github.com/nicolasacchi/gumlet.git
cd gumlet
make install
```

## Quick Start

```bash
# Configure with API key and default subdomain
gumlet config add production --api-key gumlet_xxxx --subdomain millefarmacie

# Or use environment variable
export GUMLET_API_KEY=gumlet_xxxx

# List image sources
gumlet source list

# Purge a cached image
gumlet cache purge --url https://millefarmacie.gumlet.io/products/example.jpg -s millefarmacie

# Build a transform URL (no API call)
gumlet transform url https://millefarmacie.gumlet.io/img.jpg --width 400 --format webp --quality 80

# Inspect cache status
gumlet transform inspect https://millefarmacie.gumlet.io/img.jpg?w=400
```

## Authentication

API key resolution order:

1. `--api-key` flag
2. `GUMLET_API_KEY` environment variable
3. `GUMLET_KEY` environment variable
4. Config file (`~/.config/gumlet/config.toml`)

### Multi-Project Config

```toml
default_project = "production"

[projects.production]
api_key = "gumlet_prod_xxxx"
subdomain = "millefarmacie"

[projects.staging]
api_key = "gumlet_stg_xxxx"
subdomain = "millefarmacie-staging"
```

Switch projects: `gumlet --project staging source list`

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--api-key` | | Override API key |
| `--project` | | Use named project from config |
| `--subdomain` | `-s` | Override default subdomain |
| `--json` | | Force JSON output (auto-enabled when piped) |
| `--jq` | | Apply gjson path filter to JSON output |
| `--verbose` | `-v` | Print request/response details to stderr |
| `--quiet` | `-q` | Suppress non-error output |
| `--yes` / `--confirm` | | Confirm destructive operations (`source delete`, `cache purge`) |
| `--dry-run` | | Print the intended mutation and exit without sending |

## Output

- **TTY** (interactive terminal): human-readable tables
- **Piped** (non-TTY): JSON automatically
- **`--json`**: force JSON in any context
- **`--jq`**: gjson filter (note: uses [gjson syntax](https://github.com/tidwall/gjson), NOT jq)

Examples:
```bash
gumlet source list                          # Table in terminal
gumlet source list --json                   # Force JSON
gumlet source list | jq .                   # Auto-JSON when piped
gumlet source list --jq "#.name"            # gjson filter
gumlet source list --jq '#.{id:id,name:name}'  # Extract fields
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | API or generic error |
| 2 | Auth error (401/403) |
| 3 | Validation (400) |
| 4 | Not found (404) |
| 5 | Rate limited (429) |
| 6 | Write refused — confirmation required (`write_locked`); re-run with `--yes` |

Fleet-canonical table (`clicore/cierrors.ExitCodeFor`). **Changed:** auth was `3`, now `2`; CLI usage errors surface via cobra (exit 1).

Destructive verbs (`source delete`, `cache purge`) refuse unless `--yes`/`--confirm`
is passed (exit 6); `--dry-run` previews without sending. Automation must pass `--yes`.

## Commands

### `config` — Configuration Management

```bash
gumlet config add <name> --api-key <key> [--subdomain <sub>]
gumlet config remove <name>
gumlet config list
gumlet config use <name>          # Set default project
gumlet config current             # Show active project
```

### `source` — Image Source Management

```bash
gumlet source list                                    # List all sources
gumlet source describe <source-id>                    # Source details
gumlet source create --name my-source --type s3 \
  --bucket my-bucket --region us-east-1 \
  --access-key AKID --secret-key SECRET               # Create S3 source
gumlet source create --name web --type web_folder \
  --origin-url https://example.com/images              # Create web source
gumlet source update <id> --name new-name              # Update source
gumlet source delete <id>                              # Delete source
```

Source types: `s3`, `gcs`, `do_spaces`, `web_folder`, `custom`

### `cache` — Cache Purge

```bash
# Single URL
gumlet cache purge --url https://sub.gumlet.io/image.jpg

# Multiple URLs
gumlet cache purge --urls https://sub.gumlet.io/1.jpg,https://sub.gumlet.io/2.jpg

# From file (one URL per line)
gumlet cache purge --file urls.txt

# From stdin
cat urls.txt | gumlet cache purge --stdin

# From path pattern (auto-constructs URL from subdomain)
gumlet cache purge --path "/products/12345.jpg" -s millefarmacie

# Dry run (show what would be purged)
gumlet cache purge --dry-run --file urls.txt

# Custom batch size
gumlet cache purge --file urls.txt --batch-size 25
```

| Flag | Description | Default |
|------|-------------|---------|
| `--url` | Single URL to purge | |
| `--urls` | Comma-separated URLs | |
| `--file` | File with URLs (one per line) | |
| `--stdin` | Read URLs from stdin | |
| `--path` | Path pattern (requires `--subdomain`) | |
| `--dry-run` | Show URLs without purging | false |
| `--batch-size` | URLs per API call | 50 |

### `analytics` — Usage Analytics

```bash
gumlet analytics bandwidth --source-id <id>                          # Last 24h
gumlet analytics bandwidth --source-id <id> --start 2026-03-01 --end 2026-03-10
gumlet analytics requests --source-id <id>
gumlet analytics summary --source-id <id>                            # Combined overview
```

| Flag | Description | Default |
|------|-------------|---------|
| `--source-id` | Source to query (required) | |
| `--start` | Start date (YYYY-MM-DD) | Yesterday |
| `--end` | End date (YYYY-MM-DD) | Today |

### `transform url` — URL Builder (Local)

Build Gumlet transform URLs by appending query parameters. No API call is made.

```bash
gumlet transform url https://sub.gumlet.io/img.jpg --width 400 --format webp
# → https://sub.gumlet.io/img.jpg?format=webp&w=400

gumlet transform url https://sub.gumlet.io/img.jpg --width 800 --quality 80 --crop smart
# → https://sub.gumlet.io/img.jpg?crop=smart&q=80&w=800
```

| Flag | Short | Description |
|------|-------|-------------|
| `--width` | `-w` | Target width |
| `--height` | | Target height |
| `--format` | `-f` | Output format: webp, avif, auto, jpg, png, jxl |
| `--quality` | | Quality 1-100 |
| `--crop` | | Crop mode: smart, center, north, south, east, west |
| `--compress` | | Enable lossy compression |
| `--blur` | | Blur radius |
| `--sharpen` | | Enable sharpening |
| `--mode` | | Resize mode: fit, fill, stretch, pad |

### `transform inspect` — Cache/Format Inspector

Fetch a Gumlet URL and report cache status, format, size, and response time.

```bash
gumlet transform inspect https://sub.gumlet.io/img.jpg?w=400
# Status:    200
# Cache:     HIT
# Format:    image/webp
# Size:      34.2 KB
# Response:  45ms
```

### `version`

```bash
gumlet version           # Human-readable
gumlet version --json    # JSON output
```

## Building

```bash
make build     # → bin/gumlet
make install   # → ~/go/bin/gumlet
make test      # Run all tests
make lint      # Run golangci-lint
make clean     # Remove build artifacts
```

## License

MIT
