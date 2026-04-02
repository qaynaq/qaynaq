package persistence

import (
	"errors"
	"time"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type File struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	ParentID  *int64     `json:"parent_id"`
	Key       string     `json:"key" gorm:"not null"`
	Content   []byte     `json:"-" gorm:"not null"`
	Size      int64      `json:"size" gorm:"not null;default:0"`
	IsCurrent bool       `json:"is_current" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt *time.Time `json:"updated_at"`

	ParentFile *File `json:"parent_file" gorm:"foreignKey:ParentID"`
}

func (f *File) ToProto() *pb.File {
	var updatedAt *timestamppb.Timestamp
	if f.UpdatedAt != nil {
		updatedAt = timestamppb.New(*f.UpdatedAt)
	}

	return &pb.File{
		Id:        f.ID,
		ParentId:  f.ParentID,
		Key:       f.Key,
		Content:   f.Content,
		Size:      f.Size,
		IsCurrent: f.IsCurrent,
		CreatedAt: timestamppb.New(f.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

func (f *File) ToProtoWithoutContent() *pb.File {
	var updatedAt *timestamppb.Timestamp
	if f.UpdatedAt != nil {
		updatedAt = timestamppb.New(*f.UpdatedAt)
	}

	return &pb.File{
		Id:        f.ID,
		ParentId:  f.ParentID,
		Key:       f.Key,
		Size:      f.Size,
		IsCurrent: f.IsCurrent,
		CreatedAt: timestamppb.New(f.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

func (f *File) FromProto(p *pb.File) {
	f.ID = p.Id
	f.ParentID = p.ParentId
	f.Key = p.Key
	if p.Content != nil {
		f.Content = p.Content
	} else {
		f.Content = []byte{}
	}
	f.Size = int64(len(p.Content))
	f.IsCurrent = p.GetIsCurrent()
	if p.GetCreatedAt() != nil {
		f.CreatedAt = p.GetCreatedAt().AsTime()
	}
	if p.GetUpdatedAt() != nil {
		t := p.GetUpdatedAt().AsTime()
		f.UpdatedAt = &t
	}
}

type FileRepository interface {
	Create(file *File) error
	Update(file *File) error
	FindByID(id int64) (*File, error)
	FindByKey(key string) (*File, error)
	FindByKeys(keys []string) ([]File, error)
	HasKeyConflict(key string, excludeID int64) (bool, error)
	Delete(id int64) error
	ListAll() ([]File, error)
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(file *File) error {
	file.CreatedAt = time.Now()
	file.IsCurrent = true
	file.Size = int64(len(file.Content))
	return r.db.Create(file).Error
}

func (r *fileRepository) Update(file *File) error {
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
		Model(&File{}).
		Where("id = ?", file.ID).
		Update("is_current", false).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var id = file.ID
	file.CreatedAt = time.Now()
	if file.ParentID == nil {
		file.ParentID = &id
	}
	file.ID = 0
	file.IsCurrent = true
	file.Size = int64(len(file.Content))

	if err = tx.Create(file).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *fileRepository) FindByID(id int64) (*File, error) {
	var file = &File{
		ID: id,
	}
	err := r.db.
		Preload("ParentFile").
		First(file).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return file, nil
}

func (r *fileRepository) FindByKey(key string) (*File, error) {
	var file File
	err := r.db.
		Where("key = ? AND is_current = true", key).
		First(&file).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &file, nil
}

func (r *fileRepository) FindByKeys(keys []string) ([]File, error) {
	var files []File
	if len(keys) == 0 {
		return files, nil
	}
	err := r.db.Where("key IN ? AND is_current = true", keys).Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (r *fileRepository) HasKeyConflict(key string, excludeID int64) (bool, error) {
	var count int64
	query := r.db.Model(&File{}).Where("is_current = true")
	if excludeID != 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.
		Where("key LIKE ? OR ? LIKE key || '/%'", key+"/%", key).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *fileRepository) Delete(id int64) error {
	return r.db.Delete(&File{}, id).Error
}

func (r *fileRepository) ListAll() ([]File, error) {
	var files []File
	err := r.db.
		Select("id, parent_id, key, size, is_current, created_at, updated_at").
		Where("is_current = true").
		Order("key ASC").
		Find(&files).
		Error
	if err != nil {
		return nil, err
	}
	return files, nil
}
