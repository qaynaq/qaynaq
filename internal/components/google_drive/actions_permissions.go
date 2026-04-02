package google_drive

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"
)

func (p *Processor) addFileSharing(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for add_file_sharing")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	perm := &drive.Permission{
		Role: f.role,
		Type: f.permissionType,
	}
	if f.email != "" {
		perm.EmailAddress = f.email
	}

	result, err := svc.Permissions.Create(f.fileID, perm).
		Context(ctx).
		SendNotificationEmail(f.sendNotification).
		SupportsAllDrives(true).
		Fields("id,type,role,emailAddress,displayName").
		Do()
	if err != nil {
		return nil, err
	}

	file, err := svc.Files.Get(f.fileID).
		Context(ctx).
		Fields("webViewLink").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"permission": map[string]any{
			"id":           result.Id,
			"type":         result.Type,
			"role":         result.Role,
			"email":        result.EmailAddress,
			"display_name": result.DisplayName,
		},
		"sharing_url": file.WebViewLink,
	}, nil
}

func (p *Processor) removeFilePermission(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for remove_file_permission")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	permID := f.permissionID
	if permID == "" && f.email != "" {
		perms, err := svc.Permissions.List(f.fileID).
			Context(ctx).
			Fields("permissions(id,emailAddress)").
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return nil, err
		}
		for _, p := range perms.Permissions {
			if p.EmailAddress == f.email {
				permID = p.Id
				break
			}
		}
		if permID == "" {
			return nil, fmt.Errorf("[400] no permission found for email: %s", f.email)
		}
	}

	if permID == "" {
		return nil, fmt.Errorf("either permission_id or email is required for remove_file_permission")
	}

	if err := svc.Permissions.Delete(f.fileID, permID).
		Context(ctx).
		SupportsAllDrives(true).
		Do(); err != nil {
		return nil, err
	}

	return map[string]any{
		"removed":       true,
		"file_id":       f.fileID,
		"permission_id": permID,
	}, nil
}

func (p *Processor) getPermissions(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.fileID == "" {
		return nil, fmt.Errorf("file_id is required for get_permissions")
	}

	svc, err := p.initDriveService()
	if err != nil {
		return nil, err
	}

	result, err := svc.Permissions.List(f.fileID).
		Context(ctx).
		Fields("permissions(id,type,role,emailAddress,displayName,domain)").
		SupportsAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	permissions := make([]map[string]any, 0, len(result.Permissions))
	for _, p := range result.Permissions {
		perm := map[string]any{
			"id":   p.Id,
			"type": p.Type,
			"role": p.Role,
		}
		if p.EmailAddress != "" {
			perm["email"] = p.EmailAddress
		}
		if p.DisplayName != "" {
			perm["display_name"] = p.DisplayName
		}
		if p.Domain != "" {
			perm["domain"] = p.Domain
		}
		permissions = append(permissions, perm)
	}

	return map[string]any{"permissions": permissions, "count": len(permissions)}, nil
}
