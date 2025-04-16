package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

// CommitInfo represents information about a git commit
type CommitInfo struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
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

	// Create a sanitized directory name from the repository name
	sanitizedName := strings.ReplaceAll(name, "/", "-")
	repoDir := filepath.Join(m.baseDir, sanitizedName)

	log.Printf("Creating repository directory at: %s", repoDir)
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
		// Clean up the directory if clone fails
		os.RemoveAll(repoDir)
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	// Verify the repository was cloned correctly
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); os.IsNotExist(err) {
		os.RemoveAll(repoDir)
		return nil, fmt.Errorf("repository was not cloned correctly: .git directory not found")
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

	log.Printf("Successfully cloned repository to: %s", repoDir)
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

	// Get current branch
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if we're already on the requested branch
	currentBranch := head.Name().Short()
	if currentBranch == branch {
		// We're already on the right branch, fetch and pull latest changes
		// First fetch the specific branch
		err = repo.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{
				config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", branch, branch)),
			},
			Force: true,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return fmt.Errorf("failed to fetch branch: %w", err)
		}

		// Get the remote branch reference
		remoteBranch := plumbing.NewRemoteReferenceName("origin", branch)
		remoteRef, err := repo.Reference(remoteBranch, true)
		if err != nil {
			return fmt.Errorf("failed to get remote branch reference: %w", err)
		}

		// Reset to the remote branch state
		err = worktree.Reset(&git.ResetOptions{
			Commit: remoteRef.Hash(),
			Mode:   git.HardReset,
		})
		if err != nil {
			return fmt.Errorf("failed to reset to remote branch: %w", err)
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

	// First, fetch all branches
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/remotes/origin/*"),
		},
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch branches: %w", err)
	}

	// Clean the worktree
	err = worktree.Clean(&git.CleanOptions{
		Dir: true,
	})
	if err != nil {
		return fmt.Errorf("failed to clean worktree: %w", err)
	}

	// Reset any local changes
	err = worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset worktree: %w", err)
	}

	// Check if the branch exists locally
	localBranch := plumbing.NewBranchReferenceName(branch)
	_, err = repo.Reference(localBranch, true)
	if err != nil {
		// Branch doesn't exist locally, try to create it from remote
		remoteBranch := plumbing.NewRemoteReferenceName("origin", branch)
		remoteRef, err := repo.Reference(remoteBranch, true)
		if err != nil {
			return fmt.Errorf("branch '%s' not found locally or remotely: %w", branch, err)
		}

		// Create a new branch pointing to the remote branch
		err = repo.Storer.SetReference(plumbing.NewHashReference(localBranch, remoteRef.Hash()))
		if err != nil {
			return fmt.Errorf("failed to create local branch reference: %w", err)
		}

		// Checkout the new branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: localBranch,
			Force:  true,
		})
		if err != nil {
			return fmt.Errorf("failed to checkout branch: %w", err)
		}

	} else {
		// Branch exists locally, check if it tracks remote
		remoteBranch := plumbing.NewRemoteReferenceName("origin", branch)
		remoteRef, err := repo.Reference(remoteBranch, true)
		if err != nil {
			// Local branch exists but doesn't track remote, delete it and recreate
			err = repo.Storer.RemoveReference(localBranch)
			if err != nil {
				return fmt.Errorf("failed to remove existing local branch: %w", err)
			}

			// Create new branch from remote
			err = repo.Storer.SetReference(plumbing.NewHashReference(localBranch, remoteRef.Hash()))
			if err != nil {
				return fmt.Errorf("failed to create local branch reference: %w", err)
			}

			// Set up tracking
			err = repo.CreateBranch(&config.Branch{
				Name:   branch,
				Remote: "origin",
				Merge:  remoteBranch,
			})
			if err != nil {
				return fmt.Errorf("failed to set up branch tracking: %w", err)
			}
		} else {
			// Reset the local branch to match remote
			err = worktree.Reset(&git.ResetOptions{
				Commit: remoteRef.Hash(),
				Mode:   git.HardReset,
			})
			if err != nil {
				return fmt.Errorf("failed to reset local branch: %w", err)
			}
		}

		// Checkout the branch
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: localBranch,
			Force:  true,
		})
		if err != nil {
			// If checkout fails, try to delete and recreate the branch
			err = repo.Storer.RemoveReference(localBranch)
			if err != nil {
				return fmt.Errorf("failed to remove problematic local branch: %w", err)
			}

			// Create new branch from remote
			err = repo.Storer.SetReference(plumbing.NewHashReference(localBranch, remoteRef.Hash()))
			if err != nil {
				return fmt.Errorf("failed to recreate local branch reference: %w", err)
			}

			// Try checkout again
			err = worktree.Checkout(&git.CheckoutOptions{
				Branch: localBranch,
				Create: true,
				Force:  true,
			})
			if err != nil {
				return fmt.Errorf("failed to checkout branch after recreation: %w", err)
			}
		}
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

// GetBranchCommits retrieves all commits for a specific branch
func (m *RepositoryManager) GetBranchCommits(ctx context.Context, id int64, branch string) ([]CommitInfo, error) {
	// Get repository metadata
	metadata, err := m.GetRepositoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}

	// Verify repository path exists
	if _, err := os.Stat(metadata.LocalPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository directory not found at: %s", metadata.LocalPath)
	}

	// Verify .git directory exists
	gitDir := filepath.Join(metadata.LocalPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository is not a valid git repository: .git directory not found at %s", gitDir)
	}

	log.Printf("Opening repository at: %s", metadata.LocalPath)
	// Open the repository
	repo, err := git.PlainOpen(metadata.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// List all branches to check if the requested branch exists
	branches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branchExists := false
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == branch {
			branchExists = true
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	if !branchExists {
		return nil, fmt.Errorf("branch '%s' does not exist", branch)
	}

	// Get the branch reference
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch reference: %w", err)
	}

	// Get the commit iterator
	commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit iterator: %w", err)
	}

	var commits []CommitInfo
	err = commitIter.ForEach(func(c *object.Commit) error {
		commit := CommitInfo{
			Hash:      c.Hash.String(),
			Author:    c.Author.Name,
			Message:   c.Message,
			Timestamp: c.Author.When,
		}
		commits = append(commits, commit)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// SyncRepository synchronizes a repository with its remote
func (m *RepositoryManager) SyncRepository(ctx context.Context, id int64) error {
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

	// Clean the worktree
	err = worktree.Clean(&git.CleanOptions{
		Dir: true,
	})
	if err != nil {
		return fmt.Errorf("failed to clean worktree: %w", err)
	}

	// Reset any local changes
	err = worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		return fmt.Errorf("failed to reset worktree: %w", err)
	}

	// Fetch all branches and tags
	err = repo.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{
			config.RefSpec("+refs/heads/*:refs/remotes/origin/*"),
			config.RefSpec("+refs/tags/*:refs/tags/*"),
		},
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Get all remote branches
	remoteRefs, err := repo.References()
	if err != nil {
		return fmt.Errorf("failed to get references: %w", err)
	}

	// Track which branches we've processed
	processedBranches := make(map[string]bool)

	// Process each remote branch
	err = remoteRefs.ForEach(func(ref *plumbing.Reference) error {
		// Skip non-remote branches
		if !ref.Name().IsRemote() {
			return nil
		}

		// Get the local branch name
		localBranchName := strings.TrimPrefix(ref.Name().String(), "refs/remotes/origin/")
		if processedBranches[localBranchName] {
			return nil
		}
		processedBranches[localBranchName] = true

		// Create or update local branch
		localBranch := plumbing.NewBranchReferenceName(localBranchName)

		// Check if local branch exists
		_, err := repo.Reference(localBranch, true)
		if err != nil {
			// Create new branch
			err = repo.Storer.SetReference(plumbing.NewHashReference(localBranch, ref.Hash()))
			if err != nil {
				return fmt.Errorf("failed to create local branch %s: %w", localBranchName, err)
			}

			// Set up tracking using git command
			cmd := exec.Command("git", "branch", "--set-upstream-to=origin/"+localBranchName, localBranchName)
			cmd.Dir = metadata.LocalPath
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to set up tracking for branch %s: %w", localBranchName, err)
			}
		} else {
			// Update existing branch
			err = repo.Storer.SetReference(plumbing.NewHashReference(localBranch, ref.Hash()))
			if err != nil {
				return fmt.Errorf("failed to update local branch %s: %w", localBranchName, err)
			}

			// Update tracking using git command
			cmd := exec.Command("git", "branch", "--set-upstream-to=origin/"+localBranchName, localBranchName)
			cmd.Dir = metadata.LocalPath
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update tracking for branch %s: %w", localBranchName, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to process branches: %w", err)
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
