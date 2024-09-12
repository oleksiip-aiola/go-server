package main

import (
	"log"
	"strings"

	"github.com/alexey-petrov/go-server/server/auth"
	"github.com/alexey-petrov/go-server/server/db"
	"github.com/alexey-petrov/go-server/server/jwtService"
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

	app := fiber.New();

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	todos := []structs.Todo{}

	initEndpoints(app, todos)
	handleLogFatal(app)
}

func initEndpoints(app *fiber.App, todos []structs.Todo) {
	app.Get("api/healthcheck", helloHandler)

	app.Get("api/todos", func (c *fiber.Ctx) error {
		// Check if Authorization header exists
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing Authorization header",
			})
		}

		// Extract JWT token from Authorization header
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify and parse the JWT token
		_, err := jwtService.VerifyAndParseToken(token)

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT token",
			})
		}

		// Continue with the API logic
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
	
	app.Post("api/register", func(c *fiber.Ctx) error {

		user := &structs.User{}

		if err := c.BodyParser(user); err != nil {
			return err
		}

		token, refreshToken, err := auth.Auth(*user)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to register",
				"detail": err.Error(),
			})
		}

		jwtService.SetRefreshCookie(c, refreshToken)

		return c.JSON(fiber.Map{
			"token": token,
		})
	})

	app.Post("api/refresh", ManualResetAccessTokenHandler)
	app.Post("api/login", handleLogin)
	app.Post("api/refresh-token", handleRefreshToken)
	app.Post("api/logout", handleLogout)
}

func handleLogout(c *fiber.Ctx) error {
	_, _, err := jwtService.HandleInvalidateTokenByJti(c)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to Invalidate JWT Refresh",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

func handleRefreshToken (c *fiber.Ctx) error {
	accessToken, err := jwtService.RefreshAccessToken(c)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

func handleLogin(c *fiber.Ctx) error {
	user := &structs.User{}

	if err := c.BodyParser(user); err != nil {
		return err
	}

	accessToken, refreshToken, err := auth.Login(user.Email, user.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	jwtService.SetRefreshCookie(c, refreshToken)

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}

// Handler function for refreshing the access token using the refresh token
func ManualResetAccessTokenHandler(c *fiber.Ctx) error {
	accessToken, refreshToken, _ := jwtService.ManualResetAccessToken(c)

	jwtService.SetRefreshCookie(c, refreshToken)
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