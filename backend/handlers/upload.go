package handlers

import (
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

	// Save file with a unique name
	err = c.SaveFile(file, filepath.Join("uploads", "upload-"+file.Filename))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Unable to save the file",
		})
	}

	return c.JSON(fiber.Map{
		"message": "File uploaded successfully",
	})
}
