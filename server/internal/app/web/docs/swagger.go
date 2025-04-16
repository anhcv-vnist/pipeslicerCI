package docs

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed swagger-ui
var swaggerUIFiles embed.FS

//go:embed swagger.yaml
var swaggerYAML []byte

// SetupSwagger registers the Swagger documentation endpoints
func SetupSwagger(app *fiber.App) {
	// Create a modified filesystem that intercepts index.html requests
	app.Use("/swagger", func(c *fiber.Ctx) error {
		path := c.Path()

		// If requesting the root or index.html, serve our modified version
		if path == "/swagger" || path == "/swagger/" || path == "/swagger/index.html" {
			c.Set("Content-Type", "text/html")

			// Get the original HTML (need to read it from the filesystem)
			originalHTML, err := swaggerUIFiles.ReadFile("swagger-ui/index.html")
			if err != nil {
				return c.Status(500).SendString("Error reading Swagger UI HTML")
			}

			// Modify the URL to use the absolute path
			modifiedHTML := strings.Replace(
				string(originalHTML),
				`url: "swagger.yaml"`,
				`url: "/swagger.yaml"`,
				1,
			)

			return c.SendString(modifiedHTML)
		}

		// Continue with the next handler for other files
		return c.Next()
	})

	// Serve the rest of the static files
	app.Use("/swagger", filesystem.New(filesystem.Config{
		Root:       http.FS(mustSubFS(swaggerUIFiles, "swagger-ui")),
		PathPrefix: "",
		Browse:     true,
	}))

	// Serve Swagger YAML
	app.Get("/swagger.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Send(swaggerYAML)
	})

	// Also serve at /swagger/swagger.yaml for consistency
	app.Get("/swagger/swagger.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml")
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Send(swaggerYAML)
	})
}

// mustSubFS returns a sub-filesystem or panics
func mustSubFS(f fs.FS, dir string) fs.FS {
	subFS, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return subFS
}
