package handlers

import (
	"fmt"

	"github.com/aramceballos/chat-group-server/pkg/auth"
	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Login(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.LoginInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		token, err := service.Login(input)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "logged in succesfully",
			"data":    token,
		})
	}
}

func Signup(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.SignupInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		token, err := service.Signup(input)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "signed up succesfully",
			"data":    token,
		})
	}
}

func Me(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userId := claims["user_id"].(float64)

		user, err := service.Me(fmt.Sprintf("%v", userId))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "information retrieved succesfully",
			"data":    user,
		})
	}
}

func ChangePassword(service auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.ChangePasswordInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		// Validate input
		validate := validator.New()
		err := validate.Struct(input)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		// Get bearer token
		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userId := claims["user_id"].(float64)

		// Change password
		err = service.ChangePassword(fmt.Sprintf("%v", userId), input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "password changed succesfully",
			"data":    nil,
		})
	}
}
