package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Get base URL from environment variables
	baseURL := os.Getenv("BASE_URL")

	// Use base URL for routing
	app.Get(baseURL+"/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Add upload route
	app.Post(baseURL+"/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "File upload failed",
			})
		}

		// Save file
		err = c.SaveFile(file, "./uploads/"+file.Filename)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not save file",
			})
		}

		return c.JSON(fiber.Map{
			"message": "File uploaded successfully",
		})
	})

	app.Listen(":8080")
}
