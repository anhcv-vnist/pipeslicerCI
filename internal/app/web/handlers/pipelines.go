package handlers

import (
	"fmt"
	"log"

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
	body := &RequestBody{}

	if err := c.BodyParser(body); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(400).SendString("Invalid request body: " + err.Error())
	}

	log.Printf("Received request: URL=%s, Branch=%s", body.Url, body.Branch)

	var ws ci.Workspace
	ws, err := ci.NewWorkspaceFromGit("/tmp", body.Url, body.Branch)
	if err != nil {
		log.Printf("Failed to create workspace: %v", err)
		return c.Status(500).SendString("Failed to create workspace: " + err.Error())
	}

	log.Printf("Workspace created: Dir=%s, Branch=%s, Commit=%s", ws.Dir(), ws.Branch(), ws.Commit())

	executor := ci.NewExecutor(ws)
	output, err := executor.RunDefault(c.UserContext())
	if err != nil {
		log.Printf("Pipeline execution failed: %v", err)
		return c.Status(500).SendString("Pipeline execution failed: " + err.Error())
	}

	log.Printf("Pipeline executed successfully: %s", output)

	return c.SendString(
		fmt.Sprintf(
			"Successfully executed pipeline.\n%s\n\nFrom branch: %s\nCommit: %s\nIn directory: %s\n",
			output,
			ws.Branch(),
			ws.Commit(),
			ws.Dir(),
		),
	)
}