package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"gorm.io/gorm"
)

// RepositoryMetadata contains metadata about a Git repository
type RepositoryMetadata struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	URL         string    `gorm:"not null;uniqueIndex"`
	Name        string    `gorm:"not null"`
	Description string    `gorm:""`
	LocalPath   string    `gorm:"not null"`
	LastUpdated time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

// RepositoryManager manages Git repositories
type RepositoryManager struct {
	db      *gorm.DB
	baseDir string
}

// NewRepositoryManager creates a new RepositoryManager instance
func NewRepositoryManager(db *gorm.DB, baseDir string) (*RepositoryManager, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Auto migrate the schema
	err := db.AutoMigrate(&RepositoryMetadata{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &RepositoryManager{
		db:      db,
		baseDir: baseDir,
	}, nil
}

// CloneRepository clones a Git repository and stores its metadata
func (m *RepositoryManager) CloneRepository(ctx context.Context, url, name, description string) (*RepositoryMetadata, error) {
	// Check if repository already exists
	var existingRepo RepositoryMetadata
	result := m.db.WithContext(ctx).Where("url = ?", url).First(&existingRepo)
	if result.Error == nil {
		// Repository exists, update it
		return m.UpdateRepository(ctx, &existingRepo)
	}

	// Create a directory for the repository
	repoDir := filepath.Join(m.baseDir, name)
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository directory: %w", err)
	}

	// Clone the repository
	log.Printf("Cloning repository %s to %s", url, repoDir)
	_, err = git.PlainClone(repoDir, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	// Create repository metadata
	now := time.Now()
	metadata := &RepositoryMetadata{
		URL:         url,
		Name:        name,
		Description: description,
		LocalPath:   repoDir,
		LastUpdated: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save metadata to database
	result = m.db.WithContext(ctx).Create(metadata)
	if result.Error != nil {
		// Clean up the cloned repository if metadata save fails
		os.RemoveAll(repoDir)
		return nil, fmt.Errorf("failed to save repository metadata: %w", result.Error)
	}

	return metadata, nil
}

// UpdateRepository updates an existing repository
func (m *RepositoryManager) UpdateRepository(ctx context.Context, metadata *RepositoryMetadata) (*RepositoryMetadata, error) {
	// Pull the latest changes
	repo, err := git.PlainOpen(metadata.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Pull the latest changes
	err = worktree.Pull(&git.PullOptions{
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("failed to pull repository: %w", err)
	}

	// Update metadata
	metadata.LastUpdated = time.Now()
	metadata.UpdatedAt = time.Now()

	// Save metadata to database
	result := m.db.WithContext(ctx).Save(metadata)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update repository metadata: %w", result.Error)
	}

	return metadata, nil
}

// GetRepositoryByURL gets a repository by its URL
func (m *RepositoryManager) GetRepositoryByURL(ctx context.Context, url string) (*RepositoryMetadata, error) {
	var metadata RepositoryMetadata
	result := m.db.WithContext(ctx).Where("url = ?", url).First(&metadata)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found: %s", url)
		}
		return nil, fmt.Errorf("failed to get repository: %w", result.Error)
	}

	return &metadata, nil
}

// GetRepositoryByID gets a repository by its ID
func (m *RepositoryManager) GetRepositoryByID(ctx context.Context, id int64) (*RepositoryMetadata, error) {
	var metadata RepositoryMetadata
	result := m.db.WithContext(ctx).First(&metadata, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("repository not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get repository: %w", result.Error)
	}

	return &metadata, nil
}

// ListRepositories lists all repositories
func (m *RepositoryManager) ListRepositories(ctx context.Context) ([]RepositoryMetadata, error) {
	var repositories []RepositoryMetadata
	result := m.db.WithContext(ctx).Find(&repositories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", result.Error)
	}

	return repositories, nil
}

// DeleteRepository deletes a repository
func (m *RepositoryManager) DeleteRepository(ctx context.Context, id int64) error {
	// Get the repository
	metadata, err := m.GetRepositoryByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the repository directory
	err = os.RemoveAll(metadata.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to delete repository directory: %w", err)
	}

	// Delete the repository metadata
	result := m.db.WithContext(ctx).Delete(&RepositoryMetadata{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete repository metadata: %w", result.Error)
	}

	return nil
}

// CheckoutBranch checks out a branch in a repository
func (m *RepositoryManager) CheckoutBranch(ctx context.Context, id int64, branch string) error {
	// Get the repository
	metadata, err := m.GetRepositoryByID(ctx, id)
	if err != nil {
		return err
	}

	// Open the repository
	repo, err := git.PlainOpen(metadata.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Checkout the branch
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	// Update metadata
	metadata.LastUpdated = time.Now()
	metadata.UpdatedAt = time.Now()

	// Save metadata to database
	result := m.db.WithContext(ctx).Save(metadata)
	if result.Error != nil {
		return fmt.Errorf("failed to update repository metadata: %w", result.Error)
	}

	return nil
}

// GetRepositoryPath gets the local path of a repository
func (m *RepositoryManager) GetRepositoryPath(ctx context.Context, id int64) (string, error) {
	// Get the repository
	metadata, err := m.GetRepositoryByID(ctx, id)
	if err != nil {
		return "", err
	}

	return metadata.LocalPath, nil
}
