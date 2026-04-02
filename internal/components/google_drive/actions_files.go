package google_drive

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/drive/v3"
)

func (p *Processor) copyFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for copy_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	copyFile := &drive.File{}
	if f.fileName != "" {
		copyFile.Name = f.fileName
	}
	if f.folderID != "" {
		copyFile.Parents = []string{f.folderID}
	}

	result, err := svc.Files.Copy(f.fileID, copyFile).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) deleteFilePermanent(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for delete_file_permanent")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	if err := svc.Files.Delete(f.fileID).
		Context(ctx).
		SupportsAllDrives(true).
		Do(); err != nil {
		return nil, err
	}

	return map[string]any{"deleted": true, "file_id": f.fileID, "permanent": true}, nil
}

func (p *Processor) deleteFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for delete_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	update := &drive.File{Trashed: true}
	_, err = svc.Files.Update(f.fileID, update).
		Context(ctx).
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"deleted": true, "file_id": f.fileID, "trashed": true}, nil
}

func (p *Processor) getFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for get_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	result, err := svc.Files.Get(f.fileID).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,webContentLink,createdTime,modifiedTime,size,description,starred,shared,owners,permissions,iconLink").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) listFiles(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	call := svc.Files.List().
		Context(ctx).
		PageSize(int64(f.maxResults)).
		Fields("files(id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size,description,starred,shared)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true)

	q := f.query
	if f.folderID != "" && q == "" {
		q = fmt.Sprintf("'%s' in parents and trashed = false", f.folderID)
	} else if f.folderID != "" {
		q = fmt.Sprintf("'%s' in parents and trashed = false and (%s)", f.folderID, q)
	}
	if q != "" {
		call = call.Q(q)
	}

	result, err := call.Do()
	if err != nil {
		return nil, err
	}

	files := make([]map[string]any, 0, len(result.Files))
	for _, file := range result.Files {
		files = append(files, fileToMap(file))
	}

	return map[string]any{"files": files, "count": len(files)}, nil
}

func (p *Processor) findFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for find_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("name = '%s' and mimeType != 'application/vnd.google-apps.folder' and trashed = false", escapeQuery(f.fileName))
	if f.folderID != "" {
		q = fmt.Sprintf("'%s' in parents and %s", f.folderID, q)
	}

	result, err := svc.Files.List().
		Context(ctx).
		Q(q).
		PageSize(int64(f.maxResults)).
		Fields("files(id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size,description,starred,shared)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	files := make([]map[string]any, 0, len(result.Files))
	for _, file := range result.Files {
		files = append(files, fileToMap(file))
	}

	return map[string]any{"files": files, "count": len(files)}, nil
}

func (p *Processor) findFolder(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for find_folder")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", escapeQuery(f.fileName))
	if f.folderID != "" {
		q = fmt.Sprintf("'%s' in parents and %s", f.folderID, q)
	}

	result, err := svc.Files.List().
		Context(ctx).
		Q(q).
		PageSize(int64(f.maxResults)).
		Fields("files(id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,description)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	folders := make([]map[string]any, 0, len(result.Files))
	for _, file := range result.Files {
		folders = append(folders, fileToMap(file))
	}

	return map[string]any{"folders": folders, "count": len(folders)}, nil
}

func (p *Processor) findOrCreateFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for find_or_create_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("name = '%s' and mimeType != 'application/vnd.google-apps.folder' and trashed = false", escapeQuery(f.fileName))
	if f.folderID != "" {
		q = fmt.Sprintf("'%s' in parents and %s", f.folderID, q)
	}

	existing, err := svc.Files.List().
		Context(ctx).
		Q(q).
		PageSize(1).
		Fields("files(id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	if len(existing.Files) > 0 {
		return map[string]any{"file": fileToMap(existing.Files[0]), "created": false}, nil
	}

	newFile := &drive.File{Name: f.fileName}
	if f.folderID != "" {
		newFile.Parents = []string{f.folderID}
	}
	if f.mimeType != "" {
		newFile.MimeType = f.mimeType
	}

	var created *drive.File
	if f.content != "" {
		created, err = svc.Files.Create(newFile).
			Context(ctx).
			Media(strings.NewReader(f.content)).
			Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size").
			SupportsAllDrives(true).
			Do()
	} else {
		created, err = svc.Files.Create(newFile).
			Context(ctx).
			Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size").
			SupportsAllDrives(true).
			Do()
	}
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(created), "created": true}, nil
}

func (p *Processor) findOrCreateFolder(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for find_or_create_folder")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and trashed = false", escapeQuery(f.fileName))
	if f.folderID != "" {
		q = fmt.Sprintf("'%s' in parents and %s", f.folderID, q)
	}

	existing, err := svc.Files.List().
		Context(ctx).
		Q(q).
		PageSize(1).
		Fields("files(id,name,mimeType,parents,webViewLink,createdTime,modifiedTime)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	if len(existing.Files) > 0 {
		return map[string]any{"folder": fileToMap(existing.Files[0]), "created": false}, nil
	}

	newFolder := &drive.File{
		Name:     f.fileName,
		MimeType: "application/vnd.google-apps.folder",
	}
	if f.folderID != "" {
		newFolder.Parents = []string{f.folderID}
	}

	created, err := svc.Files.Create(newFolder).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"folder": fileToMap(created), "created": true}, nil
}

func (p *Processor) rename(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for rename")
	}
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for rename")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	update := &drive.File{Name: f.fileName}
	result, err := svc.Files.Update(f.fileID, update).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,modifiedTime").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) updateMetadata(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for update_metadata")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	update := &drive.File{}
	if f.fileName != "" {
		update.Name = f.fileName
	}
	if f.description != "" {
		update.Description = f.description
	}
	update.Starred = f.starred
	if f.folderColor != "" {
		update.FolderColorRgb = f.folderColor
	}
	if f.customProperties != "" {
		props, err := parseJSONObject(f.customProperties)
		if err != nil {
			return nil, err
		}
		update.Properties = props
	}

	result, err := svc.Files.Update(f.fileID, update).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,modifiedTime,description,starred,properties").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) createShortcut(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.targetFileID == "" {
		return nil, fmt.Errorf("target_file_id is required for create_shortcut")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	shortcut := &drive.File{
		MimeType: "application/vnd.google-apps.shortcut",
		ShortcutDetails: &drive.FileShortcutDetails{
			TargetId: f.targetFileID,
		},
	}
	if f.fileName != "" {
		shortcut.Name = f.fileName
	}
	if f.folderID != "" {
		shortcut.Parents = []string{f.folderID}
	}

	result, err := svc.Files.Create(shortcut).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,shortcutDetails").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func escapeQuery(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
