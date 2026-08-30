package persistence

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	GroupSourceObserved = "observed"
	GroupSourceManual   = "manual"
)

type Group struct {
	ID         int64      `json:"id" gorm:"primaryKey"`
	Name       string     `json:"name" gorm:"not null;uniqueIndex"`
	Source     string     `json:"source" gorm:"not null;default:'observed'"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	LastSeenAt *time.Time `json:"last_seen_at"`
}

// UserGroups is the per-user snapshot of IdP group claims, refreshed on every
// OIDC login. It is the authority for MCP access checks; the groups registry
// above is autocomplete-only and never consulted for decisions.
type UserGroups struct {
	Email     string    `json:"email" gorm:"primaryKey"`
	Groups    []string  `json:"groups" gorm:"type:text;not null;default:'[]';serializer:json"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type GroupRepository interface {
	List() ([]Group, error)
	UpsertObserved(names []string) error
	ImportManual(names []string) (int, error)
	Delete(name string) error
}

type UserGroupsRepository interface {
	Get(email string) ([]string, error)
	Set(email string, groups []string) error
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) List() ([]Group, error) {
	var groups []Group
	err := r.db.Order("name ASC").Find(&groups).Error
	return groups, err
}

func (r *groupRepository) UpsertObserved(names []string) error {
	if len(names) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]Group, len(names))
	for i, name := range names {
		rows[i] = Group{Name: name, Source: GroupSourceObserved, LastSeenAt: &now}
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.Assignments(map[string]any{"source": GroupSourceObserved, "last_seen_at": now}),
	}).Create(&rows).Error
}

func (r *groupRepository) ImportManual(names []string) (int, error) {
	imported := 0
	for _, name := range names {
		res := r.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&Group{Name: name, Source: GroupSourceManual})
		if res.Error != nil {
			return imported, res.Error
		}
		imported += int(res.RowsAffected)
	}
	return imported, nil
}

func (r *groupRepository) Delete(name string) error {
	return r.db.Delete(&Group{}, "name = ?", name).Error
}

type userGroupsRepository struct {
	db *gorm.DB
}

func NewUserGroupsRepository(db *gorm.DB) UserGroupsRepository {
	return &userGroupsRepository{db: db}
}

func (r *userGroupsRepository) Get(email string) ([]string, error) {
	var row UserGroups
	err := r.db.First(&row, "email = ?", email).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return row.Groups, nil
}

func (r *userGroupsRepository) Set(email string, groups []string) error {
	if groups == nil {
		groups = []string{}
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"groups", "updated_at"}),
	}).Create(&UserGroups{Email: email, Groups: groups, UpdatedAt: time.Now()}).Error
}
