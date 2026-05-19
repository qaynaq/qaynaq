# Decompress

Decompresses the byte stream with the chosen algorithm and delegates the decompressed bytes to a child scanner. Stack this on top of any other scanner to read compressed sources: `.csv.gz`, `.json.zst`, `.tar.bz2`, and so on.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Algorithm | enum | `gzip` | One of `gzip`, `pgzip`, `zlib`, `bzip2`, `flate`, `snappy`, `lz4`, `zstd` |
| Into | scanner | To The End | The child scanner that consumes the decompressed bytes |

## Algorithm cheat sheet

| Algorithm | Typical file extensions |
|-----------|-------------------------|
| `gzip`, `pgzip` | `.gz`, `.tgz` |
| `zlib`, `flate` | raw zlib/deflate streams |
| `bzip2` | `.bz2` |
| `snappy` | `.sz`, `.snappy` |
| `lz4` | `.lz4` |
| `zstd` | `.zst`, `.zstd` |
