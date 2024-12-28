package main

import (
	"net/url"
	"os"

	"interviewme/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
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

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		panic("Could not create uploads directory")
	}

	// Parse base URL and get path component
	baseURL := "/"
	if urlStr := os.Getenv("BASE_URL"); urlStr != "" {
		if parsedURL, err := url.Parse(urlStr); err == nil {
			baseURL = parsedURL.Path
		}
	}

	// Setup routes with cleaned baseURL
	app.Get(baseURL+"/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post(baseURL+"/upload", handlers.UploadFile)
	app.Post(baseURL+"/preprocess", handlers.PreprocessResume)
	app.Post(baseURL+"/preprocess-job", handlers.PreprocessJobDescription) // Existing route
	app.Delete(baseURL+"/delete", handlers.DeleteFile)

	// Add new route for testing Gemini
	// app.Get(baseURL+"/test-gemini", handlers.TestGemini)

	port := ":8080"
	println("Server running on port", port)
	app.Listen(port)
}
