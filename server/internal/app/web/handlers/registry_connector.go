package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vanhcao3/pipeslicerCI/internal/ci/services/registry"
)

// SetupRegistryConnector registers the registry connector endpoints
func SetupRegistryConnector(app *fiber.App) {
	registryConnectorGroup := app.Group("/registry-connector")

	// Initialize registry connector
	connector := registry.NewRegistryConnector()

	// Register routes
	registryConnectorGroup.Post("/authenticate", authenticateRegistry(connector))
	registryConnectorGroup.Post("/push", pushImage(connector))
	registryConnectorGroup.Get("/repositories", listRepositories(connector))
	registryConnectorGroup.Get("/repositories/:repository/tags", listTags(connector))
	registryConnectorGroup.Delete("/repositories/:repository/tags/:tag", deleteTag(connector))
}

// RegistryConfigRequest represents the request body for registry operations
type RegistryConfigRequest struct {
	Type     string `json:"type" form:"type"`
	URL      string `json:"url" form:"url"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Insecure bool   `json:"insecure" form:"insecure"`
}

// PushImageRequest represents the request body for pushing an image
type PushImageRequest struct {
	RegistryConfigRequest
	ImageName string `json:"imageName" form:"imageName"`
}

// authenticateRegistry returns a handler for authenticating with a Docker registry
func authenticateRegistry(connector *registry.RegistryConnector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegistryConfigRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.Username == "" || req.Password == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "URL, username, and password are required",
			})
		}

		// Determine registry type if not provided
		registryType := registry.RegistryType(req.Type)
		if registryType == "" {
			registryType = registry.GetRegistryType(req.URL)
		}

		// Create registry config
		config := registry.RegistryConfig{
			Type:     registryType,
			URL:      req.URL,
			Username: req.Username,
			Password: req.Password,
			Insecure: req.Insecure,
		}

		// Authenticate with the registry
		token, err := connector.Authenticate(c.Context(), config)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication failed: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":      "Authentication successful",
			"token":        token,
			"registryType": string(registryType),
			"registryUrl":  registry.GetRegistryURL(config),
		})
	}
}

// pushImage returns a handler for pushing an image to a Docker registry
func pushImage(connector *registry.RegistryConnector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req PushImageRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate required fields
		if req.URL == "" || req.Username == "" || req.Password == "" || req.ImageName == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "URL, username, password, and imageName are required",
			})
		}

		// Determine registry type if not provided
		registryType := registry.RegistryType(req.Type)
		if registryType == "" {
			registryType = registry.GetRegistryType(req.URL)
		}

		// Create registry config
		config := registry.RegistryConfig{
			Type:     registryType,
			URL:      req.URL,
			Username: req.Username,
			Password: req.Password,
			Insecure: req.Insecure,
		}

		// Push the image
		err := connector.PushImage(c.Context(), config, req.ImageName)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to push image: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":   "Image pushed successfully",
			"imageName": req.ImageName,
			"registry":  req.URL,
		})
	}
}

// listRepositories returns a handler for listing repositories in a Docker registry
func listRepositories(connector *registry.RegistryConnector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get registry config from query parameters
		url := c.Query("url")
		username := c.Query("username")
		password := c.Query("password")
		typeStr := c.Query("type")
		insecure := c.Query("insecure") == "true"

		// Validate required fields
		if url == "" || username == "" || password == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "URL, username, and password are required",
			})
		}

		// Determine registry type if not provided
		registryType := registry.RegistryType(typeStr)
		if registryType == "" {
			registryType = registry.GetRegistryType(url)
		}

		// Create registry config
		config := registry.RegistryConfig{
			Type:     registryType,
			URL:      url,
			Username: username,
			Password: password,
			Insecure: insecure,
		}

		// List repositories
		repositories, err := connector.ListRepositories(c.Context(), config)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list repositories: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"repositories": repositories,
			"registry":     url,
		})
	}
}

// listTags returns a handler for listing tags for a repository in a Docker registry
func listTags(connector *registry.RegistryConnector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get repository from path parameter
		repository := c.Params("repository")
		if repository == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Repository parameter is required",
			})
		}

		// Get registry config from query parameters
		url := c.Query("url")
		username := c.Query("username")
		password := c.Query("password")
		typeStr := c.Query("type")
		insecure := c.Query("insecure") == "true"

		// Validate required fields
		if url == "" || username == "" || password == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "URL, username, and password are required",
			})
		}

		// Determine registry type if not provided
		registryType := registry.RegistryType(typeStr)
		if registryType == "" {
			registryType = registry.GetRegistryType(url)
		}

		// Create registry config
		config := registry.RegistryConfig{
			Type:     registryType,
			URL:      url,
			Username: username,
			Password: password,
			Insecure: insecure,
		}

		// List tags
		tags, err := connector.ListTags(c.Context(), config, repository)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to list tags: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"repository": repository,
			"tags":       tags,
			"registry":   url,
		})
	}
}

// deleteTag returns a handler for deleting a tag from a repository in a Docker registry
func deleteTag(connector *registry.RegistryConnector) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get repository and tag from path parameters
		repository := c.Params("repository")
		tag := c.Params("tag")
		if repository == "" || tag == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Repository and tag parameters are required",
			})
		}

		// Get registry config from query parameters
		url := c.Query("url")
		username := c.Query("username")
		password := c.Query("password")
		typeStr := c.Query("type")
		insecure := c.Query("insecure") == "true"

		// Validate required fields
		if url == "" || username == "" || password == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "URL, username, and password are required",
			})
		}

		// Determine registry type if not provided
		registryType := registry.RegistryType(typeStr)
		if registryType == "" {
			registryType = registry.GetRegistryType(url)
		}

		// Create registry config
		config := registry.RegistryConfig{
			Type:     registryType,
			URL:      url,
			Username: username,
			Password: password,
			Insecure: insecure,
		}

		// Delete tag
		err := connector.DeleteTag(c.Context(), config, repository, tag)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to delete tag: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":    "Tag deleted successfully",
			"repository": repository,
			"tag":        tag,
			"registry":   url,
		})
	}
}
