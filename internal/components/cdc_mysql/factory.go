package cdc_mysql

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/qaynaq/qaynaq/internal/logger"
	"github.com/warpstreamlabs/bento/public/service"
)

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*Input, error) {
	host, err := conf.FieldString(cmfHost)
	if err != nil {
		return nil, err
	}

	port, err := conf.FieldInt(cmfPort)
	if err != nil {
		return nil, err
	}

	user, err := conf.FieldString(cmfUser)
	if err != nil {
		return nil, err
	}

	password, err := conf.FieldString(cmfPassword)
	if err != nil {
		return nil, err
	}

	serverIDStr, err := conf.FieldString(cmfServerID)
	if err != nil {
		return nil, err
	}

	serverID, err := parseServerID(serverIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid server_id value: %w", err)
	}

	positionCacheName, err := conf.FieldString(cmfPositionCache)
	if err != nil {
		return nil, err
	}

	positionCacheKey, err := conf.FieldString(cmfPositionCacheKey)
	if err != nil {
		return nil, err
	}

	if positionCacheName == "" {
		return nil, fmt.Errorf("position_cache is required")
	}

	if positionCacheKey == "" {
		return nil, fmt.Errorf("position_cache_key is required")
	}

	var includeTables []string
	if conf.Contains(cmfIncludeTables) {
		includeTables, err = conf.FieldStringList(cmfIncludeTables)
		if err != nil {
			return nil, err
		}
	}

	var excludeTables []string
	if conf.Contains(cmfExcludeTables) {
		excludeTables, err = conf.FieldStringList(cmfExcludeTables)
		if err != nil {
			return nil, err
		}
	}

	useSchemaCache := false
	if conf.Contains(cmfUseSchemaCache) {
		useSchemaCache, err = conf.FieldBool(cmfUseSchemaCache)
		if err != nil {
			return nil, err
		}
	}

	schemaCacheTTLStr := "5m"
	if conf.Contains(cmfSchemaCacheTTL) {
		schemaCacheTTLStr, err = conf.FieldString(cmfSchemaCacheTTL)
		if err != nil {
			return nil, err
		}
	}

	schemaCacheTTL, err := time.ParseDuration(schemaCacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid schema_cache_ttl duration: %w", err)
	}

	var schemaCache *schemaCache
	if useSchemaCache {
		schemaCache, err = newSchemaCache(host, port, user, password, schemaCacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create schema cache: %w", err)
		}
	}

	positionMode := "gtid"
	if conf.Contains(cmfPositionMode) {
		positionMode, err = conf.FieldString(cmfPositionMode)
		if err != nil {
			return nil, err
		}
	}

	if positionMode != "gtid" && positionMode != "file" {
		return nil, fmt.Errorf("invalid position_mode: %s (must be 'gtid' or 'file')", positionMode)
	}

	flavor := "mysql"
	if conf.Contains(cmfFlavor) {
		flavor, err = conf.FieldString(cmfFlavor)
		if err != nil {
			return nil, err
		}
	}

	if flavor != "mysql" && flavor != "mariadb" {
		return nil, fmt.Errorf("invalid flavor: %s (must be 'mysql' or 'mariadb')", flavor)
	}

	maxBatchSize, err := conf.FieldInt(cmfMaxBatchSize)
	if err != nil {
		return nil, err
	}

	maxPendingCheckpoints, err := conf.FieldInt(cmfMaxPendingCheckpoints)
	if err != nil {
		return nil, err
	}

	retryInitialIntervalStr, err := conf.FieldString(cmfRetryInitialInterval)
	if err != nil {
		return nil, err
	}

	retryInitialInterval, err := time.ParseDuration(retryInitialIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid retry_initial_interval duration: %w", err)
	}

	retryMaxIntervalStr, err := conf.FieldString(cmfRetryMaxInterval)
	if err != nil {
		return nil, err
	}

	retryMaxInterval, err := time.ParseDuration(retryMaxIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid retry_max_interval duration: %w", err)
	}

	retryMultiplier, err := conf.FieldFloat(cmfRetryMultiplier)
	if err != nil {
		return nil, err
	}

	cacheSaveIntervalStr, err := conf.FieldString(cmfCacheSaveInterval)
	if err != nil {
		return nil, err
	}

	cacheSaveInterval, err := time.ParseDuration(cacheSaveIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid cache_save_interval duration: %w", err)
	}

	log := logger.NewFromBento(mgr.Logger(), nil)

	posCache := newPositionCache(mgr, positionCacheName, positionCacheKey, cacheSaveInterval, log)

	connManager := newConnectionManager(
		host,
		port,
		user,
		password,
		serverID,
		flavor,
		posCache,
		positionMode,
		retryInitialInterval,
		retryMaxInterval,
		retryMultiplier,
		log,
	)

	eventProc := newEventProcessor(serverIDStr, includeTables, excludeTables, schemaCache)
	checkpointTrack := newCheckpointTracker(posCache, positionMode, maxPendingCheckpoints, log)

	batchRead := newBatchReader(
		maxBatchSize,
		positionMode,
		posCache,
		connManager,
		eventProc,
		checkpointTrack,
		log,
	)

	return &Input{
		connManager: connManager,
		batchReader: batchRead,
		posCache:    posCache,
		schemaCache: schemaCache,
		closeChan:   make(chan struct{}),
		logger:      log,
	}, nil
}

func parseServerID(serverID string) (uint32, error) {
	if serverID == "" {
		return 0, fmt.Errorf("server_id cannot be empty")
	}

	if val, err := strconv.ParseUint(serverID, 10, 32); err == nil {
		return uint32(val), nil
	}

	h := fnv.New32a()
	h.Write([]byte(serverID))
	return h.Sum32(), nil
}
