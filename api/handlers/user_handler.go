package handlers

import (
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/gofiber/fiber/v2"
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
