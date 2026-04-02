package persistence

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type EventType string

const (
	EventTypeProduce EventType = "PRODUCE"
	EventTypeConsume EventType = "CONSUME"
	EventTypeDelete  EventType = "DELETE"
	EventTypeError   EventType = "ERROR"
	EventTypeUnknown EventType = "UNKNOWN"
)

type Event struct {
	ID             int64           `json:"id" gorm:"primaryKey"`
	WorkerFlowID int64           `json:"worker_flow_id" gorm:"not null"`
	FlowID       int64           `json:"flow_id"`
	TraceID      string          `json:"trace_id"`
	Section        string          `json:"section" gorm:"not null"`
	ComponentLabel string          `json:"component_label"`
	Type           EventType       `json:"type" gorm:"not null"`
	Content        string          `json:"content" gorm:"not null"`
	Meta           json.RawMessage `json:"meta" gorm:"not null"`
	CreatedAt      time.Time       `json:"created_at"`

	Worker WorkerFlow `json:"worker" gorm:"foreignKey:WorkerFlowID"`
}

type EventRepository interface {
	AddEvent(event *Event) error
	ListEventsByWorkerFlow(workerID int64, preload bool) ([]*Event, error)
	ListEventsByFlowIDs(flowIDs []int64, limit, offset int, startTime, endTime time.Time) ([]*Event, int64, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) AddEvent(event *Event) error {
	event.CreatedAt = time.Now()
	return r.db.Create(event).Error
}

func (r *eventRepository) ListEventsByWorkerFlow(workerFlowID int64, preload bool) ([]*Event, error) {
	var events []*Event
	query := r.db.Where("worker_flow_id = ?", workerFlowID)
	if preload {
		query = query.Preload("WorkerFlow")
	}
	err := query.Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepository) ListEventsByFlowIDs(flowIDs []int64, limit, offset int, startTime, endTime time.Time) ([]*Event, int64, error) {
	var events []*Event
	var totalFlows int64

	baseWhere := "flow_id IN ? AND trace_id != ''"
	baseArgs := []any{flowIDs}

	if !startTime.IsZero() {
		baseWhere += " AND created_at >= ?"
		baseArgs = append(baseArgs, startTime)
	}
	if !endTime.IsZero() {
		baseWhere += " AND created_at <= ?"
		baseArgs = append(baseArgs, endTime)
	}

	// Count distinct trace_ids for pagination
	if err := r.db.Model(&Event{}).
		Where(baseWhere, baseArgs...).
		Distinct("trace_id").
		Count(&totalFlows).Error; err != nil {
		return nil, 0, err
	}

	// Find the N most recent trace_ids (by their latest event timestamp)
	type flowResult struct {
		TraceID string
	}
	var traceIDs []flowResult
	if err := r.db.Model(&Event{}).
		Select("trace_id, MAX(created_at) as max_ts").
		Where(baseWhere, baseArgs...).
		Group("trace_id").
		Order("max_ts DESC").
		Limit(limit).
		Offset(offset).
		Find(&traceIDs).Error; err != nil {
		return nil, 0, err
	}

	if len(traceIDs) == 0 {
		return events, totalFlows, nil
	}

	// Fetch ALL events for the selected trace_ids
	ids := make([]string, len(traceIDs))
	for i, f := range traceIDs {
		ids[i] = f.TraceID
	}

	if err := r.db.
		Where("trace_id IN ?", ids).
		Order("created_at DESC").
		Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, totalFlows, nil
}
