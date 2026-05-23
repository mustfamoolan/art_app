package routes

import (
	"Continuous_b/app/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterAPIRoutes(app *fiber.App) {
	api := app.Group("/api")

	// V1 Grouping
	v1 := api.Group("/v1")

	// Health check
	v1.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": "v1"})
	})

	// String Art: POST /api/v1/process
	// Accepts: multipart/form-data with "image" field
	// Returns: {"steps": [...], "total_pegs": 200, "total_lines": 3000}
	artController := controllers.NewArtController()
	v1.Post("/process", artController.Process)
}
