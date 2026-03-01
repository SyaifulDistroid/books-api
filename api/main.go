package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	_ "modernc.org/sqlite"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "books.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		year INTEGER NOT NULL
	);
	`)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	app.Post("/echo", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(c.Body())
	})

	app.Post("/books", createBook)
	app.Get("/books", getBooks)
	app.Get("/books/:id", getBook)
	app.Put("/books/:id", updateBook)
	app.Delete("/books/:id", deleteBook)

	log.Fatal(app.Listen(":3000"))
}

func createBook(c *fiber.Ctx) error {
	var book Book
	if err := c.BodyParser(&book); err != nil || book.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	result, err := db.Exec(
		"INSERT INTO books (title, author, year) VALUES (?, ?, ?)",
		book.Title, book.Author, book.Year,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}

	id, _ := result.LastInsertId()
	book.ID = fmt.Sprintf("%d", id)

	return c.Status(201).JSON(book)
}

func getBooks(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT id, title, author, year FROM books")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		var id int
		rows.Scan(&id, &book.Title, &book.Author, &book.Year)
		book.ID = fmt.Sprintf("%d", id)
		books = append(books, book)
	}

	return c.JSON(books)
}

func getBook(c *fiber.Ctx) error {
	id := c.Params("id")

	row := db.QueryRow(
		"SELECT id, title, author, year FROM books WHERE id=?",
		id,
	)

	var book Book
	var bookID int
	err := row.Scan(&bookID, &book.Title, &book.Author, &book.Year)
	if err == sql.ErrNoRows {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}

	book.ID = fmt.Sprintf("%d", bookID)
	return c.JSON(book)
}

func updateBook(c *fiber.Ctx) error {
	id := c.Params("id")

	var input Book
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
	}

	result, err := db.Exec(
		"UPDATE books SET title=?, author=?, year=? WHERE id=?",
		input.Title, input.Author, input.Year, id,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	return getBook(c)
}

func deleteBook(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := db.Exec("DELETE FROM books WHERE id=?", id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db error"})
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	}

	return c.SendStatus(204)
}
