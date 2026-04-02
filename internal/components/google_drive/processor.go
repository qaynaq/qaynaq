package google_drive

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/warpstreamlabs/bento/public/service"
	"google.golang.org/api/drive/v3"
)

func init() {
	err := service.RegisterProcessor(
		"google_drive", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type Processor struct {
	serviceAccountJSON  string
	oauthConnection     string
	delegateTo          string
	action              string
	fileID              *service.InterpolatedString
	fileName            *service.InterpolatedString
	folderID            *service.InterpolatedString
	destinationFolderID *service.InterpolatedString
	mimeType            *service.InterpolatedString
	content             *service.InterpolatedString
	fileURL             *service.InterpolatedString
	description         *service.InterpolatedString
	email               *service.InterpolatedString
	role                *service.InterpolatedString
	permissionType      *service.InterpolatedString
	permissionID        *service.InterpolatedString
	query               *service.InterpolatedString
	maxResults          *service.InterpolatedString
	starred             *service.InterpolatedString
	folderColor         *service.InterpolatedString
	customProperties    *service.InterpolatedString
	sharedDriveName     *service.InterpolatedString
	targetFileID        *service.InterpolatedString
	sendNotification    *service.InterpolatedString

	driveService   *drive.Service
	serviceOnce    sync.Once
	serviceInitErr error
	logger         *service.Logger
}

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	serviceAccountJSON, err := conf.FieldString(gdfServiceAccountJSON)
	if err != nil {
		return nil, err
	}

	oauthConnection, err := conf.FieldString(gdfOAuthConnection)
	if err != nil {
		return nil, err
	}

	if serviceAccountJSON == "" && oauthConnection == "" {
		return nil, fmt.Errorf("either service_account_json or oauth_connection must be provided")
	}

	action, err := conf.FieldString(gdfAction)
	if err != nil {
		return nil, err
	}

	p := &Processor{
		serviceAccountJSON: serviceAccountJSON,
		oauthConnection:    oauthConnection,
		action:             action,
		logger:             mgr.Logger(),
	}

	if conf.Contains(gdfDelegateTo) {
		if p.delegateTo, err = conf.FieldString(gdfDelegateTo); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfFileID) {
		if p.fileID, err = conf.FieldInterpolatedString(gdfFileID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfFileName) {
		if p.fileName, err = conf.FieldInterpolatedString(gdfFileName); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfFolderID) {
		if p.folderID, err = conf.FieldInterpolatedString(gdfFolderID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfDestinationFolderID) {
		if p.destinationFolderID, err = conf.FieldInterpolatedString(gdfDestinationFolderID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfMimeType) {
		if p.mimeType, err = conf.FieldInterpolatedString(gdfMimeType); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfContent) {
		if p.content, err = conf.FieldInterpolatedString(gdfContent); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfFileURL) {
		if p.fileURL, err = conf.FieldInterpolatedString(gdfFileURL); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfDescription) {
		if p.description, err = conf.FieldInterpolatedString(gdfDescription); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfEmail) {
		if p.email, err = conf.FieldInterpolatedString(gdfEmail); err != nil {
			return nil, err
		}
	}

	if p.role, err = conf.FieldInterpolatedString(gdfRole); err != nil {
		return nil, err
	}

	if p.permissionType, err = conf.FieldInterpolatedString(gdfPermissionType); err != nil {
		return nil, err
	}

	if conf.Contains(gdfPermissionID) {
		if p.permissionID, err = conf.FieldInterpolatedString(gdfPermissionID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfQuery) {
		if p.query, err = conf.FieldInterpolatedString(gdfQuery); err != nil {
			return nil, err
		}
	}

	if p.maxResults, err = conf.FieldInterpolatedString(gdfMaxResults); err != nil {
		return nil, err
	}

	if p.starred, err = conf.FieldInterpolatedString(gdfStarred); err != nil {
		return nil, err
	}

	if conf.Contains(gdfFolderColor) {
		if p.folderColor, err = conf.FieldInterpolatedString(gdfFolderColor); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfCustomProperties) {
		if p.customProperties, err = conf.FieldInterpolatedString(gdfCustomProperties); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfSharedDriveName) {
		if p.sharedDriveName, err = conf.FieldInterpolatedString(gdfSharedDriveName); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gdfTargetFileID) {
		if p.targetFileID, err = conf.FieldInterpolatedString(gdfTargetFileID); err != nil {
			return nil, err
		}
	}

	if p.sendNotification, err = conf.FieldInterpolatedString(gdfSendNotification); err != nil {
		return nil, err
	}

	return p, nil
}

type resolvedFields struct {
	fileID              string
	fileName            string
	folderID            string
	destinationFolderID string
	mimeType            string
	content             string
	fileURL             string
	description         string
	email               string
	role                string
	permissionType      string
	permissionID        string
	query               string
	maxResults          int
	starred             bool
	folderColor         string
	customProperties    string
	sharedDriveName     string
	targetFileID        string
	sendNotification    bool
}

func (p *Processor) resolveFields(msg *service.Message) (*resolvedFields, error) {
	r := &resolvedFields{}
	var err error

	if p.fileID != nil {
		if r.fileID, err = p.fileID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate file_id: %w", err)
		}
	}

	if p.fileName != nil {
		if r.fileName, err = p.fileName.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate file_name: %w", err)
		}
	}

	if p.folderID != nil {
		if r.folderID, err = p.folderID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate folder_id: %w", err)
		}
	}

	if p.destinationFolderID != nil {
		if r.destinationFolderID, err = p.destinationFolderID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate destination_folder_id: %w", err)
		}
	}

	if p.mimeType != nil {
		if r.mimeType, err = p.mimeType.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate mime_type: %w", err)
		}
	}

	if p.content != nil {
		if r.content, err = p.content.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate content: %w", err)
		}
	}

	if p.fileURL != nil {
		if r.fileURL, err = p.fileURL.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate file_url: %w", err)
		}
	}

	if p.description != nil {
		if r.description, err = p.description.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate description: %w", err)
		}
	}

	if p.email != nil {
		if r.email, err = p.email.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate email: %w", err)
		}
	}

	if r.role, err = p.role.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate role: %w", err)
	}

	if r.permissionType, err = p.permissionType.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate permission_type: %w", err)
	}

	if p.permissionID != nil {
		if r.permissionID, err = p.permissionID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate permission_id: %w", err)
		}
	}

	if p.query != nil {
		if r.query, err = p.query.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate query: %w", err)
		}
	}

	r.maxResults = resolveInt(p.maxResults, msg)
	if r.maxResults == 0 {
		r.maxResults = 100
	}

	r.starred = resolveBool(p.starred, msg)
	r.sendNotification = resolveBool(p.sendNotification, msg)

	if p.folderColor != nil {
		if r.folderColor, err = p.folderColor.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate folder_color: %w", err)
		}
	}

	if p.customProperties != nil {
		if r.customProperties, err = p.customProperties.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate custom_properties: %w", err)
		}
	}

	if p.sharedDriveName != nil {
		if r.sharedDriveName, err = p.sharedDriveName.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate shared_drive_name: %w", err)
		}
	}

	if p.targetFileID != nil {
		if r.targetFileID, err = p.targetFileID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate target_file_id: %w", err)
		}
	}

	return r, nil
}

func (p *Processor) Process(ctx context.Context, msg *service.Message) (service.MessageBatch, error) {
	fields, err := p.resolveFields(msg)
	if err != nil {
		return nil, classifyError(err)
	}

	var result map[string]any

	switch p.action {
	case actionCopyFile:
		result, err = p.copyFile(ctx, fields)
	case actionDeleteFilePermanent:
		result, err = p.deleteFilePermanent(ctx, fields)
	case actionExportFile:
		result, err = p.exportFile(ctx, fields)
	case actionUploadFile:
		result, err = p.uploadFile(ctx, fields)
	case actionCreateFolder:
		result, err = p.createFolder(ctx, fields)
	case actionMoveFile:
		result, err = p.moveFile(ctx, fields)
	case actionCreateFileFromText:
		result, err = p.createFileFromText(ctx, fields)
	case actionRemoveFilePermission:
		result, err = p.removeFilePermission(ctx, fields)
	case actionReplaceFile:
		result, err = p.replaceFile(ctx, fields)
	case actionCreateSharedDrive:
		result, err = p.createSharedDrive(ctx, fields)
	case actionAddFileSharing:
		result, err = p.addFileSharing(ctx, fields)
	case actionCreateShortcut:
		result, err = p.createShortcut(ctx, fields)
	case actionUpdateMetadata:
		result, err = p.updateMetadata(ctx, fields)
	case actionRename:
		result, err = p.rename(ctx, fields)
	case actionDeleteFile:
		result, err = p.deleteFile(ctx, fields)
	case actionListFiles:
		result, err = p.listFiles(ctx, fields)
	case actionGetFile:
		result, err = p.getFile(ctx, fields)
	case actionGetPermissions:
		result, err = p.getPermissions(ctx, fields)
	case actionFindFile:
		result, err = p.findFile(ctx, fields)
	case actionFindFolder:
		result, err = p.findFolder(ctx, fields)
	case actionFindOrCreateFile:
		result, err = p.findOrCreateFile(ctx, fields)
	case actionFindOrCreateFolder:
		result, err = p.findOrCreateFolder(ctx, fields)
	default:
		err = fmt.Errorf("unsupported action: %s", p.action)
	}

	if err != nil {
		return nil, classifyError(err)
	}

	outMsg := msg.Copy()
	outMsg.SetStructured(result)
	return service.MessageBatch{outMsg}, nil
}

func (p *Processor) Close(ctx context.Context) error {
	return nil
}

func resolveInt(field *service.InterpolatedString, msg *service.Message) int {
	if field == nil {
		return 0
	}
	v, _ := field.TryString(msg)
	n, _ := strconv.Atoi(v)
	return n
}

func resolveBool(field *service.InterpolatedString, msg *service.Message) bool {
	if field == nil {
		return false
	}
	v, _ := field.TryString(msg)
	return strings.EqualFold(v, "true")
}

func fileToMap(f *drive.File) map[string]any {
	result := map[string]any{
		"id":        f.Id,
		"name":      f.Name,
		"mime_type":  f.MimeType,
		"web_view_link": f.WebViewLink,
	}
	if f.Parents != nil {
		result["parents"] = f.Parents
	}
	if f.Size != 0 {
		result["size"] = f.Size
	}
	if f.CreatedTime != "" {
		result["created_time"] = f.CreatedTime
	}
	if f.ModifiedTime != "" {
		result["modified_time"] = f.ModifiedTime
	}
	if f.Description != "" {
		result["description"] = f.Description
	}
	if f.Starred {
		result["starred"] = f.Starred
	}
	if f.Shared {
		result["shared"] = f.Shared
	}
	if f.WebContentLink != "" {
		result["web_content_link"] = f.WebContentLink
	}
	if f.IconLink != "" {
		result["icon_link"] = f.IconLink
	}
	if len(f.Owners) > 0 {
		owners := make([]map[string]any, 0, len(f.Owners))
		for _, o := range f.Owners {
			owners = append(owners, map[string]any{
				"email":        o.EmailAddress,
				"display_name": o.DisplayName,
			})
		}
		result["owners"] = owners
	}
	return result
}

func parseJSONObject(s string) (map[string]string, error) {
	if s == "" {
		return nil, nil
	}
	var raw map[string]any
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON object: %w", err)
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result, nil
}
