package cdc_mysql

import (
	"encoding/base64"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/replication"
)

type eventProcessor struct {
	serverID      string
	includeTables []string
	excludeTables []string
	pendingGTID   string
	gtidMutex     sync.Mutex
	schemaCache   *schemaCache

	currentFile string
	currentPos  uint32
	fileMutex   sync.Mutex
}

func newEventProcessor(serverID string, includeTables, excludeTables []string, schemaCache *schemaCache) *eventProcessor {
	return &eventProcessor{
		serverID:      serverID,
		includeTables: includeTables,
		excludeTables: excludeTables,
		schemaCache:   schemaCache,
	}
}

func (ep *eventProcessor) shouldIncludeTable(database, table string) bool {
	fullTableName := fmt.Sprintf("%s.%s", database, table)

	if slices.Contains(ep.excludeTables, fullTableName) {
		return false
	}

	if len(ep.includeTables) > 0 {
		return slices.Contains(ep.includeTables, fullTableName)
	}

	return true
}

func (ep *eventProcessor) processEvent(ev *replication.BinlogEvent) []binlogEvent {
	var events []binlogEvent

	ep.fileMutex.Lock()
	ep.currentPos = ev.Header.LogPos
	ep.fileMutex.Unlock()

	if rotateEv, ok := ev.Event.(*replication.RotateEvent); ok {
		ep.fileMutex.Lock()
		ep.currentFile = string(rotateEv.NextLogName)
		ep.currentPos = uint32(rotateEv.Position)
		ep.fileMutex.Unlock()
		return events
	}

	if gtidEv, ok := ev.Event.(*replication.GTIDEvent); ok {
		sidStr := fmt.Sprintf("%x-%x-%x-%x-%x",
			gtidEv.SID[0:4], gtidEv.SID[4:6], gtidEv.SID[6:8], gtidEv.SID[8:10], gtidEv.SID[10:16])
		gtid := fmt.Sprintf("%s:%d", sidStr, gtidEv.GNO)
		ep.gtidMutex.Lock()
		ep.pendingGTID = gtid
		ep.gtidMutex.Unlock()
		return events
	}

	switch e := ev.Event.(type) {
	case *replication.RowsEvent:
		if !ep.shouldIncludeTable(string(e.Table.Schema), string(e.Table.Table)) {
			return events
		}

		event := binlogEvent{
			Database: string(e.Table.Schema),
			Table:    string(e.Table.Table),
			Ts:       time.Now().Unix(),
			ServerID: ep.serverID,
		}

		ep.gtidMutex.Lock()
		event.GTID = ep.pendingGTID
		ep.gtidMutex.Unlock()

		switch ev.Header.EventType {
		case replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
			event.Type = eventTypeInsert
			for _, row := range e.Rows {
				event.New = ep.rowToMap(e.Table, row)
				events = append(events, event)
			}
		case replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
			event.Type = eventTypeUpdate
			for i := 0; i < len(e.Rows); i += 2 {
				if i+1 < len(e.Rows) {
					event.Old = ep.rowToMap(e.Table, e.Rows[i])
					event.New = ep.rowToMap(e.Table, e.Rows[i+1])
					events = append(events, event)
				}
			}
		case replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
			event.Type = eventTypeDelete
			for _, row := range e.Rows {
				event.Old = ep.rowToMap(e.Table, row)
				events = append(events, event)
			}
		}
	}

	return events
}

func (ep *eventProcessor) rowToMap(table *replication.TableMapEvent, row []any) map[string]any {
	result := make(map[string]any)

	var columnNames []string
	if ep.schemaCache != nil {
		if names, err := ep.schemaCache.getColumnNames(string(table.Schema), string(table.Table)); err == nil {
			columnNames = names
		}
	}

	if len(columnNames) == 0 && len(table.ColumnName) > 0 {
		for _, colName := range table.ColumnName {
			columnNames = append(columnNames, string(colName))
		}
	}

	for i, value := range row {
		var columnName string
		if i < len(columnNames) && columnNames[i] != "" {
			columnName = columnNames[i]
		} else {
			columnName = fmt.Sprintf("col_%d", i)
		}

		decodedValue := ep.decodeValue(value, table, i)
		result[columnName] = decodedValue
	}

	return result
}

func (ep *eventProcessor) decodeValue(value any, table *replication.TableMapEvent, columnIndex int) any {
	if bytes, ok := value.([]byte); ok {
		if columnIndex < len(table.ColumnType) {
			columnType := table.ColumnType[columnIndex]

			if ep.isStringType(columnType) {
				if str := string(bytes); len(str) > 0 {
					return str
				}
			}
		}

		if len(bytes) > 0 {
			if decoded, err := base64.StdEncoding.DecodeString(string(bytes)); err == nil {
				if str := string(decoded); len(str) > 0 {
					return str
				}
			}
			return string(bytes)
		}
	}

	return value
}

func (ep *eventProcessor) isStringType(columnType byte) bool {
	switch columnType {
	case 0x0f, // MYSQL_TYPE_VARCHAR
		0xf6, // MYSQL_TYPE_NEWDECIMAL
		0xf7, // MYSQL_TYPE_ENUM
		0xf8, // MYSQL_TYPE_SET
		0xf9, // MYSQL_TYPE_TINY_BLOB
		0xfa, // MYSQL_TYPE_MEDIUM_BLOB
		0xfb, // MYSQL_TYPE_LONG_BLOB
		0xfc, // MYSQL_TYPE_BLOB
		0xfd, // MYSQL_TYPE_VAR_STRING
		0xfe, // MYSQL_TYPE_STRING
		0xff: // MYSQL_TYPE_GEOMETRY
		return true
	}
	return false
}

func (ep *eventProcessor) getCurrentGTID() string {
	ep.gtidMutex.Lock()
	defer ep.gtidMutex.Unlock()
	return ep.pendingGTID
}

func (ep *eventProcessor) getCurrentFilePosition() (string, uint32) {
	ep.fileMutex.Lock()
	defer ep.fileMutex.Unlock()
	return ep.currentFile, ep.currentPos
}
