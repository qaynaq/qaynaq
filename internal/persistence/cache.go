package persistence

import (
	"errors"
	"time"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Cache struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	ParentID  *int64     `json:"parent_id"`
	Label     string     `json:"label" gorm:"not null"`
	Component string     `json:"component" gorm:"not null"`
	Config    []byte     `json:"config" gorm:"not null"`
	IsCurrent bool       `json:"is_current" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`

	ParentCache *Cache `json:"parent_cache" gorm:"foreignKey:ParentID"`
}

func (c *Cache) ToProto() *pb.Cache {
	var updatedAt *timestamppb.Timestamp
	if c.UpdatedAt != nil {
		updatedAt = timestamppb.New(*c.UpdatedAt)
	}

	return &pb.Cache{
		Id:        c.ID,
		ParentId:  c.ParentID,
		Label:     c.Label,
		Component: c.Component,
		Config:    string(c.Config),
		IsCurrent: c.IsCurrent,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

func (c *Cache) FromProto(p *pb.Cache) {
	var updatedAt *time.Time
	if p.GetUpdatedAt() != nil {
		t := p.GetUpdatedAt().AsTime()
		updatedAt = &t
	}

	c.ID = p.Id
	c.ParentID = p.ParentId
	c.Label = p.Label
	c.Component = p.Component
	c.Config = []byte(p.Config)
	c.IsCurrent = p.GetIsCurrent()
	c.CreatedAt = p.CreatedAt.AsTime()
	c.UpdatedAt = updatedAt
}

type CacheRepository interface {
	Create(cache *Cache) error
	Update(cache *Cache) error
	FindByID(id int64) (*Cache, error)
	FindByLabel(label string) (*Cache, error)
	Delete(id int64) error
	ListAll() ([]Cache, error)
}

type cacheRepository struct {
	db *gorm.DB
}

func NewCacheRepository(db *gorm.DB) CacheRepository {
	return &cacheRepository{db: db}
}

func (r *cacheRepository) Create(cache *Cache) error {
	cache.CreatedAt = time.Now()
	cache.IsCurrent = true
	return r.db.Create(cache).Error
}

func (r *cacheRepository) Update(cache *Cache) error {
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
		Model(&Cache{}).
		Where("id = ?", cache.ID).
		Update("is_current", false).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var id = cache.ID
	cache.CreatedAt = time.Now()
	if cache.ParentID == nil {
		cache.ParentID = &id
	}
	cache.ID = 0
	cache.IsCurrent = true

	if err = tx.Create(cache).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *cacheRepository) FindByID(id int64) (*Cache, error) {
	var cache = &Cache{
		ID: id,
	}
	err := r.db.
		Preload("ParentCache").
		First(cache).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return cache, nil
}

func (r *cacheRepository) FindByLabel(label string) (*Cache, error) {
	var cache Cache
	err := r.db.
		Where("label = ? AND is_current = true", label).
		First(&cache).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &cache, nil
}

func (r *cacheRepository) Delete(id int64) error {
	return r.db.Delete(&Cache{}, id).Error
}

func (r *cacheRepository) ListAll() ([]Cache, error) {
	var caches []Cache
	err := r.db.
		Where("is_current = true").
		Order("created_at DESC").
		Find(&caches).
		Error
	if err != nil {
		return nil, err
	}
	return caches, nil
}
