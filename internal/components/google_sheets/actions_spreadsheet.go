package google_sheets

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func (p *Processor) getSheetIDByName(ctx context.Context, spreadsheetID, sheetName string) (int64, error) {
	svc, err := p.initSheetsService()
	if err != nil {
		return 0, err
	}

	ss, err := svc.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	for _, sheet := range ss.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId, nil
		}
	}

	return 0, fmt.Errorf("sheet %q not found in spreadsheet", sheetName)
}

func (p *Processor) createSpreadsheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	title := f.title
	if title == "" {
		return nil, fmt.Errorf("title is required for create_spreadsheet action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	ss := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}

	if f.sheetName != "" && f.sheetName != "Sheet1" {
		ss.Sheets = []*sheets.Sheet{
			{
				Properties: &sheets.SheetProperties{
					Title: f.sheetName,
				},
			},
		}
	}

	created, err := svc.Spreadsheets.Create(ss).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create spreadsheet: %w", err)
	}

	sheetsList := make([]map[string]any, 0, len(created.Sheets))
	for _, sheet := range created.Sheets {
		sheetsList = append(sheetsList, map[string]any{
			"sheet_id": sheet.Properties.SheetId,
			"title":    sheet.Properties.Title,
			"index":    sheet.Properties.Index,
		})
	}

	return map[string]any{
		"spreadsheet_id": created.SpreadsheetId,
		"title":          created.Properties.Title,
		"url":            created.SpreadsheetUrl,
		"sheets":         sheetsList,
	}, nil
}

func (p *Processor) getSpreadsheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for get_spreadsheet action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	ss, err := svc.Spreadsheets.Get(f.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	sheetsList := make([]map[string]any, 0, len(ss.Sheets))
	for _, sheet := range ss.Sheets {
		sheetsList = append(sheetsList, map[string]any{
			"sheet_id": sheet.Properties.SheetId,
			"title":    sheet.Properties.Title,
			"index":    sheet.Properties.Index,
			"hidden":   sheet.Properties.Hidden,
		})
	}

	return map[string]any{
		"spreadsheet_id": ss.SpreadsheetId,
		"title":          ss.Properties.Title,
		"locale":         ss.Properties.Locale,
		"time_zone":      ss.Properties.TimeZone,
		"sheets":         sheetsList,
	}, nil
}

func (p *Processor) createWorksheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for create_worksheet action")
	}

	title := f.title
	if title == "" {
		title = f.sheetName
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	resp, err := svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{Title: title},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create worksheet: %w", err)
	}

	props := resp.Replies[0].AddSheet.Properties
	return map[string]any{
		"sheet_id": props.SheetId,
		"title":    props.Title,
		"index":    props.Index,
	}, nil
}

func (p *Processor) findWorksheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for find_worksheet action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	ss, err := svc.Spreadsheets.Get(f.spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	for _, sheet := range ss.Sheets {
		if sheet.Properties.Title == f.sheetName {
			return map[string]any{
				"found":    true,
				"sheet_id": sheet.Properties.SheetId,
				"title":    sheet.Properties.Title,
				"index":    sheet.Properties.Index,
				"hidden":   sheet.Properties.Hidden,
			}, nil
		}
	}

	return map[string]any{"found": false}, nil
}

func (p *Processor) findOrCreateWorksheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	result, err := p.findWorksheet(ctx, f)
	if err != nil {
		return nil, err
	}

	if result["found"].(bool) {
		result["created"] = false
		return result, nil
	}

	createResult, err := p.createWorksheet(ctx, f)
	if err != nil {
		return nil, err
	}

	createResult["found"] = true
	createResult["created"] = true
	return createResult, nil
}

func (p *Processor) deleteWorksheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for delete_worksheet action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteSheet: &sheets.DeleteSheetRequest{SheetId: sheetID},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete worksheet: %w", err)
	}

	return map[string]any{
		"deleted":    true,
		"sheet_name": f.sheetName,
	}, nil
}

func (p *Processor) renameWorksheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for rename_worksheet action")
	}
	if f.newName == "" {
		return nil, fmt.Errorf("new_name is required for rename_worksheet action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId: sheetID,
						Title:   f.newName,
					},
					Fields: "title",
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to rename worksheet: %w", err)
	}

	return map[string]any{
		"old_name": f.sheetName,
		"new_name": f.newName,
		"sheet_id": sheetID,
	}, nil
}

func (p *Processor) copyWorksheet(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for copy_worksheet action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	destID := f.spreadsheetID
	if f.destSpreadsheetID != "" {
		destID = f.destSpreadsheetID
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	resp, err := svc.Spreadsheets.Sheets.CopyTo(f.spreadsheetID, sheetID, &sheets.CopySheetToAnotherSpreadsheetRequest{
		DestinationSpreadsheetId: destID,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to copy worksheet: %w", err)
	}

	return map[string]any{
		"sheet_id":                   resp.SheetId,
		"title":                      resp.Title,
		"destination_spreadsheet_id": destID,
	}, nil
}

func (p *Processor) changeSheetProperties(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for change_sheet_properties action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	props := &sheets.SheetProperties{
		SheetId: sheetID,
	}

	var fields []string

	if f.frozenRows > 0 {
		props.GridProperties = &sheets.GridProperties{FrozenRowCount: int64(f.frozenRows)}
		fields = append(fields, "gridProperties.frozenRowCount")
	}
	if f.frozenColumns > 0 {
		if props.GridProperties == nil {
			props.GridProperties = &sheets.GridProperties{}
		}
		props.GridProperties.FrozenColumnCount = int64(f.frozenColumns)
		fields = append(fields, "gridProperties.frozenColumnCount")
	}
	if f.hidden {
		props.Hidden = true
		fields = append(fields, "hidden")
	}
	if f.sheetPosition > 0 {
		props.Index = int64(f.sheetPosition)
		fields = append(fields, "index")
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("at least one property (frozen_rows, frozen_columns, hidden, sheet_position) must be set")
	}

	fieldMask := strings.Join(fields, ",")

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: props,
					Fields:     fieldMask,
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update sheet properties: %w", err)
	}

	return map[string]any{
		"updated":  true,
		"sheet_id": sheetID,
	}, nil
}
