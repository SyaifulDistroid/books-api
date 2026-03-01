package handler

import (
	"fmt"
	"net/http"
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
		Title:  "Dune",
		Author: "George Orwell",
		Year:   1942,
	}
	seed3 := Book{
		ID:     "3",
		Title:  "Dune",
		Author: "Frank Herbert",
		Year:   2000,
	}
	seed4 := Book{
		ID:     "4",
		Title:  "Lorem ipsum",
		Author: "George Orwell",
		Year:   1949,
	}

	books[seed1.ID] = seed1
	books[seed2.ID] = seed2
	books[seed3.ID] = seed3
	books[seed4.ID] = seed4

	books["7"] = Book{"7", "The Hobbit", "J.R.R. Tolkien", 1937}
	books["8"] = Book{"8", "Fahrenheit 451", "Ray Bradbury", 1953}
	books["9"] = Book{"9", "The Catcher in the Rye", "J.D. Salinger", 1951}
	books["10"] = Book{"10", "Moby Dick", "Herman Melville", 1851}
	books["11"] = Book{"11", "To Kill a Mockingbird", "Harper Lee", 1960}
	books["12"] = Book{"12", "The Great Gatsby", "F. Scott Fitzgerald", 1925}
	books["13"] = Book{"13", "War and Peace", "Leo Tolstoy", 1869}
	books["14"] = Book{"14", "Crime and Punishment", "Fyodor Dostoevsky", 1866}
	books["15"] = Book{"15", "The Alchemist", "Paulo Coelho", 1988}
	books["16"] = Book{"16", "The Odyssey", "Homer", -700}
	books["17"] = Book{"17", "The Divine Comedy", "Dante Alighieri", 1320}
	books["18"] = Book{"18", "Les Misérables", "Victor Hugo", 1862}
	books["19"] = Book{"19", "Don Quixote", "Miguel de Cervantes", 1605}
	books["20"] = Book{"20", "Dracula", "Bram Stoker", 1897}
	books["21"] = Book{"21", "The Brothers Karamazov", "Fyodor Dostoevsky", 1880}
	books["22"] = Book{"22", "Brave New World", "Aldous Huxley", 1932}
	books["23"] = Book{"23", "The Road", "Cormac McCarthy", 2006}

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
	author := strings.TrimSpace(c.Query("author"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "2"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 2
	}

	var filtered []Book
	for _, b := range books {
		if author == "" {
			filtered = append(filtered, b)
		}
	}

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
