package handler

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

var books = make(map[string]Book)
var idCounter = 1

const token = "supersecret"

var app = fiber.New()

func init() {
	seed1 := Book{
		ID:     "1",
		Title:  "Dune",
		Author: "Frank Herbert",
		Year:   2000,
	}
	seed2 := Book{
		ID:     "2",
		Title:  "1984",
		Author: "George Orwell",
		Year:   1949,
	}
	seed3 := Book{
		ID:     "1",
		Title:  "Dune",
		Author: "Frank Herbert",
		Year:   2000,
	}
	seed4 := Book{
		ID:     "2",
		Title:  "1984",
		Author: "George Orwell",
		Year:   1949,
	}

	books[seed1.ID] = seed1
	books[seed2.ID] = seed2
	books[seed3.ID] = seed3
	books[seed4.ID] = seed4

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
		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "invalid request",
			})
		}

		if body.Username == "admin" && body.Password == "password" {
			return c.JSON(fiber.Map{
				"token": token,
			})
		}

		return c.Status(401).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	})
	api := app.Group("/books", authMiddleware)

	api.Post("/", createBook)
	api.Get("/", getBooks)
	api.Get("/:id", getBook)

	api.Put("/:id", updateBook)
	api.Delete("/:id", deleteBook)

	// app.Listen(":3000")
}

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
	limit, _ := strconv.Atoi(c.Query("limit", "2"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 2
	}

	var all []Book
	for _, b := range books {
		all = append(all, b)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].ID < all[j].ID
	})

	var filtered []Book
	for _, b := range all {
		if author == "" || strings.EqualFold(b.Author, author) {
			filtered = append(filtered, b)
		}
	}

	// STEP 4: pagination
	start := (page - 1) * limit
	if start >= len(filtered) {
		return c.JSON([]Book{})
	}

	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return c.JSON(filtered[start:end])
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
			"data":  fmt.Sprint(book),
			"error": "not found",
		})
	}

	var input Book
	if err := c.BodyParser(&input); err != nil {
		return c.Status(200).JSON(fiber.Map{
			"data":  fmt.Sprint(input),
			"error": "invalid input",
		})
	}

	book.Title = input.Title
	book.Author = input.Author
	book.Year = input.Year

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
