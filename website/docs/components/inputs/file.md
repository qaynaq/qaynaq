# File

Reads messages from files on disk. Paths support glob patterns, including super globs (`**`), so a single input can consume an entire directory tree. Each file is read sequentially and broken into messages by the chosen scanner.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Paths | list of strings | `[]` | Files to consume. Glob patterns (`*`, `**`) are expanded at startup |
| Scanner | [scanner](/docs/components/scanners) | Lines | How each file's bytes are split into messages |
| Delete On Finish | boolean | `false` | Delete each file from disk once fully consumed |
| Auto Replay Nacks | boolean | `true` | Automatically replay rejected messages |

## Metadata

Each emitted message carries these metadata fields, available to downstream processors:

| Field | Description |
|-------|-------------|
| `path` | Absolute path of the source file |
| `mod_time_unix` | File modification time as a Unix timestamp |
| `mod_time` | File modification time in RFC 3339 format |
