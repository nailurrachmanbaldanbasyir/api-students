package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"api-students/app/repository"
	"api-students/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: file .env tidak ditemukan")
	}

	db, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	studentRepo := repository.NewStudentRepository(db)

	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"message": "database tidak tersedia",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "service sehat",
		})
	})

	app.Get("/api/v1/students", listStudents(studentRepo))
	app.Get("/api/v1/students/:id", getStudent(studentRepo))
	app.Post("/api/v1/students", createStudent(studentRepo))
	app.Put("/api/v1/students/:id", replaceStudent(studentRepo))
	app.Patch("/api/v1/students/:id", patchStudent(studentRepo))
	app.Delete("/api/v1/students/:id", deleteStudent(studentRepo))

	log.Fatal(app.Listen(":3000"))
}
