package repository

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Service handles repository-related HTTP requests
type Service struct {
	manager *RepositoryManager
}

// NewService creates a new repository service
func NewService(manager *RepositoryManager) *Service {
	return &Service{
		manager: manager,
	}
}

// SyncRepository godoc
// @Summary Sync repository with remote
// @Description Synchronizes a repository with its remote, fetching all branches and updating local state
// @Tags repositories
// @Accept json
// @Produce json
// @Param id path int true "Repository ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /repositories/{id}/sync [post]
func (s *Service) SyncRepository(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid repository ID",
		})
	}

	err = s.manager.SyncRepository(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to sync repository: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "repository synced successfully",
	})
}
