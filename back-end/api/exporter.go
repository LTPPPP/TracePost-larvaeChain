package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"time"
)

// Exporter represents an exporter in the supply chain
type Exporter struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Contact   string    `json:"contact"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateExporterRequest represents the request body for creating an exporter
type CreateExporterRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Contact  string `json:"contact"`
}

// CreateExporter creates a new exporter
func CreateExporter(c *fiber.Ctx) error {
	var req CreateExporterRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	if req.Name == "" || req.Location == "" || req.Contact == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Name, location, and contact are required")
	}

	query := `
		INSERT INTO exporter (name, location, contact, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var exporter Exporter
	err := db.DB.QueryRow(query, req.Name, req.Location, req.Contact).Scan(&exporter.ID, &exporter.CreatedAt, &exporter.UpdatedAt)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create exporter")
	}

	exporter.Name = req.Name
	exporter.Location = req.Location
	exporter.Contact = req.Contact

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Exporter created successfully",
		"data":    exporter,
	})
}

// GetAllExporters retrieves all exporters
func GetAllExporters(c *fiber.Ctx) error {
	rows, err := db.DB.Query(`
		SELECT id, name, location, contact, created_at, updated_at
		FROM exporter
		ORDER BY created_at DESC
	`)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve exporters")
	}
	defer rows.Close()

	var exporters []Exporter
	for rows.Next() {
		var exporter Exporter
		if err := rows.Scan(&exporter.ID, &exporter.Name, &exporter.Location, &exporter.Contact, &exporter.CreatedAt, &exporter.UpdatedAt); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to parse exporter data")
		}
		exporters = append(exporters, exporter)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Exporters retrieved successfully",
		"data":    exporters,
	})
}