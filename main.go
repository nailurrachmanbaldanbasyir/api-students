package main

import "github.com/gofiber/fiber/v2"

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	api := app.Group("/api/v1")

	api.Get("/students", listStudents)
	api.Get("/students/:id", getStudent)

	api.Post("/students", createStudent)

	api.Put("/students/:id", replaceStudent)
	api.Patch("/students/:id", patchStudent)

	api.Delete("/students/:id", deleteStudent)

	app.Listen(":3000")
}
