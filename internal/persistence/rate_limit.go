package persistence

import (
	"errors"
	"time"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type RateLimit struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	ParentID  *int64     `json:"parent_id"`
	Label     string     `json:"label" gorm:"not null"`
	Component string     `json:"component" gorm:"not null"`
	Config    []byte     `json:"config" gorm:"not null"`
	IsCurrent bool       `json:"is_current" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`

	ParentRateLimit *RateLimit `json:"parent_rate_limit" gorm:"foreignKey:ParentID"`
}

func (rl *RateLimit) ToProto() *pb.RateLimit {
	var updatedAt *timestamppb.Timestamp
	if rl.UpdatedAt != nil {
		updatedAt = timestamppb.New(*rl.UpdatedAt)
	}

	return &pb.RateLimit{
		Id:        rl.ID,
		ParentId:  rl.ParentID,
		Label:     rl.Label,
		Component: rl.Component,
		Config:    string(rl.Config),
		IsCurrent: rl.IsCurrent,
		CreatedAt: timestamppb.New(rl.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

func (rl *RateLimit) FromProto(p *pb.RateLimit) {
	var updatedAt *time.Time
	if p.GetUpdatedAt() != nil {
		t := p.GetUpdatedAt().AsTime()
		updatedAt = &t
	}

	rl.ID = p.Id
	rl.ParentID = p.ParentId
	rl.Label = p.Label
	rl.Component = p.Component
	rl.Config = []byte(p.Config)
	rl.IsCurrent = p.GetIsCurrent()
	rl.CreatedAt = p.CreatedAt.AsTime()
	rl.UpdatedAt = updatedAt
}

type RateLimitRepository interface {
	Create(rateLimit *RateLimit) error
	Update(rateLimit *RateLimit) error
	FindByID(id int64) (*RateLimit, error)
	FindByLabel(label string) (*RateLimit, error)
	Delete(id int64) error
	ListAll() ([]RateLimit, error)
}

type rateLimitRepository struct {
	db *gorm.DB
}

func NewRateLimitRepository(db *gorm.DB) RateLimitRepository {
	return &rateLimitRepository{db: db}
}

func (r *rateLimitRepository) Create(rateLimit *RateLimit) error {
	rateLimit.CreatedAt = time.Now()
	rateLimit.IsCurrent = true
	return r.db.Create(rateLimit).Error
}

func (r *rateLimitRepository) Update(rateLimit *RateLimit) error {
	var err error

	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	err = tx.
		Model(&RateLimit{}).
		Where("id = ?", rateLimit.ID).
		Update("is_current", false).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var id = rateLimit.ID
	rateLimit.CreatedAt = time.Now()
	if rateLimit.ParentID == nil {
		rateLimit.ParentID = &id
	}
	rateLimit.ID = 0
	rateLimit.IsCurrent = true

	if err = tx.Create(rateLimit).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *rateLimitRepository) FindByID(id int64) (*RateLimit, error) {
	var rateLimit = &RateLimit{
		ID: id,
	}
	err := r.db.
		Preload("ParentRateLimit").
		First(rateLimit).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return rateLimit, nil
}

func (r *rateLimitRepository) FindByLabel(label string) (*RateLimit, error) {
	var rateLimit RateLimit
	err := r.db.
		Where("label = ? AND is_current = true", label).
		First(&rateLimit).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &rateLimit, nil
}

func (r *rateLimitRepository) Delete(id int64) error {
	return r.db.Delete(&RateLimit{}, id).Error
}

func (r *rateLimitRepository) ListAll() ([]RateLimit, error) {
	var rateLimits []RateLimit
	err := r.db.
		Where("is_current = true").
		Order("created_at DESC").
		Find(&rateLimits).
		Error
	if err != nil {
		return nil, err
	}
	return rateLimits, nil
}
