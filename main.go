package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aramceballos/chat-group-server/api/routes"
	"github.com/aramceballos/chat-group-server/pkg/entities"
	"github.com/aramceballos/chat-group-server/pkg/user"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"

	_ "github.com/lib/pq"
)

type Message struct {
	Body json.RawMessage `json:"body"`
}
type Result struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Database instance
var db *sql.DB

func Connect(
	dbHost string,
	dbPort string,
	dbUser string,
	dbPassword string,
	dbName string,
	sslMode string,
) error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPassword, dbName, sslMode))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

type Client struct {
	conn *websocket.Conn
	send chan interface{}
	quit chan struct{}
	once sync.Once
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan interface{}, 256),
		quit: make(chan struct{}),
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message := <-c.send:
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("write error [broadcast]: %v", err)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
		case <-c.quit:
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}

func (c *Client) close() {
	c.once.Do(func() {
		close(c.quit)
	})
}

var channels = make(map[int64]map[*websocket.Conn]*Client)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSL_MODE")
	if sslMode == "" {
		sslMode = "require" // Default for secure connection
	}

	if err := Connect(dbHost, dbPort, dbUser, dbPassword, dbName, sslMode); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
			return slices.Contains(allowedOrigins, origin)
		},
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        10,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello world")
	})

	api := app.Group("/api")

	v1 := api.Group("/v1")

	routes.UserRouter(v1, userService)

	v1.Use("/chat", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})

	v1.Get("/chat/:channelId", websocket.New(func(c *websocket.Conn) {
		token := c.Query("token")
		if token == "" {
			errorMessage := Result{
				Success: false,
				Message: "Token is required",
			}
			err := c.WriteJSON(errorMessage)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
				return
			}
			return
		}

		// Parse the token
		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			secret := os.Getenv("JWT_SECRET")
			if secret == "" {
				panic("JWT_SECRET env var must be set")
			}
			return []byte(secret), nil
		})
		if err != nil {
			errorMessage := Result{
				Success: false,
				Message: err.Error(),
			}
			err = c.WriteJSON(errorMessage)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
				return
			}
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			errorResponse := Result{
				Success: false,
				Message: "invalid claims format",
			}
			err := c.WriteJSON(errorResponse)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
			}
			return
		}

		userId, ok := claims["user_id"].(float64)
		fmt.Println("Claims: ", claims)
		fmt.Println("userId: ", claims["user_id"])
		fmt.Println("parsed userId: ", userId)
		if !ok {
			errorResponse := Result{
				Success: false,
				Message: "invalid user_id in token",
			}
			err := c.WriteJSON(errorResponse)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
			}
			return
		}

		channelId := c.Params("channelId")
		// Cast channelId to int64
		channelIdInt, err := strconv.ParseInt(channelId, 10, 64)
		if err != nil {
			log.Println("channelId error:", err)
			return
		}

		// Check if user is a member of the channel
		var exists bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM memberships WHERE channel_id = $1 AND user_id = $2)", channelIdInt, int64(userId)).Scan(&exists)
		if err != nil || !exists {
			fmt.Println("error checking membership:", err)
			errorMessage := Result{
				Success: false,
				Message: "You are not a member of this channel",
			}
			err = c.WriteJSON(errorMessage)
			if err != nil {
				log.Printf("write error [error response]: %v", err)
				return
			}
			return
		}

		// Create a new client
		client := NewClient(c)

		// Add client to the channel
		if _, ok := channels[channelIdInt]; !ok {
			channels[channelIdInt] = make(map[*websocket.Conn]*Client)
		}
		channels[channelIdInt][c] = client

		fmt.Printf("Client %s connected to channel %d\n", c.RemoteAddr().String(), channelIdInt)

		go client.writePump()

		// Remove client from the channel
		defer func() {
			fmt.Printf("Client %s disconnected from channel %d\n", c.RemoteAddr().String(), channelIdInt)
			client.close()
			delete(channels[channelIdInt], c)
		}()

		for {
			msg := Message{}
			err := c.ReadJSON(&msg.Body)
			if err != nil {
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}

			// Check if message is empty
			if len(msg.Body) == 0 {
				continue
			}

			// Validate message body JSON structure
			var body map[string]interface{}
			if err := json.Unmarshal(msg.Body, &body); err != nil {
				client.send <- Result{
					Success: false,
					Message: "Invalid message body JSON",
				}
				continue
			}

			msgType, ok := body["type"].(string)
			if !ok {
				client.send <- Result{
					Success: false,
					Message: "Message body must have a 'type' field",
				}
				continue
			}

			if msgType == "text" {
				content, ok := body["content"].(string)
				if !ok || len(content) == 0 {
					client.send <- Result{
						Success: false,
						Message: "Text messages must have a non-empty 'content' field",
					}
					continue
				}
			} else if msgType == "file" {
				// Validate file fields
				fileID, fileIDOk := body["file_id"].(string)
				filename, filenameOk := body["filename"].(string)
				mimeType, mimeTypeOk := body["mime_type"].(string)
				url, urlOk := body["url"].(string)
				_, sizeOk := body["size_in_bytes"]
				if !fileIDOk || fileID == "" || !filenameOk || filename == "" || !mimeTypeOk || mimeType == "" || !urlOk || url == "" || !sizeOk {
					client.send <- Result{
						Success: false,
						Message: "File messages must have non-empty 'file_id', 'filename', 'mime_type', 'url', and 'size_in_bytes' fields",
					}
					continue
				}
			} else {
				client.send <- Result{
					Success: false,
					Message: "Unsupported message type",
				}
				continue
			}

			fmt.Printf("Received message from %s content: %s\n", c.RemoteAddr().String(), string(msg.Body))

			insertedMessage := entities.Message{}
			// Insert message into database
			err = db.QueryRow("INSERT INTO messages (channel_id, user_id, body) VALUES ($1, $2, $3::jsonb) RETURNING id, user_id, channel_id, body, created_at", channelId, userId, msg.Body).Scan(&insertedMessage.ID, &insertedMessage.UserID, &insertedMessage.ChannelID, &insertedMessage.Body, &insertedMessage.CreatedAt)
			if err != nil {
				fmt.Println("error inserting message:", err)
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}

			// Query user from database to populate the message
			user := entities.User{}
			err = db.QueryRow("SELECT id, name, avatar_url, created_at FROM users WHERE id = $1", userId).Scan(&user.ID, &user.Name, &user.AvatarURL, &user.CreatedAt)
			if err != nil {
				client.send <- Result{
					Success: false,
					Message: err.Error(),
				}
				continue
			}
			insertedMessage.User = user
			fmt.Println("Broadcasting message", string(insertedMessage.Body))

			client.send <- Result{
				Success: true,
				Message: "Message sent successfully",
			}

			// Send message to all clients
			for conn, cl := range channels[channelIdInt] {
				go func(conn *websocket.Conn, cl *Client) {
					fmt.Printf("Sending message to %s content: %s\n", conn.RemoteAddr().String(), string(msg.Body))
					cl.send <- insertedMessage
				}(conn, cl)
			}

		}
	}))

	app.Listen(":4000")
}
