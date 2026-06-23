# HubSpot

Performs HubSpot CRM operations - list, get, search, create, update, and delete contacts, companies, deals, and tickets.

:::tip Template Available
Want to expose HubSpot CRM data as MCP tools for AI assistants? Use the [HubSpot template](/docs/guides/templates/hubspot) to deploy all 24 tools in one step - no manual configuration needed.
:::

## Authentication

This processor authenticates with a Qaynaq OAuth connection. Create a HubSpot connection on the [Connections](/docs/guides/connections) page, then reference it by name in `oauth_connection`. The connection's access token is injected on every request and refreshed automatically when it expires, so no token handling is needed in the flow.

The connected HubSpot app needs scopes matching the actions used: `.read` scopes (e.g. `crm.objects.contacts.read`) for list/get/search, and `.write` scopes (e.g. `crm.objects.contacts.write`) for create/update/delete.

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| OAuth Connection | string | - | HubSpot OAuth connection name (required) |
| Action | select | - | The CRM operation to perform (required) |
| Limit | string | `10` | Max records per page; list allows 1-100, search allows 1-200 |
| After | string | - | Pagination cursor for list and search actions |
| Object ID | string | - | Record ID for get, update, and delete actions |
| Properties | string | - | Comma-separated properties to return; defaults to HubSpot's standard set |
| Query | string | - | Free-text search string for search actions |
| Filters | string | - | JSON filter groups for search actions (see below) |
| Properties JSON | string | - | JSON object of property values to write; required for create and update actions |

### Actions

Each object (`contact`, `company`, `deal`, `ticket`) supports six actions: `list_<object>`, `get_<object>`, `search_<object>`, `create_<object>`, `update_<object>`, `delete_<object>`.

- **list** actions return HubSpot's `{results, paging}` response for recent records.
- **get** actions require `object_id` and return the single record.
- **search** actions POST to HubSpot's search endpoint. Pass a free-text `query` and/or a `filters` JSON array (filter groups: AND within a group, OR across groups). Operators: `EQ`, `NEQ`, `LT`, `LTE`, `GT`, `GTE`, `BETWEEN`, `IN`, `NOT_IN`, `HAS_PROPERTY`, `NOT_HAS_PROPERTY`, `CONTAINS_TOKEN`, `NOT_CONTAINS_TOKEN`.
- **create** actions require `properties_json` (a JSON object of property values) and return the created record.
- **update** actions require `object_id` and `properties_json`, and return the updated record. Only the included properties change.
- **delete** actions require `object_id` and archive the record (recoverable from the HubSpot UI for a limited window).

All parameter fields support `${!this.field}` interpolation from the incoming message.
