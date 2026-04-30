package persistence

import (
	"time"

	"gorm.io/gorm"
)

type OAuthConsent struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UserEmail  string    `gorm:"not null" json:"user_email"`
	ClientID   string    `gorm:"not null" json:"client_id"`
	Scope      string    `gorm:"not null;default:''" json:"scope"`
	ApprovedAt time.Time `gorm:"not null" json:"approved_at"`
}

func (OAuthConsent) TableName() string {
	return "oauth_consents"
}

type OAuthConsentRepository interface {
	Has(userEmail, clientID, scope string) (bool, error)
	Upsert(consent *OAuthConsent) error
	// ApprovedClientIDs returns the set of client IDs the user has any
	// active consent for.
	ApprovedClientIDs(userEmail string) (map[string]bool, error)
	// ClientIDsWithAnyConsent returns the set of client IDs that have at
	// least one consent row from any user. Used for the Settings UI's
	// "consented" column on single-user deployments where the gRPC handler
	// does not yet have access to the requesting user.
	ClientIDsWithAnyConsent() (map[string]bool, error)
	// DeleteUserClient drops every consent the user has for the client.
	DeleteUserClient(userEmail, clientID string) error
	// DeleteByClient drops all consents for a client (called when the client
	// itself is deleted).
	DeleteByClient(clientID string) error
}

type oauthConsentRepository struct {
	db *gorm.DB
}

func NewOAuthConsentRepository(db *gorm.DB) OAuthConsentRepository {
	return &oauthConsentRepository{db: db}
}

func (r *oauthConsentRepository) Has(userEmail, clientID, scope string) (bool, error) {
	var count int64
	err := r.db.Model(&OAuthConsent{}).
		Where("user_email = ? AND client_id = ? AND scope = ?", userEmail, clientID, scope).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *oauthConsentRepository) Upsert(consent *OAuthConsent) error {
	var existing OAuthConsent
	err := r.db.Where("user_email = ? AND client_id = ? AND scope = ?",
		consent.UserEmail, consent.ClientID, consent.Scope).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Update("approved_at", consent.ApprovedAt).Error
	}
	return r.db.Create(consent).Error
}

func (r *oauthConsentRepository) ApprovedClientIDs(userEmail string) (map[string]bool, error) {
	var ids []string
	err := r.db.Model(&OAuthConsent{}).
		Where("user_email = ?", userEmail).
		Distinct().
		Pluck("client_id", &ids).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

func (r *oauthConsentRepository) ClientIDsWithAnyConsent() (map[string]bool, error) {
	var ids []string
	err := r.db.Model(&OAuthConsent{}).Distinct().Pluck("client_id", &ids).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

func (r *oauthConsentRepository) DeleteUserClient(userEmail, clientID string) error {
	return r.db.Where("user_email = ? AND client_id = ?", userEmail, clientID).Delete(&OAuthConsent{}).Error
}

func (r *oauthConsentRepository) DeleteByClient(clientID string) error {
	return r.db.Where("client_id = ?", clientID).Delete(&OAuthConsent{}).Error
}
