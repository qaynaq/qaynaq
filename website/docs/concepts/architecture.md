---
sidebar_position: 1
---

# Architecture

Qaynaq uses a **Coordinator & Worker** model designed for simplicity and horizontal scalability.

## Overview

```mermaid
graph TD
    C["Coordinator<br/>(HTTP + gRPC)"]
    C -- gRPC --> W1["Worker 1"]
    C -- gRPC --> W2["Worker 2"]
    C -- gRPC --> W3["Worker 3"]
    C -- gRPC --> W4["Worker ..."]
```

## Coordinator

The coordinator is the control plane. It:

- Serves the **web UI** and **REST API** on the HTTP port (default `8080`).
- Manages **flow definitions** — create, update, delete, start, stop.
- Handles **worker registration** and health tracking via gRPC.
- Balances **workload distribution** across available workers.
- Stores all state in the configured database (SQLite or PostgreSQL).
- Exposes an **MCP server** at `/mcp` endpoint using Flowable HTTP transport, allowing AI assistants to discover and call flows configured as MCP Tools.

There is one coordinator per deployment.

## Workers

Workers are the data plane. They:

- **Register** with the coordinator on startup via gRPC.
- **Execute flows** — run the actual input, processing, and output pipelines.
- Are **stateless** — all configuration comes from the coordinator.
- Can be **added or removed** at any time for horizontal scaling.
- Send periodic **heartbeats** to the coordinator.

You can run as many workers as needed. Each worker handles one or more flows assigned by the coordinator.

## Communication

All communication between coordinator and workers uses **gRPC**. The coordinator exposes a REST/HTTP API (with gRPC-Gateway) for the web UI and external integrations.

### MCP Server

The coordinator includes a built-in **Model Context Protocol (MCP) server** that exposes flows as tools for AI assistants. This server:

- Is accessible at the `/mcp` endpoint on the coordinator's HTTP port
- Uses the **Flowable HTTP transport** protocol from the MCP specification
- Automatically syncs tools every 5 seconds based on active flows with the MCP Tool input
- Works with any MCP-compatible client (Claude Desktop, Claude Code, Cursor, etc.)
- Forwards tool calls to workers for execution and returns the response
- Supports optional **token-based authentication** to control which clients can access tools (see [MCP Authentication](/docs/getting-started/authentication#mcp-authentication))

When an AI assistant calls a tool, the MCP server:
1. Receives the tool call via the `/mcp` endpoint
2. Forwards the request to a worker running the corresponding flow
3. The flow processes the request through its processors
4. Returns the result with optional status code (via `meta status_code`) to the AI client

See the [MCP Server guide](/docs/guides/mcp-server) for connecting clients and authentication, and the [MCP Tool component](/docs/components/inputs/mcp-tool) for creating tools.

## Deployment

Since Qaynaq is a single binary, deployment is straightforward:

- **Single machine**: Run coordinator and worker(s) as separate processes with different gRPC ports.
- **Multiple machines**: Run the coordinator on one host and workers on others, pointing workers to the coordinator's address via `-discovery-uri`.
- **Kubernetes**: Deploy coordinator as a Deployment/Service and workers as a **StatefulSet**. StatefulSets give each worker pod a stable hostname (e.g., `qaynaq-worker-0`, `qaynaq-worker-1`), which prevents stale worker registrations that occur with Deployments where pods get random names on every restart.
- **Docker**: Use `docker run` commands to run the coordinator and workers as containers. See [Installation](/docs/getting-started/installation#docker) for details.

No JVM or external dependencies are required. Docker can be used for containerized deployments (Kubernetes, Docker Compose, etc.) but is not a hard dependency — unlike tools such as Airbyte, Qaynaq runs natively as a single binary.
