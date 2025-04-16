package handlers

import (
	"io"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	configservice "github.com/vanhcao3/pipeslicerCI/internal/ci/services/config"
)

// SetupConfig registers the config endpoints
func SetupConfig(app *fiber.App) {
	configGroup := app.Group("/config")

	// Initialize config manager
	manager, err := configservice.NewConfigManager(configservice.PostgresConnectionString)
	if err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}

	// Register routes
	configGroup.Get("/services", getConfigServices(manager))
	configGroup.Get("/services/:service/environments", getEnvironments(manager))
	configGroup.Get("/services/:service/environments/:environment/values", getValues(manager))
	configGroup.Post("/services/:service/environments/:environment/values", setValue(manager))
	configGroup.Delete("/services/:service/environments/:environment/values/:key", deleteValue(manager))
	configGroup.Get("/services/:service/environments/:environment/env", generateEnvFile(manager))
	configGroup.Post("/import", importConfig(manager))
	configGroup.Get("/export", exportConfig(manager))
}

// ConfigValueResponse represents a configuration value in API responses
type ConfigValueResponse struct {
	ID          int64     `json:"id"`
	Service     string    `json:"service"`
	Environment string    `json:"environment"`
	Key         string    `json:"key"`
	Value       string    `json:"value,omitempty"`
	IsSecret    bool      `json:"isSecret"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// getEnvironments returns a handler for getting all environments
func getEnvironments(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		environments, err := manager.GetEnvironments(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get environments: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"environments": environments,
		})
	}
}

// getConfigServices returns a handler for getting all services with configuration
func getConfigServices(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		services, err := manager.GetServices(c.Context())
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

// getValues returns a handler for getting all configuration values for a service and environment
func getValues(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		environment := c.Params("environment")
		if service == "" || environment == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service and environment parameters are required",
			})
		}

		config, err := manager.GetServiceConfig(c.Context(), service, environment)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to get service config: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"service":     service,
			"environment": environment,
			"config":      config,
		})
	}
}

// SetConfigValueRequest represents the request body for setting a configuration value
type SetConfigValueRequest struct {
	Key      string `json:"key" form:"key"`
	Value    string `json:"value" form:"value"`
	IsSecret bool   `json:"isSecret" form:"isSecret"`
}

// setValue returns a handler for setting a configuration value
func setValue(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		environment := c.Params("environment")
		if service == "" || environment == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service and environment parameters are required",
			})
		}

		var req SetConfigValueRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		if req.Key == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Key is required",
			})
		}

		// If setting a secret, check authorization
		if req.IsSecret {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(401).JSON(fiber.Map{
					"error": "Authorization required to set secret values",
				})
			}
		}

		err := manager.SetValue(c.Context(), service, environment, req.Key, req.Value, req.IsSecret)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to set value: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":     "Value set successfully",
			"service":     service,
			"environment": environment,
			"key":         req.Key,
			"isSecret":    req.IsSecret,
		})
	}
}

// deleteValue returns a handler for deleting a configuration value
func deleteValue(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		environment := c.Params("environment")
		key := c.Params("key")
		if service == "" || environment == "" || key == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service, environment, and key parameters are required",
			})
		}

		// Check if the value exists and if it's a secret
		value, err := manager.GetValue(c.Context(), service, environment, key)
		if err == nil && value.IsSecret {
			// If deleting a secret, check authorization
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(401).JSON(fiber.Map{
					"error": "Authorization required to delete secret values",
				})
			}
		}

		err = manager.DeleteValue(c.Context(), service, environment, key)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to delete value: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":     "Value deleted successfully",
			"service":     service,
			"environment": environment,
			"key":         key,
		})
	}
}

// importConfig returns a handler for importing configuration values
func importConfig(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authorization required to import configuration",
			})
		}

		// Get the request body
		jsonStr := string(c.Body())
		if jsonStr == "" {
			// Try to get the file from form
			file, err := c.FormFile("file")
			if err != nil {
				return c.Status(400).JSON(fiber.Map{
					"error": "No JSON data provided",
				})
			}

			// Open the file
			f, err := file.Open()
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to open uploaded file: " + err.Error(),
				})
			}
			defer f.Close()

			// Read the file
			data, err := io.ReadAll(f)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"error": "Failed to read file: " + err.Error(),
				})
			}

			jsonStr = string(data)
		}

		// Import the configuration
		err := manager.ImportConfig(c.Context(), jsonStr)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to import configuration: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Configuration imported successfully",
		})
	}
}

// exportConfig returns a handler for exporting configuration values
func exportConfig(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check authorization for exporting secrets
		includeSecrets := c.Query("includeSecrets") == "true"
		if includeSecrets {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(401).JSON(fiber.Map{
					"error": "Authorization required to export secrets",
				})
			}
		}

		// Export the configuration
		jsonStr, err := manager.ExportConfig(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to export configuration: " + err.Error(),
			})
		}

		// Check if the client wants to download the file
		if c.Query("download") == "true" {
			c.Set("Content-Disposition", "attachment; filename=config.json")
			c.Set("Content-Type", "application/json")
			return c.SendString(jsonStr)
		}

		return c.JSON(fiber.Map{
			"config": jsonStr,
		})
	}
}

// generateEnvFile returns a handler for generating a .env file
func generateEnvFile(manager *configservice.ConfigManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		service := c.Params("service")
		environment := c.Params("environment")
		if service == "" || environment == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "Service and environment parameters are required",
			})
		}

		// Check authorization for including secrets
		includeSecrets := true
		if includeSecrets {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(401).JSON(fiber.Map{
					"error": "Authorization required to include secrets in .env file",
				})
			}
		}

		// Generate the .env file
		envContent, err := manager.GenerateEnvFile(c.Context(), service, environment)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to generate .env file: " + err.Error(),
			})
		}

		// Check if the client wants to download the file
		if c.Query("download") == "true" {
			c.Set("Content-Disposition", "attachment; filename=.env")
			c.Set("Content-Type", "text/plain")
			return c.SendString(envContent)
		}

		return c.JSON(fiber.Map{
			"service":     service,
			"environment": environment,
			"env":         envContent,
		})
	}
}
