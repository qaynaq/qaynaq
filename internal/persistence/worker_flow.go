package persistence

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type WorkerFlowStatus string

const (
	WorkerFlowStatusWaiting   WorkerFlowStatus = "waiting"
	WorkerFlowStatusRunning   WorkerFlowStatus = "running"
	WorkerFlowStatusStopped   WorkerFlowStatus = "stopped"
	WorkerFlowStatusCompleted WorkerFlowStatus = "completed"
	WorkerFlowStatusFailed    WorkerFlowStatus = "failed"
)

const (
	FlowLeaseInterval = 20 * time.Second
)

type WorkerFlow struct {
	ID              int64              `json:"id" gorm:"primaryKey"`
	WorkerID        string             `json:"worker_id" gorm:"not null"`
	FlowID        int64              `json:"flow_id" gorm:"not null"`
	InputEvents     uint64             `json:"input_events" gorm:"not null" default:"0"`
	ProcessorErrors uint64             `json:"processor_errors" gorm:"not null" default:"0"`
	OutputEvents    uint64             `json:"output_events" gorm:"not null" default:"0"`
	Status          WorkerFlowStatus `json:"status" gorm:"not null"`
	LeaseExpiresAt  time.Time          `json:"lease_expires_at"`
	CreatedAt       time.Time          `json:"created_at" gorm:"not null"`
	UpdatedAt       time.Time          `json:"updated_at"`

	Worker Worker `json:"worker" gorm:"foreignKey:WorkerID"`
	Flow Flow `json:"flow" gorm:"foreignKey:FlowID"`
}

type WorkerFlowRepository interface {
	Queue(workerID string, flowID int64) (WorkerFlow, error)
	FindByID(id int64) (*WorkerFlow, error)
	FindByWorkerIDAndFlowID(workerID string, flowID int64) (*WorkerFlow, error)
	UpdateStatus(id int64, status WorkerFlowStatus) error
	UpdateMetrics(id int64, inputEvents, processorErrors, outputEvents uint64) error
	UpdateLeaseExpiry(id int64, expiresAt time.Time) error
	FindRunningWithExpiredLeases() ([]WorkerFlow, error)
	StopAllRunningAndWaitingByWorkerID(workerID string) error
	ListAllByWorkerID(workerID string) ([]WorkerFlow, error)
	ListAllByFlowID(flowID int64) ([]WorkerFlow, error)
	ListAllByStatuses(statuses ...WorkerFlowStatus) ([]WorkerFlow, error)
}

type workerFlowRepository struct {
	db *gorm.DB
}

func NewWorkerFlowRepository(db *gorm.DB) WorkerFlowRepository {
	return &workerFlowRepository{db: db}
}

func (r *workerFlowRepository) Queue(workerID string, flowID int64) (WorkerFlow, error) {
	workerFlow := WorkerFlow{
		WorkerID:       workerID,
		FlowID:       flowID,
		Status:         WorkerFlowStatusWaiting,
		LeaseExpiresAt: time.Now().Add(FlowLeaseInterval),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := r.db.Create(&workerFlow).Error

	return workerFlow, err
}

func (r *workerFlowRepository) UpdateStatus(id int64, status WorkerFlowStatus) error {
	return r.db.
		Model(&WorkerFlow{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()}).
		Error
}

func (r *workerFlowRepository) UpdateMetrics(id int64, inputEvents, processorErrors, outputEvents uint64) error {
	return r.db.
		Model(&WorkerFlow{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"input_events":     inputEvents,
			"processor_errors": processorErrors,
			"output_events":    outputEvents,
			"updated_at":       time.Now(),
		}).
		Error
}

func (r *workerFlowRepository) FindByID(id int64) (*WorkerFlow, error) {
	var workerFlow = &WorkerFlow{
		ID: id,
	}
	err := r.db.
		Preload("Worker").
		Preload("Flow").
		First(workerFlow).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return workerFlow, nil
}

func (r *workerFlowRepository) StopAllRunningAndWaitingByWorkerID(workerID string) error {
	return r.db.
		Model(&WorkerFlow{}).
		Where("worker_id = ?", workerID).
		Where("status IN ?", []WorkerFlowStatus{WorkerFlowStatusRunning, WorkerFlowStatusWaiting}).
		Updates(map[string]any{"status": WorkerFlowStatusStopped, "updated_at": time.Now()}).
		Error
}

func (r *workerFlowRepository) FindByWorkerIDAndFlowID(workerID string, flowID int64) (*WorkerFlow, error) {
	var workerFlow WorkerFlow
	err := r.db.
		Where("worker_id = ? AND flow_id = ?", workerID, flowID).
		First(&workerFlow).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &workerFlow, nil
}

func (r *workerFlowRepository) UpdateLeaseExpiry(id int64, expiresAt time.Time) error {
	return r.db.
		Model(&WorkerFlow{}).
		Where("id = ?", id).
		Update("lease_expires_at", expiresAt).
		Error
}

func (r *workerFlowRepository) FindRunningWithExpiredLeases() ([]WorkerFlow, error) {
	var flows []WorkerFlow
	err := r.db.
		Where("status = ?", WorkerFlowStatusRunning).
		Where("lease_expires_at != ?", time.Time{}).
		Where("lease_expires_at < ?", time.Now()).
		Find(&flows).
		Error
	if err != nil {
		return nil, err
	}
	return flows, nil
}

func (r *workerFlowRepository) ListAllByWorkerID(workerID string) ([]WorkerFlow, error) {
	var workerFlows []WorkerFlow
	err := r.db.
		Preload("Worker").
		Preload("Flow").
		Preload("Output").
		Where("worker_id = ?", workerID).
		Find(&workerFlows).
		Error

	return workerFlows, err
}

func (r *workerFlowRepository) ListAllByFlowID(flowID int64) ([]WorkerFlow, error) {
	var workerFlows []WorkerFlow
	err := r.db.
		Preload("Worker").
		Preload("Flow").
		Where("flow_id = ?", flowID).
		Find(&workerFlows).
		Error

	return workerFlows, err
}

func (r *workerFlowRepository) ListAllByStatuses(statuses ...WorkerFlowStatus) ([]WorkerFlow, error) {
	var workerFlows []WorkerFlow
	err := r.db.
		Preload("Flow").
		Preload("Worker").
		Where("status IN ?", statuses).
		Find(&workerFlows).
		Error

	return workerFlows, err
}
