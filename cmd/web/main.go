package main

import (
	"github.com/vanhcao3/pipeslicerCI/internal/app/web/handlers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	handlers.SetupPipelines(app)

	app.Listen(":3000")
}