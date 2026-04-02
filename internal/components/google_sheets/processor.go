package google_sheets

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/warpstreamlabs/bento/public/service"
	"google.golang.org/api/sheets/v4"
)

func init() {
	err := service.RegisterProcessor(
		"google_sheets", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type Processor struct {
	serviceAccountJSON string
	oauthConnection    string
	delegateTo         string
	action             string
	spreadsheetID      *service.InterpolatedString
	sheetName          *service.InterpolatedString
	title              *service.InterpolatedString
	rangeField         *service.InterpolatedString
	rowNumber          *service.InterpolatedString
	endRowNumber       *service.InterpolatedString
	columnName         *service.InterpolatedString
	lookupValue        *service.InterpolatedString
	values             *service.InterpolatedString
	rows               *service.InterpolatedString
	maxResults         *service.InterpolatedString
	newName            *service.InterpolatedString
	destSpreadsheetID  *service.InterpolatedString
	sourceRange        *service.InterpolatedString
	destinationRange   *service.InterpolatedString
	pasteType          *service.InterpolatedString
	sortColumnIndex    *service.InterpolatedString
	sortOrder          *service.InterpolatedString
	bold               *service.InterpolatedString
	italic             *service.InterpolatedString
	strikethrough      *service.InterpolatedString
	backgroundColor    *service.InterpolatedString
	foregroundColor    *service.InterpolatedString
	numberFormat       *service.InterpolatedString
	frozenRows         *service.InterpolatedString
	frozenColumns      *service.InterpolatedString
	hidden             *service.InterpolatedString
	sheetPosition      *service.InterpolatedString
	validationType     *service.InterpolatedString
	validationValues   *service.InterpolatedString
	conditionType      *service.InterpolatedString
	conditionValue     *service.InterpolatedString
	conditionBgColor   *service.InterpolatedString
	includeHeaders     *service.InterpolatedString

	sheetsService  *sheets.Service
	serviceOnce    sync.Once
	serviceInitErr error
	logger         *service.Logger
}

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Processor, error) {
	serviceAccountJSON, err := conf.FieldString(gsfServiceAccountJSON)
	if err != nil {
		return nil, err
	}

	oauthConnection, err := conf.FieldString(gsfOAuthConnection)
	if err != nil {
		return nil, err
	}

	if serviceAccountJSON == "" && oauthConnection == "" {
		return nil, fmt.Errorf("either service_account_json or oauth_connection must be provided")
	}

	action, err := conf.FieldString(gsfAction)
	if err != nil {
		return nil, err
	}

	p := &Processor{
		serviceAccountJSON: serviceAccountJSON,
		oauthConnection:    oauthConnection,
		action:             action,
		logger:             mgr.Logger(),
	}

	if conf.Contains(gsfDelegateTo) {
		if p.delegateTo, err = conf.FieldString(gsfDelegateTo); err != nil {
			return nil, err
		}
	}

	if p.spreadsheetID, err = conf.FieldInterpolatedString(gsfSpreadsheetID); err != nil {
		return nil, err
	}

	if p.sheetName, err = conf.FieldInterpolatedString(gsfSheetName); err != nil {
		return nil, err
	}

	if conf.Contains(gsfTitle) {
		if p.title, err = conf.FieldInterpolatedString(gsfTitle); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfRange) {
		if p.rangeField, err = conf.FieldInterpolatedString(gsfRange); err != nil {
			return nil, err
		}
	}

	if p.rowNumber, err = conf.FieldInterpolatedString(gsfRowNumber); err != nil {
		return nil, err
	}

	if p.endRowNumber, err = conf.FieldInterpolatedString(gsfEndRowNumber); err != nil {
		return nil, err
	}

	if conf.Contains(gsfColumnName) {
		if p.columnName, err = conf.FieldInterpolatedString(gsfColumnName); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfLookupValue) {
		if p.lookupValue, err = conf.FieldInterpolatedString(gsfLookupValue); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfValues) {
		if p.values, err = conf.FieldInterpolatedString(gsfValues); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfRows) {
		if p.rows, err = conf.FieldInterpolatedString(gsfRows); err != nil {
			return nil, err
		}
	}

	if p.maxResults, err = conf.FieldInterpolatedString(gsfMaxResults); err != nil {
		return nil, err
	}

	if conf.Contains(gsfNewName) {
		if p.newName, err = conf.FieldInterpolatedString(gsfNewName); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfDestSpreadsheetID) {
		if p.destSpreadsheetID, err = conf.FieldInterpolatedString(gsfDestSpreadsheetID); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfSourceRange) {
		if p.sourceRange, err = conf.FieldInterpolatedString(gsfSourceRange); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfDestinationRange) {
		if p.destinationRange, err = conf.FieldInterpolatedString(gsfDestinationRange); err != nil {
			return nil, err
		}
	}

	if p.pasteType, err = conf.FieldInterpolatedString(gsfPasteType); err != nil {
		return nil, err
	}

	if p.sortColumnIndex, err = conf.FieldInterpolatedString(gsfSortColumnIndex); err != nil {
		return nil, err
	}

	if p.sortOrder, err = conf.FieldInterpolatedString(gsfSortOrder); err != nil {
		return nil, err
	}

	if p.bold, err = conf.FieldInterpolatedString(gsfBold); err != nil {
		return nil, err
	}

	if p.italic, err = conf.FieldInterpolatedString(gsfItalic); err != nil {
		return nil, err
	}

	if p.strikethrough, err = conf.FieldInterpolatedString(gsfStrikethrough); err != nil {
		return nil, err
	}

	if conf.Contains(gsfBackgroundColor) {
		if p.backgroundColor, err = conf.FieldInterpolatedString(gsfBackgroundColor); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfForegroundColor) {
		if p.foregroundColor, err = conf.FieldInterpolatedString(gsfForegroundColor); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfNumberFormat) {
		if p.numberFormat, err = conf.FieldInterpolatedString(gsfNumberFormat); err != nil {
			return nil, err
		}
	}

	if p.frozenRows, err = conf.FieldInterpolatedString(gsfFrozenRows); err != nil {
		return nil, err
	}

	if p.frozenColumns, err = conf.FieldInterpolatedString(gsfFrozenColumns); err != nil {
		return nil, err
	}

	if p.hidden, err = conf.FieldInterpolatedString(gsfHidden); err != nil {
		return nil, err
	}

	if p.sheetPosition, err = conf.FieldInterpolatedString(gsfSheetPosition); err != nil {
		return nil, err
	}

	if conf.Contains(gsfValidationType) {
		if p.validationType, err = conf.FieldInterpolatedString(gsfValidationType); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfValidationValues) {
		if p.validationValues, err = conf.FieldInterpolatedString(gsfValidationValues); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfConditionType) {
		if p.conditionType, err = conf.FieldInterpolatedString(gsfConditionType); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfConditionValue) {
		if p.conditionValue, err = conf.FieldInterpolatedString(gsfConditionValue); err != nil {
			return nil, err
		}
	}

	if conf.Contains(gsfConditionBackgroundColor) {
		if p.conditionBgColor, err = conf.FieldInterpolatedString(gsfConditionBackgroundColor); err != nil {
			return nil, err
		}
	}

	if p.includeHeaders, err = conf.FieldInterpolatedString(gsfIncludeHeaders); err != nil {
		return nil, err
	}

	return p, nil
}

type resolvedFields struct {
	spreadsheetID     string
	sheetName         string
	title             string
	rangeField        string
	rowNumber         int
	endRowNumber      int
	columnName        string
	lookupValue       string
	values            string
	rows              string
	maxResults        int
	newName           string
	destSpreadsheetID string
	sourceRange       string
	destinationRange  string
	pasteType         string
	sortColumnIndex   int
	sortOrder         string
	bold              bool
	italic            bool
	strikethrough     bool
	backgroundColor   string
	foregroundColor   string
	numberFormat      string
	frozenRows        int
	frozenColumns     int
	hidden            bool
	sheetPosition     int
	includeHeaders    bool
	validationType    string
	validationValues  string
	conditionType     string
	conditionValue    string
	conditionBgColor  string
}

func (p *Processor) resolveFields(msg *service.Message) (*resolvedFields, error) {
	r := &resolvedFields{}
	var err error

	if r.spreadsheetID, err = p.spreadsheetID.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate spreadsheet_id: %w", err)
	}

	if r.sheetName, err = p.sheetName.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate sheet_name: %w", err)
	}

	if p.title != nil {
		if r.title, err = p.title.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate title: %w", err)
		}
	}

	if p.rangeField != nil {
		if r.rangeField, err = p.rangeField.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate range: %w", err)
		}
	}

	if p.columnName != nil {
		if r.columnName, err = p.columnName.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate column_name: %w", err)
		}
	}

	if p.lookupValue != nil {
		if r.lookupValue, err = p.lookupValue.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate lookup_value: %w", err)
		}
	}

	if p.values != nil {
		if r.values, err = p.values.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate values: %w", err)
		}
	}

	if p.rows != nil {
		if r.rows, err = p.rows.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate rows: %w", err)
		}
	}

	if p.newName != nil {
		if r.newName, err = p.newName.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate new_name: %w", err)
		}
	}

	if p.destSpreadsheetID != nil {
		if r.destSpreadsheetID, err = p.destSpreadsheetID.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate destination_spreadsheet_id: %w", err)
		}
	}

	if p.sourceRange != nil {
		if r.sourceRange, err = p.sourceRange.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate source_range: %w", err)
		}
	}

	if p.destinationRange != nil {
		if r.destinationRange, err = p.destinationRange.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate destination_range: %w", err)
		}
	}

	if r.pasteType, err = p.pasteType.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate paste_type: %w", err)
	}

	if r.sortOrder, err = p.sortOrder.TryString(msg); err != nil {
		return nil, fmt.Errorf("failed to interpolate sort_order: %w", err)
	}

	r.rowNumber = resolveInt(p.rowNumber, msg)
	r.endRowNumber = resolveInt(p.endRowNumber, msg)
	r.maxResults = resolveInt(p.maxResults, msg)
	r.sortColumnIndex = resolveInt(p.sortColumnIndex, msg)
	r.frozenRows = resolveInt(p.frozenRows, msg)
	r.frozenColumns = resolveInt(p.frozenColumns, msg)
	r.sheetPosition = resolveInt(p.sheetPosition, msg)
	r.bold = resolveBool(p.bold, msg)
	r.italic = resolveBool(p.italic, msg)
	r.strikethrough = resolveBool(p.strikethrough, msg)
	r.hidden = resolveBool(p.hidden, msg)
	r.includeHeaders = resolveBool(p.includeHeaders, msg)

	if p.backgroundColor != nil {
		if r.backgroundColor, err = p.backgroundColor.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate background_color: %w", err)
		}
	}

	if p.foregroundColor != nil {
		if r.foregroundColor, err = p.foregroundColor.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate foreground_color: %w", err)
		}
	}

	if p.numberFormat != nil {
		if r.numberFormat, err = p.numberFormat.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate number_format: %w", err)
		}
	}

	if p.validationType != nil {
		if r.validationType, err = p.validationType.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate validation_type: %w", err)
		}
	}

	if p.validationValues != nil {
		if r.validationValues, err = p.validationValues.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate validation_values: %w", err)
		}
	}

	if p.conditionType != nil {
		if r.conditionType, err = p.conditionType.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate condition_type: %w", err)
		}
	}

	if p.conditionValue != nil {
		if r.conditionValue, err = p.conditionValue.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate condition_value: %w", err)
		}
	}

	if p.conditionBgColor != nil {
		if r.conditionBgColor, err = p.conditionBgColor.TryString(msg); err != nil {
			return nil, fmt.Errorf("failed to interpolate condition_background_color: %w", err)
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
	case actionCreateSpreadsheet:
		result, err = p.createSpreadsheet(ctx, fields)
	case actionGetSpreadsheet:
		result, err = p.getSpreadsheet(ctx, fields)
	case actionCreateWorksheet:
		result, err = p.createWorksheet(ctx, fields)
	case actionFindWorksheet:
		result, err = p.findWorksheet(ctx, fields)
	case actionFindOrCreateWorksheet:
		result, err = p.findOrCreateWorksheet(ctx, fields)
	case actionDeleteWorksheet:
		result, err = p.deleteWorksheet(ctx, fields)
	case actionRenameWorksheet:
		result, err = p.renameWorksheet(ctx, fields)
	case actionCopyWorksheet:
		result, err = p.copyWorksheet(ctx, fields)
	case actionChangeSheetProperties:
		result, err = p.changeSheetProperties(ctx, fields)
	case actionCreateRow:
		result, err = p.createRow(ctx, fields)
	case actionCreateRows:
		result, err = p.createRows(ctx, fields)
	case actionCreateRowAtTop:
		result, err = p.createRowAtTop(ctx, fields)
	case actionGetRow:
		result, err = p.getRow(ctx, fields)
	case actionGetRows:
		result, err = p.getRows(ctx, fields)
	case actionGetDataRange:
		result, err = p.getDataRange(ctx, fields)
	case actionLookupRow:
		result, err = p.lookupRow(ctx, fields)
	case actionLookupRows:
		result, err = p.lookupRows(ctx, fields)
	case actionUpdateRow:
		result, err = p.updateRow(ctx, fields)
	case actionUpdateRows:
		result, err = p.updateRows(ctx, fields)
	case actionClearRows:
		result, err = p.clearRows(ctx, fields)
	case actionDeleteRows:
		result, err = p.deleteRows(ctx, fields)
	case actionCreateColumn:
		result, err = p.createColumn(ctx, fields)
	case actionCopyRange:
		result, err = p.copyRange(ctx, fields)
	case actionSortRange:
		result, err = p.sortRange(ctx, fields)
	case actionFormatCells:
		result, err = p.formatCells(ctx, fields)
	case actionFormatRow:
		result, err = p.formatRow(ctx, fields)
	case actionSetDataValidation:
		result, err = p.setDataValidation(ctx, fields)
	case actionCreateConditionalFormatting:
		result, err = p.createConditionalFormatting(ctx, fields)
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

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}
	return result
}

