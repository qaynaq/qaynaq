package persistence

import (
	"errors"
	"time"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Buffer struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	ParentID  *int64     `json:"parent_id"`
	Label     string     `json:"label" gorm:"not null"`
	Component string     `json:"component" gorm:"not null"`
	Config    []byte     `json:"config" gorm:"not null"`
	IsCurrent bool       `json:"is_current" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`

	ParentBuffer *Buffer `json:"parent_buffer" gorm:"foreignKey:ParentID"`
}

func (b *Buffer) ToProto() *pb.Buffer {
	var updatedAt *timestamppb.Timestamp
	if b.UpdatedAt != nil {
		updatedAt = timestamppb.New(*b.UpdatedAt)
	}

	return &pb.Buffer{
		Id:        b.ID,
		ParentId:  b.ParentID,
		Label:     b.Label,
		Component: b.Component,
		Config:    string(b.Config),
		IsCurrent: b.IsCurrent,
		CreatedAt: timestamppb.New(b.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

func (b *Buffer) FromProto(p *pb.Buffer) {
	var updatedAt *time.Time
	if p.GetUpdatedAt() != nil {
		t := p.GetUpdatedAt().AsTime()
		updatedAt = &t
	}

	b.ID = p.Id
	b.ParentID = p.ParentId
	b.Label = p.Label
	b.Component = p.Component
	b.Config = []byte(p.Config)
	b.IsCurrent = p.GetIsCurrent()
	b.CreatedAt = p.CreatedAt.AsTime()
	b.UpdatedAt = updatedAt
}

type BufferRepository interface {
	Create(buffer *Buffer) error
	Update(buffer *Buffer) error
	FindByID(id int64) (*Buffer, error)
	FindByLabel(label string) (*Buffer, error)
	Delete(id int64) error
	ListAll() ([]Buffer, error)
}

type bufferRepository struct {
	db *gorm.DB
}

func NewBufferRepository(db *gorm.DB) BufferRepository {
	return &bufferRepository{db: db}
}

func (r *bufferRepository) Create(buffer *Buffer) error {
	buffer.CreatedAt = time.Now()
	buffer.IsCurrent = true
	return r.db.Create(buffer).Error
}

func (r *bufferRepository) Update(buffer *Buffer) error {
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
		Model(&Buffer{}).
		Where("id = ?", buffer.ID).
		Update("is_current", false).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var id = buffer.ID
	buffer.CreatedAt = time.Now()
	if buffer.ParentID == nil {
		buffer.ParentID = &id
	}
	buffer.ID = 0
	buffer.IsCurrent = true

	if err = tx.Create(buffer).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *bufferRepository) FindByID(id int64) (*Buffer, error) {
	var buffer = &Buffer{
		ID: id,
	}
	err := r.db.
		Preload("ParentBuffer").
		First(buffer).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return buffer, nil
}

func (r *bufferRepository) FindByLabel(label string) (*Buffer, error) {
	var buffer Buffer
	err := r.db.
		Where("label = ? AND is_current = true", label).
		First(&buffer).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &buffer, nil
}

func (r *bufferRepository) Delete(id int64) error {
	return r.db.Delete(&Buffer{}, id).Error
}

func (r *bufferRepository) ListAll() ([]Buffer, error) {
	var buffers []Buffer
	err := r.db.
		Where("is_current = true").
		Order("created_at DESC").
		Find(&buffers).
		Error
	if err != nil {
		return nil, err
	}
	return buffers, nil
}
