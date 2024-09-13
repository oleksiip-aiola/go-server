package searchRoutes

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func SearchRoutes(app *fiber.App) {
	fmt.Println("Search routes initialized")
	app.Get("/search", func(c *fiber.Ctx) error {
		data := map[string]string{
			"result1": "Item 1",
			"result2": "Item 2",
			"result3": "Item 3",
		}
		return c.JSON(data)
	})
}
