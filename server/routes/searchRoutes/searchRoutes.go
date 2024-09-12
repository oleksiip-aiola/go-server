package searchRoutes

import (
	"github.com/gofiber/fiber/v2"
)

func SearchRoutes(app *fiber.App) {
	app.Get("api/search", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Search API"})
	})
}