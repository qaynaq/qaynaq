---
sidebar_position: 2
---

# Google Drive

Deploys up to 22 MCP tools for managing Google Drive files, folders, permissions, and shared drives.

## Shared Configuration

| Field | Required | Description |
|-------|----------|-------------|
| OAuth Connection | Yes | Google OAuth connection. Set up in the [Connections](/docs/guides/google-oauth-setup) page first. Make sure to enable the **Google Drive API** and add the `auth/drive` scope. |

## Included Tools

| Tool | Description |
|------|-------------|
| `google_drive_copy_file` | Create a copy of the specified file |
| `google_drive_delete_file_permanent` | Permanently delete a file (cannot be undone) |
| `google_drive_export_file` | Export Google Workspace files to different formats |
| `google_drive_upload_file` | Upload a file to Google Drive from a URL or content |
| `google_drive_create_folder` | Create a new, empty folder |
| `google_drive_move_file` | Move a file from one folder to another |
| `google_drive_create_file_from_text` | Create a new file from plain text |
| `google_drive_remove_file_permission` | Remove specific user access to a file |
| `google_drive_replace_file` | Upload new content to replace an existing file |
| `google_drive_create_shared_drive` | Create a new shared drive (Team Drive) |
| `google_drive_add_file_sharing` | Add a sharing permission to a file |
| `google_drive_create_shortcut` | Create a shortcut to a file |
| `google_drive_update_metadata` | Update file or folder metadata |
| `google_drive_rename` | Update the name of a file or folder |
| `google_drive_delete_file` | Move a file to the trash |
| `google_drive_list_files` | Retrieve a list of files based on query parameters |
| `google_drive_get_file` | Get a file or folder by its ID |
| `google_drive_get_permissions` | List all users who have access to a file |
| `google_drive_find_file` | Search for a specific file by name |
| `google_drive_find_folder` | Search for a specific folder by name |
| `google_drive_find_or_create_file` | Find a file by name, or create one if not found |
| `google_drive_find_or_create_folder` | Find a folder by name, or create one if not found |
