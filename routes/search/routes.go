package searchRoutes

import (
	"fmt"
	"time"

	"github.com/alexey-petrov/go-server/server/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
)

func SearchRoutes(app *fiber.App) {
	fmt.Println("Search routes initialized")
	app.Post("api/search", handleSearch)
	app.Use("api/search", cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Query("noCache") == "true"
		},
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))

	app.Get("api/searchSettings", func(c *fiber.Ctx) error {
		searchSettings := db.SearchSettings{}
		settings, err := searchSettings.GetSearchSettings()

		if err != nil {
			return err
		}

		return c.JSON(settings)
	})

	app.Post("api/searchSettings", func(c *fiber.Ctx) error {
		searchSettings := &db.SearchSettings{}

		if err := c.BodyParser(searchSettings); err != nil {
			return err
		}

		searchSettings.UpdateSearchSettings(searchSettings.SearchOn, searchSettings.AddNew, searchSettings.Amount)

		return c.JSON(fiber.Map{
			"addNew":   searchSettings.AddNew,
			"amount":   searchSettings.Amount,
			"searchOn": searchSettings.SearchOn,
		})
	})
}
