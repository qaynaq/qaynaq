---
sidebar_position: 4
---

# Files

The **File Manager** lets you create and manage files that flow components can reference at runtime. Instead of embedding file contents directly in component configuration, you create files once and reference them by key.

## How It Works

1. Open the File Manager in the Qaynaq UI and create a new file with a unique **key** (e.g., `schemas/order.json`) and its content.
2. In a component's file field, select the file.
3. When the flow is assigned to a worker, Qaynaq automatically transfers the file to the worker and resolves the reference to a local path.

Files are stored in the coordinator's database and distributed to workers on demand. You don't need to manually place files on worker machines.

## Versioning

Files support versioning. When you update a file, Qaynaq creates a new version while preserving the previous one. Flows that are reassigned after an update will use the latest version.

## Supported Components

| Component | Field | Description |
|-----------|-------|-------------|
| [JSON Schema](/docs/components/processors/json-schema) | Schema File | JSON Schema file for message validation |
