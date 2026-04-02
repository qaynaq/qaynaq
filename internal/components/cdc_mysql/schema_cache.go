package cdc_mysql

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
)

type tableSchema struct {
	columns   []string
	updatedAt time.Time
}

type schemaCache struct {
	conn     *client.Conn
	cache    map[string]*tableSchema
	cacheTTL time.Duration
	mutex    sync.RWMutex
}

func newSchemaCache(host string, port int, user, password string, cacheTTL time.Duration) (*schemaCache, error) {
	conn, err := client.Connect(fmt.Sprintf("%s:%d", host, port), user, password, "information_schema")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to ping MySQL: %w", err)
	}

	return &schemaCache{
		conn:     conn,
		cache:    make(map[string]*tableSchema),
		cacheTTL: cacheTTL,
	}, nil
}

func (sc *schemaCache) getColumnNames(database, table string) ([]string, error) {
	tableKey := fmt.Sprintf("%s.%s", database, table)

	sc.mutex.RLock()
	schema, exists := sc.cache[tableKey]
	sc.mutex.RUnlock()

	if exists && time.Since(schema.updatedAt) < sc.cacheTTL {
		return schema.columns, nil
	}

	columns, err := sc.fetchColumnNames(database, table)
	if err != nil {
		return nil, err
	}

	sc.mutex.Lock()
	sc.cache[tableKey] = &tableSchema{
		columns:   columns,
		updatedAt: time.Now(),
	}
	sc.mutex.Unlock()

	return columns, nil
}

func (sc *schemaCache) fetchColumnNames(database, table string) ([]string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	result, err := sc.conn.Execute(query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query column names: %w", err)
	}

	var columns []string
	for i := 0; i < result.RowNumber(); i++ {
		columnName, err := result.GetString(i, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to get column name at row %d: %w", i, err)
		}
		columns = append(columns, columnName)
	}

	return columns, nil
}

func (sc *schemaCache) close() error {
	if sc.conn != nil {
		return sc.conn.Close()
	}
	return nil
}
