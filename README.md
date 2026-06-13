<div align="center">

<img src="https://qaynaq.io/img/mascot.png" alt="Qaynaq" width="180" />

# Qaynaq

### The fastest way to connect your data to AI

Connect any database, API, or service to AI assistants like Claude and Cursor.<br />
Open-source, runs on your machine, no coding required.

[![license](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/qaynaq/qaynaq)](https://github.com/qaynaq/qaynaq/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/qaynaq/qaynaq)](https://goreportcard.com/report/github.com/qaynaq/qaynaq)

**[Website](https://qaynaq.io)** · **[Documentation](https://qaynaq.io/docs/getting-started/installation)** · **[Playbooks](https://qaynaq.io/playbooks)**

<br />

<img src="https://qaynaq.io/img/screens/flow-builder.png" alt="Qaynaq visual flow builder" width="820" />

</div>

<br />

## Why Qaynaq

Turn APIs, databases, and scripts into MCP tools without writing a single MCP server. Pick a connector, define your parameters, and get an instant endpoint your AI assistant can call.

- **Install in 30 seconds, build your first tool in 5 minutes**
- **Your data stays on your machine** - nothing leaves your network
- **66+ connectors** out of the box - databases, APIs, queues, and more
- **Free forever** - open-source and self-hosted

## Quick Start

```bash
curl -Lsf https://qaynaq.io/sh/install | bash
```

```bash
# Start the coordinator
qaynaq -role coordinator -grpc-port 50000

# Start a worker (in a separate terminal)
qaynaq -role worker -grpc-port 50001
```

Open **http://localhost:8080** and start building.

> Prefer Docker? `docker pull ghcr.io/qaynaq/qaynaq`. See the [Installation docs](https://qaynaq.io/docs/getting-started/installation) for all options.

## How It Works

| 1. Connect | 2. Define | 3. Use With AI |
| --- | --- | --- |
| Pick from 66+ built-in connectors for APIs, databases, queues, and services. | Set the tool name, description, and input parameters. Qaynaq generates the MCP tool for you. | Tools are instantly callable by Claude, Cursor, agents, and copilots via the `/mcp` endpoint. |

<div align="center">
<img src="https://qaynaq.io/img/screens/dashboard.png" alt="Qaynaq dashboard" width="820" />
</div>

## Features

- **Instant MCP endpoint** - flows auto-register as MCP tools, discoverable by any MCP client
- **External MCP server proxy** - register remote (HTTP) or local (npx/command) MCP servers and serve their tools through the same `/mcp` endpoint, namespaced and aggregated, with OAuth and API key auth plus automatic token refresh
- **Visual tool builder** - drag-and-drop DAG editor for tools and data flows
- **66+ connectors** - Kafka, HTTP, AMQP, MySQL CDC, PostgreSQL, and more
- **Built-in transformations** - shape data with the Bloblang DSL and JSON Schema validation
- **Smart parameter validation** - types, descriptions, and required flags visible to AI assistants
- **Secure credentials** - encrypted secrets management for API keys and tokens
- **Automation flows** - move and route data between systems, or orchestrate AI workflows
- **Horizontal scaling** - coordinator and worker architecture
- **Single binary** - no Docker, JVM, or external dependencies required
- **Self-hosted and open-source** - full control over your data, Apache 2.0 licensed

## Playbooks

- [Build a Weather Tool for AI Assistants](https://qaynaq.io/playbooks/mcp-weather-tool) - tools for AI assistants without writing MCP servers.
- [Kafka to PostgreSQL](https://qaynaq.io/playbooks/kafka-to-postgresql) - flow events from Kafka through Avro schema decoding into PostgreSQL.
- [HTTP Webhooks](https://qaynaq.io/playbooks/http-webhooks) - accept webhook data over HTTP and store it in a database.

## Contributing

We welcome contributions! Please check out [CONTRIBUTING](CONTRIBUTING.md) for guidelines.

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
