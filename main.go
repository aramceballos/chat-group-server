package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/aramceballos/chat-group-server/api/routes"
	"github.com/aramceballos/chat-group-server/pkg/channel"
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	_ "github.com/lib/pq"
)

type Message struct {
	Content string `json:"content"`
}
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbUser     = "postgres"
	dbPassword = "mysecretpassword"
	dbName     = "postgres"
)

// Database instance
var db *sql.DB

func Connect() error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

type Client struct {
	isClosing bool
	mu        sync.Mutex
}

var channels = make(map[int64]map[*websocket.Conn]*Client)

func chatProcess() {

}

func main() {
	if err := Connect(); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	channelRepo := channel.NewRepository(db)
	channelService := channel.NewService(channelRepo)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
	}))

	app.Use("/chat", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello world")
	})

	api := app.Group("/api")

	v1 := api.Group("/v1")

	routes.UserRouter(v1, userService)

	routes.ChannelRouter(v1, channelService)

	v1.Get("/chat/:channelId", websocket.New(func(c *websocket.Conn) {
		channelId := c.Params("channelId")
		// Cast channelId to int64
		channelIdInt, err := strconv.ParseInt(channelId, 10, 64)
		if err != nil {
			log.Println("channelId error:", err)
			return
		}

		// Create a new client
		client := &Client{}

		// Add client to the channel
		if _, ok := channels[channelIdInt]; !ok {
			channels[channelIdInt] = make(map[*websocket.Conn]*Client)
		}
		channels[channelIdInt][c] = client

		// Remove client from the channel
		defer func() {
			delete(channels[channelIdInt], c)
		}()

		for {
			msg := Message{}
			err := c.ReadJSON(&msg)
			if err != nil {
				errorMessage := Result{
					Success: false,
					Message: err.Error(),
				}
				err = c.WriteJSON(errorMessage)
				if err != nil {
					log.Println("write error:", err)
					break
				}
				continue
			}

			// Check if message is empty
			if msg.Content == "" {
				continue
			}

			// Insert message into database
			_, err = db.Exec("INSERT INTO messages (channel_id, user_id, content) VALUES ($1, $2, $3)", channelId, 1, string(msg.Content))
			if err != nil {
				errorMessage := Result{
					Success: false,
					Message: err.Error(),
				}
				err = c.WriteJSON(errorMessage)
				if err != nil {
					log.Println("write error:", err)
					break
				}
				continue
			}

			res := Result{
				Success: true,
				Message: "Message sent successfully",
			}
			err = c.WriteJSON(res)
			if err != nil {
				log.Println("write error:", err)
				break
			}

			// Send message to all clients
			for conn, cl := range channels[channelIdInt] {
				// Check if the client is the sender
				if cl == client {
					continue
				}

				go func(conn *websocket.Conn, cl *Client) {
					cl.mu.Lock()
					defer cl.mu.Unlock()
					if cl.isClosing {
						return
					}

					err = conn.WriteJSON(msg)
					if err != nil {
						cl.isClosing = true
						log.Println("write error:", err)

						conn.WriteMessage(websocket.CloseMessage, []byte{})
						conn.Close()
					}
				}(conn, cl)
			}

		}
	}))

	app.Listen(":4000")
}
