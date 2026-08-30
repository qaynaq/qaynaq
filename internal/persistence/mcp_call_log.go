package persistence

import (
	"time"

	"gorm.io/gorm"
)

const (
	MCPActorOAuth     = "oauth"
	MCPActorAPIToken  = "api_token"
	MCPActorAnonymous = "anonymous"

	MCPCallStatusOK     = "ok"
	MCPCallStatusError  = "error"
	MCPCallStatusDenied = "denied"

	MCPTargetNative   = "native"
	MCPTargetUpstream = "upstream"
	MCPTargetStdio    = "stdio"
)

type MCPCallLog struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	Actor      string    `json:"actor" gorm:"not null"`
	ActorKind  string    `json:"actor_kind" gorm:"not null"`
	Groups     []string  `json:"groups" gorm:"type:text;not null;default:'[]';serializer:json"`
	ToolName   string    `json:"tool_name" gorm:"not null"`
	TargetKind string    `json:"target_kind" gorm:"not null"`
	TargetID   int64     `json:"target_id" gorm:"not null;default:0"`
	Status     string    `json:"status" gorm:"not null"`
	Error      string    `json:"error" gorm:"not null;default:''"`
	DurationMs int64     `json:"duration_ms" gorm:"not null;default:0"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type MCPCallLogRepository interface {
	Add(entry *MCPCallLog) error
	List(limit, offset int) ([]MCPCallLog, int64, error)
}

type mcpCallLogRepository struct {
	db *gorm.DB
}

func NewMCPCallLogRepository(db *gorm.DB) MCPCallLogRepository {
	return &mcpCallLogRepository{db: db}
}

func (r *mcpCallLogRepository) Add(entry *MCPCallLog) error {
	return r.db.Create(entry).Error
}

func (r *mcpCallLogRepository) List(limit, offset int) ([]MCPCallLog, int64, error) {
	var total int64
	if err := r.db.Model(&MCPCallLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entries []MCPCallLog
	err := r.db.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&entries).Error
	return entries, total, err
}
