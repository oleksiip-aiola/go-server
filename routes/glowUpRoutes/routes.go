package glowUpRoutes

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/oleksiip-aiola/go-server/db"
)

func InitGlowUpRoutes(app *fiber.App) {
	fmt.Println("Initializing glowUp routes")
	app.Post(`api/glowUp/rate`, handleCreateRate)
	app.Patch(`api/glowUp/rate/:id`, handleUpdateRate)
	app.Get(`api/glowUp/rates/:userId/:year/:month`, getMoodScores)
}

func handleCreateRate(c *fiber.Ctx) error {
	moodDto := &db.MoodScore{}

	if err := c.BodyParser(moodDto); err != nil {
		return errors.New("failed to parse mood score")
	}

	moodScore, err := db.CreateMoodScore(moodDto.UserId, moodDto.Year, moodDto.Month, moodDto.Day, moodDto.MoodId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to create mood score",
			"detail": err.Error(),
		})
	}

	return c.JSON(moodScore)
}

type Rate struct {
	MoodId int32
}

func handleUpdateRate(c *fiber.Ctx) error {
	moodDto := &Rate{}
	id := c.Params("id")

	if err := c.BodyParser(moodDto); err != nil {
		return errors.New("failed to parse mood score")
	}

	err := db.UpdateMoodScore(id, moodDto.MoodId)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update mood score",
			"detail": err.Error(),
		})
	}

	return c.JSON(moodDto)
}

type GetMoodsStruct struct {
	UserId string `json:"userId"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
}

func getMoodScores(c *fiber.Ctx) error {
	moodDto := &GetMoodsStruct{}

	if err := c.ParamsParser(moodDto); err != nil {
		return errors.New("failed to parse get moods query params")
	}

	moods, err := db.GetMoodScores(moodDto.UserId, moodDto.Year, moodDto.Month)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to update mood score",
			"detail": err.Error(),
		})
	}

	return c.JSON(moods)
}
