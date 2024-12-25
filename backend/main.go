package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	app := fiber.New()

	// Add logger middleware
	app.Use(logger.New())

	// Get base URL from environment variables
	baseURL := os.Getenv("BASE_URL")

	// Use base URL for routing
	app.Get(baseURL+"/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":8080")
}
