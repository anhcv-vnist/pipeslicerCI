package handlers

import (
	//"fmt"
	"log"
	"io"

	"github.com/vanhcao3/pipeslicerCI/internal/ci"
	"github.com/gofiber/fiber/v2"
)

func SetupPipelines(app *fiber.App) {
	pipelinesGroup := app.Group("/pipelines")

	pipelinesGroup.Post("/build", postBuild)
}

type RequestBody struct {
	Url    string `json:"url" xml:"url" form:"url"`
	Branch string `json:"branch" xml:"branch" form:"branch"`
}

func postBuild(c *fiber.Ctx) error {
	url := c.FormValue("url")
	branch := c.FormValue("branch")

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("Failed to read uploaded file: %v", err)
		return c.Status(400).SendString("Invalid file upload: " + err.Error())
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).SendString("Failed to open uploaded file: " + err.Error())
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(500).SendString("Failed to read file: " + err.Error())
	}

	log.Printf("Received request: URL=%s, Branch=%s, File Size=%d", url, branch, len(data))

	ws, err := ci.NewWorkspaceFromGit("/tmp", url, branch)
	if err != nil {
		return c.Status(500).SendString("Failed to create workspace: " + err.Error())
	}

	pipeline, err := ws.LoadPipeline(data)
	if err != nil {
		return c.Status(400).SendString("Invalid YAML file: " + err.Error())
	}

	executor := ci.NewExecutor(ws)
	output, err := executor.Run(c.UserContext(), pipeline)
	if err != nil {
		return c.Status(500).SendString("Pipeline execution failed: " + err.Error())
	}

	return c.SendString("Successfully executed pipeline.\n" + output)
}