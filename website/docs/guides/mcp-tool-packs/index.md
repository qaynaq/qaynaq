---
sidebar_position: 2
---

# MCP Tool Packs

MCP Tool Packs let you deploy a full set of pre-built MCP tools for a service in one step. Instead of manually creating each tool, select a pack, configure shared credentials, pick the tools you need, and deploy them all at once.

## How It Works

1. Navigate to **Flows** > **Add New** > **MCP Tool Pack**.
2. Select a pack from the list.
3. Fill in the shared configuration (credentials, delegation settings).
4. Check or uncheck individual tools - all are selected by default.
5. Click **Deploy** to create all selected tools as active flows.

Each deployed tool becomes a standard MCP Tool flow. You can edit, pause, or delete individual tools after deployment just like any other flow. The tools appear on the `/mcp` endpoint and sync automatically within 5 seconds.

## Managing Tool Packs

Flows created from a tool pack are grouped together on the **Flows** page. Each pack appears as a collapsible section showing the pack name, number of tools, and a status summary.

### Viewing Individual Tools

Click the pack header to expand it and see all individual flows. Each flow has the same actions as standalone flows - edit, pause, resume, preview, and delete.

### Editing a Tool

You can open and edit any tool in a pack just like a regular flow. The tool stays in its pack group after editing.

### Bulk Delete

Click **Delete All** on a pack header to remove every flow in the pack at once. A confirmation dialog shows how many flows will be deleted. You can also delete individual tools from inside the expanded view.

### Redeploying a Pack

Click **Redeploy** on a pack header to open the tool pack wizard with that pack pre-selected. By default, tools that are already deployed are skipped.

To update existing tools with fresh configuration from the template, enable the **Override existing** checkbox. This deletes each existing tool and recreates it from the template.

:::warning
Override mode replaces existing tools with fresh copies from the template. Any manual edits you made to individual tools will be lost. Only use this when you want to reset tools back to their template defaults or update shared configuration like credentials.
:::

## Available Packs

| Pack | Tools | Description |
|------|-------|-------------|
| [Google Calendar](./google-calendar) | 13 | Manage events, calendars, and scheduling |
| [Google Drive](./google-drive) | 22 | Manage files, folders, permissions, and shared drives |
| [Google Sheets](./google-sheets) | 28 | Read, write, and manage spreadsheets |

Each tool exposes only the parameters relevant to its action. Required parameters are enforced by the MCP schema. Optional parameters default to sensible values.

## Tips

- You do not need to deploy all tools from a pack. Deploying only the tools your AI assistants need reduces noise and helps them pick the right tool more reliably.
- You can deploy the same pack multiple times with different credentials (e.g. different Google Workspace accounts). Each deployment creates a separate group.
- All deployed tools are immediately active and visible to connected MCP clients.
- Deleting a flow removes the corresponding tool from the MCP endpoint.
