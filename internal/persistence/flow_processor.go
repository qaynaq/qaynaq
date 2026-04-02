package persistence

import (
	"time"

	"gorm.io/gorm"
)

type FlowProcessor struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	FlowID  int64     `json:"flow_id" gorm:"not null"`
	Component string    `json:"component" gorm:"not null"`
	Label     string    `json:"label"`
	Config    []byte    `json:"config"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	Flow Flow `json:"flow" gorm:"foreignKey:FlowID"`
}

type FlowProcessorRepository interface {
	Create(flowProcessor FlowProcessor) error
	DeleteByFlowID(flowID int64) error
}

type flowProcessorRepository struct {
	db *gorm.DB
}

func NewFlowProcessorRepository(db *gorm.DB) FlowProcessorRepository {
	return &flowProcessorRepository{db: db}
}

func (r *flowProcessorRepository) Create(flowProcessor FlowProcessor) error {
	flowProcessor.CreatedAt = time.Now()
	return r.db.Create(&flowProcessor).Error
}

func (r *flowProcessorRepository) DeleteByFlowID(flowID int64) error {
	return r.db.Where("flow_id = ?", flowID).Delete(&FlowProcessor{}).Error
}
