package handlers

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"crypto/md5"

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
	imageBuilderGroup.Post("/detect-commit-changes", postDetectCommitChanges(repoManager))
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

// DetectChangesRequest represents a request to detect changes between branches
type DetectChangesRequest struct {
	URL           string `json:"url"`
	BaseBranch    string `json:"baseBranch"`
	CurrentBranch string `json:"currentBranch"`
	Registry      string `json:"registry"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

// DetectChangesResponse represents a response from the detect changes endpoint
type DetectChangesResponse struct {
	ChangedServices []struct {
		Path          string `json:"path"`
		HasDockerfile bool   `json:"hasDockerfile"`
	} `json:"changedServices"`
}

// DetectCommitChangesRequest represents the request body for detecting changed services between commits
type DetectCommitChangesRequest struct {
	URL           string `json:"url" form:"url"`
	BaseCommit    string `json:"baseCommit" form:"baseCommit"`
	CurrentCommit string `json:"currentCommit" form:"currentCommit"`
	Registry      string `json:"registry" form:"registry"`
	Username      string `json:"username" form:"username"`
	Password      string `json:"password" form:"password"`
}

// DetectCommitChangesResponse represents a response from the detect commit changes endpoint
type DetectCommitChangesResponse struct {
	ChangedServices []struct {
		Path          string `json:"path"`
		HasDockerfile bool   `json:"hasDockerfile"`
	} `json:"changedServices"`
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

// postDetectChanges handles requests to detect which services have changed between branches
func postDetectChanges(repoManager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req DetectChangesRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		// Validate required fields
		if req.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Repository URL is required",
			})
		}
		if req.BaseBranch == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Base branch is required",
			})
		}
		if req.CurrentBranch == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Current branch is required",
			})
		}

		// Clone the repository if it doesn't exist
		repoID := fmt.Sprintf("%x", md5.Sum([]byte(req.URL)))
		repo, err := repoManager.GetRepositoryByURL(c.Context(), req.URL)
		if err != nil {
			// Repository doesn't exist, clone it
			repo, err = repoManager.CloneRepository(c.Context(), req.URL, repoID, "Repository for detecting changes")
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to clone repository: %v", err),
				})
			}
		}

		// Checkout the current branch
		err = repoManager.CheckoutBranch(c.Context(), repo.ID, req.CurrentBranch)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to checkout branch: %v", err),
			})
		}

		// Create workspace from repository
		ws, err := ci.NewWorkspaceFromPath(repo.LocalPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create workspace: %v", err),
			})
		}

		// Create image builder
		builder := imagebuilder.NewImageBuilder(ws, req.Registry, req.Username, req.Password)

		// Detect changed services
		changedServices, err := builder.DetectChangedServices(c.Context(), req.BaseBranch, req.CurrentBranch)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to detect changes: %v", err),
			})
		}

		// Convert to response format
		response := DetectChangesResponse{
			ChangedServices: make([]struct {
				Path          string `json:"path"`
				HasDockerfile bool   `json:"hasDockerfile"`
			}, len(changedServices)),
		}

		for i, service := range changedServices {
			response.ChangedServices[i] = struct {
				Path          string `json:"path"`
				HasDockerfile bool   `json:"hasDockerfile"`
			}{
				Path:          service.Path,
				HasDockerfile: service.HasDockerfile,
			}
		}

		return c.JSON(response)
	}
}

// postDetectCommitChanges handles requests to detect which services have changed between commits
func postDetectCommitChanges(repoManager *repository.RepositoryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req DetectCommitChangesRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.BaseCommit == "" || req.CurrentCommit == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Missing required fields: url, baseCommit, and currentCommit are required",
			})
		}

		log.Printf("Detecting changes between commits %s and %s in %s", req.BaseCommit, req.CurrentCommit, req.URL)

		// Get repository from database
		repo, err := repoManager.GetRepositoryByURL(c.Context(), req.URL)
		if err != nil {
			log.Printf("Repository not found in database: %v", err)
			return c.Status(404).JSON(fiber.Map{
				"error": "Repository not found in database. Please clone it first using the repository API.",
			})
		}

		log.Printf("Repository found at: %s", repo.LocalPath)

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
		changedServices, err := builder.DetectChangedServices(c.Context(), req.BaseCommit, req.CurrentCommit)
		if err != nil {
			log.Printf("Failed to detect changed services: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to detect changed services: " + err.Error(),
			})
		}

		log.Printf("Detected changed services: %v", changedServices)

		// Convert to response format
		response := DetectCommitChangesResponse{
			ChangedServices: make([]struct {
				Path          string `json:"path"`
				HasDockerfile bool   `json:"hasDockerfile"`
			}, len(changedServices)),
		}

		for i, service := range changedServices {
			response.ChangedServices[i] = struct {
				Path          string `json:"path"`
				HasDockerfile bool   `json:"hasDockerfile"`
			}{
				Path:          service.Path,
				HasDockerfile: service.HasDockerfile,
			}
		}

		return c.JSON(response)
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
