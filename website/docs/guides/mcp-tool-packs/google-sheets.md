---
sidebar_position: 3
---

# Google Sheets

Deploys up to 28 MCP tools for reading, writing, and managing Google Sheets spreadsheets.

## Shared Configuration

| Field | Required | Description |
|-------|----------|-------------|
| OAuth Connection | Yes | Google OAuth connection. Set up in the [Connections](/docs/guides/google-oauth-setup) page first. Make sure to enable the **Google Sheets API** and add the `auth/spreadsheets` scope. |

## Included Tools

| Tool | Description |
|------|-------------|
| `google_sheets_create_spreadsheet` | Create a new Google Sheets spreadsheet |
| `google_sheets_create_column` | Create a new column in a specific spreadsheet |
| `google_sheets_create_row` | Create a new row in a specific spreadsheet |
| `google_sheets_create_rows` | Create one or more new rows in a specific spreadsheet (with line item support) |
| `google_sheets_create_row_at_top` | Create a new spreadsheet row at the top (after the header row) |
| `google_sheets_change_sheet_properties` | Update properties like frozen rows/columns, sheet position, and visibility |
| `google_sheets_create_conditional_formatting` | Apply conditional formatting to cells based on their values |
| `google_sheets_copy_range` | Copy data from one range to another within a spreadsheet |
| `google_sheets_copy_worksheet` | Create a new worksheet by copying an existing one |
| `google_sheets_create_worksheet` | Create a new worksheet in a spreadsheet |
| `google_sheets_clear_rows` | Clear row contents while keeping the rows intact |
| `google_sheets_delete_worksheet` | Permanently delete a worksheet from a spreadsheet |
| `google_sheets_delete_rows` | Delete selected rows and all associated data |
| `google_sheets_format_cells` | Apply date, number, or style formatting to a range of cells |
| `google_sheets_format_row` | Format a row in a specific spreadsheet |
| `google_sheets_rename_worksheet` | Rename a worksheet in a spreadsheet |
| `google_sheets_set_data_validation` | Set data validation rules on a range of cells |
| `google_sheets_sort_range` | Sort data within a specified range by a chosen column |
| `google_sheets_update_row` | Update a row in a specific spreadsheet |
| `google_sheets_update_rows` | Update one or more rows in a specific spreadsheet (with line item support) |
| `google_sheets_lookup_rows` | Find up to 500 rows based on a column and value |
| `google_sheets_find_worksheet` | Find a worksheet by title |
| `google_sheets_get_data_range` | Get the data range of a worksheet |
| `google_sheets_get_rows` | Return up to 1,500 rows as JSON or line items |
| `google_sheets_get_row` | Get a specific row by its row number |
| `google_sheets_get_spreadsheet` | Get a specific spreadsheet by its ID |
| `google_sheets_lookup_row` | Find a specific row based on a column and value |
| `google_sheets_find_or_create_worksheet` | Find or create a specific worksheet |
