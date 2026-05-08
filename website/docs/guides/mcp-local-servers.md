---
sidebar_position: 5
---

# Local MCP Servers (npx / command)

Qaynaq proxies two flavors of upstream MCP server: **remote** (an HTTP URL you point at, covered in [Remote MCP Servers](/docs/guides/mcp-remote-servers)) and **local** (a CLI binary Qaynaq runs as a child process, covered here). Both show up in the same MCP Servers list, both surface their tools through `/mcp`, and AI clients see one unified endpoint either way.

Most MCP servers in the wild ship as a CLI you run with `npx` (or `python -m`). They speak MCP over stdin/stdout, expect to be spawned by a parent process, and live for the lifetime of that parent. Qaynaq runs and supervises these for you.

## When to use local vs remote

**We recommend remote (HTTP) whenever a remote version exists.** Remote servers run out-of-process from the coordinator, so they don't compete for memory or CPU on the same host, scale independently, and survive coordinator restarts without a cold-start penalty. Local servers are a great fit when you don't have another option, but they are heavier on the host the coordinator runs on.

Use **local** when:
- There is no remote alternative for the MCP you need (the upstream only publishes a CLI).
- The MCP only ships as a CLI (e.g. `@modelcontextprotocol/server-filesystem`).
- You're testing locally and don't want to host the upstream over HTTP.

Use **remote** when:
- A remote version exists (almost always preferred, see above).
- The upstream already runs as a service (cloud-hosted vendor MCPs, internal team services).
- You want OAuth-connection-based auth.

Both can coexist in the same Qaynaq instance.

## The Catalog

Local servers come from a curated allowlist. You don't type a free-form `npx <pkg>` command, you pick from the catalog. This is intentional: spawning arbitrary executables is a footgun, especially in shared deployments. Each entry is tagged with a maintainer tier so you can see at a glance who stands behind it:

- **official** — vendor-maintained (Anthropic, Microsoft, Sentry, etc.).
- **community** — single maintainer or unaffiliated team. Useful but lower bus factor.

| Catalog ID | Maintainer | What it does | Required env |
|------------|------------|--------------|--------------|
| `filesystem` | official | Read/write files in a sandboxed dir | `ALLOWED_DIR` |
| `slack` | community | Read/post Slack messages | `SLACK_MCP_XOXP_TOKEN` (or alternative token) |
| `playwright` | official | Browser automation (headless or headed) | none |
| `redash` | community | Run SQL queries, browse schemas, manage dashboards | `REDASH_URL`, `REDASH_API_KEY` |

If you need an entry that isn't there, file an issue. We treat the catalog as something the project owns.

## Setup

### 1. Open MCP Servers

Same page as remote servers. The list shows a `local` or `remote` badge plus a process state badge (`running`, `idle`, `starting`, `failed`, `stopped`) on local rows.

### 2. Add a Server

Click **Add Server**. In the dialog:

- **Name** - a short, stable identifier. Used to namespace tool names: a tool `read_file` from a server named `fs` becomes `fs__read_file` on `/mcp`.
- **Server type** - pick **Local (command / npx)**.
- **Catalog** - pick the entry you want.
- **Env vars** - one input per declared env spec. Each var is one of:
  - **Required** - marked with `*`, must be filled before save.
  - **Optional** - marked `(optional)`, can stay empty.
  - **Advanced** - hidden behind a "Show advanced settings" toggle. Tuning knobs the typical user does not need to touch (cache TTL, safety mode, alternate hosts).

### 3. Use Secrets for Sensitive Values

You can pass an env value either as a literal string or as a reference to a saved secret. References look like `${SECRET_KEY}`:

```
GITHUB_PERSONAL_ACCESS_TOKEN = ${GH_TOKEN}
```

This pulls `GH_TOKEN` from **Settings → Secrets** and decrypts it at spawn. Why bother?

- **One place to rotate.** If you have ten servers using `${GH_TOKEN}`, you rotate the secret once and every server picks up the new value on next spawn.
- **No leak on backup or DB dump.** Literal env values are encrypted, but referenced secrets keep the credential out of the per-server row entirely.

Inline composition works the same way as Bento pipelines: any `${NAME}` inside a value is substituted, and you can mix multiple references in one string. Useful when the upstream MCP wants a fully-formed connection string:

```
DSN = postgres://${DB_USER}:${DB_PASS}@db.example.com/qaynaq
```

Both references resolve from the secrets table at spawn time. To literally pass `${SOMETHING}` to the child without resolution, escape with a double dollar sign: `$${SOMETHING}` becomes the literal `${SOMETHING}`.

## Process Lifecycle

When you save a local server, the coordinator spawns the configured command in its own process group, runs the MCP `initialize` handshake, and (on success) lists its tools. Tools appear on `/mcp` within ~30 seconds.

The supervisor handles three steady-state behaviors:

- **Idle stop.** A server with no tool calls for 15 minutes is stopped to save memory. The next call lazy-starts it (1-3 seconds while npm cache is warm). The state badge shows `idle` while stopped.
- **Crash recovery.** If the child exits unexpectedly, the supervisor restarts it. Three crashes in a 5-minute window flip the state to `failed`; manual restart is required from there.
- **Process group cleanup.** Children of the spawned command (e.g. Node spawned by `npx`) die with the parent. No zombies.

## When Things Fail

The state badge is your first signal. Click the **terminal icon** on a local row to see:

- **Process state** - what the supervisor thinks (`running`, `idle`, `failed`, ...).
- **Last error** - the most recent failure message from spawn, handshake, or liveness ping.
- **Stderr (last 4 KB)** - tail of the child's stderr. Most upstream servers log helpful errors here.

Common shapes:

- **Missing secret.** State stays `failed`, last error reads `missing secret: GH_TOKEN`. Create the secret in **Settings → Secrets**, then click **Restart**.
- **Bad token.** Process starts, upstream returns 401 on every call. Stderr shows the upstream error. Edit the env, save - the supervisor auto-restarts the process to pick up the new value.
- **Command not found.** Stderr is your friend - `npx` not in `$PATH`, or the package name typoed. The catalog only lists known-good package names, so this should be rare.

`Restart` (the refresh icon on local rows) clears any failed state and triggers a fresh spawn.

## Resource Caps

There's a coordinator-wide cap of 100 concurrent local processes. Idle servers don't count against it - they only count while running. If you hit the cap, the next attempt fails with a clear "process cap reached" message. Idle some servers or remove ones you don't need.

This is deliberately coarse: count-only, not per-process memory. If a single MCP eats too much memory, the kernel OOM-killer takes it and the supervisor restarts. Tighter limits are on the roadmap.

## What's Out of Scope

- Worker-distributed processes - everything runs on the coordinator. If you scale beyond ~50 local servers in active use, [open an issue](https://github.com/qaynaq/qaynaq/issues).
- Container or seccomp isolation. The processes share the coordinator's filesystem and network.
- Streaming notifications from upstream MCP back to the client.
