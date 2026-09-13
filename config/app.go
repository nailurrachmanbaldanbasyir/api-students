package config

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/repository"
	"api-students/middleware"
	"api-students/route"
)

func NewApp(
	logger *log.Logger,
	db *pgxpool.Pool,
	studentRepo *repository.StudentRepository,
) *fiber.App {
	app := fiber.New()

	middleware.Register(app, logger)

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

	route.Register(app, studentRepo)

	logger.Println("Fiber application initialized")

	return app
}
