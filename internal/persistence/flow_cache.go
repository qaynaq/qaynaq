package persistence

import (
	"time"

	"gorm.io/gorm"
)

type FlowCache struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	FlowID  int64     `json:"flow_id" gorm:"not null"`
	CacheID   int64     `json:"cache_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	Flow Flow `json:"flow" gorm:"foreignKey:FlowID"`
	Cache  Cache  `json:"cache" gorm:"foreignKey:CacheID"`
}

type FlowCacheRepository interface {
	Create(flowCache FlowCache) error
	FindByFlowID(flowID int64) ([]FlowCache, error)
	DeleteByFlowID(flowID int64) error
}

type flowCacheRepository struct {
	db *gorm.DB
}

func NewFlowCacheRepository(db *gorm.DB) FlowCacheRepository {
	return &flowCacheRepository{db: db}
}

func (r *flowCacheRepository) Create(flowCache FlowCache) error {
	flowCache.CreatedAt = time.Now()
	return r.db.Create(&flowCache).Error
}

func (r *flowCacheRepository) FindByFlowID(flowID int64) ([]FlowCache, error) {
	var flowCaches []FlowCache
	err := r.db.Preload("Cache").Where("flow_id = ?", flowID).Find(&flowCaches).Error
	return flowCaches, err
}

func (r *flowCacheRepository) DeleteByFlowID(flowID int64) error {
	return r.db.Where("flow_id = ?", flowID).Delete(&FlowCache{}).Error
}
