package cdc_mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qaynaq/qaynaq/internal/logger"
)

type connectionManager struct {
	host     string
	port     int
	user     string
	password string
	serverID uint32
	flavor   string
	posCache *positionCache

	positionMode         string
	retryInitialInterval time.Duration
	retryMaxInterval     time.Duration
	retryMultiplier      float64
	currentRetryInterval time.Duration
	retryMutex           sync.Mutex

	syncer    *replication.BinlogSyncer
	streamer  *replication.BinlogStreamer
	connected bool
	connMutex sync.RWMutex

	logger *logger.Logger
}

func newConnectionManager(
	host string,
	port int,
	user string,
	password string,
	serverID uint32,
	flavor string,
	posCache *positionCache,
	positionMode string,
	retryInitialInterval time.Duration,
	retryMaxInterval time.Duration,
	retryMultiplier float64,
	log *logger.Logger,
) *connectionManager {
	return &connectionManager{
		host:                 host,
		port:                 port,
		user:                 user,
		password:             password,
		serverID:             serverID,
		flavor:               flavor,
		posCache:             posCache,
		positionMode:         positionMode,
		retryInitialInterval: retryInitialInterval,
		retryMaxInterval:     retryMaxInterval,
		retryMultiplier:      retryMultiplier,
		currentRetryInterval: retryInitialInterval,
		logger:               log,
	}
}

func (cm *connectionManager) connect(ctx context.Context) error {
	retryCount := 0
	for {
		err := cm.connectWithRetry(ctx, false)
		if err == nil {
			cm.resetRetryInterval()
			if retryCount > 0 {
				cm.logger.Info("Successfully connected after retries", "retry_count", retryCount)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		retryInterval := cm.getNextRetryInterval()
		retryCount++

		cm.logger.Warn("Failed to connect to MySQL, will retry with exponential backoff",
			"error", err,
			"retry_count", retryCount,
			"retry_interval", retryInterval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
		}
	}
}

func (cm *connectionManager) connectWithRetry(ctx context.Context, isRetry bool) error {
	cm.logger.Debug("Connect method called")
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	if cm.connected {
		cm.logger.Debug("Already connected, returning")
		return nil
	}

	cfg := replication.BinlogSyncerConfig{
		ServerID:             cm.serverID,
		Flavor:               cm.flavor,
		Host:                 cm.host,
		Port:                 uint16(cm.port),
		User:                 cm.user,
		Password:             cm.password,
		MaxReconnectAttempts: 0,
		DisableRetrySync:     true,
	}

	syncer := replication.NewBinlogSyncer(cfg)
	cm.syncer = syncer

	var streamer *replication.BinlogStreamer
	var err error

	pos, loadErr := cm.posCache.loadPosition(ctx)

	if cm.positionMode == "gtid" {
		var gtidSet mysql.GTIDSet
		if loadErr == nil && pos.GTIDSet != "" && pos.Mode == "gtid" {
			storedGtidSet, parseErr := mysql.ParseGTIDSet(cm.flavor, pos.GTIDSet)
			if parseErr != nil {
				cm.logger.Error("Failed to parse stored GTID set, will query earliest available", "gtid_set", pos.GTIDSet, "error", parseErr)
				earliestGTID, gtidErr := cm.getEarliestAvailableGTIDSet(ctx)
				if gtidErr != nil {
					return fmt.Errorf("failed to get earliest available GTID: %w", gtidErr)
				}
				gtidSet = earliestGTID
			} else {
				cm.logger.Info("Resuming from stored GTID set", "gtid_set", pos.GTIDSet)
				gtidSet = storedGtidSet
			}
		} else {
			if loadErr != nil {
				cm.logger.Debug("No valid position file found", "error", loadErr)
			}
			cm.logger.Info("No position file found, querying earliest available GTID from MySQL")
			earliestGTID, gtidErr := cm.getEarliestAvailableGTIDSet(ctx)
			if gtidErr != nil {
				return fmt.Errorf("failed to get earliest available GTID: %w", gtidErr)
			}
			gtidSet = earliestGTID
		}
		cm.logger.Info("Starting GTID sync with final GTID set", "final_gtid_set", gtidSet.String())
		streamer, err = syncer.StartSyncGTID(gtidSet)
	} else {
		var position mysql.Position

		if loadErr == nil && pos.BinlogFile != "" && pos.Mode == "file" {
			position = mysql.Position{Name: pos.BinlogFile, Pos: pos.BinlogPos}
			cm.logger.Info("Resuming from file position", "binlog_file", pos.BinlogFile, "binlog_pos", pos.BinlogPos)
		} else {
			position = mysql.Position{}
			cm.logger.Info("Starting from current binlog position (file mode)")
		}

		streamer, err = syncer.StartSync(position)
	}

	if err != nil {
		syncer.Close()

		if isBinlogNotAvailableError(err) {
			if isRetry {
				return fmt.Errorf("failed to start sync after position recovery: %w", err)
			}

			cm.logger.Warn("Binlog position/GTID no longer available on master, purging position and retrying from current position", "error", err)

			if purgeErr := cm.posCache.purgePosition(ctx); purgeErr != nil {
				cm.logger.Error("Failed to purge position from cache", "error", purgeErr)
				return fmt.Errorf("failed to start sync: %w, and failed to purge position: %v", err, purgeErr)
			}

			cm.connMutex.Unlock()
			cm.logger.Info("Retrying connection from current position after purging corrupted position")
			retryErr := cm.connectWithRetry(ctx, true)
			cm.connMutex.Lock()
			return retryErr
		}

		return fmt.Errorf("failed to start sync: %w", err)
	}

	cm.streamer = streamer
	cm.connected = true
	cm.logger.Info("Connected to MySQL binlog")
	return nil
}

func (cm *connectionManager) getEarliestAvailableGTIDSet(ctx context.Context) (mysql.GTIDSet, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", cm.user, cm.password, cm.host, cm.port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	defer db.Close() //nolint:errcheck

	var variableName, purgedGTIDs string
	err = db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'gtid_purged'").Scan(&variableName, &purgedGTIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query gtid_purged: %w", err)
	}

	cm.logger.Info("Found purged GTIDs from MySQL", "purged_gtids", purgedGTIDs)

	if purgedGTIDs == "" {
		return mysql.ParseGTIDSet(cm.flavor, "")
	}

	purgedSet, err := mysql.ParseGTIDSet(cm.flavor, purgedGTIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse purged GTID set: %w", err)
	}

	cm.logger.Info("Starting from purged GTID set to skip unavailable transactions", "starting_gtid_set", purgedSet.String())
	return purgedSet, nil
}

func (cm *connectionManager) close() {
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	if cm.syncer != nil {
		cm.logger.Debug("Closing syncer in Close method")
		cm.syncer.Close()
		cm.syncer = nil
	}

	cm.connected = false
}

func (cm *connectionManager) markDisconnected() {
	cm.connMutex.Lock()
	defer cm.connMutex.Unlock()

	if cm.syncer != nil {
		cm.logger.Debug("Closing syncer due to ErrNeedSyncAgain")
		cm.syncer.Close()
		cm.syncer = nil
	}
	cm.streamer = nil
	cm.connected = false
}

func (cm *connectionManager) getConnectionState() (connected bool, streamer *replication.BinlogStreamer) {
	cm.connMutex.RLock()
	defer cm.connMutex.RUnlock()
	return cm.connected, cm.streamer
}

func (cm *connectionManager) getNextRetryInterval() time.Duration {
	cm.retryMutex.Lock()
	defer cm.retryMutex.Unlock()

	current := cm.currentRetryInterval
	next := min(time.Duration(float64(current)*cm.retryMultiplier), cm.retryMaxInterval)

	cm.currentRetryInterval = next
	return current
}

func (cm *connectionManager) resetRetryInterval() {
	cm.retryMutex.Lock()
	defer cm.retryMutex.Unlock()
	cm.currentRetryInterval = cm.retryInitialInterval
}

func isBinlogNotAvailableError(err error) bool {
	var mysqlErr *mysql.MyError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Code == mysql.ER_MASTER_FATAL_ERROR_READING_BINLOG
	}
	return false
}
