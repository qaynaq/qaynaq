package persistence

import (
	"errors"
	"time"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type FlowStatus string
type FlowSection string

const (
	FlowStatusActive    FlowStatus = "active"
	FlowStatusCompleted FlowStatus = "completed"
	FlowStatusPaused    FlowStatus = "paused"
	FlowStatusFailed    FlowStatus = "failed"

	FlowSectionInput    FlowSection = "input"
	FlowSectionPipeline FlowSection = "pipeline"
	FlowSectionOutput   FlowSection = "output"
)

type Flow struct {
	ID              int64      `json:"id" gorm:"primaryKey"`
	ParentID        *int64     `json:"parent_id"`
	Name            string     `json:"name" gorm:"not null"`
	InputLabel      string     `json:"input_label"`
	InputComponent  string     `json:"input_component" gorm:"not null"`
	InputConfig     []byte     `json:"input_config" gorm:"not null"`
	OutputLabel     string     `json:"output_label"`
	OutputComponent string     `json:"output_component" gorm:"not null"`
	OutputConfig    []byte     `json:"output_config" gorm:"not null"`
	BufferID        *int64     `json:"buffer_id"`
	IsCurrent       bool       `json:"is_current" gorm:"default:true"`
	IsReady         bool       `json:"is_ready" gorm:"default:false"`
	BuilderState    []byte     `json:"builder_state"`
	ManagedBy       *string    `json:"managed_by"`
	Status          FlowStatus `json:"status" gorm:"not null"`
	CreatedAt       time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt       *time.Time `json:"updated_at"`

	ParentFlow *Flow           `json:"parent_flow" gorm:"foreignKey:ParentID"`
	Buffer     *Buffer         `json:"buffer" gorm:"foreignKey:BufferID"`
	Processors []FlowProcessor `json:"processors" gorm:"foreignKey:FlowID;references:ID"`
	Caches     []FlowCache     `json:"caches" gorm:"foreignKey:FlowID;references:ID"`
}

func (s *Flow) ToProto() *pb.Flow {
	var updatedAt *timestamppb.Timestamp
	if s.UpdatedAt != nil {
		updatedAt = timestamppb.New(*s.UpdatedAt)
	}

	result := &pb.Flow{
		Id:              s.ID,
		ParentId:        s.ParentID,
		Name:            s.Name,
		InputLabel:      s.InputLabel,
		InputComponent:  s.InputComponent,
		InputConfig:     string(s.InputConfig),
		Processors:      make([]*pb.Flow_Processor, len(s.Processors)),
		OutputLabel:     s.OutputLabel,
		OutputComponent: s.OutputComponent,
		OutputConfig:    string(s.OutputConfig),
		BufferId:        s.BufferID,
		IsCurrent:       s.IsCurrent,
		IsReady:         s.IsReady,
		BuilderState:    string(s.BuilderState),
		ManagedBy:       s.ManagedBy,
		Status:          string(s.Status),
		CreatedAt:       timestamppb.New(s.CreatedAt),
		UpdatedAt:       updatedAt,
		IsHttpServer:    s.InputComponent == "http_server",
		IsMcpTool:       s.InputComponent == "mcp_tool",
	}

	for i, processor := range s.Processors {
		result.Processors[i] = &pb.Flow_Processor{
			Label:     processor.Label,
			Component: processor.Component,
			Config:    string(processor.Config),
		}
	}

	return result
}

func (s *Flow) FromProto(p *pb.Flow) {
	updatedAt := p.GetUpdatedAt().AsTime()

	s.ID = p.Id
	s.ParentID = p.ParentId
	s.Name = p.Name
	s.InputLabel = p.InputLabel
	s.InputComponent = p.InputComponent
	s.InputConfig = []byte(p.InputConfig)
	s.OutputLabel = p.OutputLabel
	s.OutputComponent = p.OutputComponent
	s.OutputConfig = []byte(p.OutputConfig)
	s.BufferID = p.BufferId
	s.IsCurrent = p.GetIsCurrent()
	s.IsReady = p.GetIsReady()
	s.BuilderState = []byte(p.GetBuilderState())
	s.ManagedBy = p.ManagedBy
	s.Status = FlowStatus(p.GetStatus())
	s.CreatedAt = p.CreatedAt.AsTime()
	s.UpdatedAt = &updatedAt
}

type FlowRepository interface {
	Create(flow *Flow) error
	Update(flow *Flow) error
	FindByID(id int64) (*Flow, error)
	FindByNameAndManagedBy(name string, managedBy string) (*Flow, error)
	UpdateStatus(id int64, status FlowStatus) error
	Delete(id int64) error
	ListAllByStatuses(...FlowStatus) ([]Flow, error)
	ListAllActiveAndNonAssigned() ([]Flow, error)
	ListAllVersionsByParentID(parentID int64) ([]Flow, error)
}

type flowRepository struct {
	db *gorm.DB
}

func NewFlowRepository(db *gorm.DB) FlowRepository {
	return &flowRepository{db: db}
}

func (r *flowRepository) FindByNameAndManagedBy(name string, managedBy string) (*Flow, error) {
	var flow Flow
	err := r.db.
		Where("name = ? AND managed_by = ? AND is_current = ?", name, managedBy, true).
		First(&flow).
		Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

func (r *flowRepository) Create(flow *Flow) error {
	flow.CreatedAt = time.Now()
	return r.db.Create(flow).Error
}

func (r *flowRepository) Update(flow *Flow) error {
	var err error

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	err = tx.
		Model(&Flow{}).
		Where("id = ?", flow.ID).
		Update("is_current", false).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var id = flow.ID
	flow.CreatedAt = time.Now()
	if flow.ParentID == nil {
		flow.ParentID = &id
	}
	flow.ID = 0

	if err = tx.Create(flow).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *flowRepository) FindByID(id int64) (*Flow, error) {
	var flow = &Flow{
		ID: id,
	}
	err := r.db.
		Preload("Processors").
		Preload("Caches").
		Preload("Buffer").
		Preload("ParentFlow").
		First(flow).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return flow, nil
}

func (r *flowRepository) UpdateStatus(id int64, status FlowStatus) error {
	return r.db.
		Model(&Flow{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()}).
		Error
}

func (r *flowRepository) Delete(id int64) error {
	return r.db.Delete(&Flow{}, id).Error
}

func (r *flowRepository) ListAllByStatuses(statuses ...FlowStatus) ([]Flow, error) {
	var flows []Flow
	db := r.db.
		Preload("Processors").
		Preload("Caches").
		Preload("Buffer")

	if len(statuses) > 0 {
		db = db.Where("is_current = true AND status IN ?", statuses)
	}

	if err := db.Find(&flows).Error; err != nil {
		return nil, err
	}
	return flows, nil
}

func (r *flowRepository) ListAllActiveAndNonAssigned() ([]Flow, error) {
	var flows []Flow

	recentCutoff := time.Now().Add(-30 * time.Second)

	err := r.db.
		Preload("Processors").
		Preload("Caches").
		Preload("Buffer").
		Where("is_current = true AND is_ready = true AND status = ?", FlowStatusActive).
		Where(
			"NOT EXISTS (?)",
			r.db.
				Model(&WorkerFlow{}).Select("1").
				Where(
					"worker_flows.flow_id = flows.id AND (worker_flows.status IN ? OR worker_flows.created_at > ?)",
					[]WorkerFlowStatus{WorkerFlowStatusWaiting, WorkerFlowStatusRunning, WorkerFlowStatusCompleted},
					recentCutoff,
				),
		).
		Find(&flows).Error

	if err != nil {
		return nil, err
	}

	return flows, nil
}

func (r *flowRepository) ListAllVersionsByParentID(parentID int64) ([]Flow, error) {
	var flows []Flow
	err := r.db.
		Where("id = ? OR parent_id = ?", parentID, parentID).
		Find(&flows).Error
	if err != nil {
		return nil, err
	}
	return flows, nil
}
