package persistence

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type RateLimitState struct {
	ID             int64     `json:"id" gorm:"primaryKey"`
	RateLimitLabel string    `json:"rate_limit_label" gorm:"not null;index:idx_label_key"`
	Key            string    `json:"key" gorm:"not null;index:idx_label_key"`
	Tokens         float64   `json:"tokens" gorm:"not null"`
	LastRefillAt   time.Time `json:"last_refill_at" gorm:"not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"not null"`
}

type RateLimitStateRepository interface {
	GetOrCreate(label, key string, initialTokens float64) (*RateLimitState, error)
	Update(state *RateLimitState) error
	DeleteOlderThan(duration time.Duration) error
}

type rateLimitStateRepository struct {
	db *gorm.DB
}

func NewRateLimitStateRepository(db *gorm.DB) RateLimitStateRepository {
	return &rateLimitStateRepository{db: db}
}

func (r *rateLimitStateRepository) GetOrCreate(label, key string, initialTokens float64) (*RateLimitState, error) {
	var state RateLimitState

	err := r.db.
		Where("rate_limit_label = ? AND key = ?", label, key).
		First(&state).
		Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		state = RateLimitState{
			RateLimitLabel: label,
			Key:            key,
			Tokens:         initialTokens,
			LastRefillAt:   now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := r.db.Create(&state).Error; err != nil {
			return nil, err
		}
	}

	return &state, nil
}

func (r *rateLimitStateRepository) Update(state *RateLimitState) error {
	state.UpdatedAt = time.Now()
	return r.db.Save(state).Error
}

func (r *rateLimitStateRepository) DeleteOlderThan(duration time.Duration) error {
	cutoff := time.Now().Add(-duration)
	return r.db.
		Where("updated_at < ?", cutoff).
		Delete(&RateLimitState{}).
		Error
}
