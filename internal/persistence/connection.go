package persistence

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Connection struct {
	Name            string    `gorm:"primaryKey" json:"name"`
	Provider        string    `gorm:"not null" json:"provider"`
	EncryptedConfig string    `gorm:"not null" json:"-"`
	EncryptedToken  string    `gorm:"not null" json:"-"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type ConnectionRepository interface {
	List() ([]Connection, error)
	GetByName(name string) (*Connection, error)
	Create(conn *Connection) (bool, error)
	UpdateToken(name, encryptedToken string) error
	UpdateConfig(name, encryptedConfig string) error
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
		Select("name, provider, encrypted_config, created_at, updated_at").
		Order("created_at DESC").
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

func (r *connectionRepository) UpdateToken(name, encryptedToken string) error {
	return r.db.Model(&Connection{}).
		Where("name = ?", name).
		Update("encrypted_token", encryptedToken).
		Error
}

func (r *connectionRepository) UpdateConfig(name, encryptedConfig string) error {
	return r.db.Model(&Connection{}).
		Where("name = ?", name).
		Update("encrypted_config", encryptedConfig).
		Error
}

func (r *connectionRepository) Delete(name string) error {
	return r.db.Delete(&Connection{}, "name = ?", name).Error
}
