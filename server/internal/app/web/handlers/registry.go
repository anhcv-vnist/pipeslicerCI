package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/models"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/repository"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupRegistry registers the registry endpoints
func SetupRegistry(app *fiber.App) {
	// Initialize database connection
	db, err := gorm.Open(postgres.Open(config.PostgresConnectionString), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Registry{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Initialize repositories
	registryRepo := repository.NewRegistryRepository(db)

	// Initialize services
	registryService := services.NewRegistryService(registryRepo)

	// Initialize handlers
	registryHandler := NewRegistryHandler(registryService)

	// Register routes
	registryHandler.RegisterRoutes(app)
}

// RegistryHandler handles HTTP requests for registry operations
type RegistryHandler struct {
	service *services.RegistryService
}

// NewRegistryHandler creates a new RegistryHandler instance
func NewRegistryHandler(service *services.RegistryService) *RegistryHandler {
	return &RegistryHandler{service: service}
}

// RegisterRoutes registers the registry routes
func (h *RegistryHandler) RegisterRoutes(app *fiber.App) {
	registry := app.Group("/registries")
	registry.Post("", h.CreateRegistry)
	registry.Get("", h.ListRegistries)
	registry.Get("/:id", h.GetRegistry)
	registry.Put("/:id", h.UpdateRegistry)
	registry.Delete("/:id", h.DeleteRegistry)

	// Connection test endpoint
	registry.Post("/:id/test-connection", h.TestConnection)

	// WebSocket connection test endpoint
	registry.Get("/:id/test-connection-ws", websocket.New(h.TestConnectionWS))

	// List Docker images endpoint
	registry.Get("/:id/images", h.ListImages)

	// Image management endpoints
	registry.Get("/:id/images/:image/:tag", h.GetImageDetail)
	registry.Post("/:id/images/retag", h.RetagImage)
	registry.Delete("/:id/images/:image/:tag", h.DeleteImage)
	registry.Post("/images/copy", h.CopyImage)
}

// CreateRegistryRequest represents the request body for creating a registry
type CreateRegistryRequest struct {
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Description string `json:"description"`
}

// CreateRegistry handles the creation of a new registry
func (h *RegistryHandler) CreateRegistry(c *fiber.Ctx) error {
	var req CreateRegistryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	registry := &models.Registry{
		Name:        req.Name,
		URL:         req.URL,
		Username:    req.Username,
		Password:    req.Password,
		Description: req.Description,
	}

	if err := h.service.CreateRegistry(c.Context(), registry); err != nil {
		// Check for duplicate name error
		if err.Error() == "registry with this name already exists" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "A registry with this name already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(registry)
}

// ListRegistries handles retrieving all registries
func (h *RegistryHandler) ListRegistries(c *fiber.Ctx) error {
	registries, err := h.service.ListRegistries(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(registries)
}

// GetRegistry handles retrieving a specific registry
func (h *RegistryHandler) GetRegistry(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	registry, err := h.service.GetRegistry(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(registry)
}

// UpdateRegistryRequest represents the request body for updating a registry
type UpdateRegistryRequest struct {
	Name        string `json:"name" validate:"required"`
	URL         string `json:"url" validate:"required"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Description string `json:"description"`
}

// UpdateRegistry handles updating a registry
func (h *RegistryHandler) UpdateRegistry(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	var req UpdateRegistryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	registry := &models.Registry{
		ID:          uint(id),
		Name:        req.Name,
		URL:         req.URL,
		Username:    req.Username,
		Password:    req.Password,
		Description: req.Description,
	}

	if err := h.service.UpdateRegistry(c.Context(), registry); err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(registry)
}

// DeleteRegistry handles deleting a registry
func (h *RegistryHandler) DeleteRegistry(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	if err := h.service.DeleteRegistry(c.Context(), uint(id)); err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// TestConnection handles testing the connection to a registry
func (h *RegistryHandler) TestConnection(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	status, err := h.service.TestConnection(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(status)
}

// TestConnectionWS handles WebSocket connection for testing registry connectivity
func (h *RegistryHandler) TestConnectionWS(c *websocket.Conn) {
	// Get the Fiber context from the WebSocket connection
	ctx := c.Locals("ctx").(*fiber.Ctx)

	// Extract registry ID from path
	id, err := ctx.ParamsInt("id")
	if err != nil {
		c.WriteJSON(fiber.Map{
			"error": "Invalid registry ID",
		})
		c.Close()
		return
	}

	// Get the registry
	_, err = h.service.GetRegistry(ctx.Context(), uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			c.WriteJSON(fiber.Map{
				"error": "Registry not found",
			})
			c.Close()
			return
		}
		c.WriteJSON(fiber.Map{
			"error": err.Error(),
		})
		c.Close()
		return
	}

	// Send initial message
	c.WriteJSON(fiber.Map{
		"status":  "connecting",
		"message": "Testing connection to registry...",
	})

	// Test connection in a loop
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial test
	status, err := h.service.TestConnection(ctx.Context(), uint(id))
	if err != nil {
		c.WriteJSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	} else {
		c.WriteJSON(status)
	}

	// Listen for client messages (for closing the connection)
	go func() {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// Continue testing connection periodically
	for {
		select {
		case <-ticker.C:
			status, err := h.service.TestConnection(ctx.Context(), uint(id))
			if err != nil {
				c.WriteJSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
					"time":    time.Now().Format(time.RFC3339),
				})
			} else {
				// Add timestamp to the response
				response := fiber.Map{
					"status":  status.Status,
					"message": status.Message,
					"time":    time.Now().Format(time.RFC3339),
				}
				c.WriteJSON(response)
			}
		case <-ctx.Context().Done():
			return
		}
	}
}

// ListImages handles retrieving all Docker images from a registry
func (h *RegistryHandler) ListImages(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	images, err := h.service.ListImages(c.Context(), uint(id))
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(images)
}

// GetImageDetail handles retrieving detailed information about a Docker image
func (h *RegistryHandler) GetImageDetail(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	imageName := c.Params("image")
	tag := c.Params("tag")

	detail, err := h.service.GetImageDetail(c.Context(), uint(id), imageName, tag)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(detail)
}

// RetagImageRequest represents the request to retag a Docker image
type RetagImageRequest struct {
	SourceImage      string `json:"source_image"`
	SourceTag        string `json:"source_tag"`
	DestinationImage string `json:"destination_image"`
	DestinationTag   string `json:"destination_tag"`
}

// RetagImage handles retagging a Docker image
func (h *RegistryHandler) RetagImage(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	var req RetagImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Log the retag request
	fmt.Printf("Retag request: registry ID=%d, source=%s:%s, destination=%s:%s\n",
		id, req.SourceImage, req.SourceTag, req.DestinationImage, req.DestinationTag)

	// Convert handler's RetagImageRequest to service's RetagImageRequest
	serviceReq := services.RetagImageRequest{
		SourceImage:      req.SourceImage,
		SourceTag:        req.SourceTag,
		DestinationImage: req.DestinationImage,
		DestinationTag:   req.DestinationTag,
	}

	err = h.service.RetagImage(c.Context(), uint(id), serviceReq)
	if err != nil {
		// Log the detailed error
		fmt.Printf("Error retagging image: %v\n", err)

		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// DeleteImage handles deleting a Docker image
func (h *RegistryHandler) DeleteImage(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid registry ID",
		})
	}

	imageName := c.Params("image")
	tag := c.Params("tag")

	err = h.service.DeleteImage(c.Context(), uint(id), imageName, tag)
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// CopyImageRequest represents the request to copy a Docker image between registries
type CopyImageRequest struct {
	SourceRegistryID      uint   `json:"source_registry_id"`
	SourceImage           string `json:"source_image"`
	SourceTag             string `json:"source_tag"`
	DestinationRegistryID uint   `json:"destination_registry_id"`
	DestinationImage      string `json:"destination_image"`
	DestinationTag        string `json:"destination_tag"`
}

// CopyImage handles copying a Docker image between registries
func (h *RegistryHandler) CopyImage(c *fiber.Ctx) error {
	var req CopyImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.service.CopyImage(c.Context(), services.CopyImageRequest(req))
	if err != nil {
		if errors.Is(err, repository.ErrRegistryNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Registry not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
