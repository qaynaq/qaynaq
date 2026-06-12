---
sidebar_position: 2
---

# Templates

Templates let you deploy a full set of pre-built flows for a service in one step. Instead of manually creating each flow, select a template, configure shared settings once, pick the flows you need, and deploy them all at once.

A template can contain MCP tools (exposed on the `/mcp` endpoint for AI assistants) and automations (scheduled or event-driven flows). Each flow in a template is labeled with its kind in the wizard.

## How It Works

1. Navigate to **Flows** and click **Install Template**.
2. Select a template from the list.
3. Fill in the shared configuration. Depending on the template this can be plain text fields, an OAuth connection picker, or a secret picker.
4. Check or uncheck individual flows - flows that are not yet deployed are selected by default.
5. Click **Deploy** to create all selected flows as active flows.

Each deployed flow becomes a standard flow. You can edit, pause, or delete individual flows after deployment just like any other flow. MCP tools appear on the `/mcp` endpoint and sync automatically within 5 seconds.

## Configuration Types

Templates declare which settings they need, and the wizard renders the right input for each:

- **Text** - plain values like a store name.
- **Connection** - a picker for your existing OAuth connections. Set these up on the [Connections](/docs/guides/connections) page first.
- **Secret** - a picker for your existing secrets, with an inline option to create a new one. The deployed flows reference the secret by name and never contain the raw value, so flows stay safe to export and share.

## Managing Templates

Flows created from a template are grouped together on the **Flows** page. Each template appears as a collapsible section showing the template name, number of flows, and a status summary.

### Viewing Individual Flows

Click the template header to expand it and see all individual flows. Each flow has the same actions as standalone flows - edit, pause, resume, preview, and delete.

### Editing a Flow

You can open and edit any flow in a template group just like a regular flow. The flow stays in its group after editing.

### Bulk Delete

Click **Delete All** on a template header to remove every flow in the group at once. A confirmation dialog shows how many flows will be deleted. You can also delete individual flows from inside the expanded view.

### Redeploying a Template

Click **Redeploy** on a template header to open the template wizard with that template pre-selected. By default, flows that are already deployed are skipped.

To update existing flows with fresh configuration from the template, enable the **Override existing** checkbox. This replaces the configuration of each existing flow with the template's current configuration.

:::warning
Override mode replaces the configuration of existing flows. Any manual edits you made to individual flows will be lost. Only use this when you want to reset flows back to their template defaults or update shared configuration like credentials.
:::

## Available Templates

| Template | Flows | Description |
|----------|-------|-------------|
| [Google Calendar](./google-calendar) | 13 | Manage events, calendars, and scheduling |
| [Google Drive](./google-drive) | 22 | Manage files, folders, permissions, and shared drives |
| [Google Sheets](./google-sheets) | 28 | Read, write, and manage spreadsheets |
| [Shopify](./shopify) | 7 | Read orders, products, customers, and inventory |

Each MCP tool exposes only the parameters relevant to its action. Required parameters are enforced by the MCP schema. Optional parameters default to sensible values.

## Tips

- You do not need to deploy all flows from a template. Deploying only the tools your AI assistants need reduces noise and helps them pick the right tool more reliably.
- You can deploy the same template multiple times with different credentials (e.g. different Google Workspace accounts). Each deployment creates a separate group.
- All deployed flows are immediately active, and MCP tools are visible to connected MCP clients.
- Deleting a flow removes the corresponding tool from the MCP endpoint.
