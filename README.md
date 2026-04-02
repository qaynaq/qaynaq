# Qaynaq

> The fastest way to connect your data to AI

[![license](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE) [![Release](https://img.shields.io/github/v/release/qaynaq/qaynaq)](https://github.com/qaynaq/qaynaq/releases) [![Go Report Card](https://goreportcard.com/badge/github.com/qaynaq/qaynaq)](https://goreportcard.com/report/github.com/qaynaq/qaynaq)

Turn APIs, databases, and scripts into MCP tools - without writing MCP servers. Qaynaq is an open-source, self-hosted platform that lets you visually build tools for AI assistants like Claude, Cursor, and AI agents. Connect 66+ sources, define parameters, and get an instant MCP endpoint. Single binary, no Docker or JVM required.

**[Documentation](https://qaynaq.io/)** | **[Playbooks](https://qaynaq.io/playbooks)**

## Quick Start

### Install

```bash
curl -Lsf https://qaynaq.io/sh/install | bash
```

See the [Installation docs](https://qaynaq.io/docs/getting-started/installation) for more options.

### Run

```bash
# Start coordinator
qaynaq -role coordinator -grpc-port 50000

# Start worker (in a separate terminal)
qaynaq -role worker -grpc-port 50001
```

Open **http://localhost:8080** and start building MCP tools.

### Docker

```bash
docker pull ghcr.io/qaynaq/qaynaq
```

See the [Installation docs](https://qaynaq.io/docs/getting-started/installation) for Docker run commands and the [Configuration docs](https://qaynaq.io/docs/getting-started/configuration) for database setup.

## How It Works

1. **Connect** - Pick from 66+ built-in connectors for APIs, databases, queues, and services.
2. **Define Tool** - Set the tool name, description, and input parameters. Qaynaq generates the MCP tool automatically.
3. **Use With AI** - Tools are instantly callable by Claude, Cursor, AI agents, and internal copilots via the `/mcp` endpoint.

## Features

- **Instant MCP Endpoint** - Flows are auto-registered as MCP tools, discoverable by any MCP client
- **Visual Tool Builder** - Drag-and-drop DAG editor to build and manage tools and data flows
- **66+ Connectors** - Kafka, HTTP, AMQP, MySQL CDC, PostgreSQL, and more
- **Built-in Transformations** - Transform data with Bloblang DSL and JSON Schema validation
- **Parameter Validation** - Define types, descriptions, and required flags visible to AI assistants
- **Secure Credentials** - Encrypted secrets management for API keys and tokens
- **Automation Flows** - Move and route data between systems, or orchestrate AI workflows that call your MCP tools
- **Horizontal Scaling** - Coordinator & worker architecture for easy scaling
- **Single Binary** - No Docker, JVM, or external dependencies required
- **Self-hosted & Open Source** - Full control over your data, Apache 2.0 licensed

## Playbooks

- [Build a Weather Tool for AI Assistants](https://qaynaq.io/playbooks/mcp-weather-tool) - Build tools for AI assistants without writing MCP servers.
- [Kafka to PostgreSQL](https://qaynaq.io/playbooks/kafka-to-postgresql) - Flow events from Kafka through Avro schema decoding into PostgreSQL.
- [HTTP Webhooks](https://qaynaq.io/playbooks/http-webhooks) - Accept webhook data over HTTP and store it in a database.

## Contributing

We welcome contributions! Please check out [CONTRIBUTING](CONTRIBUTING) (coming soon) for guidelines.

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
