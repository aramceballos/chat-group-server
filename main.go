package main

import "github.com/gofiber/fiber/v2"

type Group struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	ImageURL  string `json:"image_url"`
	CreatedAt string `json:"created_at"`
}

type User struct {
	Id        string `json:"id"`
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
	GroupId   string `json:"group_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

func main() {
	groups := []Group{
		{
			Id:        "1",
			Name:      "Front-End Developers",
			ImageURL:  "",
			CreatedAt: "2023-10-20 13:29:00",
		},
		{
			Id:        "2",
			Name:      "Random",
			ImageURL:  "",
			CreatedAt: "2023-10-20 13:29:00",
		},
		{
			Id:        "3",
			Name:      "Backend",
			ImageURL:  "",
			CreatedAt: "2023-10-20 13:29:00",
		},
		{
			Id:        "4",
			Name:      "Cats and Dogs",
			ImageURL:  "",
			CreatedAt: "2023-10-20 13:29:00",
		},
		{
			Id:        "5",
			Name:      "Welcome",
			ImageURL:  "",
			CreatedAt: "2023-10-20 13:29:00",
		},
	}

	users := []User{
		{
			Id:        "1",
			Name:      "Shaunna Firth",
			CreatedAt: "2023-10-20 13:29:00",
			AvatarUrl: "https://picsum.photos/seed/ShaunnaFirth/42/42",
		},
		{
			Id:        "2",
			Name:      "Nellie Francis",
			CreatedAt: "2023-10-20 13:29:00",
			AvatarUrl: "https://picsum.photos/seed/NellieFrancis/42/42",
		},
		{
			Id:        "2",
			Name:      "Denzel Barrett",
			CreatedAt: "2023-10-20 13:29:00",
			AvatarUrl: "https://picsum.photos/seed/DenzelBarret/42/42",
		},
	}

	messages := []Message{
		{
			Id:        "1",
			UserId:    "1",
			Content:   "Morbi eget turpis ut massa luctus cursus. Sed sit amet risus quis neque condimentum aliquet. Phasellus consequat et justo eu accumsan 🙌. Proin pretium id nunc eu molestie. Nam consectetur, ligula vel mattis facilisis, ex mauris venenatis nulla, eget tempor enim neque eget massa 🤣",
			CreatedAt: "2023-10-27 13:29:00",
		},
		{
			Id:        "2",
			UserId:    "2",
			Content:   "Class aptent taciti sociosqu ad litora torquent per conubia nostra 😀",
			CreatedAt: "2023-10-27 14:29:00",
		},
		{
			Id:        "3",
			UserId:    "3",
			Content:   "Aenean tempus nibh vel est lobortis euismod. Vivamus laoreet viverra nunc 🐶",
			CreatedAt: "2023-10-27 14:39:00",
		},
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString("Hello world")
	})

	api := app.Group("/api")

	v1 := api.Group("/v1")

	v1.Get("/users", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "Ok",
			"data":   users,
		})
	})
	v1.Get("/groups", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "Ok",
			"data":   groups,
		})
	})
	v1.Get("/messages", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "Ok",
			"data":   messages,
		})
	})

	app.Listen(":4000")
}
