package persistence

import (
	"time"

	"gorm.io/gorm"
)

type APIToken struct {
	ID         int64      `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"not null;uniqueIndex" json:"name"`
	TokenHash  string     `gorm:"not null" json:"-"`
	Scopes     []string   `gorm:"type:text;not null;default:'[]';serializer:json" json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

type APITokenRepository interface {
	List() ([]APIToken, error)
	Create(token *APIToken) error
	Delete(id int64) error
	FindByHash(hash string) (*APIToken, error)
	BatchUpdateLastUsedAt(updates map[int64]time.Time) error
}

type apiTokenRepository struct {
	db *gorm.DB
}

func NewAPITokenRepository(db *gorm.DB) APITokenRepository {
	return &apiTokenRepository{db: db}
}

func (r *apiTokenRepository) List() ([]APIToken, error) {
	var tokens []APIToken
	err := r.db.Order("created_at DESC").Find(&tokens).Error
	return tokens, err
}

func (r *apiTokenRepository) Create(token *APIToken) error {
	return r.db.Create(token).Error
}

func (r *apiTokenRepository) Delete(id int64) error {
	return r.db.Delete(&APIToken{}, "id = ?", id).Error
}

func (r *apiTokenRepository) FindByHash(hash string) (*APIToken, error) {
	var token APIToken
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *apiTokenRepository) BatchUpdateLastUsedAt(updates map[int64]time.Time) error {
	for id, usedAt := range updates {
		t := usedAt
		if err := r.db.Model(&APIToken{}).Where("id = ?", id).Update("last_used_at", &t).Error; err != nil {
			return err
		}
	}
	return nil
}
