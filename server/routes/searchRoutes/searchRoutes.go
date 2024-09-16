package searchRoutes

import (
	"fmt"

	"github.com/alexey-petrov/go-server/server/gormdb"
	"github.com/gofiber/fiber/v2"
)


func SearchRoutes(app *fiber.App) {
	fmt.Println("Search routes initialized")
	app.Get("api/search", func(c *fiber.Ctx) error {
		data := map[string]string{
			"result1": "Item 1",
			"result2": "Item 2",
			"result3": "Item 3",
		}
		return c.JSON(data)
	})
	app.Get("api/searchSettings", func(c *fiber.Ctx) error {
		searchSettings := gormdb.SearchSettings{}
		settings, err := searchSettings.GetSearchSettings()

		if err != nil {
			return err
		}

		return c.JSON(settings)
	})

	app.Post("api/searchSettings", func(c *fiber.Ctx) error {
		searchSettings := &gormdb.SearchSettings{}

		if err := c.BodyParser(searchSettings); err != nil {
			return err
		}

		searchSettings.UpdateSearchSettings(searchSettings.SearchOn, searchSettings.AddNew, searchSettings.Amount)

		return c.JSON(fiber.Map{
			"addNew": searchSettings.AddNew,
			"amount": searchSettings.Amount,
			"searchOn": searchSettings.SearchOn,
		})
	})
}
