# CSV

Parses a CSV stream row by row. Each row becomes one message.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Custom Delimiter | string | `,` | Field separator. Override for TSV (`\t`) or pipe-delimited files |
| Parse Header Row | boolean | `true` | Use the first row as field names. When off, rows are emitted as arrays |
| Lazy Quotes | boolean | `false` | Allow quotes inside unquoted fields and unescaped quotes inside quoted fields |
| Continue On Error | boolean | `false` | On a parsing error, emit an error message and keep going instead of stopping |
| Expected Headers | list of strings | `[]` | Optional. Assert that the header row matches these names exactly. Requires Parse Header Row |
| Expected Number Of Fields | integer | `0` | Optional. Assert every row has exactly this many fields. `0` disables the check |

## Metadata

| Field | Description |
|-------|-------------|
| `csv_row` | Index of each row in the file, starting at 0 |
