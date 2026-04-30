package persistence

import (
	"time"

	"gorm.io/gorm"
)

type OAuthClient struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	SecretHash   string     `gorm:"not null" json:"-"`
	Name         string     `gorm:"not null" json:"name"`
	RedirectURIs []string   `gorm:"type:text;not null;default:'[]';serializer:json" json:"redirect_uris"`
	CreatedAt    time.Time  `gorm:"not null" json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at"`
}

func (OAuthClient) TableName() string {
	return "oauth_clients"
}

type OAuthClientRepository interface {
	List() ([]OAuthClient, error)
	Create(client *OAuthClient) error
	Delete(id string) error
	FindByID(id string) (*OAuthClient, error)
	UpdateLastUsedAt(id string, t time.Time) error
}

type oauthClientRepository struct {
	db *gorm.DB
}

func NewOAuthClientRepository(db *gorm.DB) OAuthClientRepository {
	return &oauthClientRepository{db: db}
}

func (r *oauthClientRepository) List() ([]OAuthClient, error) {
	var clients []OAuthClient
	err := r.db.Order("created_at DESC").Find(&clients).Error
	return clients, err
}

func (r *oauthClientRepository) Create(client *OAuthClient) error {
	return r.db.Create(client).Error
}

func (r *oauthClientRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&OAuthClient{}).Error
}

func (r *oauthClientRepository) FindByID(id string) (*OAuthClient, error) {
	var client OAuthClient
	err := r.db.Where("id = ?", id).First(&client).Error
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *oauthClientRepository) UpdateLastUsedAt(id string, t time.Time) error {
	return r.db.Model(&OAuthClient{}).Where("id = ?", id).Update("last_used_at", &t).Error
}
