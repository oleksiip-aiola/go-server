package main

import (
	"fmt"
	"log"

	"github.com/alexey-petrov/go-server/server/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Todo struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Done bool `json:"done"`
	Body string `json:"body"`
}


func findTodoIndexByID(id int, todos []Todo) int {
    for i, todo := range todos {
        if todo.ID == id {
            return i
        }
    }
    return -1
}

func initDBConnection() {
	// Initialize DB connection
	database := db.ConnectDB()

	defer database.Close()

	db.CreateTable(database)
	db.InsertUser(database, "John Doe", "john.doe@example.com")

	db.GetUsers(database)

	// Close the database connection when done
	defer db.CloseDB()
}


func main() {
	initDBConnection()
	fmt.Println("Hello, Test")
	app := fiber.New();

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	todos := []Todo{}

	initEndpoints(app, todos)
	handleLogFatal(app)
}

func initEndpoints(app *fiber.App, todos []Todo) {
	app.Get("api/healthcheck", helloHandler)

	app.Get("api/todos", func (c *fiber.Ctx) error {
		return c.JSON(todos)
	})

	app.Post("api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

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
				"error": "Todo not found",
			})
		}

		
		todo := &Todo{ID: id}

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
				"error": "Todo is missing fields",
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
                "error": "Todo not found",
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
				"error": "Todo not found",
			})
		}
		
		todos = append(todos[:todoIndex], todos[todoIndex + 1:]...)

		return c.JSON(todos)
	})
	
}

func handleLogFatal(app *fiber.App) {
	log.Fatal(app.Listen(":4000"))

}

func helloHandler(c *fiber.Ctx) error {
	return c.SendString("Access Granted")
}