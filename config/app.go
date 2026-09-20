package config

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/middleware"
	"api-students/route"
)

func NewApp(
	logger *log.Logger,
	db *pgxpool.Pool,
	deps route.Dependencies,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnvWithDefault("APP_NAME", "API Students"),
		BodyLimit:    1 * 1024 * 1024, // 1 MB — cegah DoS via payload besar
		ErrorHandler: newErrorHandler(logger),
	})

	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS"))

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

	route.Register(app, deps)

	logger.Println("Fiber application initialized")
	return app
}

func newErrorHandler(logger *log.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		if code == fiber.StatusRequestEntityTooLarge {
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": "payload terlalu besar (maks 1 MB)",
			})
		}
		logger.Printf("unhandled error: %v", err)
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"message": "terjadi kesalahan internal",
		})
	}
}
