package google_drive

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"google.golang.org/api/drive/v3"
)

func (p *Processor) createFolder(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for create_folder")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	folder := &drive.File{
		Name:     f.fileName,
		MimeType: "application/vnd.google-apps.folder",
	}
	if f.folderID != "" {
		folder.Parents = []string{f.folderID}
	}

	result, err := svc.Files.Create(folder).
		Context(ctx).
		Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"folder": fileToMap(result)}, nil
}

func (p *Processor) moveFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for move_file")
	}
	if f.destinationFolderID == "" {
		return nil, fmt.Errorf("destination_folder_id is required for move_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	existing, err := svc.Files.Get(f.fileID).
		Context(ctx).
		Fields("parents").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	var previousParents string
	if len(existing.Parents) > 0 {
		previousParents = existing.Parents[0]
		for i := 1; i < len(existing.Parents); i++ {
			previousParents += "," + existing.Parents[i]
		}
	}

	result, err := svc.Files.Update(f.fileID, &drive.File{}).
		Context(ctx).
		AddParents(f.destinationFolderID).
		RemoveParents(previousParents).
		Fields("id,name,mimeType,parents,webViewLink,modifiedTime").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) createSharedDrive(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.sharedDriveName == "" {
		return nil, fmt.Errorf("shared_drive_name is required for create_shared_drive")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	sharedDrive := &drive.Drive{
		Name: f.sharedDriveName,
	}

	reqID := make([]byte, 16)
	_, _ = rand.Read(reqID)
	result, err := svc.Drives.Create(hex.EncodeToString(reqID), sharedDrive).
		Context(ctx).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"shared_drive": map[string]any{
			"id":   result.Id,
			"name": result.Name,
		},
	}, nil
}
