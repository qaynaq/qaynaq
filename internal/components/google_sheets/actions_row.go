package google_sheets

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

func parseJSONArray(s string) ([]any, error) {
	var result []any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}
	return result, nil
}

func parseJSONArrayOfArrays(s string) ([][]any, error) {
	var result [][]any
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array of arrays: %w", err)
	}
	return result, nil
}

func buildRowMap(headers []any, values []any) map[string]any {
	row := make(map[string]any, len(headers))
	for i, h := range headers {
		key := fmt.Sprintf("%v", h)
		if i < len(values) {
			row[key] = values[i]
		} else {
			row[key] = ""
		}
	}
	return row
}

func (p *Processor) createRow(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for create_row action")
	}
	if f.values == "" {
		return nil, fmt.Errorf("values is required for create_row action")
	}

	vals, err := parseJSONArray(f.values)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	rangeStr := f.sheetName
	resp, err := svc.Spreadsheets.Values.Append(f.spreadsheetID, rangeStr, &sheets.ValueRange{
		Values: [][]any{vals},
	}).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append row: %w", err)
	}

	return map[string]any{
		"updated_range":  resp.Updates.UpdatedRange,
		"updated_rows":   resp.Updates.UpdatedRows,
		"updated_cells":  resp.Updates.UpdatedCells,
	}, nil
}

func (p *Processor) createRows(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for create_rows action")
	}
	if f.rows == "" {
		return nil, fmt.Errorf("rows is required for create_rows action")
	}

	rowData, err := parseJSONArrayOfArrays(f.rows)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	var values [][]any
	for _, row := range rowData {
		values = append(values, row)
	}

	rangeStr := f.sheetName
	resp, err := svc.Spreadsheets.Values.Append(f.spreadsheetID, rangeStr, &sheets.ValueRange{
		Values: values,
	}).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to append rows: %w", err)
	}

	return map[string]any{
		"updated_range":  resp.Updates.UpdatedRange,
		"updated_rows":   resp.Updates.UpdatedRows,
		"updated_cells":  resp.Updates.UpdatedCells,
	}, nil
}

func (p *Processor) createRowAtTop(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for create_row_at_top action")
	}
	if f.values == "" {
		return nil, fmt.Errorf("values is required for create_row_at_top action")
	}

	vals, err := parseJSONArray(f.values)
	if err != nil {
		return nil, err
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
				InsertDimension: &sheets.InsertDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  "ROWS",
						StartIndex: 1,
						EndIndex:   2,
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to insert row: %w", err)
	}

	rangeStr := fmt.Sprintf("%s!2:2", f.sheetName)
	resp, err := svc.Spreadsheets.Values.Update(f.spreadsheetID, rangeStr, &sheets.ValueRange{
		Values: [][]any{vals},
	}).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to write row data: %w", err)
	}

	return map[string]any{
		"updated_range": resp.UpdatedRange,
		"updated_cells": resp.UpdatedCells,
	}, nil
}

func (p *Processor) getRow(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for get_row action")
	}
	if f.rowNumber <= 0 {
		return nil, fmt.Errorf("row_number is required for get_row action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	headerRange := fmt.Sprintf("%s!1:1", f.sheetName)
	headerResp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, headerRange).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	var headers []any
	if len(headerResp.Values) > 0 {
		headers = headerResp.Values[0]
	}

	rowRange := fmt.Sprintf("%s!%d:%d", f.sheetName, f.rowNumber, f.rowNumber)
	rowResp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, rowRange).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get row: %w", err)
	}

	if len(rowResp.Values) == 0 {
		return map[string]any{
			"row_number": f.rowNumber,
			"found":      false,
		}, nil
	}

	return map[string]any{
		"row_number": f.rowNumber,
		"found":      true,
		"row":        buildRowMap(headers, rowResp.Values[0]),
		"values":     rowResp.Values[0],
	}, nil
}

func (p *Processor) getRows(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for get_rows action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	rangeStr := f.sheetName
	resp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, rangeStr).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(resp.Values) == 0 {
		return map[string]any{"rows": []map[string]any{}, "count": 0}, nil
	}

	headers := resp.Values[0]
	dataRows := resp.Values[1:]

	if len(dataRows) > f.maxResults {
		dataRows = dataRows[:f.maxResults]
	}

	rows := make([]map[string]any, 0, len(dataRows))
	for i, row := range dataRows {
		entry := buildRowMap(headers, row)
		entry["_row_number"] = i + 2
		rows = append(rows, entry)
	}

	result := map[string]any{
		"rows":  rows,
		"count": len(rows),
	}

	if f.includeHeaders {
		result["headers"] = headers
	}

	return result, nil
}

func (p *Processor) getDataRange(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for get_data_range action")
	}
	if f.rangeField == "" {
		return nil, fmt.Errorf("range is required for get_data_range action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	rangeStr := fmt.Sprintf("%s!%s", f.sheetName, f.rangeField)
	resp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, rangeStr).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get data range: %w", err)
	}

	return map[string]any{
		"range":  resp.Range,
		"values": resp.Values,
		"count":  len(resp.Values),
	}, nil
}

func (p *Processor) lookupRow(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for lookup_row action")
	}
	if f.columnName == "" {
		return nil, fmt.Errorf("column_name is required for lookup_row action")
	}
	if f.lookupValue == "" {
		return nil, fmt.Errorf("lookup_value is required for lookup_row action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	resp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, f.sheetName).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	if len(resp.Values) < 2 {
		return map[string]any{"found": false}, nil
	}

	headers := resp.Values[0]
	colIdx := -1
	for i, h := range headers {
		if fmt.Sprintf("%v", h) == f.columnName {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return nil, fmt.Errorf("column %q not found in headers", f.columnName)
	}

	for i, row := range resp.Values[1:] {
		if colIdx < len(row) && fmt.Sprintf("%v", row[colIdx]) == f.lookupValue {
			return map[string]any{
				"found":      true,
				"row_number": i + 2,
				"row":        buildRowMap(headers, row),
				"values":     row,
			}, nil
		}
	}

	return map[string]any{"found": false}, nil
}

func (p *Processor) lookupRows(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for lookup_rows action")
	}
	if f.columnName == "" {
		return nil, fmt.Errorf("column_name is required for lookup_rows action")
	}
	if f.lookupValue == "" {
		return nil, fmt.Errorf("lookup_value is required for lookup_rows action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	resp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, f.sheetName).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get sheet data: %w", err)
	}

	if len(resp.Values) < 2 {
		return map[string]any{"rows": []map[string]any{}, "count": 0}, nil
	}

	headers := resp.Values[0]
	colIdx := -1
	for i, h := range headers {
		if fmt.Sprintf("%v", h) == f.columnName {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return nil, fmt.Errorf("column %q not found in headers", f.columnName)
	}

	var matches []map[string]any
	for i, row := range resp.Values[1:] {
		if len(matches) >= f.maxResults {
			break
		}
		if colIdx < len(row) && fmt.Sprintf("%v", row[colIdx]) == f.lookupValue {
			entry := buildRowMap(headers, row)
			entry["_row_number"] = i + 2
			matches = append(matches, entry)
		}
	}

	if matches == nil {
		matches = []map[string]any{}
	}

	return map[string]any{
		"rows":  matches,
		"count": len(matches),
	}, nil
}

func (p *Processor) updateRow(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for update_row action")
	}
	if f.rowNumber <= 0 {
		return nil, fmt.Errorf("row_number is required for update_row action")
	}
	if f.values == "" {
		return nil, fmt.Errorf("values is required for update_row action")
	}

	vals, err := parseJSONArray(f.values)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	rangeStr := fmt.Sprintf("%s!%d:%d", f.sheetName, f.rowNumber, f.rowNumber)
	resp, err := svc.Spreadsheets.Values.Update(f.spreadsheetID, rangeStr, &sheets.ValueRange{
		Values: [][]any{vals},
	}).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update row: %w", err)
	}

	return map[string]any{
		"updated_range": resp.UpdatedRange,
		"updated_rows":  resp.UpdatedRows,
		"updated_cells": resp.UpdatedCells,
	}, nil
}

func (p *Processor) updateRows(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for update_rows action")
	}
	if f.rowNumber <= 0 {
		return nil, fmt.Errorf("row_number is required for update_rows action")
	}
	if f.rows == "" {
		return nil, fmt.Errorf("rows is required for update_rows action")
	}

	rowData, err := parseJSONArrayOfArrays(f.rows)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	var data []*sheets.ValueRange
	for i, row := range rowData {
		rowNum := f.rowNumber + i
		rangeStr := fmt.Sprintf("%s!%d:%d", f.sheetName, rowNum, rowNum)
		data = append(data, &sheets.ValueRange{
			Range:  rangeStr,
			Values: [][]any{row},
		})
	}

	resp, err := svc.Spreadsheets.Values.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             data,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update rows: %w", err)
	}

	return map[string]any{
		"total_updated_rows":  resp.TotalUpdatedRows,
		"total_updated_cells": resp.TotalUpdatedCells,
	}, nil
}

func (p *Processor) clearRows(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for clear_rows action")
	}
	if f.rowNumber <= 0 {
		return nil, fmt.Errorf("row_number is required for clear_rows action")
	}

	endRow := f.endRowNumber
	if endRow <= 0 {
		endRow = f.rowNumber
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	rangeStr := fmt.Sprintf("%s!%d:%d", f.sheetName, f.rowNumber, endRow)
	resp, err := svc.Spreadsheets.Values.Clear(f.spreadsheetID, rangeStr, &sheets.ClearValuesRequest{}).
		Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to clear rows: %w", err)
	}

	return map[string]any{
		"cleared_range": resp.ClearedRange,
	}, nil
}

func (p *Processor) deleteRows(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for delete_rows action")
	}
	if f.rowNumber <= 0 {
		return nil, fmt.Errorf("row_number is required for delete_rows action")
	}

	endRow := f.endRowNumber
	if endRow <= 0 {
		endRow = f.rowNumber
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
				DeleteDimension: &sheets.DeleteDimensionRequest{
					Range: &sheets.DimensionRange{
						SheetId:    sheetID,
						Dimension:  "ROWS",
						StartIndex: int64(f.rowNumber - 1),
						EndIndex:   int64(endRow),
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to delete rows: %w", err)
	}

	return map[string]any{
		"deleted":    true,
		"start_row":  f.rowNumber,
		"end_row":    endRow,
		"rows_count": endRow - f.rowNumber + 1,
	}, nil
}

func (p *Processor) createColumn(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for create_column action")
	}
	if f.columnName == "" {
		return nil, fmt.Errorf("column_name is required for create_column action")
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	headerRange := fmt.Sprintf("%s!1:1", f.sheetName)
	headerResp, err := svc.Spreadsheets.Values.Get(f.spreadsheetID, headerRange).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	nextCol := 0
	if len(headerResp.Values) > 0 {
		nextCol = len(headerResp.Values[0])
	}

	colLetter := columnIndexToLetter(nextCol)
	headerCell := fmt.Sprintf("%s!%s1", f.sheetName, colLetter)

	_, err = svc.Spreadsheets.Values.Update(f.spreadsheetID, headerCell, &sheets.ValueRange{
		Values: [][]any{{f.columnName}},
	}).ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to set column header: %w", err)
	}

	if f.values != "" {
		vals, err := parseJSONArray(f.values)
		if err != nil {
			return nil, err
		}

		var colValues [][]any
		for _, v := range vals {
			colValues = append(colValues, []any{v})
		}

		dataRange := fmt.Sprintf("%s!%s2:%s%d", f.sheetName, colLetter, colLetter, len(vals)+1)
		_, err = svc.Spreadsheets.Values.Update(f.spreadsheetID, dataRange, &sheets.ValueRange{
			Values: colValues,
		}).ValueInputOption("USER_ENTERED").Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to write column data: %w", err)
		}
	}

	return map[string]any{
		"column_name":   f.columnName,
		"column_letter": colLetter,
		"column_index":  nextCol,
	}, nil
}

func columnIndexToLetter(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}
