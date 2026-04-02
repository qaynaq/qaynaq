---
sidebar_position: 1
slug: /
---

# Introduction

Qaynaq is a modern, open-source data pipeline tool designed to be a powerful and efficient alternative to tools like Airbyte and Fivetran. It empowers data analysts and scientists to easily build and manage data flows with a visual pipeline builder UI.

At its core, Qaynaq uses [Bento](https://warpstreamlabs.github.io/bento/) as the flow processing engine. Bento provides battle-tested connectors, at-least-once delivery guarantees, and a rich set of processors — all driven by Qaynaq's coordinator-worker architecture and managed through the UI without writing any configuration by hand.

## Key Features

- **Visual Flow Builder** — Intuitive UI to visually create and manage data pipelines with support for conditional routing and fan-out patterns.
- **Powerful In-Pipeline Transformations** — Bloblang DSL for efficient data transformation and enrichment within the pipeline, replacing the need for separate tools like dbt.
- **MCP Integration** — Expose flows as tools for AI assistants via the Model Context Protocol. Connect with Claude Desktop, Claude Code, Cursor, and other MCP-compatible clients.
- **Flexible Subprocess Processor** — Integrate processors written in any programming language. Communication via stdin/stdout ensures language-agnostic compatibility.
- **Native HTTP Input** — Accept data over HTTP, ideal for webhooks and flowing data sources.
- **Horizontally Scalable Worker Pool** — Scale your data processing with a horizontally scalable worker pool architecture.
- **Delivery Guarantees** — Reliable data delivery with buffering, caching, and robust error handling.

## Why Qaynaq?

- **Completely free** — Apache 2.0 license with no usage limits or paid tiers.
- **Zero operational overhead** — Runs as a single lightweight binary with no Docker, JVM, or external dependencies required.
- **Native transformations** — The built-in Bloblang DSL handles mapping, filtering, and conditional logic, eliminating the need for separate tools like dbt.
- **MCP Server built-in** — Turn any flow into an AI tool with the built-in Model Context Protocol server. No external infrastructure needed.
- **Language-agnostic processors** — Write custom processors in any language through simple stdin/stdout communication.
- **Built-in observability** — Metrics, tracing, and logs out of the box.
- **High performance** — Written in Go with low memory and CPU footprint. Supports both real-time and batch workloads with fine-grained parallel execution control.

## Who is Qaynaq for?

- **Data Engineers** who want a lightweight, self-hosted ETL tool without Docker overhead.
- **Backend Developers** who need to move data between systems with transformations.
- **AI Developers** building agentic systems that need custom tools for AI assistants via MCP.
- **Small Teams** that want powerful data pipelines without enterprise complexity or cost.

## Next Steps

- [Install Qaynaq](/docs/getting-started/installation) and get running in minutes.
- Follow the [Quickstart](/docs/getting-started/quickstart) to build your first pipeline.
- Learn about the [Architecture](/docs/concepts/architecture) behind Qaynaq.
