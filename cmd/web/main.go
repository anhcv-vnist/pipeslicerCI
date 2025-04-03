package main

import (
	"github.com/vanhcao3/pipeslicerCI/internal/app/web/docs"
	"github.com/vanhcao3/pipeslicerCI/internal/app/web/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New(fiber.Config{
		// Increase the maximum request body size to handle large file uploads
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// Add middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup routes
	handlers.SetupPipelines(app)
	handlers.SetupImageBuilder(app)
	handlers.SetupRegistry(app)
	handlers.SetupConfig(app)
	handlers.SetupRegistryConnector(app)

	// Setup Swagger documentation
	docs.SetupSwagger(app)

	// Start server
	app.Listen(":3000")
}
