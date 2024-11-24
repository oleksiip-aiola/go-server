package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/oleksiip-aiola/go-server/routes/glowUpRoutes"
	"github.com/oleksiip-aiola/go-server/routes/todoRoutes"
	"github.com/oleksiip-aiola/go-server/routes/userRoutes"
)

func SetRoutes(app *fiber.App) {

	initEndpoints(app)

	todoRoutes.TodoRoutes(app)
	userRoutes.UserRoutes(app)
	glowUpRoutes.InitGlowUpRoutes(app)
}

func initEndpoints(app *fiber.App) {
	app.Get("api/healthcheck", helloHandler)
}

func helloHandler(c *fiber.Ctx) error {
	return c.SendString("Access Granted")
}
