package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParseUUIDPathParameters validate ids passing into path parameters values
func ParseUUIDPathParameters(c *fiber.Ctx) error {
	params := c.AllParams()

	var invalidUUIDs []string

	for param, value := range params {
		parsedUUID, err := uuid.Parse(value)
		if err != nil {
			invalidUUIDs = append(invalidUUIDs, param)
			continue
		}

		c.Locals(param, parsedUUID)
	}

	if len(invalidUUIDs) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUIDs", "invalid_uuids": invalidUUIDs})
	}

	return c.Next()
}
