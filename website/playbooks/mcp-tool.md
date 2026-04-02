---
slug: mcp-weather-tool
description: Build a weather tool that AI assistants can call via Model Context Protocol.
---

# Build a Weather Tool for AI Assistants

This playbook walks through creating a weather tool that AI assistants can discover and call via MCP.

For general setup, connecting clients, and authentication, see the [MCP Server guide](/docs/guides/mcp-server).

## Create the Weather Tool

Open the Qaynaq UI, click **Create New Flow**, and configure each section:

### Input - select **MCP Tool**

| Field | Value |
|-------|-------|
| Name | `get_weather` |
| Description | `Get current weather for a city` |
| Input Parameters | `city` (string, required) - "City name" |

### Processor - select **Mapping**

For this example, we return a random temperature. In practice, you would call an external API or use other processors.

| Field | Value |
|-------|-------|
| Mapping | `root.temperature = random_int(min:0, max:40).string() + "C"` / `root.city = this.city` |

### Output

The output is automatically set to **Sync Response** and locked when using MCP Tool input.

Click **Save** and then **Start** the flow.

## Test the Tool

Once the flow is running, the `get_weather` tool appears automatically on the `/mcp` endpoint. You can verify it by asking your AI assistant something like "What's the weather in London?" and it will call the tool.

See [MCP Server - Verify with curl](/docs/guides/mcp-server#verify-with-curl) for testing via the command line.
