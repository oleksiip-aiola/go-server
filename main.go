package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-petrov/go-server/db"
	"github.com/alexey-petrov/go-server/routes"
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

	//Glow up
	db.CreateUserMoodRecordsTable()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	// Connect to the database
	establishdbConnection()
	app := fiber.New(fiber.Config{
		IdleTimeout: 5 * time.Second,
	})

	publicUrl := os.Getenv("PUBLIC_URL")
	allowedOrigins := "http://localhost:3000,https://localhost:3000"

	if publicUrl != "" {
		allowedOrigins = fmt.Sprintf("%s, %s", allowedOrigins, publicUrl)
	}

	app.Options("*", cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET, POST, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Set-Cookie, connect-protocol-version",
		AllowCredentials: true,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Set-Cookie, connect-protocol-version",
		AllowMethods:     "GET, POST, OPTIONS",
		AllowCredentials: true,
	}))

	app.Use(compress.New())

	routes.SetRoutes(app)

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
