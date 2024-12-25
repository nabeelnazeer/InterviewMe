package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func CreateTask(c *fiber.Ctx) error {
	type Task struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	task := new(Task)
	if err := c.BodyParser(task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Here you would normally save the task to a database
	// For now, we'll just return the task as a confirmation
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Task created successfully",
		"task":    task,
	})
}
