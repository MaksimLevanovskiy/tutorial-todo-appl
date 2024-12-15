package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
)

type Todo struct {
	ID        int
	Completed bool   `json:"completed"`
	Body      string `json:"body"`
}

func main() {
	app := fiber.New()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")

	var todos []Todo
	app.Get("/api/todos", func(ctx fiber.Ctx) error {
		return ctx.Status(200).JSON(todos)
	})

	app.Post("/api/todos", func(ctx fiber.Ctx) error {
		if string(ctx.Body()) == "" {
			return ctx.Status(400).JSON(fiber.Map{"error": "empty body"})
		}

		var todo Todo
		if err := ctx.Bind().Body(&todo); err != nil {
			return err
		}
		if todo.Body == "" {
			return ctx.Status(400).JSON(fiber.Map{"error": "empty TODO's body"})
		}
		todo.ID = len(todos) + 1
		todos = append(todos, todo)

		return ctx.Status(201).JSON(todo)
	})

	app.Patch("/api/todos/:id", func(ctx fiber.Ctx) error {
		id := ctx.Params("id")

		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos[i].Completed = true
				return ctx.Status(200).JSON(todos[i])
			}
		}
		return ctx.Status(404).JSON(fiber.Map{"error": "todo not found"})
	})

	////delete a todo
	app.Delete("/api/todos/:id", func(ctx fiber.Ctx) error {
		id := ctx.Params("id")

		for i, todo := range todos {
			if fmt.Sprint(todo.ID) == id {
				todos = append(todos[:i], todos[i+1:]...)
				return ctx.Status(200).JSON(fiber.Map{"success": "true"})
			}
		}
		return ctx.Status(404).JSON(fiber.Map{"error": "todo not found"})
	})

	log.Fatal(app.Listen(":" + port))
}
