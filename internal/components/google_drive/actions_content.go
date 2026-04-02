package google_drive

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/api/drive/v3"
)

func (p *Processor) exportFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for export_file")
	}
	if f.mimeType == "" {
		return nil, fmt.Errorf("mime_type is required for export_file (e.g. 'application/pdf', 'text/csv')")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	resp, err := svc.Files.Export(f.fileID, f.mimeType).
		Context(ctx).
		Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read export response: %w", err)
	}

	result := map[string]any{
		"file_id":   f.fileID,
		"mime_type": f.mimeType,
		"size":      len(data),
	}

	if isTextMimeType(f.mimeType) {
		result["content"] = string(data)
	} else {
		result["content_base64"] = base64.StdEncoding.EncodeToString(data)
	}

	return result, nil
}

func (p *Processor) uploadFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for upload_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if f.fileURL != "" {
		resp, err := http.Get(f.fileURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download file from URL: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download file from URL: HTTP %d", resp.StatusCode)
		}
		reader = resp.Body
	} else if f.content != "" {
		decoded, err := base64.StdEncoding.DecodeString(f.content)
		if err != nil {
			reader = strings.NewReader(f.content)
		} else {
			reader = strings.NewReader(string(decoded))
		}
	} else {
		return nil, fmt.Errorf("either file_url or content is required for upload_file")
	}

	newFile := &drive.File{Name: f.fileName}
	if f.folderID != "" {
		newFile.Parents = []string{f.folderID}
	}
	if f.mimeType != "" {
		newFile.MimeType = f.mimeType
	}
	if f.description != "" {
		newFile.Description = f.description
	}

	result, err := svc.Files.Create(newFile).
		Context(ctx).
		Media(reader).
		Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) createFileFromText(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileName == "" {
		return nil, fmt.Errorf("file_name is required for create_file_from_text")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	newFile := &drive.File{Name: f.fileName}
	if f.folderID != "" {
		newFile.Parents = []string{f.folderID}
	}
	if f.mimeType != "" {
		newFile.MimeType = f.mimeType
	} else {
		newFile.MimeType = "text/plain"
	}
	if f.description != "" {
		newFile.Description = f.description
	}

	result, err := svc.Files.Create(newFile).
		Context(ctx).
		Media(strings.NewReader(f.content)).
		Fields("id,name,mimeType,parents,webViewLink,createdTime,modifiedTime,size").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func (p *Processor) replaceFile(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for replace_file")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if f.fileURL != "" {
		resp, err := http.Get(f.fileURL)
		if err != nil {
			return nil, fmt.Errorf("failed to download file from URL: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to download file from URL: HTTP %d", resp.StatusCode)
		}
		reader = resp.Body
	} else if f.content != "" {
		decoded, err := base64.StdEncoding.DecodeString(f.content)
		if err != nil {
			reader = strings.NewReader(f.content)
		} else {
			reader = strings.NewReader(string(decoded))
		}
	} else {
		return nil, fmt.Errorf("either file_url or content is required for replace_file")
	}

	update := &drive.File{}
	result, err := svc.Files.Update(f.fileID, update).
		Context(ctx).
		Media(reader).
		Fields("id,name,mimeType,parents,webViewLink,modifiedTime,size").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{"file": fileToMap(result)}, nil
}

func isTextMimeType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "text/") ||
		mimeType == "application/json" ||
		mimeType == "application/xml" ||
		mimeType == "application/csv" ||
		mimeType == "text/csv"
}
