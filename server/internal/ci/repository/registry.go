package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/vanhcao3/pipeslicerCI/internal/ci/models"
)

// ErrRegistryNotFound is returned when a registry is not found
var ErrRegistryNotFound = errors.New("registry not found")

// RegistryRepository handles database operations for registries
type RegistryRepository struct {
	db *gorm.DB
}

// NewRegistryRepository creates a new RegistryRepository instance
func NewRegistryRepository(db *gorm.DB) *RegistryRepository {
	return &RegistryRepository{db: db}
}

// Create creates a new registry
func (r *RegistryRepository) Create(ctx context.Context, registry *models.Registry) error {
	return r.db.WithContext(ctx).Create(registry).Error
}

// GetByID retrieves a registry by its ID
func (r *RegistryRepository) GetByID(ctx context.Context, id uint) (*models.Registry, error) {
	var registry models.Registry
	err := r.db.WithContext(ctx).First(&registry, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegistryNotFound
		}
		return nil, err
	}
	return &registry, nil
}

// GetByName retrieves a registry by its name
func (r *RegistryRepository) GetByName(ctx context.Context, name string) (*models.Registry, error) {
	var registry models.Registry
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&registry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegistryNotFound
		}
		return nil, err
	}
	return &registry, nil
}

// List retrieves all registries
func (r *RegistryRepository) List(ctx context.Context) ([]models.Registry, error) {
	var registries []models.Registry
	err := r.db.WithContext(ctx).Find(&registries).Error
	if err != nil {
		return nil, err
	}
	return registries, nil
}

// Update updates an existing registry
func (r *RegistryRepository) Update(ctx context.Context, registry *models.Registry) error {
	return r.db.WithContext(ctx).Save(registry).Error
}

// Delete deletes a registry by its ID
func (r *RegistryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&models.Registry{}, id).Error
}
