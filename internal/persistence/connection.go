package persistence

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Connection struct {
	Name                string     `gorm:"primaryKey" json:"name"`
	Provider            string     `gorm:"not null" json:"provider"`
	EncryptedConfig     string     `gorm:"not null" json:"-"`
	EncryptedToken      string     `gorm:"not null" json:"-"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	LastError           string     `json:"last_error,omitempty"`
	LastErrorAt         *time.Time `json:"last_error_at,omitempty"`
	FirstFailedAt       *time.Time `json:"first_failed_at,omitempty"`
	ConsecutiveFailures int        `gorm:"not null;default:0" json:"consecutive_failures"`
	CreatedAt           time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type ConnectionRepository interface {
	List() ([]Connection, error)
	GetByName(name string) (*Connection, error)
	ListExpiringBefore(threshold time.Time) ([]Connection, error)
	ListFailing() ([]Connection, error)
	Create(conn *Connection) (bool, error)
	UpdateToken(name, encryptedToken string, expiresAt *time.Time) error
	UpdateConfig(name, encryptedConfig string) error
	RecordFailure(name, lastError string) error
	ClearError(name string) error
	Delete(name string) error
}

type connectionRepository struct {
	db *gorm.DB
}

func NewConnectionRepository(db *gorm.DB) ConnectionRepository {
	return &connectionRepository{db: db}
}

func (r *connectionRepository) List() ([]Connection, error) {
	var connections []Connection
	err := r.db.
		Select("name, provider, encrypted_config, expires_at, last_error, last_error_at, first_failed_at, consecutive_failures, created_at, updated_at").
		Order("created_at DESC").
		Find(&connections).
		Error
	if err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *connectionRepository) ListFailing() ([]Connection, error) {
	var connections []Connection
	err := r.db.
		Select("name, provider, last_error, last_error_at, first_failed_at, consecutive_failures, created_at, updated_at").
		Where("last_error <> ''").
		Order("last_error_at DESC").
		Find(&connections).
		Error
	if err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *connectionRepository) GetByName(name string) (*Connection, error) {
	var conn Connection
	err := r.db.
		Where("name = ?", name).
		First(&conn).
		Error
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *connectionRepository) Create(conn *Connection) (bool, error) {
	err := r.db.Create(conn).Error
	return errors.Is(err, gorm.ErrDuplicatedKey), err
}

func (r *connectionRepository) UpdateToken(name, encryptedToken string, expiresAt *time.Time) error {
	return r.db.Model(&Connection{}).
		Where("name = ?", name).
		Updates(map[string]any{
			"encrypted_token": encryptedToken,
			"expires_at":      expiresAt,
		}).
		Error
}

// ListExpiringBefore returns connections whose token expires before
// threshold. NULL expires_at means a non-expiring token (Slack without
// rotation), which is excluded.
func (r *connectionRepository) ListExpiringBefore(threshold time.Time) ([]Connection, error) {
	var connections []Connection
	err := r.db.
		Where("expires_at IS NOT NULL AND expires_at < ?", threshold).
		Find(&connections).
		Error
	if err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *connectionRepository) UpdateConfig(name, encryptedConfig string) error {
	return r.db.Model(&Connection{}).
		Where("name = ?", name).
		Update("encrypted_config", encryptedConfig).
		Error
}

// COALESCE preserves first_failed_at across a failure streak so the UI can
// show "failing for 6h" instead of resetting on each tick.
func (r *connectionRepository) RecordFailure(name, lastError string) error {
	return r.db.Model(&Connection{}).
		Where("name = ?", name).
		Updates(map[string]any{
			"last_error":           lastError,
			"last_error_at":        gorm.Expr("CURRENT_TIMESTAMP"),
			"first_failed_at":      gorm.Expr("COALESCE(first_failed_at, CURRENT_TIMESTAMP)"),
			"consecutive_failures": gorm.Expr("consecutive_failures + 1"),
		}).
		Error
}

func (r *connectionRepository) ClearError(name string) error {
	return r.db.Model(&Connection{}).
		Where("name = ? AND (last_error <> '' OR consecutive_failures > 0)", name).
		Updates(map[string]any{
			"last_error":           "",
			"last_error_at":        nil,
			"first_failed_at":      nil,
			"consecutive_failures": 0,
		}).
		Error
}

func (r *connectionRepository) Delete(name string) error {
	return r.db.Delete(&Connection{}, "name = ?", name).Error
}
