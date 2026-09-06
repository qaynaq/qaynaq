---
sidebar_position: 5
---

# HubSpot

Deploys 24 MCP tools for reading, searching, creating, updating, and deleting contacts, companies, deals, and tickets in your HubSpot CRM.

## Setting Up the HubSpot Connection

HubSpot authenticates with OAuth. You create a HubSpot app once, register it as a Qaynaq connection, and the template references that connection. Qaynaq refreshes the short-lived HubSpot access token automatically, so there is no token to copy or rotate by hand.

1. In a [HubSpot developer account](https://developers.hubspot.com/), create a **public app**.
2. Open the app's **Auth** tab and copy the **Client ID** and **Client Secret**.
3. Add the scopes for the actions you intend to use. Read scopes cover list/get/search; write scopes cover create/update/delete:

   | Object | Read scope | Write scope |
   |--------|-----------|-------------|
   | Contacts | `crm.objects.contacts.read` | `crm.objects.contacts.write` |
   | Companies | `crm.objects.companies.read` | `crm.objects.companies.write` |
   | Deals | `crm.objects.deals.read` | `crm.objects.deals.write` |
   | Tickets | `crm.objects.tickets.read` | `crm.objects.tickets.write` |

   If you only want the read tools to work, add just the read scopes - the create/update/delete tools will return a permission error until you add the matching write scope. To deploy a read-only set, deselect the create/update/delete tools in the install wizard.

:::warning
The create, update, and delete tools let an AI agent change and remove records in your live HubSpot CRM. Delete archives the record (recoverable from the HubSpot UI for a limited window), it does not permanently erase it. Grant write scopes deliberately, and consider deploying only the read tools for agents that should not modify data.
:::

:::note
Every tool carries MCP annotations describing its effect. Read tools (list/get/search) are marked `readOnlyHint`; update and delete tools are marked `destructiveHint`. MCP clients that respect these hints (such as Claude) can warn or ask for confirmation before running a destructive tool, giving you a second safety layer on top of scopes and tool selection.
:::

4. In Qaynaq, go to **Connections** and create a connection with provider **HubSpot**, pasting the Client ID and Client Secret. Authorize when prompted - this links your HubSpot account.
5. When you install the template, pick this connection in the **OAuth Connection** field.

:::note
HubSpot access tokens expire after 30 minutes. The connection stores the refresh token and renews access automatically, so your deployed flows keep working without manual intervention.
:::

## Shared Configuration

| Field | Required | Description |
|-------|----------|-------------|
| OAuth Connection | Yes | The HubSpot connection from the steps above. Set it up on the [Connections](/docs/guides/connections) page first. |

## Included Tools

Each of the four objects (contacts, companies, deals, tickets) has the same six tools - list, get, search, create, update, delete:

| Tool pattern | Description |
|--------------|-------------|
| `hubspot_list_<object>` | List recent records with pagination |
| `hubspot_get_<object>` | Get a specific record by ID |
| `hubspot_search_<object>` | Find records by free-text query or property filters |
| `hubspot_create_<object>` | Create a new record from a JSON property object |
| `hubspot_update_<object>` | Update an existing record's properties by ID |
| `hubspot_delete_<object>` | Delete (archive) a record by ID |

So the full set is `hubspot_list_contacts`, `hubspot_get_contact`, `hubspot_search_contacts`, `hubspot_create_contact`, `hubspot_update_contact`, `hubspot_delete_contact`, and the same six for `company`, `deal`, and `ticket`.

## Creating and Updating Records

Create and update tools take a `properties_json` argument - a JSON object of HubSpot property values. For example, to create a contact:

```json
{"email":"sarah@acme.com","firstname":"Sarah","lastname":"Lee"}
```

Update tools additionally take the record's `<object>_id`. Only the properties you include are changed; others are left as-is. The create and update calls return the full record (including its new ID) so an agent can chain follow-up actions. Delete tools take only the `<object>_id` and archive the record.

## Listing, Searching, and Pagination

- **List tools** return recent records. They accept `limit` (default 10, max 100) and an `after` pagination cursor.
- **Search tools** find specific records. Pass a free-text `query` (matches the object's default searchable properties, e.g. name or email) for simple lookups, or a `filters` JSON array for precise matching. They accept `limit` (default 10, max 200) and `after`.
- **Pagination**: both list and search responses include `paging.next.after` when more records exist. Pass that value back as `after` to fetch the next page. Search results are capped at 10,000 total records per query.

### Filter syntax

The `filters` parameter is a JSON array of filter groups. Filters within a group are combined with AND; groups are combined with OR. Example - find deals in the "closed won" stage worth more than 1000:

```json
[{"filters":[
  {"propertyName":"dealstage","operator":"EQ","value":"closedwon"},
  {"propertyName":"amount","operator":"GT","value":"1000"}
]}]
```

Supported operators: `EQ`, `NEQ`, `LT`, `LTE`, `GT`, `GTE`, `BETWEEN`, `IN`, `NOT_IN`, `HAS_PROPERTY`, `NOT_HAS_PROPERTY`, `CONTAINS_TOKEN`, `NOT_CONTAINS_TOKEN`.
