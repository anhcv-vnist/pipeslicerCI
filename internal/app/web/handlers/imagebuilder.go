package handlers

import (
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanhcao3/pipeslicerCI/internal/ci"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/config"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/imagebuilder"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupImageBuilder registers the image builder endpoints
func SetupImageBuilder(app *fiber.App) {
	imageBuilderGroup := app.Group("/imagebuilder")

	// Initialize repository manager
	baseDir := "/home/anhcv/workspace/repositories"
	db, err := gorm.Open(postgres.Open(config.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	repoManager, err := repository.NewRepositoryManager(db, baseDir)
	if err != nil {
		log.Fatalf("Failed to initialize repository manager: %v", err)
	}

	imageBuilderGroup.Post("/build", postBuildImage(repoManager))
	imageBuilderGroup.Post("/build-multiple", postBuildMultipleImages(repoManager))
	imageBuilderGroup.Post("/detect-changes", postDetectChanges(repoManager))
	imageBuilderGroup.Get("/branches", getBranches(repoManager))
}

// BuildImageRequest represents the request body for building a Docker image
type BuildImageRequest struct {
	URL         string `json:"url" form:"url"`
	Branch      string `json:"branch" form:"branch"`
	ServicePath string `json:"servicePath" form:"servicePath"`
	Tag         string `json:"tag" form:"tag"`
	Registry    string `json:"registry" form:"registry"`
	Username    string `json:"username" form:"username"`
	Password    string `json:"password" form:"password"`
}

// BuildImageResponse represents the response for a Docker image build
type BuildImageResponse struct {
	Service   string    `json:"service"`
	Tag       string    `json:"tag"`
	Commit    string    `json:"commit"`
	Branch    string    `json:"branch"`
	BuildTime time.Time `json:"buildTime"`
	Success   bool      `json:"success"`
	Output    string    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// postBuildImage handles requests to build and push a Docker image
func postBuildImage(repoManager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req BuildImageRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.Branch == "" || req.ServicePath == "" || req.Registry == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Missing required fields: url, branch, servicePath, and registry are required",
			})
		}

		// Set default tag if not provided
		if req.Tag == "" {
			req.Tag = "latest"
		}

		log.Printf("Building image for service %s from %s (branch: %s)", req.ServicePath, req.URL, req.Branch)

		// Get or clone repository
		repo, err := repoManager.GetRepositoryByURL(c.Context(), req.URL)
		if err != nil {
			// Repository not found, clone it
			repo, err = repoManager.CloneRepository(c.Context(), req.URL, req.ServicePath, "Repository for building Docker images")
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to clone repository: " + err.Error(),
				})
			}
		}

		// Checkout the specified branch
		err = repoManager.CheckoutBranch(c.Context(), repo.ID, req.Branch)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to checkout branch: " + err.Error(),
			})
		}

		// Create workspace from repository
		ws, err := ci.NewWorkspaceFromPath(repo.LocalPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create workspace: " + err.Error(),
			})
		}

		// Create image builder
		builder := imagebuilder.NewImageBuilder(ws, req.Registry, req.Username, req.Password)

		// Build and push image
		result, err := builder.BuildAndPushImage(c.Context(), req.ServicePath, req.Tag)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":  "Failed to build image: " + err.Error(),
				"output": result.Output,
			})
		}

		// Convert result to response
		response := BuildImageResponse{
			Service:   result.Service,
			Tag:       result.Tag,
			Commit:    result.Commit,
			Branch:    result.Branch,
			BuildTime: result.BuildTime,
			Success:   result.Success,
			Output:    result.Output,
		}

		return c.JSON(response)
	}
}

// BuildMultipleRequest represents the request body for building multiple Docker images
type BuildMultipleRequest struct {
	URL          string   `json:"url" form:"url"`
	Branch       string   `json:"branch" form:"branch"`
	ServicePaths []string `json:"servicePaths" form:"servicePaths"`
	Tag          string   `json:"tag" form:"tag"`
	Registry     string   `json:"registry" form:"registry"`
	Username     string   `json:"username" form:"username"`
	Password     string   `json:"password" form:"password"`
}

// postBuildMultipleImages handles requests to build and push multiple Docker images
func postBuildMultipleImages(repoManager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req BuildMultipleRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.Branch == "" || len(req.ServicePaths) == 0 || req.Registry == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Missing required fields: url, branch, servicePaths, and registry are required",
			})
		}

		// Set default tag if not provided
		if req.Tag == "" {
			req.Tag = "latest"
		}

		log.Printf("Building images for %d services from %s (branch: %s)", len(req.ServicePaths), req.URL, req.Branch)

		// Get or clone repository
		repo, err := repoManager.GetRepositoryByURL(c.Context(), req.URL)
		if err != nil {
			// Repository not found, clone it
			repo, err = repoManager.CloneRepository(c.Context(), req.URL, req.ServicePaths[0], "Repository for building Docker images")
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to clone repository: " + err.Error(),
				})
			}
		}

		// Checkout the specified branch
		err = repoManager.CheckoutBranch(c.Context(), repo.ID, req.Branch)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to checkout branch: " + err.Error(),
			})
		}

		// Create workspace from repository
		ws, err := ci.NewWorkspaceFromPath(repo.LocalPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create workspace: " + err.Error(),
			})
		}

		// Create image builder
		builder := imagebuilder.NewImageBuilder(ws, req.Registry, req.Username, req.Password)

		// Build and push images
		results, err := builder.BuildMultipleServices(c.Context(), req.ServicePaths, req.Tag)
		if err != nil {
			// Find the failed service
			var failedResult *imagebuilder.ImageBuildResult
			for _, result := range results {
				if !result.Success {
					failedResult = result
					break
				}
			}

			if failedResult != nil {
				return c.Status(500).JSON(fiber.Map{
					"error":   "Failed to build image for service " + failedResult.Service + ": " + err.Error(),
					"output":  failedResult.Output,
					"service": failedResult.Service,
				})
			}

			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to build images: " + err.Error(),
			})
		}

		// Convert results to responses
		var responses []BuildImageResponse
		for _, result := range results {
			responses = append(responses, BuildImageResponse{
				Service:   result.Service,
				Tag:       result.Tag,
				Commit:    result.Commit,
				Branch:    result.Branch,
				BuildTime: result.BuildTime,
				Success:   result.Success,
				Output:    result.Output,
			})
		}

		return c.JSON(fiber.Map{
			"results": responses,
		})
	}
}

// DetectChangesRequest represents the request body for detecting changed services
type DetectChangesRequest struct {
	URL           string `json:"url" form:"url"`
	BaseBranch    string `json:"baseBranch" form:"baseBranch"`
	CurrentBranch string `json:"currentBranch" form:"currentBranch"`
	Registry      string `json:"registry" form:"registry"`
	Username      string `json:"username" form:"username"`
	Password      string `json:"password" form:"password"`
}

// postDetectChanges handles requests to detect which services have changed between branches
func postDetectChanges(repoManager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req DetectChangesRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.BaseBranch == "" || req.CurrentBranch == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Missing required fields: url, baseBranch, and currentBranch are required",
			})
		}

		log.Printf("Detecting changes between %s and %s in %s", req.BaseBranch, req.CurrentBranch, req.URL)

		// Get or clone repository
		repo, err := repoManager.GetRepositoryByURL(c.Context(), req.URL)
		if err != nil {
			log.Printf("Repository not found, attempting to clone: %v", err)
			// Repository not found, clone it
			repo, err = repoManager.CloneRepository(c.Context(), req.URL, "temp", "Repository for detecting changes")
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to clone repository: " + err.Error(),
				})
			}
		}

		log.Printf("Repository found/cloned at: %s", repo.LocalPath)

		// List available branches before checkout
		cmd := exec.Command("git", "branch", "-a")
		cmd.Dir = repo.LocalPath
		branches, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to list branches: %v", err)
		} else {
			log.Printf("Available branches:\n%s", string(branches))
		}

		// Checkout the current branch
		log.Printf("Attempting to checkout branch: %s", req.CurrentBranch)
		err = repoManager.CheckoutBranch(c.Context(), repo.ID, req.CurrentBranch)
		if err != nil {
			log.Printf("Failed to checkout branch: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to checkout branch: " + err.Error(),
			})
		}

		// Verify current branch after checkout
		cmd = exec.Command("git", "branch", "--show-current")
		cmd.Dir = repo.LocalPath
		currentBranch, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to get current branch: %v", err)
		} else {
			log.Printf("Current branch after checkout: %s", string(currentBranch))
		}

		// Create workspace from repository
		ws, err := ci.NewWorkspaceFromPath(repo.LocalPath)
		if err != nil {
			log.Printf("Failed to create workspace: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create workspace: " + err.Error(),
			})
		}

		log.Printf("Workspace created successfully at: %s", ws.Dir())

		// Create image builder
		builder := imagebuilder.NewImageBuilder(ws, req.Registry, req.Username, req.Password)

		// Detect changed services
		changedServices, err := builder.DetectChangedServices(c.Context(), req.BaseBranch, req.CurrentBranch)
		if err != nil {
			log.Printf("Failed to detect changed services: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to detect changed services: " + err.Error(),
			})
		}

		log.Printf("Detected changed services: %v", changedServices)

		return c.JSON(fiber.Map{
			"changedServices": changedServices,
		})
	}
}

// ListBranchesRequest represents the request body for listing branches
type ListBranchesRequest struct {
	URL string `json:"url" query:"url"`
}

// BranchInfo represents information about a Git branch
type BranchInfo struct {
	Name      string `json:"name"`
	IsRemote  bool   `json:"isRemote"`
	IsCurrent bool   `json:"isCurrent"`
}

// getBranches handles requests to list all branches in a repository
func getBranches(repoManager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req ListBranchesRequest
		if err := c.QueryParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request parameters: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Missing required field: url",
			})
		}

		log.Printf("Listing branches for repository: %s", req.URL)

		// Get or clone repository
		repo, err := repoManager.GetRepositoryByURL(c.Context(), req.URL)
		if err != nil {
			log.Printf("Repository not found, attempting to clone: %v", err)
			// Repository not found, clone it
			repo, err = repoManager.CloneRepository(c.Context(), req.URL, "temp", "Repository for listing branches")
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to clone repository: " + err.Error(),
				})
			}
		}

		log.Printf("Repository found/cloned at: %s", repo.LocalPath)

		// Get current branch
		cmd := exec.Command("git", "branch", "--show-current")
		cmd.Dir = repo.LocalPath
		currentBranch, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to get current branch: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get current branch: " + err.Error(),
			})
		}
		currentBranchStr := strings.TrimSpace(string(currentBranch))

		// List all branches
		cmd = exec.Command("git", "branch", "-a")
		cmd.Dir = repo.LocalPath
		branchesOutput, err := cmd.Output()
		if err != nil {
			log.Printf("Failed to list branches: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list branches: " + err.Error(),
			})
		}

		// Parse branch output
		branches := strings.Split(string(branchesOutput), "\n")
		var branchInfos []BranchInfo
		for _, branch := range branches {
			branch = strings.TrimSpace(branch)
			if branch == "" {
				continue
			}

			// Remove the * prefix if present
			isCurrent := strings.HasPrefix(branch, "* ")
			if isCurrent {
				branch = strings.TrimPrefix(branch, "* ")
			}

			// Check if it's a remote branch
			isRemote := strings.HasPrefix(branch, "remotes/origin/")
			if isRemote {
				branch = strings.TrimPrefix(branch, "remotes/origin/")
			}

			// Skip HEAD references
			if branch == "HEAD" {
				continue
			}

			branchInfos = append(branchInfos, BranchInfo{
				Name:      branch,
				IsRemote:  isRemote,
				IsCurrent: branch == currentBranchStr,
			})
		}

		return c.JSON(fiber.Map{
			"branches": branchInfos,
		})
	}
}
