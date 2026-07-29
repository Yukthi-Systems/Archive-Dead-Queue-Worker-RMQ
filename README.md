# Archive Dead Queue Worker (RMQ)

A small RabbitMQ worker that drains a dead-letter queue and archives every
message to disk as durable, human-inspectable files, instead of letting
dead-lettered messages sit in the broker or get discarded.

Each message is written to a per-month directory alongside a
`metadata.jsonl` index describing it (message ID, dead-letter reason,
source exchange/routing key, file size, etc.), so archived mail can be
audited or replayed later.

A companion CLI tool, `republish`, lets you take a message captured via the
RabbitMQ management API and re-publish it against another broker (e.g. a
local/test instance) for debugging.

## Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [consumer](#consumer)
  - [republish](#republish)
- [How it works](#how-it-works)
- [Project layout](#project-layout)
- [Development](#development)
- [License](#license)

## Features

- Consumes a durable RabbitMQ queue with configurable prefetch and
  concurrency.
- Archives each message body under `<ARCHIVE_BASE_PATH>/<YYYY-MM>/`,
  extracting the `email_content` field when present, falling back to the
  raw message body when it isn't.
- Appends one JSON line per archived message to that month's
  `metadata.jsonl`, capturing the dead-letter reason, source exchange and
  routing key, timestamp, and file size.
- Filesystem errors trigger a requeue (`nack` + requeue) instead of message
  loss; malformed payloads are archived raw rather than dropped.
- Structured JSON logging with automatic log rotation, retention, and
  compression.
- Graceful shutdown on `SIGINT`/`SIGTERM`.

## Requirements

- Go 1.26 or later
- Access to a RabbitMQ broker with a durable queue to consume from
  (typically a dead-letter queue)

## Installation

```bash
git clone <this-repository>
cd Archive-Dead-Queue-Worker-RMQ
go build -o bin/consumer ./cmd/consumer
go build -o bin/republish ./cmd/republish
```

Or run directly without building a binary:

```bash
go run ./cmd/consumer
```

## Configuration

The worker is configured entirely through environment variables, optionally
loaded from a `.env` file in the working directory. Copy the example file
and fill in your broker details:

```bash
cp .env.example .env
```

| Variable             | Default            | Description                                             |
| --------------------- | ------------------ | -------------------------------------------------------- |
| `RB_HOST`             | *(required)*        | RabbitMQ broker hostname or IP                            |
| `RB_PORT`             | `5672`              | RabbitMQ broker AMQP port                                 |
| `RB_VHOST`            | `/`                 | RabbitMQ virtual host                                     |
| `RB_USERNAME`         | *(required)*        | Broker login username                                     |
| `RB_PASSWD`           | *(required)*        | Broker login password                                     |
| `RB_CONSUME_QUEUE`    | *(required)*        | Queue to consume (dead-letter) messages from              |
| `RB_CONSUMER_TAG`     | *(empty)*           | Consumer tag reported to the broker                       |
| `RB_WORKER_COUNT`     | `1`                 | Number of concurrent message handlers                     |
| `RB_PREFETCH`         | `1`                 | Per-consumer QoS prefetch count                            |
| `AMQPR_CON_DELAY`     | `5s`                | Delay between reconnect attempts                          |
| `ARCHIVE_BASE_PATH`   | `./archive`         | Root directory archived messages are written under        |
| `LOG_FILE`            | `./logs/app.log`    | Log file path                                              |
| `LOG_MAX_SIZE_MB`     | `50`                | Log file size (MB) that triggers rotation                 |
| `LOG_MAX_BACKUPS`     | `5`                 | Number of rotated log files to retain                      |
| `LOG_MAX_AGE_DAYS`    | `14`                | Days to retain rotated log files                            |
| `LOG_COMPRESS`        | `true`              | Gzip-compress rotated log files                             |
| `LOG_LEVEL`           | `debug`             | Minimum log level (`trace`, `debug`, `info`, `warn`, `error`) |
| `LOG_CONSOLE`         | `false`             | Also log to stderr in addition to the log file              |

See [internal/config/config.go](internal/config/config.go) for the
authoritative list and defaults.

## Usage

### consumer

Runs the long-lived worker that consumes and archives dead-lettered
messages. Stop it with `Ctrl+C` (`SIGINT`) or `SIGTERM`; in-flight
deliveries are allowed to finish before it exits.

```bash
go run ./cmd/consumer
```

Archived files are written to:

```
<ARCHIVE_BASE_PATH>/<YYYY-MM>/<message-id>.eml       # parsed successfully
<ARCHIVE_BASE_PATH>/<YYYY-MM>/<message-id>.raw.json  # unparseable body, archived raw
<ARCHIVE_BASE_PATH>/<YYYY-MM>/metadata.jsonl         # one JSON line per archived message
```

### republish

A standalone CLI for replaying a single message captured from the RabbitMQ
management API's "get message(s)" endpoint
(`POST /api/queues/<vhost>/<queue>/get`) into a queue on another broker —
handy for reproducing a dead-lettered message against a local/test
instance.

```bash
go run ./cmd/republish \
  -file message.json \
  -url  amqp://user:pass@host:5672/ \
  -queue test_queue \
  -index 0
```

| Flag      | Default                    | Description                                              |
| --------- | --------------------------- | ---------------------------------------------------------- |
| `-file`   | `message.txt`                | Path to the management API JSON export                     |
| `-index`  | `0`                           | Index of the message in the export array to republish      |
| `-url`    | `$TEST_AMQP_URL`              | AMQP URL of the target broker                               |
| `-queue`  | *(required)*                  | Target queue name (delivered via the default exchange)      |

## How it works

1. `cmd/consumer` loads configuration and initializes logging, then opens a
   RabbitMQ connection and starts a durable consumer
   ([internal/rmq](internal/rmq)).
2. Each delivery is handed to a `Handler`
   ([internal/rmq/handler.go](internal/rmq/handler.go)), which:
   - reads the dead-letter reason from the AMQP headers
     ([internal/utils](internal/utils)),
   - attempts to parse the body as an email payload,
   - writes the extracted (or raw) content to
     `<ARCHIVE_BASE_PATH>/<YYYY-MM>/`,
   - appends a metadata record describing the archived file.
3. The message is acknowledged once archived; a filesystem error requeues
   it instead of losing it.

## Project layout

```
cmd/
  consumer/   entry point for the archiving worker
  republish/  entry point for the message-replay CLI
internal/
  config/     environment-variable configuration loading
  logger/     rotating, structured (zerolog) logging
  model/      shared data shapes (archive metadata, email payload)
  rmq/        RabbitMQ connection, consumer, and archiving handler
  utils/      payload parsing and filename/header helpers
```

## Development

```bash
go build ./...   # compile everything
go vet ./...      # static checks
gofmt -l .        # formatting check
```

There are currently no automated tests; when adding features, prefer
covering `internal/utils` and `internal/rmq` handler logic with table-driven
tests.

## License

Licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).
