package google_sheets

import "github.com/warpstreamlabs/bento/public/service"

const (
	gsfServiceAccountJSON      = "service_account_json"
	gsfDelegateTo              = "delegate_to"
	gsfAction                  = "action"
	gsfSpreadsheetID           = "spreadsheet_id"
	gsfSheetName               = "sheet_name"
	gsfTitle                   = "title"
	gsfRange                   = "range"
	gsfRowNumber               = "row_number"
	gsfEndRowNumber            = "end_row_number"
	gsfColumnName              = "column_name"
	gsfLookupValue             = "lookup_value"
	gsfValues                  = "values"
	gsfRows                    = "rows"
	gsfMaxResults              = "max_results"
	gsfNewName                 = "new_name"
	gsfDestSpreadsheetID       = "destination_spreadsheet_id"
	gsfSourceRange             = "source_range"
	gsfDestinationRange        = "destination_range"
	gsfPasteType               = "paste_type"
	gsfSortColumnIndex         = "sort_column_index"
	gsfSortOrder               = "sort_order"
	gsfBold                    = "bold"
	gsfItalic                  = "italic"
	gsfStrikethrough           = "strikethrough"
	gsfBackgroundColor         = "background_color"
	gsfForegroundColor         = "foreground_color"
	gsfNumberFormat            = "number_format"
	gsfFrozenRows              = "frozen_rows"
	gsfFrozenColumns           = "frozen_columns"
	gsfHidden                  = "hidden"
	gsfSheetPosition           = "sheet_position"
	gsfValidationType          = "validation_type"
	gsfValidationValues        = "validation_values"
	gsfConditionType           = "condition_type"
	gsfConditionValue          = "condition_value"
	gsfConditionBackgroundColor = "condition_background_color"
	gsfIncludeHeaders          = "include_headers"
	gsfOAuthConnection         = "oauth_connection"
)

const (
	actionGetSpreadsheet           = "get_spreadsheet"
	actionCreateWorksheet          = "create_worksheet"
	actionFindWorksheet            = "find_worksheet"
	actionFindOrCreateWorksheet    = "find_or_create_worksheet"
	actionDeleteWorksheet          = "delete_worksheet"
	actionRenameWorksheet          = "rename_worksheet"
	actionCopyWorksheet            = "copy_worksheet"
	actionChangeSheetProperties    = "change_sheet_properties"
	actionCreateRow                = "create_row"
	actionCreateRows               = "create_rows"
	actionCreateRowAtTop           = "create_row_at_top"
	actionGetRow                   = "get_row"
	actionGetRows                  = "get_rows"
	actionGetDataRange             = "get_data_range"
	actionLookupRow                = "lookup_row"
	actionLookupRows               = "lookup_rows"
	actionUpdateRow                = "update_row"
	actionUpdateRows               = "update_rows"
	actionClearRows                = "clear_rows"
	actionDeleteRows               = "delete_rows"
	actionCreateColumn             = "create_column"
	actionCopyRange                = "copy_range"
	actionSortRange                = "sort_range"
	actionFormatCells              = "format_cells"
	actionFormatRow                = "format_row"
	actionSetDataValidation        = "set_data_validation"
	actionCreateConditionalFormatting = "create_conditional_formatting"
	actionCreateSpreadsheet          = "create_spreadsheet"
)

func Config() *service.ConfigSpec {
	return service.NewConfigSpec().
		Beta().
		Categories("Integration").
		Summary("Performs Google Sheets operations - create, read, update, and format spreadsheets, worksheets, rows, and cells.").
		Description(`
This processor interacts with the Google Sheets API using service account authentication.
It supports creating and managing spreadsheets, worksheets, rows, columns, and cell formatting,
as well as data lookups, sorting, validation, and conditional formatting.

Store your Google service account JSON as a secret in Settings > Secrets, then reference
it in the Service Account JSON field. For accessing spreadsheets owned by other users in a
Google Workspace domain, enable Domain-Wide Delegation and set the Delegate To field.

Most fields support interpolation functions, allowing dynamic values from message content
using the ` + "`${!this.field_name}`" + ` syntax.`).
		Field(service.NewStringField(gsfServiceAccountJSON).
			Description("Google service account credentials JSON. Store as a secret and reference via ${SECRET_NAME}.").
			Secret().
			Optional().
			Default("")).
		Field(service.NewStringField(gsfOAuthConnection).
			Description("OAuth connection for user authentication. Set up in Settings > Connections, then reference via ${CONN_NAME}. Alternative to Service Account JSON.").
			Secret().
			Optional().
			Default("")).
		Field(service.NewStringField(gsfDelegateTo).
			Description("Email address to impersonate via Domain-Wide Delegation. Required for accessing spreadsheets owned by other users in Google Workspace.").
			Default("").
			Optional()).
		Field(service.NewStringEnumField(gsfAction,
			actionCreateSpreadsheet,
			actionGetSpreadsheet,
			actionCreateWorksheet,
			actionFindWorksheet,
			actionFindOrCreateWorksheet,
			actionDeleteWorksheet,
			actionRenameWorksheet,
			actionCopyWorksheet,
			actionChangeSheetProperties,
			actionCreateRow,
			actionCreateRows,
			actionCreateRowAtTop,
			actionGetRow,
			actionGetRows,
			actionGetDataRange,
			actionLookupRow,
			actionLookupRows,
			actionUpdateRow,
			actionUpdateRows,
			actionClearRows,
			actionDeleteRows,
			actionCreateColumn,
			actionCopyRange,
			actionSortRange,
			actionFormatCells,
			actionFormatRow,
			actionSetDataValidation,
			actionCreateConditionalFormatting,
		).Description("The spreadsheet operation to perform.")).
		Field(service.NewInterpolatedStringField(gsfSpreadsheetID).
			Description("The spreadsheet identifier. Required for all actions.").
			Default("")).
		Field(service.NewInterpolatedStringField(gsfSheetName).
			Description("The worksheet/tab name. Used by most row and formatting actions to target a specific sheet.").
			Default("Sheet1")).
		Field(service.NewInterpolatedStringField(gsfTitle).
			Description("Name for new worksheets. Used by: create_worksheet, find_or_create_worksheet.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfRange).
			Description("A1 notation range (e.g. 'A1:D10'). Used by: get_data_range, format_cells, sort_range, set_data_validation, create_conditional_formatting.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfRowNumber).
			Description("Specific row number (1-based). Used by: get_row, update_row, update_rows (start row), clear_rows (start row), delete_rows (start row), format_row.").
			Default("0")).
		Field(service.NewInterpolatedStringField(gsfEndRowNumber).
			Description("End row number for range operations (1-based, inclusive). Used by: clear_rows, delete_rows.").
			Default("0")).
		Field(service.NewInterpolatedStringField(gsfColumnName).
			Description("Column header name for lookup operations. Used by: lookup_row, lookup_rows, create_column.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfLookupValue).
			Description("Value to search for in the specified column. Used by: lookup_row, lookup_rows.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfValues).
			Description("JSON array of values for a single row (e.g. '[\"val1\",\"val2\"]'). Used by: create_row, create_row_at_top, update_row, create_column.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfRows).
			Description("JSON array of arrays for multiple rows (e.g. '[[\"a\",\"b\"],[\"c\",\"d\"]]'). Used by: create_rows, update_rows.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfMaxResults).
			Description("Maximum number of rows to return. Used by: get_rows, lookup_rows.").
			Default("100")).
		Field(service.NewInterpolatedStringField(gsfNewName).
			Description("New name for renaming a worksheet. Required for: rename_worksheet.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfDestSpreadsheetID).
			Description("Target spreadsheet ID for copying a worksheet. Used by: copy_worksheet.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfSourceRange).
			Description("Source range in A1 notation for copy operations. Used by: copy_range.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfDestinationRange).
			Description("Destination range in A1 notation for copy operations. Used by: copy_range.").
			Default("").
			Optional()).
		Field(service.NewInterpolatedStringField(gsfPasteType).
			Description("Paste type for copy_range: PASTE_NORMAL, PASTE_VALUES, PASTE_FORMAT, PASTE_NO_BORDERS, PASTE_FORMULA, PASTE_DATA_VALIDATION, PASTE_CONDITIONAL_FORMATTING.").
			Default("PASTE_NORMAL").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfSortColumnIndex).
			Description("0-based column index to sort by. Used by: sort_range.").
			Default("0").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfSortOrder).
			Description("Sort direction: ASCENDING or DESCENDING. Used by: sort_range.").
			Default("ASCENDING").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfBold).
			Description("Apply bold formatting (true/false). Used by: format_cells, format_row.").
			Default("false").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfItalic).
			Description("Apply italic formatting (true/false). Used by: format_cells, format_row.").
			Default("false").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfStrikethrough).
			Description("Apply strikethrough formatting (true/false). Used by: format_cells, format_row.").
			Default("false").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfBackgroundColor).
			Description("Background color as hex (e.g. '#FF0000'). Used by: format_cells, format_row.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfForegroundColor).
			Description("Text color as hex (e.g. '#000000'). Used by: format_cells, format_row.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfNumberFormat).
			Description("Number format pattern (e.g. '#,##0.00'). Used by: format_cells, format_row.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfFrozenRows).
			Description("Number of rows to freeze. Used by: change_sheet_properties.").
			Default("0").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfFrozenColumns).
			Description("Number of columns to freeze. Used by: change_sheet_properties.").
			Default("0").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfHidden).
			Description("Whether the sheet is hidden (true/false). Used by: change_sheet_properties.").
			Default("false").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfSheetPosition).
			Description("Sheet tab position (0-based). Used by: change_sheet_properties.").
			Default("0").
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfValidationType).
			Description("Data validation type: ONE_OF_LIST, ONE_OF_RANGE, NUMBER_BETWEEN, TEXT_CONTAINS, etc. Used by: set_data_validation.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfValidationValues).
			Description("Comma-separated validation values (e.g. 'yes,no,maybe'). Used by: set_data_validation.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfConditionType).
			Description("Conditional format type: NUMBER_GREATER, TEXT_CONTAINS, CUSTOM_FORMULA, etc. Used by: create_conditional_formatting.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfConditionValue).
			Description("Condition value or formula. Used by: create_conditional_formatting.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfConditionBackgroundColor).
			Description("Background color as hex for conditional formatting (e.g. '#00FF00'). Used by: create_conditional_formatting.").
			Default("").
			Optional().
			Advanced()).
		Field(service.NewInterpolatedStringField(gsfIncludeHeaders).
			Description("Include header row in get operations (true/false). Used by: get_rows, get_data_range.").
			Default("true").
			Advanced()).
		Version("1.0.0")
}
