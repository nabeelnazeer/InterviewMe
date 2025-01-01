package handlers

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func ClearFiles(c *fiber.Ctx) error {
	// Define base directories
	processedTextsDir := "processed_texts"
	uploadsDir := "uploads"

	// Clear and recreate processed_texts subdirectories
	subDirs := []string{"resume", "job"}
	for _, dir := range subDirs {
		fullPath := filepath.Join(processedTextsDir, dir)

		// Remove directory and its contents
		if err := os.RemoveAll(fullPath); err != nil {
			log.Printf("Error removing directory %s: %v", fullPath, err)
			continue
		}

		// Recreate directory
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			log.Printf("Error creating directory %s: %v", fullPath, err)
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to recreate directories",
			})
		}
	}

	// Clear and recreate uploads directory
	if err := os.RemoveAll(uploadsDir); err != nil {
		log.Printf("Error removing uploads directory: %v", err)
	}
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Error creating uploads directory: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to recreate uploads directory",
		})
	}

	return c.JSON(fiber.Map{
		"message": "All directories cleared successfully",
	})
}
