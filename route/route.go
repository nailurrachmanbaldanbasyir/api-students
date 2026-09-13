package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/repository"
)

func Register(app *fiber.App, studentRepo *repository.StudentRepository) {
	app.Get("/api/v1/students", listStudents(studentRepo))
	app.Get("/api/v1/students/:id", getStudent(studentRepo))
	app.Post("/api/v1/students", createStudent(studentRepo))
	app.Put("/api/v1/students/:id", replaceStudent(studentRepo))
	app.Patch("/api/v1/students/:id", patchStudent(studentRepo))
	app.Delete("/api/v1/students/:id", deleteStudent(studentRepo))
}
