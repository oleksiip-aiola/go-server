package todoRoutes

import (
	"github.com/alexey-petrov/go-server/jwtService"
	"github.com/alexey-petrov/go-server/structs"
	"github.com/gofiber/fiber/v2"
)

func findTodoIndexByID(id int, todos []structs.Todo) int {
	for i, todo := range todos {
		if todo.ID == id {
			return i
		}
	}
	return -1
}

func TodoRoutes(app *fiber.App) {
	todos := []structs.Todo{}

	app.Get("api/todos", func(c *fiber.Ctx) error {
		jwtService.VerifyTokenProtectedRoute(c)

		// Continue with the API logic
		return c.JSON(todos)
	})

	app.Post("api/todos", func(c *fiber.Ctx) error {
		jwtService.VerifyTokenProtectedRoute(c)

		todo := &structs.Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		todo.ID = len(todos) + 1

		todos = append(todos, *todo)

		return c.JSON(todos)
	})

	app.Put("api/todos/:id", func(c *fiber.Ctx) error {
		jwtService.VerifyTokenProtectedRoute(c)

		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(401).SendString("Invalid todo ID")
		}

		todoIndex := findTodoIndexByID(id, todos)

		if todoIndex == -1 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "structs.Todo not found",
			})
		}

		todo := &structs.Todo{ID: id}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		if todo.Title == "" || todo.Body == "" {
			missingFields := []string{}
			if todo.Title == "" {
				missingFields = append(missingFields, "title")
			}
			if todo.Body == "" {
				missingFields = append(missingFields, "body")
			}

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "structs.Todo is missing fields",
				"fields": missingFields,
			})
		}

		todos[todoIndex] = *todo

		return c.JSON(todos)
	})

	app.Patch("api/todos/:id/status", func(c *fiber.Ctx) error {
		jwtService.VerifyTokenProtectedRoute(c)

		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(401).SendString("Invalid todo ID")
		}

		// Find the todo by ID
		todoIndex := findTodoIndexByID(id, todos)

		if todoIndex == -1 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "structs.Todo not found",
			})
		}

		todos[todoIndex].Done = !todos[todoIndex].Done

		return c.JSON(todos)
	})

	app.Delete("api/todos/:id", func(c *fiber.Ctx) error {
		jwtService.VerifyTokenProtectedRoute(c)

		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(401).SendString("Invalid todo ID")
		}

		todoIndex := findTodoIndexByID(id, todos)

		if todoIndex == -1 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "structs.Todo not found",
			})
		}

		todos = append(todos[:todoIndex], todos[todoIndex+1:]...)

		return c.JSON(todos)
	})

}
