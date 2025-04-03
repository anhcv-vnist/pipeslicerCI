package registry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ImageMetadata contains metadata about a Docker image
type ImageMetadata struct {
	ID        int64
	Service   string
	Tag       string
	Commit    string
	Branch    string
	BuildTime time.Time
	Status    string // "success", "failed", etc.
	Registry  string
	ImageName string
}

// RegistryManager manages Docker image metadata
type RegistryManager struct {
	db *sql.DB
}

// NewRegistryManager creates a new RegistryManager instance
func NewRegistryManager(dbPath string) (*RegistryManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS images (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			service TEXT NOT NULL,
			tag TEXT NOT NULL,
			"commit" TEXT NOT NULL,
			branch TEXT NOT NULL,
			build_time TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			registry TEXT NOT NULL,
			image_name TEXT NOT NULL,
			UNIQUE(service, tag)
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &RegistryManager{db: db}, nil
}

// Close closes the database connection
func (m *RegistryManager) Close() error {
	return m.db.Close()
}

// RecordImage records metadata about a built Docker image
func (m *RegistryManager) RecordImage(ctx context.Context, metadata ImageMetadata) error {
	_, err := m.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO images (service, tag, "commit", branch, build_time, status, registry, image_name)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, metadata.Service, metadata.Tag, metadata.Commit, metadata.Branch, metadata.BuildTime, metadata.Status, metadata.Registry, metadata.ImageName)
	
	if err != nil {
		return fmt.Errorf("failed to record image: %w", err)
	}
	
	return nil
}

// GetLatestImage gets the latest successful image for a service and branch
func (m *RegistryManager) GetLatestImage(ctx context.Context, service, branch string) (*ImageMetadata, error) {
	row := m.db.QueryRowContext(ctx, `
		SELECT id, service, tag, "commit", branch, build_time, status, registry, image_name
		FROM images
		WHERE service = ? AND branch = ? AND status = 'success'
		ORDER BY build_time DESC
		LIMIT 1
	`, service, branch)

	var img ImageMetadata
	err := row.Scan(&img.ID, &img.Service, &img.Tag, &img.Commit, &img.Branch, &img.BuildTime, &img.Status, &img.Registry, &img.ImageName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no image found for service %s on branch %s", service, branch)
		}
		return nil, fmt.Errorf("failed to get latest image: %w", err)
	}

	return &img, nil
}

// GetImageByTag gets an image by service and tag
func (m *RegistryManager) GetImageByTag(ctx context.Context, service, tag string) (*ImageMetadata, error) {
	row := m.db.QueryRowContext(ctx, `
		SELECT id, service, tag, "commit", branch, build_time, status, registry, image_name
		FROM images
		WHERE service = ? AND tag = ?
		LIMIT 1
	`, service, tag)

	var img ImageMetadata
	err := row.Scan(&img.ID, &img.Service, &img.Tag, &img.Commit, &img.Branch, &img.BuildTime, &img.Status, &img.Registry, &img.ImageName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no image found for service %s with tag %s", service, tag)
		}
		return nil, fmt.Errorf("failed to get image by tag: %w", err)
	}

	return &img, nil
}

// GetImageByCommit gets an image by service and commit
func (m *RegistryManager) GetImageByCommit(ctx context.Context, service, commit string) (*ImageMetadata, error) {
	row := m.db.QueryRowContext(ctx, `
		SELECT id, service, tag, "commit", branch, build_time, status, registry, image_name
		FROM images
		WHERE service = ? AND "commit" = ? AND status = 'success'
		ORDER BY build_time DESC
		LIMIT 1
	`, service, commit)

	var img ImageMetadata
	err := row.Scan(&img.ID, &img.Service, &img.Tag, &img.Commit, &img.Branch, &img.BuildTime, &img.Status, &img.Registry, &img.ImageName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no image found for service %s with commit %s", service, commit)
		}
		return nil, fmt.Errorf("failed to get image by commit: %w", err)
	}

	return &img, nil
}

// GetImageHistory gets the build history for a service
func (m *RegistryManager) GetImageHistory(ctx context.Context, service string, limit int) ([]ImageMetadata, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	rows, err := m.db.QueryContext(ctx, `
		SELECT id, service, tag, "commit", branch, build_time, status, registry, image_name
		FROM images
		WHERE service = ?
		ORDER BY build_time DESC
		LIMIT ?
	`, service, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get image history: %w", err)
	}
	defer rows.Close()

	var images []ImageMetadata
	for rows.Next() {
		var img ImageMetadata
		err := rows.Scan(&img.ID, &img.Service, &img.Tag, &img.Commit, &img.Branch, &img.BuildTime, &img.Status, &img.Registry, &img.ImageName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image row: %w", err)
		}
		images = append(images, img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating image rows: %w", err)
	}

	return images, nil
}

// DeleteImage deletes an image by ID
func (m *RegistryManager) DeleteImage(ctx context.Context, id int64) error {
	_, err := m.db.ExecContext(ctx, `
		DELETE FROM images
		WHERE id = ?
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// TagImage adds a new tag to an existing image
func (m *RegistryManager) TagImage(ctx context.Context, service, sourceTag, newTag string) error {
	// Get the source image
	sourceImage, err := m.GetImageByTag(ctx, service, sourceTag)
	if err != nil {
		return fmt.Errorf("failed to get source image: %w", err)
	}

	// Create a new metadata entry with the new tag
	newMetadata := ImageMetadata{
		Service:   sourceImage.Service,
		Tag:       newTag,
		Commit:    sourceImage.Commit,
		Branch:    sourceImage.Branch,
		BuildTime: time.Now(), // Use current time for the tagging operation
		Status:    sourceImage.Status,
		Registry:  sourceImage.Registry,
		ImageName: sourceImage.ImageName,
	}

	// Record the new tag
	err = m.RecordImage(ctx, newMetadata)
	if err != nil {
		return fmt.Errorf("failed to record new tag: %w", err)
	}

	return nil
}

// GetServiceList gets a list of all services that have images
func (m *RegistryManager) GetServiceList(ctx context.Context) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT DISTINCT service
		FROM images
		ORDER BY service
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get service list: %w", err)
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		err := rows.Scan(&service)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service row: %w", err)
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service rows: %w", err)
	}

	return services, nil
}

// GetTagsForService gets all tags for a service
func (m *RegistryManager) GetTagsForService(ctx context.Context, service string) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT DISTINCT tag
		FROM images
		WHERE service = ?
		ORDER BY build_time DESC
	`, service)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for service: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tag rows: %w", err)
	}

	return tags, nil
}
