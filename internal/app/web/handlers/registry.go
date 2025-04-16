package handlers

import (
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/config"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/registry"
)

// SetupRegistry registers the registry endpoints
func SetupRegistry(app *fiber.App) {
	registryGroup := app.Group("/registry")

	// Initialize registry manager
	manager, err := registry.NewRegistryManager(config.PostgresConnectionString)
	if err != nil {
		log.Fatalf("Failed to initialize registry manager: %v", err)
	}

	// Register routes
	registryGroup.Get("/services", getServices(manager))
	registryGroup.Get("/services/:service/tags", getServiceTags(manager))
	registryGroup.Get("/services/:service/history", getServiceHistory(manager))
	registryGroup.Get("/services/:service/latest", getLatestImage(manager))
	registryGroup.Get("/services/:service/tags/:tag", getImageByTag(manager))
	registryGroup.Post("/services/:service/tags", createTag(manager))
	registryGroup.Post("/images", recordImage(manager))
	registryGroup.Delete("/images/:id", deleteImage(manager))
}

// ImageResponse represents the response for an image
type ImageResponse struct {
	ID        int64     `json:"id"`
	Service   string    `json:"service"`
	Tag       string    `json:"tag"`
	Commit    string    `json:"commit"`
	Branch    string    `json:"branch"`
	BuildTime time.Time `json:"buildTime"`
	Status    string    `json:"status"`
	Registry  string    `json:"registry"`
	ImageName string    `json:"imageName"`
}

// getServices returns a handler for getting all services
func getServices(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		services, err := manager.GetServiceList(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get services: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"services": services,
		})
	}
}

// getServiceTags returns a handler for getting all tags for a service
func getServiceTags(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		if service == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service parameter is required",
			})
		}

		tags, err := manager.GetTagsForService(c.Context(), service)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get tags: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"service": service,
			"tags":    tags,
		})
	}
}

// getServiceHistory returns a handler for getting the build history for a service
func getServiceHistory(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		if service == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service parameter is required",
			})
		}

		limitStr := c.Query("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid limit parameter: " + err.Error(),
			})
		}

		images, err := manager.GetImageHistory(c.Context(), service, limit)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get image history: " + err.Error(),
			})
		}

		// Convert to response format
		var response []ImageResponse
		for _, img := range images {
			response = append(response, ImageResponse{
				ID:        img.ID,
				Service:   img.Service,
				Tag:       img.Tag,
				Commit:    img.Commit,
				Branch:    img.Branch,
				BuildTime: img.BuildTime,
				Status:    img.Status,
				Registry:  img.Registry,
				ImageName: img.ImageName,
			})
		}

		return c.JSON(fiber.Map{
			"service": service,
			"images":  response,
		})
	}
}

// getLatestImage returns a handler for getting the latest image for a service
func getLatestImage(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		if service == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service parameter is required",
			})
		}

		branch := c.Query("branch", "main")

		image, err := manager.GetLatestImage(c.Context(), service, branch)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Convert to response format
		response := ImageResponse{
			ID:        image.ID,
			Service:   image.Service,
			Tag:       image.Tag,
			Commit:    image.Commit,
			Branch:    image.Branch,
			BuildTime: image.BuildTime,
			Status:    image.Status,
			Registry:  image.Registry,
			ImageName: image.ImageName,
		}

		return c.JSON(response)
	}
}

// getImageByTag returns a handler for getting an image by tag
func getImageByTag(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		tag := c.Params("tag")
		if service == "" || tag == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service and tag parameters are required",
			})
		}

		image, err := manager.GetImageByTag(c.Context(), service, tag)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Convert to response format
		response := ImageResponse{
			ID:        image.ID,
			Service:   image.Service,
			Tag:       image.Tag,
			Commit:    image.Commit,
			Branch:    image.Branch,
			BuildTime: image.BuildTime,
			Status:    image.Status,
			Registry:  image.Registry,
			ImageName: image.ImageName,
		}

		return c.JSON(response)
	}
}

// TagRequest represents the request body for creating a new tag
type TagRequest struct {
	SourceTag string `json:"sourceTag" form:"sourceTag"`
	NewTag    string `json:"newTag" form:"newTag"`
}

// createTag returns a handler for creating a new tag
func createTag(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		if service == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service parameter is required",
			})
		}

		var req TagRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		if req.SourceTag == "" || req.NewTag == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Source tag and new tag are required",
			})
		}

		err := manager.TagImage(c.Context(), service, req.SourceTag, req.NewTag)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create tag: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":   "Tag created successfully",
			"service":   service,
			"sourceTag": req.SourceTag,
			"newTag":    req.NewTag,
		})
	}
}

// RecordImageRequest represents the request body for recording an image
type RecordImageRequest struct {
	Service   string    `json:"service" form:"service"`
	Tag       string    `json:"tag" form:"tag"`
	Commit    string    `json:"commit" form:"commit"`
	Branch    string    `json:"branch" form:"branch"`
	BuildTime time.Time `json:"buildTime" form:"buildTime"`
	Status    string    `json:"status" form:"status"`
	Registry  string    `json:"registry" form:"registry"`
	ImageName string    `json:"imageName" form:"imageName"`
}

// recordImage returns a handler for recording an image
func recordImage(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RecordImageRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.Service == "" || req.Tag == "" || req.Commit == "" || req.Branch == "" || req.Status == "" || req.Registry == "" || req.ImageName == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "All fields are required",
			})
		}

		// If build time is not provided, use current time
		if req.BuildTime.IsZero() {
			req.BuildTime = time.Now()
		}

		// Convert to metadata
		metadata := registry.ImageMetadata{
			Service:   req.Service,
			Tag:       req.Tag,
			Commit:    req.Commit,
			Branch:    req.Branch,
			BuildTime: req.BuildTime,
			Status:    req.Status,
			Registry:  req.Registry,
			ImageName: req.ImageName,
		}

		// Record the image
		err := manager.RecordImage(c.Context(), metadata)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to record image: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Image recorded successfully",
			"service": req.Service,
			"tag":     req.Tag,
		})
	}
}

// deleteImage returns a handler for deleting an image
func deleteImage(manager *registry.RegistryManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		idStr := c.Params("id")
		if idStr == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "ID parameter is required",
			})
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid ID parameter: " + err.Error(),
			})
		}

		err = manager.DeleteImage(c.Context(), id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to delete image: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Image deleted successfully",
			"id":      id,
		})
	}
}
