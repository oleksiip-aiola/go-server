package searchRoutes

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/alexey-petrov/go-server/server/views"
	"github.com/gofiber/fiber/v2"
)

func SearchRoutes(app *fiber.App) {
	fmt.Println("Search routes initialized")
	app.Get("search", func(c *fiber.Ctx) error {
		return render(c, views.Home())
	})
}

func render(c *fiber.Ctx, component templ.Component) error {
	fmt.Println("Content rendered")

	c.Set("Content-Type", "text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}