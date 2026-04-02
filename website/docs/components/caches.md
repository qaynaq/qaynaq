---
sidebar_position: 4
---

# Caches

Cache resources provide key-value storage that components can use to persist state. They are created through the Qaynaq UI and linked to flows that need them.

Components like [CDC MySQL](/docs/components/inputs/cdc-mysql) and [Shopify](/docs/components/inputs/shopify) reference a cache by its **label** to store and retrieve data such as binlog positions or pagination cursors.

## How caches work

Each cache resource has:

- **Label**: A unique name used by components to reference this cache (e.g., `positions`).
- **Component**: The cache type (e.g., `memory`, `redis`, `file`).
- **Config**: Type-specific settings.

When a flow is linked to a cache resource, the cache becomes available to the flow's components at runtime. Components access the cache by label and store data under specific keys.

Multiple flows can share a single cache resource as long as they use different keys.

---

## Memory

In-process in-memory cache. Fast but does not survive restarts. Good for development and testing.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Default TTL | string | `5m` | TTL for each item. After this period the item is eligible for removal on next compaction |
| Compaction Interval | string | `60s` | How often to compact and remove expired items. Set to empty string to disable |
| Init Values | object | `{}` | Key/value pairs to pre-populate the cache on initialization |
| Shards | integer | `1` | Number of logical shards. Increasing shards can improve performance under high concurrency |

---

## Redis

Persistent, distributed cache backed by Redis. Recommended for production use with components that need to survive restarts (like CDC position tracking).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| URL | string | — | **(required)** Redis server URL. Database can be specified as URL path (e.g., `redis://localhost:6379/1`) |
| Kind | string | `simple` | Client type: `simple`, `cluster`, or `failover` |
| Master | string | — | Redis master name (when Kind is `failover`) |
| Prefix | string | — | Key prefix to prevent collisions with other services |
| Default TTL | string | — | Default TTL for items. Empty means no expiry |

**TLS settings:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Enabled | boolean | `false` | Whether custom TLS settings are enabled |
| Skip Cert Verify | boolean | `false` | Whether to skip server-side certificate verification |
| Enable Renegotiation | boolean | `false` | Whether to allow the remote server to repeatedly request renegotiation |
| Root CAs | string | — | An optional root certificate authority to use |
| Root CAs File | string | — | An optional path of a root certificate authority file to use |
| Client Certs | array | `[]` | A list of client certificates to use |

**Retry settings:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Initial Interval | string | `500ms` | The initial period to wait between retry attempts |
| Max Interval | string | `1s` | The maximum period to wait between retry attempts |
| Max Elapsed Time | string | `5s` | The maximum overall period of time to spend on retry attempts before aborting |

---

## Memcached

Distributed cache backed by Memcached.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Addresses | array | — | **(required)** List of Memcached server addresses |
| Prefix | string | — | Key prefix to prevent collisions |
| Default TTL | string | `300s` | Default TTL for items |

**Retry settings:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Initial Interval | string | `1s` | The initial period to wait between retry attempts |
| Max Interval | string | `5s` | The maximum period to wait between retry attempts |
| Max Elapsed Time | string | `30s` | The maximum overall period of time to spend on retry attempts before aborting |

---

## File

File-system-based cache. Each key is stored as a file in the specified directory. Survives restarts but is local to the machine.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Directory | string | — | **(required)** Directory to store cache files in |

---

## LRU

In-memory Least Recently Used cache with a fixed capacity. Does not survive restarts.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Capacity | integer | `1024` | **(required)** Maximum number of items |
| Init Values | object | `{}` | Key/value pairs to pre-populate the cache on initialization |

---

## TTLRU

In-memory LRU cache with per-item TTL expiry. Does not survive restarts.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Capacity | integer | `1024` | **(required)** Maximum number of items |
| Default TTL | string | `5m` | Default TTL for items |
| Init Values | object | `{}` | Key/value pairs to pre-populate the cache on initialization |

---

## Ristretto

High-performance in-memory cache using the Ristretto library. Provides better hit ratios than simple LRU under many workloads. Does not survive restarts.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Default TTL | string | `5m` | Default TTL for items |
| Max Cost | integer | `1073741824` | The maximum size of the cache in bytes (default ~1 GB) |
| Num Counters | integer | `10000000` | The number of 4-bit access counters to keep for admission and eviction |

---

## Noop

A no-operation cache that discards all writes and returns not-found for all reads. Useful for testing or disabling caching without changing the pipeline structure.

No configuration fields.

---

## Choosing a cache type

| Use case | Recommended cache |
|----------|-------------------|
| Development / testing | Memory |
| Production CDC position tracking | Redis or File |
| Single-node deployment | File |
| Multi-node / distributed deployment | Redis or Memcached |
| High-throughput state with eviction | Ristretto or TTLRU |
