package models

import (
	"time"

	"gorm.io/gorm"
)

// Registry represents a container registry
type Registry struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	URL         string         `json:"url" gorm:"not null"`
	Username    string         `json:"username"`
	Password    string         `json:"password"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName specifies the table name for the Registry model
func (Registry) TableName() string {
	return "registries"
}

// BeforeCreate is a GORM hook that sets timestamps before creating a record
func (r *Registry) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook that sets the updated timestamp before updating a record
func (r *Registry) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}
