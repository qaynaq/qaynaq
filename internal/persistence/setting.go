package persistence

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Setting struct {
	Key       string     `gorm:"primaryKey" json:"key"`
	Value     string     `gorm:"not null" json:"value"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type SettingRepository interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) Get(key string) (string, error) {
	var setting Setting
	err := r.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *settingRepository) Set(key, value string) error {
	now := time.Now()
	setting := Setting{Key: key, Value: value, UpdatedAt: &now}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&setting).Error
}
