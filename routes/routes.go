package routes

import (
	"github.com/gofiber/fiber/v2"
)

func Register(app *fiber.App) {
	// API routes MUST come before web routes.
	// The web router has a wildcard "/*" fallback that would otherwise
	// catch all /api/... requests and return index.html instead.
	RegisterAPIRoutes(app)
	RegisterWebRoutes(app)
}
