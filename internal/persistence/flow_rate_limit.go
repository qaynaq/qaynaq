package persistence

import (
	"time"

	"gorm.io/gorm"
)

type FlowRateLimit struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	FlowID    int64     `json:"flow_id" gorm:"not null"`
	RateLimitID int64     `json:"rate_limit_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`

	Flow    Flow    `json:"flow" gorm:"foreignKey:FlowID"`
	RateLimit RateLimit `json:"rate_limit" gorm:"foreignKey:RateLimitID"`
}

type FlowRateLimitRepository interface {
	Create(flowRateLimit FlowRateLimit) error
	FindByFlowID(flowID int64) ([]FlowRateLimit, error)
	DeleteByFlowID(flowID int64) error
}

type flowRateLimitRepository struct {
	db *gorm.DB
}

func NewFlowRateLimitRepository(db *gorm.DB) FlowRateLimitRepository {
	return &flowRateLimitRepository{db: db}
}

func (r *flowRateLimitRepository) Create(flowRateLimit FlowRateLimit) error {
	flowRateLimit.CreatedAt = time.Now()
	return r.db.Create(&flowRateLimit).Error
}

func (r *flowRateLimitRepository) FindByFlowID(flowID int64) ([]FlowRateLimit, error) {
	var flowRateLimits []FlowRateLimit
	err := r.db.Preload("RateLimit").Where("flow_id = ?", flowID).Find(&flowRateLimits).Error
	return flowRateLimits, err
}

func (r *flowRateLimitRepository) DeleteByFlowID(flowID int64) error {
	return r.db.Where("flow_id = ?", flowID).Delete(&FlowRateLimit{}).Error
}
