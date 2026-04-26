# Python

Executes a Python script for each message inside a sandboxed WebAssembly runtime (Python 3.12). The incoming message is exposed as the global variable `this`, and the value assigned to `root` becomes the new message.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Script | Python | - | The Python script to execute (required) |
| Imports | list of strings | `[]` | Optional Python modules to pre-import for the script |

## How messages flow

Each message is decoded into Python as `this`. Whatever you assign to `root` becomes the outgoing message. To drop a message, assign `None` to `root`.

A typical script looks like:

```python
root = {
    "full_name": f"{this['first_name']} {this['last_name']}",
    "age": this["age"],
}
```

## Imports

The runtime ships with a curated set of standard libraries. To use one in your script, list it in Imports. For example, to use `math` and `json`, add both as separate entries.

## Runtime

This processor is experimental. The script runs in a WASM-hosted Python interpreter, so it is isolated from the host filesystem and network. C extensions and arbitrary pip packages are not available - only modules that ship with the embedded runtime.

## When to use

- You need transformation logic that is awkward to express in Bloblang
- The team is more comfortable maintaining Python than learning Bloblang
- You want a sandboxed alternative to the Command processor that does not spawn host processes

For simple field renames and conversions, prefer Mapping - Bloblang has no per-message interpreter startup cost.
