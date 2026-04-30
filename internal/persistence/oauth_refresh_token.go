package persistence

import (
	"time"

	"gorm.io/gorm"
)

type OAuthRefreshToken struct {
	ID        int64      `gorm:"primaryKey" json:"id"`
	TokenHash string     `gorm:"not null;uniqueIndex" json:"-"`
	ClientID  string     `gorm:"not null;index" json:"client_id"`
	UserEmail string     `gorm:"not null;index" json:"user_email"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
}

func (OAuthRefreshToken) TableName() string {
	return "oauth_refresh_tokens"
}

// OAuthSession is an active refresh token joined with client display info.
type OAuthSession struct {
	ID         int64     `json:"id"`
	ClientID   string    `json:"client_id"`
	ClientName string    `json:"client_name"`
	UserEmail  string    `json:"user_email"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type OAuthRefreshTokenRepository interface {
	Create(token *OAuthRefreshToken) error
	FindByHash(hash string) (*OAuthRefreshToken, error)
	FindByID(id int64) (*OAuthRefreshToken, error)
	Revoke(id int64) error
	// DeleteByClient hard-deletes every refresh token belonging to the
	// client. Used when the client itself is being removed - we don't want
	// orphan rows surviving in the table.
	DeleteByClient(clientID string) error
	DeleteExpired(before time.Time) error
	ListActiveSessions() ([]OAuthSession, error)
}

type oauthRefreshTokenRepository struct {
	db *gorm.DB
}

func NewOAuthRefreshTokenRepository(db *gorm.DB) OAuthRefreshTokenRepository {
	return &oauthRefreshTokenRepository{db: db}
}

func (r *oauthRefreshTokenRepository) Create(token *OAuthRefreshToken) error {
	return r.db.Create(token).Error
}

func (r *oauthRefreshTokenRepository) FindByHash(hash string) (*OAuthRefreshToken, error) {
	var token OAuthRefreshToken
	err := r.db.Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *oauthRefreshTokenRepository) FindByID(id int64) (*OAuthRefreshToken, error) {
	var token OAuthRefreshToken
	err := r.db.Where("id = ?", id).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *oauthRefreshTokenRepository) ListActiveSessions() ([]OAuthSession, error) {
	var sessions []OAuthSession
	err := r.db.Table("oauth_refresh_tokens AS rt").
		Select("rt.id, rt.client_id, c.name AS client_name, rt.user_email, rt.created_at, rt.expires_at").
		Joins("LEFT JOIN oauth_clients c ON c.id = rt.client_id").
		Where("rt.revoked_at IS NULL AND rt.expires_at > ?", time.Now()).
		Order("rt.created_at DESC").
		Scan(&sessions).Error
	return sessions, err
}


func (r *oauthRefreshTokenRepository) Revoke(id int64) error {
	now := time.Now()
	return r.db.Model(&OAuthRefreshToken{}).Where("id = ?", id).Update("revoked_at", &now).Error
}

func (r *oauthRefreshTokenRepository) DeleteByClient(clientID string) error {
	return r.db.Where("client_id = ?", clientID).Delete(&OAuthRefreshToken{}).Error
}

func (r *oauthRefreshTokenRepository) DeleteExpired(before time.Time) error {
	return r.db.Where("expires_at < ?", before).Delete(&OAuthRefreshToken{}).Error
}
