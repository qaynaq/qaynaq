# Generate

Generates messages on a schedule. Primarily used for synthetic data in testing and development, but also works well as a **recurring job trigger** — use it to drive any flow that should run on a fixed interval or cron schedule.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Mapping | Bloblang | — | Bloblang mapping to generate message content |
| Interval | string | `1s` | How often to emit a message. Accepts a duration (`1s`, `5m`, `1h`), a `@every <duration>` shorthand, or a full 6-field cron expression (`0 0 * * * *`) |
| Count | integer | `0` | Number of messages to generate (0 = unlimited) |
| Batch Size | integer | `1` | Number of messages per batch |
| Auto Replay Nacks | boolean | `true` | Automatically replay rejected messages |

## Interval formats

| Format | Example | Meaning |
|--------|---------|---------|
| Duration | `30s` | Every 30 seconds |
| `@every` shorthand | `@every 1h` | Every 1 hour |
| Cron expression | `0 0 9 * * *` | Every day at 09:00 |

Cron expressions use 6 fields: `seconds minutes hours day-of-month month day-of-week`.

## As a recurring job trigger

Set the Mapping to a static payload that carries the context your downstream processors or outputs need (e.g., a date range, a job name, a resource ID). The pipeline then acts as a scheduled job — fetching data, running transforms, and writing results — triggered entirely by Generate's interval.

Set **Count** to `1` for one-shot runs, or leave it at `0` to repeat indefinitely on schedule.
