package cdc_mysql

import (
	"context"
	"errors"
	"sync"

	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	err := service.RegisterBatchInput(
		"cdc_mysql", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.BatchInput, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

var ErrBinlogNotAvailable = errors.New("binlog position no longer available")

type Input struct {
	connManager *connectionManager
	batchReader *batchReader
	posCache    *positionCache
	schemaCache *schemaCache
	closeChan   chan struct{}
	closeOnce   sync.Once
	logger      *logger.Logger
}

func (m *Input) Connect(ctx context.Context) error {
	m.posCache.Start(ctx)

	return m.connManager.connect(ctx)
}

func (m *Input) ReadBatch(ctx context.Context) (service.MessageBatch, service.AckFunc, error) {
	return m.batchReader.readBatch(ctx)
}

func (m *Input) Close(ctx context.Context) error {
	m.closeOnce.Do(func() {
		close(m.closeChan)

		m.posCache.Stop(ctx)

		m.connManager.close()

		if m.schemaCache != nil {
			err := m.schemaCache.close()
			if err != nil {
				m.logger.Error("Error closing schema cache", "error", err)
			}
		}

		m.logger.Info("MySQL CDC input closed")
	})
	return nil
}
