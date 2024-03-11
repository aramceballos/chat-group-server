package handlers

import (
	"fmt"

	"github.com/aramceballos/chat-group-server/pkg/channel"
	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GetChannels(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		channels, err := service.FetchChannels()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "channels retrieved",
			"data":    channels,
		})
	}

}

func GetChannelById(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		channelId := c.Params("id")
		channel, err := service.FetchChannelById(channelId)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "channel retrieved",
			"data":    channel,
		})
	}
}

func CreateChannel(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.CreateChannelInput

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		validate := validator.New()
		err := validate.Struct(input)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		err = service.CreateChannel(input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":  "success",
			"message": "channel created",
			"data":    nil,
		})
	}
}

func JoinChannel(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.JoinChannelInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

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

		err = service.JoinChannel(fmt.Sprintf("%v", userId), input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status":  "success",
			"message": "joined to channel",
			"data":    nil,
		})
	}
}

func LeaveChannel(service channel.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input entities.JoinChannelInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		validate := validator.New()
		err := validate.Struct(input)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		token := c.Locals("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)
		userId := claims["user_id"].(float64)

		err = service.LeaveChannel(fmt.Sprintf("%v", userId), input)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "channel left",
			"data":    nil,
		})
	}
}
