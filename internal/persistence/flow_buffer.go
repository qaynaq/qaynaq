package persistence

import (
	"time"

	"gorm.io/gorm"
)

type FlowBuffer struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	FlowID  int64     `json:"flow_id" gorm:"not null"`
	BufferID  int64     `json:"buffer_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`

	Flow Flow `json:"flow" gorm:"foreignKey:FlowID"`
	Buffer Buffer `json:"buffer" gorm:"foreignKey:BufferID"`
}

type FlowBufferRepository interface {
	Create(flowBuffer FlowBuffer) error
	FindByFlowID(flowID int64) ([]FlowBuffer, error)
	DeleteByFlowID(flowID int64) error
}

type flowBufferRepository struct {
	db *gorm.DB
}

func NewFlowBufferRepository(db *gorm.DB) FlowBufferRepository {
	return &flowBufferRepository{db: db}
}

func (r *flowBufferRepository) Create(flowBuffer FlowBuffer) error {
	flowBuffer.CreatedAt = time.Now()
	return r.db.Create(&flowBuffer).Error
}

func (r *flowBufferRepository) FindByFlowID(flowID int64) ([]FlowBuffer, error) {
	var flowBuffers []FlowBuffer
	err := r.db.Preload("Buffer").Where("flow_id = ?", flowID).Find(&flowBuffers).Error
	return flowBuffers, err
}

func (r *flowBufferRepository) DeleteByFlowID(flowID int64) error {
	return r.db.Where("flow_id = ?", flowID).Delete(&FlowBuffer{}).Error
}
