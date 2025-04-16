package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/config"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupRepository registers the repository endpoints
func SetupRepository(app *fiber.App) {
	repositoryGroup := app.Group("/repository")

	// Initialize repository manager with a directory in the workspace
	baseDir := "/home/anhcv/workspace/repositories"

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	manager, err := repository.NewRepositoryManager(db, baseDir)
	if err != nil {
		log.Fatalf("Failed to initialize repository manager: %v", err)
	}

	// Register routes
	repositoryGroup.Post("/clone", cloneRepository(manager))
	repositoryGroup.Get("/list", listAllRepositories(manager))
	repositoryGroup.Get("/:id", getRepository(manager))
	repositoryGroup.Put("/:id", updateRepository(manager))
	repositoryGroup.Delete("/:id", deleteRepository(manager))
	repositoryGroup.Post("/:id/checkout", checkoutBranch(manager))
	repositoryGroup.Get("/:id/commits", getBranchCommits(manager))
	repositoryGroup.Post("/:id/sync", syncRepository(manager))
}

// RepositoryResponse represents a repository in API responses
type RepositoryResponse struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LocalPath   string    `json:"localPath"`
	LastUpdated time.Time `json:"lastUpdated"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CloneRequest represents the request body for cloning a repository
type CloneRequest struct {
	URL         string `json:"url"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// cloneRepository returns a handler for cloning a Git repository
func cloneRepository(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CloneRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.Name == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "URL and name are required",
			})
		}

		// Clone the repository
		metadata, err := manager.CloneRepository(c.Context(), req.URL, req.Name, req.Description)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to clone repository: " + err.Error(),
			})
		}

		// Convert to response format
		response := RepositoryResponse{
			ID:          metadata.ID,
			URL:         metadata.URL,
			Name:        metadata.Name,
			Description: metadata.Description,
			LocalPath:   metadata.LocalPath,
			LastUpdated: metadata.LastUpdated,
			CreatedAt:   metadata.CreatedAt,
			UpdatedAt:   metadata.UpdatedAt,
		}

		return c.JSON(response)
	}
}

// listAllRepositories returns a handler for listing all repositories
func listAllRepositories(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		repositories, err := manager.ListRepositories(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list repositories: " + err.Error(),
			})
		}

		// Convert to response format
		var response []RepositoryResponse
		for _, repo := range repositories {
			response = append(response, RepositoryResponse{
				ID:          repo.ID,
				URL:         repo.URL,
				Name:        repo.Name,
				Description: repo.Description,
				LocalPath:   repo.LocalPath,
				LastUpdated: repo.LastUpdated,
				CreatedAt:   repo.CreatedAt,
				UpdatedAt:   repo.UpdatedAt,
			})
		}

		return c.JSON(fiber.Map{
			"repositories": response,
		})
	}
}

// getRepository returns a handler for getting a repository by ID
func getRepository(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid repository ID",
			})
		}

		metadata, err := manager.GetRepositoryByID(c.Context(), int64(id))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Convert to response format
		response := RepositoryResponse{
			ID:          metadata.ID,
			URL:         metadata.URL,
			Name:        metadata.Name,
			Description: metadata.Description,
			LocalPath:   metadata.LocalPath,
			LastUpdated: metadata.LastUpdated,
			CreatedAt:   metadata.CreatedAt,
			UpdatedAt:   metadata.UpdatedAt,
		}

		return c.JSON(response)
	}
}

// updateRepository returns a handler for updating a repository
func updateRepository(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid repository ID",
			})
		}

		metadata, err := manager.GetRepositoryByID(c.Context(), int64(id))
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Update the repository
		updated, err := manager.UpdateRepository(c.Context(), metadata)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update repository: " + err.Error(),
			})
		}

		// Convert to response format
		response := RepositoryResponse{
			ID:          updated.ID,
			URL:         updated.URL,
			Name:        updated.Name,
			Description: updated.Description,
			LocalPath:   updated.LocalPath,
			LastUpdated: updated.LastUpdated,
			CreatedAt:   updated.CreatedAt,
			UpdatedAt:   updated.UpdatedAt,
		}

		return c.JSON(response)
	}
}

// deleteRepository returns a handler for deleting a repository
func deleteRepository(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid repository ID",
			})
		}

		err = manager.DeleteRepository(c.Context(), int64(id))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to delete repository: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Repository deleted successfully",
		})
	}
}

// CheckoutRequest represents the request body for checking out a branch
type CheckoutRequest struct {
	Branch string `json:"branch"`
}

// checkoutBranch returns a handler for checking out a branch in a repository
func checkoutBranch(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid repository ID",
			})
		}

		var req CheckoutRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		if req.Branch == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Branch name is required",
			})
		}

		err = manager.CheckoutBranch(c.Context(), int64(id), req.Branch)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to checkout branch: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Branch checked out successfully",
		})
	}
}

// getBranchCommits returns a handler for getting commits of a specific branch
func getBranchCommits(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get repository ID from path parameter
		id, err := c.ParamsInt("id")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid repository ID",
			})
		}

		// Get branch name from query parameter
		branch := c.Query("branch")
		if branch == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Branch name is required. Use ?branch=your-branch-name",
			})
		}

		// Get commits for the branch
		commits, err := manager.GetBranchCommits(c.Context(), int64(id), branch)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get branch commits: " + err.Error(),
			})
		}

		return c.JSON(commits)
	}
}

// syncRepository returns a handler for syncing a repository with its remote
func syncRepository(manager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid repository ID",
			})
		}

		err = manager.SyncRepository(c.Context(), id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to sync repository: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Repository synced successfully",
		})
	}
}
