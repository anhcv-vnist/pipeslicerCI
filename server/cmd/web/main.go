package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/vanhcao3/pipeslicerCI/internal/app/web/docs"
	"github.com/vanhcao3/pipeslicerCI/internal/app/web/handlers"
)

func main() {
	app := fiber.New(fiber.Config{
		// Increase the maximum request body size to handle large file uploads
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// Add middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup routes
	handlers.SetupPipelines(app)
	handlers.SetupImageBuilder(app)
	handlers.SetupRegistry(app)
	handlers.SetupConfig(app)
	handlers.SetupRegistryConnector(app)
	handlers.SetupRepository(app)
	handlers.SetupMetrics(app)

	// Setup Swagger documentation
	docs.SetupSwagger(app)

	// Start server
	app.Listen(":8080")
}
