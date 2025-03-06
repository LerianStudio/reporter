package handlers

import (
	"fmt"
	"k8s-golang-addons-boilerplate/pkg/models"
	"k8s-golang-addons-boilerplate/pkg/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func CreateExample(svc *service.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input = &models.ExampleInput{}
		if err := c.BodyParser(input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := validate.Struct(input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		result, err := svc.CreateExample(input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	}
}

func GetExample(svc *service.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		exampleID := fmt.Sprintf("%v", c.Locals("example_id"))
		result, err := svc.GetExample(exampleID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	}
}

func GetAllExample(svc *service.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		result, err := svc.GetAllExample()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	}
}

func UpdateExample(svc *service.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		exampleID := fmt.Sprintf("%v", c.Locals("example_id"))

		var input = &models.ExampleInput{}
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := validate.Struct(input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		result, err := svc.UpdateExample(exampleID, input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	}
}

func DeleteExample(svc *service.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		exampleID := fmt.Sprintf("%v", c.Locals("example_id"))

		err := svc.DeleteExample(exampleID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}
