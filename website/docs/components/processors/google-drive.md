# Google Drive

Performs Google Drive operations - manage files, folders, permissions, and shared drives.

:::tip MCP Tool Pack Available
Want to expose Google Drive actions as MCP tools for AI assistants? Use the [Google Drive MCP Tool Pack](/docs/guides/mcp-tool-packs/google-drive) to deploy all 22 tools in one step - no manual configuration needed.
:::

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| Service Account JSON | secret | - | Google service account credentials JSON (required) |
| Delegate To | string | - | Email to impersonate via Domain-Wide Delegation |
| Action | select | - | The Google Drive operation to perform (required) |
| File ID | string | - | File or folder ID for file-specific operations |
| File Name | string | - | File or folder name |
| Folder ID | string | - | Parent folder ID (defaults to root) |
| Destination Folder ID | string | - | Target folder for move operations |
| MIME Type | string | - | MIME type for export, upload, or create operations |
| Content | string | - | Text content for file creation |
| File URL | string | - | URL to download file content from |
| Description | string | - | File description |
| Email | string | - | Email address for permission operations |
| Role | select | `reader` | Permission role: reader, writer, commenter, owner |
| Permission Type | select | `user` | Permission type: user, group, domain, anyone |
| Permission ID | string | - | Permission ID for removal |
| Query | string | - | Search query using Drive query syntax |
| Max Results | integer | `100` | Maximum results to return |
| Starred | boolean | `false` | Whether the file is starred |
| Folder Color | string | - | Folder color as hex |
| Custom Properties | string | - | JSON object of custom properties |
| Shared Drive Name | string | - | Name for new shared drive |
| Target File ID | string | - | Target file for shortcut creation |
| Send Notification | boolean | `true` | Send email when sharing |

## Authentication

This processor supports two authentication methods: **OAuth Connection** (recommended for personal accounts) and **Service Account** (for server-to-server automation).

### Option A: OAuth Connection (recommended)

OAuth lets you authorize Qaynaq to act as your Google account. Files are owned by you, using your quota. You authorize once and it works indefinitely.

Follow the [Google OAuth Setup](/docs/guides/google-oauth-setup) guide to create a connection. Make sure to enable the **Google Drive API** and add the `auth/drive` scope. Then select your connection from the **OAuth Connection** dropdown in the processor configuration.

### Option B: Service Account

Service accounts are best for Google Workspace organizations with Domain-Wide Delegation.

:::warning
Service accounts have their own 15GB storage quota. Files created by a service account count against its quota, even when placed in shared folders. For personal Google accounts, use OAuth instead.
:::

#### Step 1: Create a Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project or select an existing one
3. Navigate to **IAM & Admin** > **Service Accounts**
4. Click **Create Service Account**, give it a name, and click **Done**
5. Click the service account, go to the **Keys** tab
6. Click **Add Key** > **Create new key** > **JSON**
7. Download the JSON key file

#### Step 2: Enable the Google Drive API

1. In Google Cloud Console, go to **APIs & Services** > **Library**
2. Search for **Google Drive API** and click **Enable**

#### Step 3: Store the Credentials

1. Open the downloaded JSON key file and copy its entire contents
2. In Qaynaq, go to **Settings** > **Secrets**
3. Create a new secret (e.g. key: `GOOGLE_DRIVE_SA`) and paste the JSON as the value
4. In the Google Drive processor, select `GOOGLE_DRIVE_SA` from the Service Account JSON dropdown

#### Step 4: Share Files with the Service Account

A service account cannot access any files by default. You must explicitly share files or folders with the service account's email address:

1. Open the JSON key file and find the `client_email` field (e.g. `my-service@my-project.iam.gserviceaccount.com`)
2. Open a Google Drive folder or file and click **Share**
3. Paste the service account email and grant **Editor** access
4. Click **Send** (you can uncheck "Notify people" since it's a service account)

#### Domain-Wide Delegation (Google Workspace only)

As an alternative to sharing individual files, Google Workspace administrators can grant the service account access to all users' files via Domain-Wide Delegation:

1. In Google Cloud Console, go to the service account details
2. Enable **Domain-Wide Delegation** and note the Client ID
3. In Google Workspace Admin Console, go to **Security** > **API Controls** > **Domain-wide Delegation**
4. Add the Client ID with the scope `https://www.googleapis.com/auth/drive`
5. Set the **Delegate To** field to the email of the user whose files you want to access

## Actions

### File Management

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="8" height="18" rx="2"/><rect x="14" y="3" width="8" height="18" rx="2"/><path d="M10 12h4"/><polyline points="12 10 14 12 12 14"/></svg>
<div class="action-card-content">
<h4>Copy File</h4>
<p>Create a copy of the specified file.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><circle cx="12" cy="15" r="1"/></svg>
<div class="action-card-content">
<h4>Get File</h4>
<p>Retrieve file or folder metadata by its ID.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
<div class="action-card-content">
<h4>Find File</h4>
<p>Search for a specific file by name.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
<div class="action-card-content">
<h4>Find or Create File</h4>
<p>Find a file by name, or create a new one if not found.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
<div class="action-card-content">
<h4>List Files</h4>
<p>Retrieve a list of files based on query parameters.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"/></svg>
<div class="action-card-content">
<h4>Rename</h4>
<p>Update the name of a file or folder.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><path d="M9 14h6"/></svg>
<div class="action-card-content">
<h4>Update Metadata</h4>
<p>Update file or folder metadata including name, description, starred status, and custom properties.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 3 21 3 21 9"/><polyline points="9 21 3 21 3 15"/><line x1="21" y1="3" x2="14" y2="10"/><line x1="3" y1="21" x2="10" y2="14"/></svg>
<div class="action-card-content">
<h4>Move File</h4>
<p>Move a file from one folder to another.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="6" x2="21" y2="6"/><line x1="5" y1="6" x2="5" y2="20"/><line x1="19" y1="6" x2="19" y2="20"/><path d="M5 20h14"/><line x1="9" y1="10" x2="15" y2="16"/><line x1="15" y1="10" x2="9" y2="16"/></svg>
<div class="action-card-content">
<h4>Delete File (Trash)</h4>
<p>Move a file to the trash.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="3" y1="6" x2="21" y2="6"/><line x1="5" y1="6" x2="5" y2="20"/><line x1="19" y1="6" x2="19" y2="20"/><path d="M5 20h14"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>
<div class="action-card-content">
<h4>Delete File (Permanent)</h4>
<p>Permanently delete a file. This action cannot be undone.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
<div class="action-card-content">
<h4>Create Shortcut</h4>
<p>Create a shortcut to a file.</p>
</div>
</div>
</div>

### Content Operations

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="12" x2="12" y2="18"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
<div class="action-card-content">
<h4>Create File From Text</h4>
<p>Create a new file from plain text content.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 16 12 12 8 16"/><line x1="12" y1="12" x2="12" y2="21"/><path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"/></svg>
<div class="action-card-content">
<h4>Upload File</h4>
<p>Upload a file to Google Drive from a URL or content.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
<div class="action-card-content">
<h4>Export File</h4>
<p>Export Google Workspace files (Docs, Sheets, Slides) to different formats like PDF, Word, Excel.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><path d="M12 18v-6"/><path d="M9 15l3 3 3-3"/></svg>
<div class="action-card-content">
<h4>Replace File</h4>
<p>Upload new content to replace an existing file.</p>
</div>
</div>
</div>

### Folder Operations

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/><line x1="12" y1="11" x2="12" y2="17"/><line x1="9" y1="14" x2="15" y2="14"/></svg>
<div class="action-card-content">
<h4>Create Folder</h4>
<p>Create a new, empty folder.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><path d="M22 19a2 2 0 0 1-2 2H4"/></svg>
<div class="action-card-content">
<h4>Find Folder</h4>
<p>Search for a specific folder by name.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/><line x1="11" y1="8" x2="11" y2="14"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
<div class="action-card-content">
<h4>Find or Create Folder</h4>
<p>Find a folder by name, or create a new one if not found.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/><path d="M12 11v6"/></svg>
<div class="action-card-content">
<h4>Create Shared Drive</h4>
<p>Create a new shared drive (Team Drive).</p>
</div>
</div>
</div>

### Permissions & Sharing

<div class="action-grid">
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="19" y1="8" x2="19" y2="14"/><line x1="22" y1="11" x2="16" y2="11"/></svg>
<div class="action-card-content">
<h4>Add File Sharing</h4>
<p>Add a sharing permission to a file. Provides a sharing URL.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><line x1="17" y1="11" x2="23" y2="11"/></svg>
<div class="action-card-content">
<h4>Remove File Permission</h4>
<p>Remove specific user access to a file.</p>
</div>
</div>
<div class="action-card">
<svg class="action-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/></svg>
<div class="action-card-content">
<h4>Get Permissions</h4>
<p>List all users who have access to a file.</p>
</div>
</div>
</div>

## Dynamic Fields

All action parameter fields support Bento interpolation functions, allowing dynamic values from message content using `${!this.field_name}` syntax. This enables processing batches of file operations or reacting to upstream data. For example, setting File ID to `${!this.file_id}` will read the file ID from the incoming message.

Static fields (not interpolated): Service Account JSON, Delegate To, Action.

## Output Format

All actions return a structured JSON object:

- **File actions** return a `file` object containing file metadata (id, name, mime_type, web_view_link, etc.)
- **Folder actions** return a `folder` object with folder metadata
- **List actions** (list_files, find_file, find_folder) return an array and a `count` field
- **Permission actions** return permission details or a permissions list
- **find_or_create** actions include a `created` boolean field
- **Delete actions** return `{deleted: true, file_id: "..."}`

:::tip
Use a Mapping processor after Google Drive to extract specific fields from the response, such as the file ID or the sharing URL.
:::
