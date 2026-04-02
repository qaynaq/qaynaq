package google_sheets

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func parseHexColor(hex string) *sheets.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return nil
	}
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil
	}
	return &sheets.Color{
		Red:   float64(r) / 255.0,
		Green: float64(g) / 255.0,
		Blue:  float64(b) / 255.0,
		Alpha: 1.0,
	}
}

func columnLetterToIndex(col string) int64 {
	col = strings.ToUpper(col)
	result := int64(0)
	for _, c := range col {
		result = result*26 + int64(c-'A') + 1
	}
	return result
}

func parseA1Range(a1 string, sheetID int64) (*sheets.GridRange, error) {
	gr := &sheets.GridRange{SheetId: sheetID}

	parts := strings.Split(a1, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid A1 notation: %s", a1)
	}

	startCol, startRow, err := parseCell(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid start cell %q: %w", parts[0], err)
	}
	endCol, endRow, err := parseCell(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid end cell %q: %w", parts[1], err)
	}

	gr.StartColumnIndex = startCol
	gr.StartRowIndex = startRow
	gr.EndColumnIndex = endCol + 1
	gr.EndRowIndex = endRow + 1

	return gr, nil
}

func parseCell(cell string) (col int64, row int64, err error) {
	cell = strings.TrimSpace(cell)
	i := 0
	for i < len(cell) && cell[i] >= 'A' && cell[i] <= 'Z' || i < len(cell) && cell[i] >= 'a' && cell[i] <= 'z' {
		i++
	}
	if i == 0 || i == len(cell) {
		return 0, 0, fmt.Errorf("invalid cell reference: %s", cell)
	}
	colStr := cell[:i]
	rowStr := cell[i:]

	col = columnLetterToIndex(colStr) - 1

	rowNum, err := strconv.ParseInt(rowStr, 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid row number in %q: %w", cell, err)
	}
	row = rowNum - 1

	return col, row, nil
}

func (p *Processor) buildCellFormat(f *resolvedFields) *sheets.CellFormat {
	cf := &sheets.CellFormat{}
	hasFormat := false

	if f.bold || f.italic || f.strikethrough {
		cf.TextFormat = &sheets.TextFormat{
			Bold:          f.bold,
			Italic:        f.italic,
			Strikethrough: f.strikethrough,
		}
		if f.foregroundColor != "" {
			cf.TextFormat.ForegroundColor = parseHexColor(f.foregroundColor)
		}
		hasFormat = true
	} else if f.foregroundColor != "" {
		cf.TextFormat = &sheets.TextFormat{
			ForegroundColor: parseHexColor(f.foregroundColor),
		}
		hasFormat = true
	}

	if f.backgroundColor != "" {
		cf.BackgroundColor = parseHexColor(f.backgroundColor)
		hasFormat = true
	}

	if f.numberFormat != "" {
		cf.NumberFormat = &sheets.NumberFormat{
			Type:    "NUMBER",
			Pattern: f.numberFormat,
		}
		hasFormat = true
	}

	if !hasFormat {
		return nil
	}
	return cf
}

func (p *Processor) copyRange(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for copy_range action")
	}
	if f.sourceRange == "" {
		return nil, fmt.Errorf("source_range is required for copy_range action")
	}
	if f.destinationRange == "" {
		return nil, fmt.Errorf("destination_range is required for copy_range action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	srcGrid, err := parseA1Range(f.sourceRange, sheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source_range: %w", err)
	}

	dstGrid, err := parseA1Range(f.destinationRange, sheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse destination_range: %w", err)
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				CopyPaste: &sheets.CopyPasteRequest{
					Source:      srcGrid,
					Destination: dstGrid,
					PasteType:   f.pasteType,
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to copy range: %w", err)
	}

	return map[string]any{
		"copied":            true,
		"source_range":      f.sourceRange,
		"destination_range": f.destinationRange,
		"paste_type":        f.pasteType,
	}, nil
}

func (p *Processor) sortRange(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for sort_range action")
	}
	if f.rangeField == "" {
		return nil, fmt.Errorf("range is required for sort_range action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	gridRange, err := parseA1Range(f.rangeField, sheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range: %w", err)
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				SortRange: &sheets.SortRangeRequest{
					Range: gridRange,
					SortSpecs: []*sheets.SortSpec{
						{
							DimensionIndex: int64(f.sortColumnIndex),
							SortOrder:      f.sortOrder,
						},
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to sort range: %w", err)
	}

	return map[string]any{
		"sorted":       true,
		"range":        f.rangeField,
		"sort_column":  f.sortColumnIndex,
		"sort_order":   f.sortOrder,
	}, nil
}

func (p *Processor) formatCells(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for format_cells action")
	}
	if f.rangeField == "" {
		return nil, fmt.Errorf("range is required for format_cells action")
	}

	cellFormat := p.buildCellFormat(f)
	if cellFormat == nil {
		return nil, fmt.Errorf("at least one formatting option (bold, italic, strikethrough, background_color, foreground_color, number_format) must be set")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	gridRange, err := parseA1Range(f.rangeField, sheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range: %w", err)
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	var fieldMaskParts []string
	if cellFormat.TextFormat != nil {
		fieldMaskParts = append(fieldMaskParts, "userEnteredFormat.textFormat")
	}
	if cellFormat.BackgroundColor != nil {
		fieldMaskParts = append(fieldMaskParts, "userEnteredFormat.backgroundColor")
	}
	if cellFormat.NumberFormat != nil {
		fieldMaskParts = append(fieldMaskParts, "userEnteredFormat.numberFormat")
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: gridRange,
					Cell: &sheets.CellData{
						UserEnteredFormat: cellFormat,
					},
					Fields: strings.Join(fieldMaskParts, ","),
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to format cells: %w", err)
	}

	return map[string]any{
		"formatted": true,
		"range":     f.rangeField,
	}, nil
}

func (p *Processor) formatRow(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for format_row action")
	}
	if f.rowNumber <= 0 {
		return nil, fmt.Errorf("row_number is required for format_row action")
	}

	cellFormat := p.buildCellFormat(f)
	if cellFormat == nil {
		return nil, fmt.Errorf("at least one formatting option (bold, italic, strikethrough, background_color, foreground_color, number_format) must be set")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	gridRange := &sheets.GridRange{
		SheetId:       sheetID,
		StartRowIndex: int64(f.rowNumber - 1),
		EndRowIndex:   int64(f.rowNumber),
	}

	var fieldMaskParts []string
	if cellFormat.TextFormat != nil {
		fieldMaskParts = append(fieldMaskParts, "userEnteredFormat.textFormat")
	}
	if cellFormat.BackgroundColor != nil {
		fieldMaskParts = append(fieldMaskParts, "userEnteredFormat.backgroundColor")
	}
	if cellFormat.NumberFormat != nil {
		fieldMaskParts = append(fieldMaskParts, "userEnteredFormat.numberFormat")
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				RepeatCell: &sheets.RepeatCellRequest{
					Range: gridRange,
					Cell: &sheets.CellData{
						UserEnteredFormat: cellFormat,
					},
					Fields: strings.Join(fieldMaskParts, ","),
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to format row: %w", err)
	}

	return map[string]any{
		"formatted":  true,
		"row_number": f.rowNumber,
	}, nil
}

func (p *Processor) setDataValidation(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for set_data_validation action")
	}
	if f.rangeField == "" {
		return nil, fmt.Errorf("range is required for set_data_validation action")
	}
	if f.validationType == "" {
		return nil, fmt.Errorf("validation_type is required for set_data_validation action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	gridRange, err := parseA1Range(f.rangeField, sheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range: %w", err)
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	rule := &sheets.DataValidationRule{
		Condition: &sheets.BooleanCondition{
			Type: f.validationType,
		},
		Strict:         true,
		ShowCustomUi:   true,
	}

	if f.validationValues != "" {
		values := splitCSV(f.validationValues)
		for _, v := range values {
			rule.Condition.Values = append(rule.Condition.Values, &sheets.ConditionValue{
				UserEnteredValue: v,
			})
		}
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				SetDataValidation: &sheets.SetDataValidationRequest{
					Range: gridRange,
					Rule:  rule,
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to set data validation: %w", err)
	}

	return map[string]any{
		"validated":       true,
		"range":           f.rangeField,
		"validation_type": f.validationType,
	}, nil
}

func (p *Processor) createConditionalFormatting(ctx context.Context, f *resolvedFields) (map[string]any, error) {
	if f.spreadsheetID == "" {
		return nil, fmt.Errorf("spreadsheet_id is required for create_conditional_formatting action")
	}
	if f.rangeField == "" {
		return nil, fmt.Errorf("range is required for create_conditional_formatting action")
	}
	if f.conditionType == "" {
		return nil, fmt.Errorf("condition_type is required for create_conditional_formatting action")
	}

	sheetID, err := p.getSheetIDByName(ctx, f.spreadsheetID, f.sheetName)
	if err != nil {
		return nil, err
	}

	gridRange, err := parseA1Range(f.rangeField, sheetID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse range: %w", err)
	}

	svc, err := p.initSheetsService()
	if err != nil {
		return nil, err
	}

	condition := &sheets.BooleanCondition{
		Type: f.conditionType,
	}
	if f.conditionValue != "" {
		condition.Values = []*sheets.ConditionValue{
			{UserEnteredValue: f.conditionValue},
		}
	}

	format := &sheets.CellFormat{}
	if f.conditionBgColor != "" {
		format.BackgroundColor = parseHexColor(f.conditionBgColor)
	}

	_, err = svc.Spreadsheets.BatchUpdate(f.spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
					Rule: &sheets.ConditionalFormatRule{
						Ranges: []*sheets.GridRange{gridRange},
						BooleanRule: &sheets.BooleanRule{
							Condition: condition,
							Format:    format,
						},
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create conditional formatting: %w", err)
	}

	return map[string]any{
		"created":        true,
		"range":          f.rangeField,
		"condition_type": f.conditionType,
	}, nil
}
