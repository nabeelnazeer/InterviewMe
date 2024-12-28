package handlers

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Error retrieving the file",
		})
	}

	// Check file extension
	if filepath.Ext(file.Filename) != ".pdf" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Only PDF files are allowed",
		})
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create uploads directory",
		})
	}

	// Save file with a unique name
	filename := "upload-" + file.Filename
	err = c.SaveFile(file, filepath.Join("uploads", filename))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Unable to save the file",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "File uploaded successfully",
		"filename": filename,
	})
}
