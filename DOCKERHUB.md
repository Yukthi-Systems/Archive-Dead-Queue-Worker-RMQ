# Archive Dead Queue Worker (RMQ)

A lightweight worker that consumes a RabbitMQ dead-letter queue and
archives every message to disk, grouped into month folders with a
`metadata.jsonl` index for auditing or replay.

Source, full configuration reference, and the companion `republish`
debugging tool: https://github.com/Yukthi-Systems/Archive-Dead-Queue-Worker-RMQ

## Quick start

```bash
docker run -d \
  --name archive-dead-queue-worker \
  -e RB_HOST=your-rabbitmq-host \
  -e RB_PORT=5672 \
  -e RB_VHOST=/ \
  -e RB_USERNAME=your-username \
  -e RB_PASSWD=your-password \
  -e RB_CONSUME_QUEUE=dead_letter_queue \
  -v "$(pwd)/logs:/app/logs" \
  -v "$(pwd)/archive:/app/archive" \
  --restart always \
  rjyspl/archive-dead-queue-worker:latest
```

Or with `docker compose`:

```yaml
services:
  archive-dead-queue-worker-rmq:
    image: rjyspl/archive-dead-queue-worker:latest
    container_name: archive-dead-queue-worker
    env_file:
      - .env
    volumes:
      - ./logs:/app/logs
      - ./archive:/app/archive
    restart: always
```

```bash
docker compose up -d
```

## Configuration

All settings are passed as environment variables (directly, or via
`--env-file` / `env_file:`).

| Variable            | Default          | Description                                        |
| -------------------- | ----------------- | ---------------------------------------------------- |
| `RB_HOST`            | *(required)*       | RabbitMQ broker hostname or IP                        |
| `RB_PORT`            | `5672`             | RabbitMQ broker AMQP port                              |
| `RB_VHOST`           | `/`                | RabbitMQ virtual host                                  |
| `RB_USERNAME`        | *(required)*       | Broker login username                                  |
| `RB_PASSWD`          | *(required)*       | Broker login password                                  |
| `RB_CONSUME_QUEUE`   | *(required)*       | Queue to consume (dead-letter) messages from            |
| `RB_CONSUMER_TAG`    | *(empty)*          | Consumer tag reported to the broker                     |
| `RB_WORKER_COUNT`    | `1`                | Number of concurrent message handlers                   |
| `RB_PREFETCH`        | `1`                | Per-consumer QoS prefetch count                          |
| `AMQPR_CON_DELAY`    | `5s`               | Delay between reconnect attempts                        |
| `ARCHIVE_BASE_PATH`  | `./archive`        | Root directory archived messages are written under (keep under `/app`, see below) |
| `LOG_FILE`           | `./logs/app.log`   | Log file path (keep under `/app`, see below)             |
| `LOG_MAX_SIZE_MB`    | `50`               | Log file size (MB) that triggers rotation               |
| `LOG_MAX_BACKUPS`    | `5`                | Number of rotated log files to retain                    |
| `LOG_MAX_AGE_DAYS`   | `14`               | Days to retain rotated log files                          |
| `LOG_COMPRESS`       | `true`             | Gzip-compress rotated log files                           |
| `LOG_LEVEL`          | `debug`            | Minimum log level (`trace`, `debug`, `info`, `warn`, `error`) |
| `LOG_CONSOLE`        | `false`            | Also log to stderr in addition to the log file            |

## Volumes

The container's working directory is `/app`, and `ARCHIVE_BASE_PATH` /
`LOG_FILE` default to relative paths under it. Mount these paths so
archived messages and logs survive container restarts/recreation:

| Container path  | Purpose                          |
| ---------------- | --------------------------------- |
| `/app/archive`   | Archived message files + `metadata.jsonl` indexes |
| `/app/logs`      | Rotating application log          |

## Tags

- `latest` — most recent build from the default branch

## License

GNU General Public License v3.0.
