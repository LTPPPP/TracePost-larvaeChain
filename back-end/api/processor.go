package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/vietchain/tracepost-larvae/db"
	"time"
)

// Processor represents a processing facility in the supply chain
type Processor struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateProcessorRequest represents the request body for creating a processor
type CreateProcessorRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Contact  string `json:"contact"`
}

// CreateProcessor creates a new processor
func CreateProcessor(c *fiber.Ctx) error {
	var req CreateProcessorRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	if req.Name == "" || req.Location == "" || req.Contact == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, location, and contact are required")
	}

	query := `
		INSERT INTO processor (name, location, contact, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var processor Processor
	err := db.DB.QueryRow(query, req.Name, req.Location, req.Contact).Scan(&processor.ID, &processor.CreatedAt, &processor.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create processor")
	}

	processor.Name = req.Name
	processor.Location = req.Location
	processor.Contact = req.Contact

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Processor created successfully",
		"data":    processor,
	})
}

// GetAllProcessors retrieves all processors
func GetAllProcessors(c *fiber.Ctx) error {
	rows, err := db.DB.Query(`
		SELECT id, name, location, contact, created_at, updated_at
		FROM processor
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve processors")
	}
	defer rows.Close()

	var processors []Processor
	for rows.Next() {
		var processor Processor
		if err := rows.Scan(&processor.ID, &processor.Name, &processor.Location, &processor.Contact, &processor.CreatedAt, &processor.UpdatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse processor data")
		}
		processors = append(processors, processor)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Processors retrieved successfully",
		"data":    processors,
	})
}