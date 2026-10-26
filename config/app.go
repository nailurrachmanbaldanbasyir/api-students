package config

import (
	"errors"
	"log"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
	"api-students/helper"
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
		BodyLimit:    1 * 1024 * 1024,
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

	// 404 handler terpusat
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})

	logger.Println("Fiber application initialized")
	return app
}

func newErrorHandler(logger *log.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		requestID := helper.RequestID(c)
		var appErr *helper.AppError
		switch {
		case errors.As(err, &appErr):
			// sudah AppError
		case errors.Is(err, fiber.ErrRequestEntityTooLarge):
			appErr = &helper.AppError{
				Status: fiber.StatusRequestEntityTooLarge,
				Code:   helper.CodePayloadTooLarge,
				Message: "ukuran body melebihi batas yang diizinkan",
			}
		default:
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				appErr = &helper.AppError{
					Status: fiberErr.Code, Code: helper.CodeHTTPError,
					Message: fiberErr.Message,
				}
			} else {
				appErr = helper.Internal(err)
			}
		}

		// 5xx ERROR, 4xx WARN (jangan tertukar)
		if appErr.Status >= fiber.StatusInternalServerError {
			causeMsg := ""
			if appErr.Unwrap() != nil {
				causeMsg = appErr.Unwrap().Error()
			}
			// slog for structured log
			slog.Error("request_failed",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
				slog.String("error", causeMsg),
			)
			logger.Printf(`{"level":"ERROR","msg":"request_failed","request_id":"%s","path":"%s","code":"%s","status":%d,"error":"%s"}`, requestID, c.Path(), appErr.Code, appErr.Status, causeMsg)
		} else {
			slog.Warn("request_rejected",
				slog.String("request_id", requestID),
				slog.String("path", c.Path()),
				slog.String("code", appErr.Code),
				slog.Int("status", appErr.Status),
			)
			logger.Printf(`{"level":"WARN","msg":"request_rejected","request_id":"%s","path":"%s","code":"%s","status":%d}`, requestID, c.Path(), appErr.Code, appErr.Status)
		}

		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}
}
