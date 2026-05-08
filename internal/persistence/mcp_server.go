package persistence

import (
	"time"

	"gorm.io/gorm"
)

// MCP transport kinds.
const (
	MCPTransportHTTP  = "http"
	MCPTransportStdio = "stdio"
)

type MCPServer struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name               string     `gorm:"uniqueIndex;not null" json:"name"`
	URL                string     `gorm:"default:''" json:"url"`
	AuthType           string     `gorm:"not null;default:'none'" json:"auth_type"`
	AuthHeader         string     `gorm:"default:''" json:"auth_header"`
	EncryptedAuthValue string     `gorm:"default:''" json:"-"`
	ConnectionName     string     `gorm:"default:''" json:"connection_name"`
	Status             string     `gorm:"not null;default:'active'" json:"status"`
	ToolCount          int        `gorm:"default:0" json:"tool_count"`
	LastSyncAt         *time.Time `json:"last_sync_at"`
	LastError          string     `gorm:"default:''" json:"last_error"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Transport    string `gorm:"not null;default:'http'" json:"transport"`
	CatalogID    string `gorm:"default:''" json:"catalog_id"`
	Command      string `gorm:"default:''" json:"command"`
	Args         string `gorm:"default:''" json:"args"`          // JSON array of strings
	EncryptedEnv string `gorm:"default:''" json:"-"`             // JSON map of {ENV_NAME: value}, AES-GCM encrypted whole-blob
	ProcessState string `gorm:"default:'stopped'" json:"process_state"`
}

type MCPServerRepository interface {
	List() ([]MCPServer, error)
	ListByStatus(status string) ([]MCPServer, error)
	GetByID(id int64) (*MCPServer, error)
	Create(server *MCPServer) error
	Update(server *MCPServer) error
	Delete(id int64) error
	UpdateSyncStatus(id int64, toolCount int, lastError string) error
	UpdateStatus(id int64, status string) error
	UpdateProcessState(id int64, processState string) error
}

type mcpServerRepository struct {
	db *gorm.DB
}

func NewMCPServerRepository(db *gorm.DB) MCPServerRepository {
	return &mcpServerRepository{db: db}
}

func (r *mcpServerRepository) List() ([]MCPServer, error) {
	var servers []MCPServer
	err := r.db.Order("created_at DESC").Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *mcpServerRepository) ListByStatus(status string) ([]MCPServer, error) {
	var servers []MCPServer
	err := r.db.Where("status = ?", status).Order("created_at DESC").Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (r *mcpServerRepository) GetByID(id int64) (*MCPServer, error) {
	var server MCPServer
	err := r.db.Where("id = ?", id).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

func (r *mcpServerRepository) Create(server *MCPServer) error {
	return r.db.Create(server).Error
}

func (r *mcpServerRepository) Update(server *MCPServer) error {
	return r.db.Save(server).Error
}

func (r *mcpServerRepository) Delete(id int64) error {
	return r.db.Delete(&MCPServer{}, "id = ?", id).Error
}

func (r *mcpServerRepository) UpdateSyncStatus(id int64, toolCount int, lastError string) error {
	now := time.Now()
	return r.db.Model(&MCPServer{}).Where("id = ?", id).Updates(map[string]any{
		"tool_count":   toolCount,
		"last_sync_at": &now,
		"last_error":   lastError,
	}).Error
}

func (r *mcpServerRepository) UpdateStatus(id int64, status string) error {
	return r.db.Model(&MCPServer{}).Where("id = ?", id).Update("status", status).Error
}

func (r *mcpServerRepository) UpdateProcessState(id int64, processState string) error {
	return r.db.Model(&MCPServer{}).Where("id = ?", id).Update("process_state", processState).Error
}
