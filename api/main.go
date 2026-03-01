package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   string `json:"year"`
}

var books = make(map[string]Book)
var idCounter = 1

const token = "supersecret"

var app = fiber.New()

func init() {
	seed := Book{
		ID:     "e67d1777-99e9-4597-a33d-9cc2aa9ee44e",
		Title:  "Dune",
		Author: "Frank Herbert",
		Year:   "2000",
	}
	books[seed.ID] = seed

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"success": true,
		})
	})

	app.Post("/echo", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(c.Body())
	})

	app.Post("/auth/token", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"token": token})
	})

	app.Post("/books", createBook)
	app.Get("/books", getBooks)
	app.Get("/books/:id", getBook)

	app.Put("/books/:id", updateBook)
	app.Delete("/books/:id", deleteBook)
	// _ = app.Group("/", authMiddleware)

}

// func main() {

// 	app.Listen(":3000")
// }

func Handler(w http.ResponseWriter, r *http.Request) {
	adaptor.FiberApp(app)(w, r)
}

func authMiddleware(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if auth != "Bearer "+token {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	return c.Next()
}

func createBook(c *fiber.Ctx) error {
	var book Book
	if err := c.BodyParser(&book); err != nil || book.Title == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	book.ID = strconv.Itoa(idCounter)
	idCounter++

	books[book.ID] = book

	return c.Status(201).JSON(book)
}

func getBooks(c *fiber.Ctx) error {
	author := c.Query("author")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	var result []Book
	for _, b := range books {
		if author == "" || b.Author == author {
			result = append(result, b)
		}
	}

	start := (page - 1) * limit
	end := start + limit

	if start > len(result) {
		return c.JSON([]Book{})
	}
	if end > len(result) {
		end = len(result)
	}

	return c.JSON(result[start:end])
}

func getBook(c *fiber.Ctx) error {
	id := c.Params("id")
	fmt.Println(id)
	book, ok := books[id]
	if !ok {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(book)
}

func updateBook(c *fiber.Ctx) error {
	id := c.Params("id")

	book, ok := books[id]
	if !ok {
		return c.Status(200).JSON(fiber.Map{
			"error": "not found",
		})
	}

	var input Book
	if err := c.BodyParser(&input); err != nil {
		return c.Status(200).JSON(fiber.Map{
			"error": "invalid input",
		})
	}

	book.Title = input.Title
	book.Author = input.Author

	books[id] = book

	return c.JSON(book)
}
func deleteBook(c *fiber.Ctx) error {
	id := c.Params("id")

	if _, ok := books[id]; !ok {
		return c.Status(200).JSON(fiber.Map{
			"error": "not found",
		})
	}

	delete(books, id)

	return c.SendStatus(204)
}
