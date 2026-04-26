# Command

Executes an external command for each message. The message contents are piped to the command's stdin, and the command's stdout becomes the new message contents. Useful for shelling out to existing scripts, CLIs, or tools available on the worker host.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Name | string | — | The command to execute. Supports interpolation, e.g. a metadata field can be used to pick the binary at runtime (required) |
| Args Mapping | Bloblang | — | A mapping that should evaluate to an array of strings used as command arguments |

## Dynamic Commands

Both Name and Args Mapping can be derived from message contents. Use Name with interpolation to pick the binary per message, and Args Mapping to build the argument list from message fields. This makes it possible to drive the command entirely from the upstream payload.

## Performance

Each message spawns a process. For high-throughput flows, the per-message process startup cost is significant. Prefer batching upstream, or use a more specialized processor where one exists.

## Security

This processor executes arbitrary binaries on the worker host. In multi-tenant or hosted deployments, restrict who can edit flows that use it. Treat any flow containing Command as privileged.
