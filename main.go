package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "mysecretpassword"
	dbname   = "postgres"
)

type Channel struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	CreatedAt   string    `json:"created_at,omitempty"`
	Members     []User    `json:"members,omitempty"`
	Messages    []Message `json:"messages,omitempty"`
}

type User struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	AvatarUrl string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}

type Message struct {
	Id        string `json:"id"`
	UserId    string `json:"user_id"`
	ChannelId string `json:"channel_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type Membership struct {
	Id        int `json:"id"`
	UserId    int `json:"user_id"`
	ChannelId int `json:"channel_id"`
}

// Database instance
var db *sql.DB

func Connect() error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello world")
	})

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Get("/users", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, avatar_url, created_at FROM users")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}
		defer rows.Close()
		result := []User{}

		for rows.Next() {
			user := User{}
			err := rows.Scan(&user.Id, &user.Name, &user.AvatarUrl, &user.CreatedAt)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}
			result = append(result, user)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   result,
		})
	})
	v1.Get("/channels", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT id, name, description, image_url FROM channels")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		result := []Channel{}

		for rows.Next() {
			channel := Channel{}
			err := rows.Scan(&channel.Id, &channel.Name, &channel.Description, &channel.ImageURL)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}
			result = append(result, channel)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   result,
		})
	})
	v1.Get("/channels/:id", func(c *fiber.Ctx) error {
		channel := Channel{}
		err := db.QueryRow("SELECT id, name, description, image_url FROM channels WHERE id = $1", c.Params("id")).Scan(&channel.Id, &channel.Name, &channel.Description, &channel.ImageURL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		rows, err := db.Query("SELECT id, user_id, channel_id FROM memberships WHERE channel_id = $1", c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		memberships := []Membership{}

		for rows.Next() {
			membership := Membership{}
			err := rows.Scan(&membership.Id, &membership.UserId, &membership.ChannelId)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}
			memberships = append(memberships, membership)
		}

		for _, membership := range memberships {
			user := User{}
			err := db.QueryRow("SELECT id, name, avatar_url, created_at FROM users WHERE id = $1", membership.UserId).Scan(&user.Id, &user.Name, &user.AvatarUrl, &user.CreatedAt)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}
			channel.Members = append(channel.Members, user)
		}

		rows, err = db.Query("SELECT id, user_id, channel_id, content, created_at FROM messages WHERE channel_id = $1", c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		}

		for rows.Next() {
			message := Message{}
			err := rows.Scan(&message.Id, &message.UserId, &message.ChannelId, &message.Content, &message.CreatedAt)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"status":  "error",
					"message": err.Error(),
				})
			}
			channel.Messages = append(channel.Messages, message)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
			"data":   channel,
		})
	})

	app.Listen(":4000")
}
