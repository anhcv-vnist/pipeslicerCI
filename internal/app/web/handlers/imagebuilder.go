package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanhcao3/pipeslicerCI/internal/ci"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/imagebuilder"
)

// SetupImageBuilder registers the image builder endpoints
func SetupImageBuilder(app *fiber.App) {
	imageBuilderGroup := app.Group("/imagebuilder")

	imageBuilderGroup.Post("/build", postBuildImage)
	imageBuilderGroup.Post("/build-multiple", postBuildMultipleImages)
	imageBuilderGroup.Post("/detect-changes", postDetectChanges)
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
func postBuildImage(c *fiber.Ctx) error {
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

	// Create workspace from Git repository
	ws, err := ci.NewWorkspaceFromGit("/tmp", req.URL, req.Branch)
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
func postBuildMultipleImages(c *fiber.Ctx) error {
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

	// Create workspace from Git repository
	ws, err := ci.NewWorkspaceFromGit("/tmp", req.URL, req.Branch)
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
func postDetectChanges(c *fiber.Ctx) error {
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

	// Create workspace from Git repository
	ws, err := ci.NewWorkspaceFromGit("/tmp", req.URL, req.CurrentBranch)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create workspace: " + err.Error(),
		})
	}

	// Create image builder
	builder := imagebuilder.NewImageBuilder(ws, req.Registry, req.Username, req.Password)

	// Detect changed services
	changedServices, err := builder.DetectChangedServices(c.Context(), req.BaseBranch, req.CurrentBranch)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to detect changed services: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"changedServices": changedServices,
	})
}
