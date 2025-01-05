package main

import (
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
		panic(" jamalu Error loading .env file")
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

	// Routes without baseURL prefix
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Post("/upload", handlers.UploadFile)
	app.Post("/preprocess", handlers.PreprocessResume)
	app.Post("/preprocess-job", handlers.PreprocessJobDescription)
	app.Post("/score-resume", handlers.ScoreResume)
	app.Delete("/delete", handlers.DeleteFile)

	app.Get("/pdf/display", handlers.DisplayPDF)

	app.Post("/score", handlers.ScoreResume)
	app.Post("/clear", handlers.ClearFiles)

	port := ":8080"
	println("Server running on port", port)
	app.Listen(port)
}
