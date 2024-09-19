package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexey-petrov/go-server/db"
	"github.com/alexey-petrov/go-server/routes"
	"github.com/alexey-petrov/go-server/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func establishdbConnection() {
	fmt.Println("Establishing Gorm DB connection")
	// Initialize DB connection
	db.InitDB()

	db.CreateTable()
	db.CreateJTITable()
	db.CreateSearchSettingsTable()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	// Connect to the database
	establishdbConnection()
	app := fiber.New(fiber.Config{
		IdleTimeout: 5,
	})

	publicUrl := os.Getenv("PUBLIC_URL")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     fmt.Sprintf("http://localhost:3000, %s", publicUrl),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	app.Use(compress.New())

	routes.SetRoutes(app)
	utils.StartCronJobs()
	handleLogFatal(app)

	go func() {
		if error := app.Listen(":" + os.Getenv("PORT")); error != nil {
			log.Panic(error)
		}
	}()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // Block the main thread until a signal is received/interrupted

	app.Shutdown()
	fmt.Println("Shutting down the server")
}

func handleLogFatal(app *fiber.App) {
	log.Fatal(app.Listen(":" + os.Getenv("PORT")))
}
