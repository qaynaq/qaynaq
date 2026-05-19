# Skip BOM

Strips a leading UTF-8, UTF-16, or UTF-32 byte order mark from the stream and delegates the rest to a child scanner. Use this for files exported from Windows tools (Notepad, Excel) that prepend a BOM, then layer it on top of [Lines](/docs/components/scanners/lines), [CSV](/docs/components/scanners/csv), or any other scanner.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Into | scanner | To The End | The child scanner that consumes the stream once the BOM has been stripped |
