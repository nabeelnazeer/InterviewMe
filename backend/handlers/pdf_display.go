package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func DisplayPDF(c *fiber.Ctx) error {
	filename := c.Query("filename")
	if filename == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Filename is required",
		})
	}

	// Ensure the path is within uploads directory
	filePath := filepath.Join("uploads", filename)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Invalid file path",
		})
	}

	// Log the file path being accessed
	println("Attempting to access file:", absPath)

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("PDF file not found: %s", filename),
			"path":  filePath,
		})
	}

	// Open and serve the file
	file, err := os.Open(filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Unable to open file",
			"details": err.Error(),
		})
	}
	defer file.Close()

	// Set PDF headers
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "inline; filename="+filename)

	// Stream the file
	_, err = io.Copy(c.Response().BodyWriter(), file)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Error streaming file",
			"details": err.Error(),
		})
	}

	return nil
}
