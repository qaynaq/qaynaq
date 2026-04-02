package google_drive

import "github.com/warpstreamlabs/bento/public/service"

const (
	gdfServiceAccountJSON  = "service_account_json"
	gdfDelegateTo          = "delegate_to"
	gdfAction              = "action"
	gdfFileID              = "file_id"
	gdfFileName            = "file_name"
	gdfFolderID            = "folder_id"
	gdfDestinationFolderID = "destination_folder_id"
	gdfMimeType            = "mime_type"
	gdfContent             = "content"
	gdfFileURL             = "file_url"
	gdfDescription         = "description"
	gdfEmail               = "email"
	gdfRole                = "role"
	gdfPermissionType      = "permission_type"
	gdfPermissionID        = "permission_id"
	gdfQuery               = "query"
	gdfMaxResults          = "max_results"
	gdfStarred             = "starred"
	gdfFolderColor         = "folder_color"
	gdfCustomProperties    = "custom_properties"
	gdfSharedDriveName     = "shared_drive_name"
	gdfTargetFileID        = "target_file_id"
	gdfSendNotification    = "send_notification"
	gdfOAuthConnection     = "oauth_connection"
)

const (
	actionCopyFile             = "copy_file"
	actionDeleteFilePermanent  = "delete_file_permanent"
	actionExportFile           = "export_file"
	actionUploadFile           = "upload_file"
	actionCreateFolder         = "create_folder"
	actionMoveFile             = "move_file"
	actionCreateFileFromText   = "create_file_from_text"
	actionRemoveFilePermission = "remove_file_permission"
	actionReplaceFile          = "replace_file"
	actionCreateSharedDrive    = "create_shared_drive"
	actionAddFileSharing       = "add_file_sharing"
	actionCreateShortcut       = "create_shortcut"
	actionUpdateMetadata       = "update_metadata"
	actionRename               = "rename"
	actionDeleteFile           = "delete_file"
	actionListFiles            = "list_files"
	actionGetFile              = "get_file"
	actionGetPermissions       = "get_permissions"
	actionFindFile             = "find_file"
	actionFindFolder           = "find_folder"
	actionFindOrCreateFile     = "find_or_create_file"
	actionFindOrCreateFolder   = "find_or_create_folder"
)

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Integration").
		Summary("Performs Google Drive operations - manage files, folders, permissions, and shared drives.").
		Description(`
This processor interacts with the Google Drive API using service account authentication.
It supports creating, reading, updating, and deleting files and folders, managing permissions,
exporting Google Workspace files, and working with shared drives.

Store your Google service account JSON as a secret in Settings > Secrets, then reference
it in the Service Account JSON field. For accessing files owned by other users in a
Google Workspace domain, enable Domain-Wide Delegation and set the Delegate To field.

Most fields support interpolation functions, allowing dynamic values from message content
using the ` + "`${!this.field_name}`" + ` syntax.`).
		Field(service.NewStringField(gdfServiceAccountJSON).
			Description("Google service account credentials JSON. Store as a secret and reference via ${SECRET_NAME}.").
			Secret().
			Optional().
			Default("")).
		Field(service.NewStringField(gdfOAuthConnection).
			Description("OAuth connection for user authentication. Set up in Settings > Connections, then reference via ${CONN_NAME}. Alternative to Service Account JSON.").
			Secret().
			Optional().
			Default("")).
		Field(service.NewStringField(gdfDelegateTo).
			Description("Email address to impersonate via Domain-Wide Delegation. Required for accessing files owned by other users in Google Workspace.").
			Default("").
			Optional()).
		Field(service.NewStringEnumField(gdfAction,
			actionCopyFile,
			actionDeleteFilePermanent,
			actionExportFile,
			actionUploadFile,
			actionCreateFolder,
			actionMoveFile,
			actionCreateFileFromText,
			actionRemoveFilePermission,
			actionReplaceFile,
			actionCreateSharedDrive,
			actionAddFileSharing,
			actionCreateShortcut,
			actionUpdateMetadata,
			actionRename,
			actionDeleteFile,
			actionListFiles,
			actionGetFile,
			actionGetPermissions,
			actionFindFile,
			actionFindFolder,
			actionFindOrCreateFile,
			actionFindOrCreateFolder,
		).Description("The Google Drive operation to perform.")).
		Field(service.NewInterpolatedStringField(gdfFileID).
			Description("The file or folder ID. Required for most file-specific operations.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfFileName).
			Description("File or folder name. Used by: create_folder, create_file_from_text, rename, find_file, find_folder, copy_file, upload_file, create_shortcut, find_or_create_file, find_or_create_folder.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfFolderID).
			Description("Parent folder ID. Defaults to root. Used by: create_folder, create_file_from_text, upload_file, find_file, find_folder, copy_file, list_files, create_shortcut, find_or_create_file, find_or_create_folder.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfDestinationFolderID).
			Description("Target folder ID for move operations. Required for: move_file.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfMimeType).
			Description("MIME type. Used by: export_file (target format, e.g. 'application/pdf'), upload_file, create_file_from_text, find_or_create_file.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfContent).
			Description("Text content for file creation. Used by: create_file_from_text, find_or_create_file.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfFileURL).
			Description("URL to download file content from. Used by: upload_file, replace_file.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfDescription).
			Description("File description. Used by: update_metadata, create_file_from_text, upload_file.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfEmail).
			Description("Email address for permission operations. Used by: add_file_sharing, remove_file_permission.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfRole).
			Description("Permission role: reader, writer, commenter, owner, organizer, fileOrganizer. Used by: add_file_sharing.").
			Default("reader").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfPermissionType).
			Description("Permission type: user, group, domain, anyone. Used by: add_file_sharing.").
			Default("user").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfPermissionID).
			Description("Permission ID for removal. Used by: remove_file_permission (alternative to email).").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfQuery).
			Description("Search query using Drive query syntax. Used by: list_files.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfMaxResults).
			Description("Maximum number of results to return. Used by: list_files, find_file, find_folder.").
			Default("100")).
		Field(service.NewInterpolatedStringField(gdfStarred).
			Description("Whether the file is starred (true/false). Used by: update_metadata.").
			Default("false").
			Advanced()).
		Field(service.NewInterpolatedStringField(gdfFolderColor).
			Description("Folder color as hex (e.g. '#FF0000'). Used by: update_metadata.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gdfCustomProperties).
			Description("JSON object of custom properties (e.g. '{\"key\":\"value\"}'). Used by: update_metadata.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gdfSharedDriveName).
			Description("Name for the new shared drive. Required for: create_shared_drive.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfTargetFileID).
			Description("Target file ID for creating a shortcut. Required for: create_shortcut.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gdfSendNotification).
			Description("Send notification email when sharing (true/false). Used by: add_file_sharing.").
			Default("true").
			Advanced()).
		Version("1.0.0")
}
