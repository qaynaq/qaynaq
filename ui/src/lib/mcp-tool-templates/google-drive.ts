import { param, type McpToolTemplate, type TemplatePack } from "./types";

function gdTool(
  id: string,
  name: string,
  description: string,
  action: string,
  parameters: ReturnType<typeof param>[],
  configOverrides: Record<string, string | number | boolean> = {},
): McpToolTemplate {
  return {
    id,
    name,
    description,
    parameters,
    processor: {
      component: "google_drive",
      config: { action, ...configOverrides },
    },
  };
}

export const googleDrivePack: TemplatePack = {
  id: "google_drive",
  name: "Google Drive",
  description: "22 MCP tools for managing Google Drive files, folders, permissions, and shared drives.",
  sharedConfig: [
    {
      key: "oauth_connection",
      title: "OAuth Connection",
      description: "Google OAuth connection. Set up in Connections page first.",
      type: "dynamic_select",
      required: true,
      dataSource: "connections",
    },
  ],
  templates: [
    gdTool(
      "gd_copy_file",
      "google_drive_copy_file",
      "Create a copy of the specified file in Google Drive.",
      "copy_file",
      [
        param("file_id", "The ID of the file to copy", true),
        param("file_name", "Name for the copy. Defaults to 'Copy of <original name>' if not provided."),
        param("folder_id", "Destination folder ID. Defaults to the configured folder if not provided."),
      ],
      {
        file_id: "${!this.file_id}",
        file_name: '${!this.file_name.or("")}',
        folder_id: '${!this.folder_id.or("")}',
      },
    ),
    gdTool(
      "gd_delete_file_permanent",
      "google_drive_delete_file_permanent",
      "Permanently delete a file from Google Drive. This action cannot be undone.",
      "delete_file_permanent",
      [
        param("file_id", "The ID of the file to permanently delete", true),
      ],
      {
        file_id: "${!this.file_id}",
      },
    ),
    gdTool(
      "gd_export_file",
      "google_drive_export_file",
      "Export Google Workspace files (Docs, Sheets, Slides) to different formats like PDF, Word, Excel, etc.",
      "export_file",
      [
        param("file_id", "The ID of the Google Workspace file to export", true),
        param("mime_type", "Target format MIME type (e.g. application/pdf, text/csv, application/vnd.openxmlformats-officedocument.wordprocessingml.document)", true),
      ],
      {
        file_id: "${!this.file_id}",
        mime_type: "${!this.mime_type}",
      },
    ),
    gdTool(
      "gd_upload_file",
      "google_drive_upload_file",
      "Upload a file to Google Drive from a URL or content.",
      "upload_file",
      [
        param("file_name", "Name for the uploaded file", true),
        param("file_url", "URL to download the file from"),
        param("content", "File content (text or base64 encoded)"),
        param("folder_id", "Destination folder ID. Defaults to the configured folder if not provided."),
        param("mime_type", "MIME type of the file"),
        param("description", "File description"),
      ],
      {
        file_name: "${!this.file_name}",
        file_url: '${!this.file_url.or("")}',
        content: '${!this.content.or("")}',
        folder_id: '${!this.folder_id.or("")}',
        mime_type: '${!this.mime_type.or("")}',
        description: '${!this.description.or("")}',
      },
    ),
    gdTool(
      "gd_create_folder",
      "google_drive_create_folder",
      "Create a new, empty folder in Google Drive.",
      "create_folder",
      [
        param("file_name", "Name for the new folder", true),
        param("folder_id", "Parent folder ID. Defaults to the configured folder if not provided."),
      ],
      {
        file_name: "${!this.file_name}",
        folder_id: '${!this.folder_id.or("")}',
      },
    ),
    gdTool(
      "gd_move_file",
      "google_drive_move_file",
      "Move a file from one folder to another in Google Drive.",
      "move_file",
      [
        param("file_id", "The ID of the file to move", true),
        param("destination_folder_id", "The ID of the destination folder", true),
      ],
      {
        file_id: "${!this.file_id}",
        destination_folder_id: "${!this.destination_folder_id}",
      },
    ),
    gdTool(
      "gd_create_file_from_text",
      "google_drive_create_file_from_text",
      "Create a new file from plain text content in Google Drive.",
      "create_file_from_text",
      [
        param("file_name", "Name for the new file", true),
        param("content", "Text content for the file", true),
        param("folder_id", "Parent folder ID. Defaults to the configured folder if not provided."),
        param("mime_type", "MIME type (defaults to text/plain)"),
        param("description", "File description"),
      ],
      {
        file_name: "${!this.file_name}",
        content: "${!this.content}",
        folder_id: '${!this.folder_id.or("")}',
        mime_type: '${!this.mime_type.or("text/plain")}',
        description: '${!this.description.or("")}',
      },
    ),
    gdTool(
      "gd_remove_file_permission",
      "google_drive_remove_file_permission",
      "Remove specific user access to a file in Google Drive. Requires either email address or permission ID.",
      "remove_file_permission",
      [
        param("file_id", "The ID of the file", true),
        param("email", "Email address of the user to remove"),
        param("permission_id", "Permission ID to remove (alternative to email)"),
      ],
      {
        file_id: "${!this.file_id}",
        email: '${!this.email.or("")}',
        permission_id: '${!this.permission_id.or("")}',
      },
    ),
    gdTool(
      "gd_replace_file",
      "google_drive_replace_file",
      "Upload new content to replace an existing file in Google Drive.",
      "replace_file",
      [
        param("file_id", "The ID of the file to replace", true),
        param("file_url", "URL to download the replacement file from"),
        param("content", "Replacement content (text or base64 encoded)"),
      ],
      {
        file_id: "${!this.file_id}",
        file_url: '${!this.file_url.or("")}',
        content: '${!this.content.or("")}',
      },
    ),
    gdTool(
      "gd_create_shared_drive",
      "google_drive_create_shared_drive",
      "Create a new shared drive (also known as Team Drive) in Google Drive.",
      "create_shared_drive",
      [
        param("shared_drive_name", "Name for the new shared drive", true),
      ],
      {
        shared_drive_name: "${!this.shared_drive_name}",
      },
    ),
    gdTool(
      "gd_add_file_sharing",
      "google_drive_add_file_sharing",
      "Add a sharing permission to a file in Google Drive. Provides a sharing URL.",
      "add_file_sharing",
      [
        param("file_id", "The ID of the file to share", true),
        param("email", "Email address to share with (for user/group type)"),
        param("role", "Permission role: reader, writer, commenter, owner"),
        param("permission_type", "Permission type: user, group, domain, anyone"),
        param("send_notification", "Send notification email (true/false)"),
      ],
      {
        file_id: "${!this.file_id}",
        email: '${!this.email.or("")}',
        role: '${!this.role.or("reader")}',
        permission_type: '${!this.permission_type.or("user")}',
        send_notification: '${!this.send_notification.or("true")}',
      },
    ),
    gdTool(
      "gd_create_shortcut",
      "google_drive_create_shortcut",
      "Create a shortcut to a file in Google Drive.",
      "create_shortcut",
      [
        param("target_file_id", "The ID of the file to create a shortcut to", true),
        param("file_name", "Name for the shortcut"),
        param("folder_id", "Folder to create the shortcut in. Defaults to the configured folder if not provided."),
      ],
      {
        target_file_id: "${!this.target_file_id}",
        file_name: '${!this.file_name.or("")}',
        folder_id: '${!this.folder_id.or("")}',
      },
    ),
    gdTool(
      "gd_update_metadata",
      "google_drive_update_metadata",
      "Update file or folder metadata including name, description, starred status, folder color, and custom properties.",
      "update_metadata",
      [
        param("file_id", "The ID of the file or folder to update", true),
        param("file_name", "New name for the file or folder"),
        param("description", "New description"),
        param("starred", "Whether the file is starred (true/false)"),
        param("folder_color", "Folder color as hex (e.g. #FF0000)"),
        param("custom_properties", 'JSON object of custom properties (e.g. {"key":"value"})'),
      ],
      {
        file_id: "${!this.file_id}",
        file_name: '${!this.file_name.or("")}',
        description: '${!this.description.or("")}',
        starred: '${!this.starred.or("false")}',
        folder_color: '${!this.folder_color.or("")}',
        custom_properties: '${!this.custom_properties.or("")}',
      },
    ),
    gdTool(
      "gd_rename",
      "google_drive_rename",
      "Update the name of a file or folder in Google Drive.",
      "rename",
      [
        param("file_id", "The ID of the file or folder to rename", true),
        param("file_name", "The new name", true),
      ],
      {
        file_id: "${!this.file_id}",
        file_name: "${!this.file_name}",
      },
    ),
    gdTool(
      "gd_delete_file",
      "google_drive_delete_file",
      "Move a file to the trash in Google Drive.",
      "delete_file",
      [
        param("file_id", "The ID of the file to trash", true),
      ],
      {
        file_id: "${!this.file_id}",
      },
    ),
    gdTool(
      "gd_list_files",
      "google_drive_list_files",
      "Retrieve a list of files from Google Drive based on query parameters.",
      "list_files",
      [
        param("query", "Search query using Drive query syntax (e.g. name contains 'report')"),
        param("folder_id", "List files in a specific folder. Defaults to the configured folder if not provided."),
        param("max_results", "Maximum number of files to return", false, "number"),
      ],
      {
        query: '${!this.query.or("")}',
        folder_id: '${!this.folder_id.or("")}',
        max_results: '${!this.max_results.or("100")}',
      },
    ),
    gdTool(
      "gd_get_file",
      "google_drive_get_file",
      "Get a file or folder by its ID from Google Drive.",
      "get_file",
      [
        param("file_id", "The ID of the file or folder to retrieve", true),
      ],
      {
        file_id: "${!this.file_id}",
      },
    ),
    gdTool(
      "gd_get_permissions",
      "google_drive_get_permissions",
      "List all users who have access to a file in Google Drive.",
      "get_permissions",
      [
        param("file_id", "The ID of the file to check permissions for", true),
      ],
      {
        file_id: "${!this.file_id}",
      },
    ),
    gdTool(
      "gd_find_file",
      "google_drive_find_file",
      "Search for a specific file by name in Google Drive.",
      "find_file",
      [
        param("file_name", "The file name to search for", true),
        param("folder_id", "Search within a specific folder. Defaults to the configured folder if not provided."),
        param("max_results", "Maximum number of results", false, "number"),
      ],
      {
        file_name: "${!this.file_name}",
        folder_id: '${!this.folder_id.or("")}',
        max_results: '${!this.max_results.or("100")}',
      },
    ),
    gdTool(
      "gd_find_folder",
      "google_drive_find_folder",
      "Search for a specific folder by name in Google Drive.",
      "find_folder",
      [
        param("file_name", "The folder name to search for", true),
        param("folder_id", "Search within a specific parent folder. Defaults to the configured folder if not provided."),
        param("max_results", "Maximum number of results", false, "number"),
      ],
      {
        file_name: "${!this.file_name}",
        folder_id: '${!this.folder_id.or("")}',
        max_results: '${!this.max_results.or("100")}',
      },
    ),
    gdTool(
      "gd_find_or_create_file",
      "google_drive_find_or_create_file",
      "Find an existing file by name, or create a new one if not found.",
      "find_or_create_file",
      [
        param("file_name", "The file name to search for or create", true),
        param("folder_id", "Folder to search in and create within. Defaults to the configured folder if not provided."),
        param("content", "Text content if creating a new file"),
        param("mime_type", "MIME type if creating a new file"),
      ],
      {
        file_name: "${!this.file_name}",
        folder_id: '${!this.folder_id.or("")}',
        content: '${!this.content.or("")}',
        mime_type: '${!this.mime_type.or("")}',
      },
    ),
    gdTool(
      "gd_find_or_create_folder",
      "google_drive_find_or_create_folder",
      "Find an existing folder by name, or create a new one if not found.",
      "find_or_create_folder",
      [
        param("file_name", "The folder name to search for or create", true),
        param("folder_id", "Parent folder to search in and create within. Defaults to the configured folder if not provided."),
      ],
      {
        file_name: "${!this.file_name}",
        folder_id: '${!this.folder_id.or("")}',
      },
    ),
  ],
};
