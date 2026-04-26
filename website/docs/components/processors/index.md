---
sidebar_position: 2
---

# Processors

Processors transform, validate, or route messages within a pipeline. They are applied in order between the input and output. Add processors when creating or editing a flow in the UI.

| Component | Description |
|-----------|-------------|
| [AI Gateway](/docs/components/processors/ai-gateway) | Calls an AI chat completion API (OpenAI, Anthropic) |
| [Google Calendar](/docs/components/processors/google-calendar) | Google Calendar operations (events, calendars, availability) |
| [Google Drive](/docs/components/processors/google-drive) | Google Drive operations (files, folders, permissions, shared drives) |
| [Google Sheets](/docs/components/processors/google-sheets) | Google Sheets operations (spreadsheets, worksheets, rows, formatting) |
| [Shopify](/docs/components/processors/shopify) | Shopify Admin API operations (orders, products, customers, inventory) |
| [Mapping](/docs/components/processors/mapping) | Bloblang transformations |
| [Command](/docs/components/processors/command) | Executes an external command, piping the message through stdin/stdout |
| [JSON Schema](/docs/components/processors/json-schema) | Validates messages against a JSON schema |
| [Catch](/docs/components/processors/catch) | Error handling — runs processors on failure |
| [Switch](/docs/components/processors/switch) | Conditional processing based on message content |
| [Schema Registry Decode](/docs/components/processors/schema-registry-decode) | Decodes Avro messages via Schema Registry |
| [HTTP Client](/docs/components/processors/http-client) | Sends a message as an HTTP request and replaces it with the response |
| [SQL Raw](/docs/components/processors/sql-raw) | Runs an arbitrary SQL query against a database |
| [SQL Select](/docs/components/processors/sql-select) | Runs a SELECT query and replaces the message with the rows returned |
| [SQL Insert](/docs/components/processors/sql-insert) | Inserts a row into an SQL database for each message |
