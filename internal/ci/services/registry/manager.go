package registry

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ImageMetadata contains metadata about a Docker image
type ImageMetadata struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Service   string    `gorm:"not null"`
	Tag       string    `gorm:"not null"`
	Commit    string    `gorm:"not null"`
	Branch    string    `gorm:"not null"`
	BuildTime time.Time `gorm:"not null"`
	Status    string    `gorm:"not null"` // "success", "failed", etc.
	Registry  string    `gorm:"not null"`
	ImageName string    `gorm:"not null"`
}

// RegistryManager manages Docker image metadata
type RegistryManager struct {
	db *gorm.DB
}

// NewRegistryManager creates a new RegistryManager instance
func NewRegistryManager(dbPath string) (*RegistryManager, error) {
	db, err := gorm.Open(postgres.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&ImageMetadata{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &RegistryManager{db: db}, nil
}

// Close closes the database connection
func (m *RegistryManager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RecordImage records metadata for a Docker image
func (m *RegistryManager) RecordImage(ctx context.Context, metadata ImageMetadata) error {
	result := m.db.WithContext(ctx).Create(&metadata)
	return result.Error
}

// GetLatestImage gets the latest image for a service and branch
func (m *RegistryManager) GetLatestImage(ctx context.Context, service, branch string) (*ImageMetadata, error) {
	var image ImageMetadata
	result := m.db.WithContext(ctx).
		Where("service = ? AND branch = ?", service, branch).
		Order("build_time DESC").
		First(&image)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no image found for service %s and branch %s", service, branch)
		}
		return nil, fmt.Errorf("failed to get latest image: %w", result.Error)
	}

	return &image, nil
}

// GetImageByTag gets an image by its tag
func (m *RegistryManager) GetImageByTag(ctx context.Context, service, tag string) (*ImageMetadata, error) {
	var image ImageMetadata
	result := m.db.WithContext(ctx).
		Where("service = ? AND tag = ?", service, tag).
		First(&image)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no image found for service %s and tag %s", service, tag)
		}
		return nil, fmt.Errorf("failed to get image by tag: %w", result.Error)
	}

	return &image, nil
}

// GetImageByCommit gets an image by its commit hash
func (m *RegistryManager) GetImageByCommit(ctx context.Context, service, commit string) (*ImageMetadata, error) {
	var image ImageMetadata
	result := m.db.WithContext(ctx).
		Where("service = ? AND commit = ?", service, commit).
		First(&image)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no image found for service %s and commit %s", service, commit)
		}
		return nil, fmt.Errorf("failed to get image by commit: %w", result.Error)
	}

	return &image, nil
}

// GetImageHistory gets the image history for a service
func (m *RegistryManager) GetImageHistory(ctx context.Context, service string, limit int) ([]ImageMetadata, error) {
	var images []ImageMetadata
	result := m.db.WithContext(ctx).
		Where("service = ?", service).
		Order("build_time DESC").
		Limit(limit).
		Find(&images)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get image history: %w", result.Error)
	}

	return images, nil
}

// DeleteImage deletes an image by its ID
func (m *RegistryManager) DeleteImage(ctx context.Context, id int64) error {
	result := m.db.WithContext(ctx).Delete(&ImageMetadata{}, id)
	return result.Error
}

// TagImage tags an image with a new tag
func (m *RegistryManager) TagImage(ctx context.Context, service, sourceTag, newTag string) error {
	var sourceImage ImageMetadata
	result := m.db.WithContext(ctx).
		Where("service = ? AND tag = ?", service, sourceTag).
		First(&sourceImage)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("no image found for service %s and tag %s", service, sourceTag)
		}
		return fmt.Errorf("failed to get source image: %w", result.Error)
	}

	newImage := sourceImage
	newImage.ID = 0
	newImage.Tag = newTag
	newImage.BuildTime = time.Now()

	result = m.db.WithContext(ctx).Create(&newImage)
	return result.Error
}

// GetServiceList gets a list of all services
func (m *RegistryManager) GetServiceList(ctx context.Context) ([]string, error) {
	var services []string
	result := m.db.WithContext(ctx).
		Model(&ImageMetadata{}).
		Distinct().
		Pluck("service", &services)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get service list: %w", result.Error)
	}

	return services, nil
}

// GetTagsForService gets all tags for a service
func (m *RegistryManager) GetTagsForService(ctx context.Context, service string) ([]string, error) {
	var tags []string
	result := m.db.WithContext(ctx).
		Model(&ImageMetadata{}).
		Where("service = ?", service).
		Distinct().
		Pluck("tag", &tags)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get tags for service: %w", result.Error)
	}

	return tags, nil
}
