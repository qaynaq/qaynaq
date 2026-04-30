# Google Sheets

Performs Google Sheets operations - create, read, update, and format spreadsheets, worksheets, rows, and cells.

:::tip MCP Tool Pack Available
Want to expose Google Sheets actions as MCP tools for AI assistants? Use the [Google Sheets MCP Tool Pack](/docs/guides/mcp-tool-packs/google-sheets) to deploy all 28 tools in one step - no manual configuration needed.
:::

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Service Account JSON | secret | - | Google service account credentials JSON (required) |
| Delegate To | string | - | Email to impersonate via Domain-Wide Delegation |
| Action | select | - | The spreadsheet operation to perform (required) |
| Spreadsheet ID | string | - | The ID of the target spreadsheet (required for most actions) |
| Sheet Name | string | - | Name of the worksheet/sheet tab |
| New Sheet Name | string | - | New name when renaming a worksheet |
| Range | string | - | Cell range in A1 notation (e.g. `A1:D10`) |
| Row Index | integer | - | Row number for single-row operations |
| Start Row | integer | - | Starting row for range operations |
| End Row | integer | - | Ending row for range operations |
| Values | string | - | JSON array of values for row/cell operations |
| Column Name | string | - | Column header name for lookup or create operations |
| Lookup Value | string | - | Value to search for in lookup operations |
| Lookup Column | string | - | Column to search in for lookup operations |
| Sort Column | string | - | Column to sort by |
| Sort Order | string | `ascending` | Sort direction: `ascending` or `descending` |
| Title | string | - | Spreadsheet or worksheet title |
| Destination Spreadsheet ID | string | - | Target spreadsheet for copy operations |
| Format | string | - | JSON formatting object for format actions |
| Validation Rule | string | - | JSON data validation rule |
| Condition Format Rule | string | - | JSON conditional formatting rule |
| Max Results | integer | `100` | Maximum number of rows to return |

## Authentication

This processor supports two authentication methods: **OAuth Connection** (recommended for personal accounts) and **Service Account** (for server-to-server automation).

### Option A: OAuth Connection (recommended)

OAuth lets you authorize Qaynaq to act as your Google account. Data is owned by you, using your quota. You authorize once and it works indefinitely.

Follow the [Google OAuth Setup](/docs/guides/google-oauth-setup) guide to create a connection. Make sure to enable the **Google Sheets API** and add the `auth/spreadsheets` scope. Then select your connection from the **OAuth Connection** dropdown in the processor configuration.

### Option B: Service Account

Service accounts are best for Google Workspace organizations with Domain-Wide Delegation.

#### Step 1: Create a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project or select an existing one
3. Navigate to **IAM & Admin** > **Service Accounts**
4. Click **Create Service Account**, give it a name, and click **Done**
5. Click the service account, go to the **Keys** tab
6. Click **Add Key** > **Create new key** > **JSON**
7. Download the JSON key file

#### Step 2: Enable the Google Sheets API

1. In Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for **Google Sheets API** and click **Enable**

#### Step 3: Store the Credentials

1. Open the downloaded JSON key file and copy its entire contents
2. In Qaynaq, go to **Secrets**
3. Create a new secret (e.g. key: `GOOGLE_SHEETS_SA`) and paste the JSON as the value
4. In the Google Sheets processor, select `GOOGLE_SHEETS_SA` from the Service Account JSON dropdown

#### Step 4: Share Your Spreadsheet with the Service Account

A service account cannot access any spreadsheet by default. You must explicitly share each spreadsheet with the service account's email address:

1. Open the JSON key file and find the `client_email` field (e.g. `my-service@my-project.iam.gserviceaccount.com`)
2. Open your Google Spreadsheet and click **Share**
3. Paste the service account email and grant **Editor** access
4. Click **Send** (you can uncheck "Notify people" since it's a service account)

:::warning
Without this step, all operations will fail with a permission error. The service account can only read and write spreadsheets that have been shared with it.
:::

#### Domain-Wide Delegation (Google Workspace only)

As an alternative to sharing individual spreadsheets, Google Workspace administrators can grant the service account access to all users' spreadsheets via Domain-Wide Delegation:

1. In Google Cloud Console, go to the service account details
2. Enable **Domain-Wide Delegation** and note the Client ID
3. In Google Workspace Admin Console, go to **Security** > **API Controls** > **Domain-wide Delegation**
4. Add the Client ID with the scope `https://www.googleapis.com/auth/spreadsheets`
5. Set the **Delegate To** field to the email of the user whose spreadsheets you want to access

## Actions

### Spreadsheet Management

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><line x1="12" y1="12" x2="12" y2="18"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
<div class="action-card-content">
<h4>Create Spreadsheet</h4>
<p>Creates a new Google Sheets spreadsheet with a title and optional initial worksheet name.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="3" y1="15" x2="21" y2="15"/><line x1="9" y1="3" x2="9" y2="21"/><circle cx="15" cy="12" r="1"/></svg>
<div class="action-card-content">
<h4>Get Spreadsheet by ID</h4>
<p>Retrieves spreadsheet metadata and properties by its ID.</p>
</div>
</div>
</div>

### Worksheet Management

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><line x1="14" y1="14" x2="14" y2="18"/><line x1="12" y1="16" x2="16" y2="16"/></svg>
<div class="action-card-content">
<h4>Create Worksheet</h4>
<p>Adds a new sheet tab to an existing spreadsheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><rect x="5" y="5" width="4" height="4" rx="1"/></svg>
<div class="action-card-content">
<h4>Find Worksheet</h4>
<p>Finds a worksheet by name within a spreadsheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
<div class="action-card-content">
<h4>Find or Create Worksheet</h4>
<p>Finds a worksheet by name, or creates it if it does not exist.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><line x1="13" y1="14" x2="17" y2="18"/><line x1="17" y1="14" x2="13" y2="18"/></svg>
<div class="action-card-content">
<h4>Delete Worksheet</h4>
<p>Removes a sheet tab from a spreadsheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><path d="M13 14l4 0"/><path d="M13 17l2 0"/></svg>
<div class="action-card-content">
<h4>Rename Worksheet</h4>
<p>Changes the name of an existing sheet tab.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="8" height="18" rx="2"/><rect x="14" y="3" width="8" height="18" rx="2"/><path d="M10 12h4"/><polyline points="12 10 14 12 12 14"/></svg>
<div class="action-card-content">
<h4>Copy Worksheet</h4>
<p>Copies a worksheet to the same or another spreadsheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><circle cx="15" cy="15" r="2"/><path d="M17 13v4h-4"/></svg>
<div class="action-card-content">
<h4>Change Sheet Properties</h4>
<p>Updates sheet-level properties such as grid size, tab color, or frozen rows and columns.</p>
</div>
</div>
</div>

### Row Operations

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
<div class="action-card-content">
<h4>Create Row</h4>
<p>Appends a new row at the bottom of the sheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/><line x1="5" y1="16" x2="19" y2="16"/><line x1="5" y1="8" x2="19" y2="8"/></svg>
<div class="action-card-content">
<h4>Create Multiple Rows</h4>
<p>Appends multiple rows at the bottom of the sheet in a single request.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/><line x1="5" y1="19" x2="19" y2="19"/></svg>
<div class="action-card-content">
<h4>Create Row at Top</h4>
<p>Inserts a new row at the top of the sheet, below the header row.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="8" y1="13" x2="16" y2="13"/></svg>
<div class="action-card-content">
<h4>Get Row</h4>
<p>Retrieves a single row by its row index.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
<div class="action-card-content">
<h4>Get Many Rows</h4>
<p>Retrieves multiple rows from a range.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="3" y1="15" x2="21" y2="15"/><line x1="9" y1="3" x2="9" y2="21"/><line x1="15" y1="3" x2="15" y2="21"/></svg>
<div class="action-card-content">
<h4>Get Data Range</h4>
<p>Returns the full data range of the sheet, including header row and all populated rows.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
<div class="action-card-content">
<h4>Lookup Row</h4>
<p>Finds the first row where a specified column matches a given value.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="8" y1="11" x2="14" y2="11"/><line x1="8" y1="14" x2="14" y2="14"/></svg>
<div class="action-card-content">
<h4>Lookup Rows</h4>
<p>Finds all rows where a specified column matches a given value.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
<div class="action-card-content">
<h4>Update Row</h4>
<p>Updates a single row at the specified row index.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/><line x1="2" y1="18" x2="10" y2="18"/><line x1="2" y1="14" x2="6" y2="14"/></svg>
<div class="action-card-content">
<h4>Update Rows</h4>
<p>Updates multiple rows in a specified range.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
<div class="action-card-content">
<h4>Clear Rows</h4>
<p>Clears the content of rows in a specified range without deleting them.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="6" x2="21" y2="6"/><line x1="5" y1="6" x2="5" y2="20"/><line x1="19" y1="6" x2="19" y2="20"/><path d="M5 20h14"/><line x1="9" y1="10" x2="15" y2="16"/><line x1="15" y1="10" x2="9" y2="16"/></svg>
<div class="action-card-content">
<h4>Delete Rows</h4>
<p>Permanently removes rows from the sheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="3" x2="12" y2="21"/><line x1="5" y1="12" x2="19" y2="12"/><line x1="12" y1="3" x2="12" y2="7"/><line x1="3" y1="9" x2="21" y2="9"/></svg>
<div class="action-card-content">
<h4>Create Column</h4>
<p>Adds a new column header to the sheet.</p>
</div>
</div>
</div>

### Formatting & Advanced

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="8" height="18" rx="2"/><rect x="14" y="3" width="8" height="18" rx="2"/><path d="M10 12h4"/><polyline points="12 10 14 12 12 14"/></svg>
<div class="action-card-content">
<h4>Copy Range</h4>
<p>Copies a range of cells to a new location within the same or another spreadsheet.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 9 3 15 6 21 3 21 18 15 21 9 18 3 21"/><line x1="9" y1="3" x2="9" y2="18"/><line x1="15" y1="6" x2="15" y2="21"/></svg>
<div class="action-card-content">
<h4>Sort Range</h4>
<p>Sorts a range of rows by a specified column in ascending or descending order.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><rect x="11" y="11" width="8" height="8" rx="1" fill="currentColor" opacity="0.15"/></svg>
<div class="action-card-content">
<h4>Format Cells</h4>
<p>Applies formatting (bold, color, borders, number format, etc.) to a range of cells.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="9" y1="3" x2="9" y2="21"/><rect x="3" y="11" width="18" height="4" fill="currentColor" opacity="0.15"/></svg>
<div class="action-card-content">
<h4>Format Row</h4>
<p>Applies formatting to an entire row.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>
<div class="action-card-content">
<h4>Set Data Validation</h4>
<p>Sets data validation rules on a range (e.g. dropdown lists, number ranges, custom formulas).</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="3" y1="15" x2="21" y2="15"/><line x1="9" y1="3" x2="9" y2="21"/><rect x="10" y="10" width="5" height="4" rx="0.5" fill="currentColor" opacity="0.3"/></svg>
<div class="action-card-content">
<h4>Create Conditional Formatting</h4>
<p>Creates a conditional formatting rule that highlights cells based on their values or a custom formula.</p>
</div>
</div>
</div>

## Dynamic Fields

All action parameter fields support Bento interpolation functions, allowing dynamic values from message content using `${!this.field_name}` syntax. This enables processing batches of spreadsheet operations or reacting to upstream data. For example, setting Spreadsheet ID to `${!this.spreadsheet_id}` will read the spreadsheet ID from the incoming message.

JSON fields (Values, Format, Validation Rule, Condition Format Rule) also support interpolation. For example, set Values to `${!this.row_data}` to pass through a JSON array from the message.

Static fields (not interpolated): Service Account JSON, Delegate To, Action, Max Results.

## Output Format

All actions return a structured JSON object:

- **Spreadsheet actions** return a `spreadsheet` object containing the spreadsheet metadata (id, title, sheets, etc.)
- **Worksheet actions** return a `sheet` object with the sheet properties (sheetId, title, index, gridProperties)
- **Row actions** (get_row, lookup_row) return a `row` object with column-value pairs
- **List actions** (get_many_rows, lookup_rows, get_data_range) return a `rows` array and a `count` field
- **Mutating actions** (create_row, update_row, clear_rows, delete_rows) return confirmation with affected range information
- **find_or_create_worksheet** includes a `created` boolean field

:::tip
Use a Mapping processor after Google Sheets to extract specific fields from the response, such as cell values or the spreadsheet ID.
:::
