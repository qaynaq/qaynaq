package cdc_mysql

type eventType string

const (
	eventTypeInsert eventType = "insert"
	eventTypeUpdate eventType = "update"
	eventTypeDelete eventType = "delete"
)

type binlogEvent struct {
	Database string         `json:"database"`
	Table    string         `json:"table"`
	Type     eventType      `json:"type"`
	Ts       int64          `json:"ts"`
	ServerID string         `json:"server_id"`
	New      map[string]any `json:"new,omitempty"`
	Old      map[string]any `json:"old,omitempty"`
	GTID     string         `json:"gtid,omitempty"`
}

func (be *binlogEvent) ToMap() map[string]any {
	return map[string]any{
		"database": be.Database,
		"table":    be.Table,
		"type":     string(be.Type),
		"ts":       be.Ts,
		"server_id": be.ServerID,
		"new":      be.New,
		"old":      be.Old,
		"gtid":     be.GTID,
	}
}

type binlogPosition struct {
	GTIDSet    string `json:"gtid_set,omitempty"`
	BinlogFile string `json:"binlog_file,omitempty"`
	BinlogPos  uint32 `json:"binlog_pos,omitempty"`
	Mode       string `json:"mode"`
}
