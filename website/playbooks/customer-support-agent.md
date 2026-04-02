---
description: Build an AI customer support agent with three MCP tools that look up orders, find customer history, and draft replies.
---

# Customer Support Agent

In this playbook you will build three MCP tools that work together as a customer support agent. An AI assistant like Claude will be able to look up orders, find customer history, and draft support replies, all through Qaynaq flows.

By the end you will have:

- A SQLite database with sample order data
- Three MCP tool flows that the AI can call
- A working setup where Claude (or any MCP client) handles support questions on its own

## Prerequisites

- Qaynaq coordinator and worker running ([Installation](/docs/getting-started/installation))
- `sqlite3` command line tool (comes pre-installed on macOS, install via `apt install sqlite3` on Ubuntu/Debian)
- An OpenAI or Anthropic API key (for the reply drafting tool)

## 1. Create the Database

First, create a SQLite database with some sample orders. Run this in your terminal from the same directory where Qaynaq is running:

```bash
sqlite3 support.sqlite <<'SQL'
CREATE TABLE orders (
    order_id TEXT PRIMARY KEY,
    customer_name TEXT NOT NULL,
    customer_email TEXT NOT NULL,
    status TEXT NOT NULL,
    items TEXT NOT NULL,
    total REAL NOT NULL,
    shipping_address TEXT NOT NULL,
    ordered_at TEXT NOT NULL,
    shipped_at TEXT,
    tracking_number TEXT
);

INSERT INTO orders VALUES
('ORD-1001', 'Alice Johnson', 'alice@example.com', 'shipped',
 'Wireless Headphones x1, Phone Case x2', 89.97,
 '123 Oak Street, Portland, OR 97201',
 '2025-03-05', '2025-03-07', 'TRK-998877'),

('ORD-1002', 'Sanan Guliyev', 'sanan@example.com', 'processing',
 'Mechanical Keyboard x1', 149.99,
 '456 Elm Avenue, Austin, TX 78701',
 '2025-03-10', NULL, NULL),

('ORD-1003', 'Alice Johnson', 'alice@example.com', 'delivered',
 'USB-C Hub x1, HDMI Cable x1', 45.50,
 '123 Oak Street, Portland, OR 97201',
 '2025-02-20', '2025-02-22', 'TRK-556644'),

('ORD-1004', 'Carol Davis', 'carol@example.com', 'refunded',
 'Laptop Stand x1', 79.99,
 '789 Pine Road, Seattle, WA 98101',
 '2025-03-01', '2025-03-03', 'TRK-112233'),

('ORD-1005', 'Sanan Guliyev', 'sanan@example.com', 'backordered',
 'Monitor Arm x1, Desk Mat x1', 124.50,
 '456 Elm Avenue, Austin, TX 78701',
 '2025-03-08', NULL, NULL);
SQL
```

This gives you five orders across three customers with different statuses: shipped, processing, delivered, refunded, and backordered. Feel free to add more rows to test different scenarios.

## 2. Create the "lookup_order" Tool

This tool lets the AI look up a specific order by its ID.

Open the Qaynaq UI, click **Create New Flow**, and give it a name like `lookup-order`.

### Input: select MCP Tool

| Field | Value |
|-------|-------|
| Name | `lookup_order` |
| Description | `Look up an order by its order ID. Returns order details including status, items, total, shipping address, and tracking number.` |
| Input Parameters | `order_id` (string, required) - "The order ID, e.g. ORD-1001" |

### Processor: select SQL Select

| Field | Value |
|-------|-------|
| Driver | `sqlite` |
| DSN | `file:./support.sqlite?mode=ro` |
| Table | `orders` |
| Columns | `order_id`, `customer_name`, `customer_email`, `status`, `items`, `total`, `shipping_address`, `ordered_at`, `shipped_at`, `tracking_number` |
| Where | `order_id = ?` |
| Args Mapping | `root = [this.order_id]` |

### Output

Automatically set to **Sync Response**.

Click **Save** and then **Start** the flow.

## 3. Create the "lookup_customer" Tool

This tool lets the AI find all orders for a given customer email. Useful when the customer does not have their order ID handy.

Create another flow called `lookup-customer`.

### Input: select MCP Tool

| Field | Value |
|-------|-------|
| Name | `lookup_customer` |
| Description | `Find all orders for a customer by their email address. Returns order IDs and statuses. Get full details for each order in case it's needed.` |
| Input Parameters | `email` (string, required) - "Customer email address" |

### Processor: select SQL Select

| Field | Value |
|-------|-------|
| Driver | `sqlite` |
| DSN | `file:./support.sqlite?mode=ro` |
| Table | `orders` |
| Columns | `order_id`, `status`, `ordered_at` |
| Where | `customer_email = ?` |
| Args Mapping | `root = [this.email]` |

### Output

Automatically set to **Sync Response**.

Click **Save** and then **Start** the flow.

## 4. Create the "draft_support_reply" Tool

This tool takes what the AI learned from the other tools and generates a professional support reply.

Create a flow called `draft-support-reply`.

### Input: select MCP Tool

| Field | Value |
|-------|-------|
| Name | `draft_support_reply` |
| Description | `Draft a customer support email reply. Provide the customer name, their message, and any relevant order details you found.` |
| Input Parameters | `customer_name` (string, required) - "Customer's name", `customer_message` (string, required) - "The customer's original message", `order_details` (string, required) - "Relevant order information you looked up" |

### Processor: select AI Gateway

| Field | Value |
|-------|-------|
| Provider | `openai` or `anthropic` |
| Model | `gpt-5-mini` or `claude-sonnet-4-6` |
| API Key | Your API key |
| System Prompt | `You are a friendly customer support agent. Write a helpful, concise reply to the customer. Be empathetic but keep it short. Do not make up information. Only use the order details provided.` |
| Prompt | Customer name: %v<br/><br/>Customer message: %v<br/><br/>Order details: %v<br/><br/>Write a reply to this customer. |
| Args Mapping | `root = [this.customer_name, this.customer_message, this.order_details]` |
| Result Map | `root.reply = this.content` |

### Output

Automatically set to **Sync Response**.

Click **Save** and then **Start** the flow.

## 5. Connect Your AI Client

Now connect an MCP client to the Qaynaq endpoint so it can discover all three tools.

### Claude Code

```bash
claude mcp add qaynaq -- npx mcp-remote http://localhost:8080/mcp
```

### Claude Desktop

Add to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "qaynaq": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp"]
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json` in your project:

```json
{
  "mcpServers": {
    "qaynaq": {
      "command": "npx",
      "args": ["mcp-remote", "http://localhost:8080/mcp"]
    }
  }
}
```

## 6. Try It Out

Open your AI client and try some prompts. Here are a few to get you started:

**Simple order lookup:**

> What is the status of order ORD-1002?

The AI will call `lookup_order` and tell you it is still processing.

**Customer history:**

> Can you check all orders for alice@example.com?

The AI will call `lookup_customer` and list both of Alice's orders.

**Full support flow:**

> A customer named Sanan Guliyev emailed us saying: "I placed two orders last week and neither has shipped yet. What is going on?" His email is sanan@example.com. Can you look into it and draft a reply and use it as is?

This is where it gets interesting. The AI will call `lookup_customer` to find Sanan's orders and their statuses, then call `lookup_order` on each one to get the full details like items and shipping info, and finally call `draft_support_reply` to write a response using everything it found.

## Adding More Data

To add more orders for testing, just insert rows into the database:

```bash
sqlite3 support.sqlite <<'SQL'
INSERT INTO orders VALUES
('ORD-1006', 'Diana Lee', 'diana@example.com', 'shipped',
 'Webcam x1, Ring Light x1', 65.00,
 '321 Maple Drive, Denver, CO 80201',
 '2025-03-09', '2025-03-11', 'TRK-445566');
SQL
```

The MCP tools query the database on every call, so new data is available immediately. No need to restart any flows.

## What is Happening Under the Hood

When the AI calls one of your tools, here is the flow:

1. The AI client sends a tool call to the Qaynaq MCP endpoint at `/mcp`
2. Qaynaq routes the call to the matching flow based on the tool name
3. The flow's processor runs (SQL query or AI generation)
4. The result goes back through Sync Response to the AI client

Each tool is just a regular Qaynaq flow with an MCP Tool input. You can add more processors, change the database, or swap in a different AI provider without touching the client configuration.
