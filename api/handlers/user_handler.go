package handlers

import (
	"fmt"

	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetUsers(service user.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		users, err := service.FetchUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   users,
		})
	}
}

func GetUserById(service user.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userId := c.Params("id")
		user, err := service.FetchUserById(userId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   user,
		})
	}
}

func UpdateUser(service user.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.UpdateUserInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		// Validate input
		validate := validator.New()
		err := validate.Struct(input)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		fmt.Println("input: ", input)

		// Get bearer token
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userId := claims["user_id"].(float64)

		// Update user
		err = service.UpdateUser(fmt.Sprintf("%v", userId), input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   "user",
		})
	}
}
