package main

import (
	"fmt"
	"log"

	"github.com/alexey-petrov/go-server/server/auth"
	"github.com/alexey-petrov/go-server/server/db"
	jwtService "github.com/alexey-petrov/go-server/server/jwt"
	"github.com/alexey-petrov/go-server/server/structs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func findTodoIndexByID(id int, todos []structs.Todo) int {
    for i, todo := range todos {
        if todo.ID == id {
            return i
        }
    }
    return -1
}

func establishDBConnection() {
	// Initialize DB connection
	database := db.ConnectDB()

	defer database.Close()

	db.CreateTable(database)
	db.CreateJTITable(database)

	// Close the database connection when done
	defer db.CloseDB()
}


func main() {
	// Connect to the database
	establishDBConnection()

	fmt.Println("Hello, Test")
	app := fiber.New();

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	todos := []structs.Todo{}

	initEndpoints(app, todos)
	handleLogFatal(app)
}

func initEndpoints(app *fiber.App, todos []structs.Todo) {
	app.Get("api/healthcheck", helloHandler)

	app.Get("api/todos", func (c *fiber.Ctx) error {
		return c.JSON(todos)
	})

	app.Post("api/todos", func(c *fiber.Ctx) error {
		todo := &structs.Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		todo.ID = len(todos) + 1

		todos = append(todos, *todo)


		return c.JSON(todos)
	})

	app.Put("api/todos/:id", func(c *fiber.Ctx) error {
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
				"error": "structs.Todo is missing fields",
				"fields": missingFields,
			})
		}

		todos[todoIndex] = *todo;

		return c.JSON(todos)
	})

	app.Patch("api/todos/:id/status", func(c *fiber.Ctx) error {
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
		
		todos = append(todos[:todoIndex], todos[todoIndex + 1:]...)

		return c.JSON(todos)
	})
	
	app.Post("api/auth", func(c *fiber.Ctx) error {

		user := &structs.User{}

		if err := c.BodyParser(user); err != nil {
			return err
		}

		token, err := auth.Auth(*user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed togenerate JWT",
			})
		}

		return c.JSON(fiber.Map{
			"token": token,
		})
	})

	app.Post("api/refresh", refreshAccessTokenHandler)
}

// Handler function for refreshing the access token using the refresh token
func refreshAccessTokenHandler(c *fiber.Ctx) error {
	accessToken, refreshToken, _ := jwtService.RefreshAccessToken(c)

	// Return the new tokens as JSON response
	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func handleLogFatal(app *fiber.App) {
	log.Fatal(app.Listen(":4000"))

}

func helloHandler(c *fiber.Ctx) error {
	return c.SendString("Access Granted")
}