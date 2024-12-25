package handlers

import (
	"log" // For detailed logging
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// DeleteFile handles file deletion
func DeleteFile(c *fiber.Ctx) error {
	// Get filename from query parameter
	fileName := c.Query("filename")
	if fileName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Filename parameter is required",
		})
	}

	// Prevent path traversal vulnerabilities
	if filepath.Base(fileName) != fileName {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid filename",
		})
	}

	// Construct file path
	filePath := filepath.Join("uploads", "upload-"+fileName)
	log.Printf("Attempting to delete file at path: %s", filePath)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("File does not exist: %s", filePath)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "File does not exist",
		})
	} else if err != nil {
		log.Printf("Error while checking file existence: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error checking file",
		})
	}

	// Delete the file
	err := os.Remove(filePath)
	if err != nil {
		log.Printf("Failed to delete file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete file",
		})
	}

	log.Printf("File deleted successfully: %s", filePath)
	return c.JSON(fiber.Map{
		"message": "File deleted successfully",
	})
}
