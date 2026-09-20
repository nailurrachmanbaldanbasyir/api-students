package middleware

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"

	"api-students/helper"
)

func Register(app *fiber.App, logger *log.Logger, allowedOrigins string) {
	app.Use(RequestID())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(corsPolicy(allowedOrigins))
	app.Use(RequestLogger(logger))
	app.Use(RequireJSON())
}

func corsPolicy(allowedOrigins string) fiber.Handler {
	if strings.TrimSpace(allowedOrigins) == "" {
		allowedOrigins = "http://localhost:5173"
	}

	return cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
	})
}

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

func RequestLogger(logger *log.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		entry := struct {
			RequestID string `json:"request_id"`
			Method    string `json:"method"`
			Path      string `json:"path"`
			Status    int    `json:"status"`
			Duration  string `json:"duration"`
			UserID    int    `json:"user_id,omitempty"`
			Role      string `json:"role,omitempty"`
		}{
			RequestID: c.Locals("request_id").(string),
			Method:    c.Method(),
			Path:      c.Path(),
			Status:    c.Response().StatusCode(),
			Duration:  time.Since(start).String(),
		}

		if user, ok := helper.CurrentUser(c); ok {
			entry.UserID = user.UserID
			entry.Role = user.Role
		}

		data, errJSON := json.Marshal(entry)
		if errJSON == nil {
			logger.Println(string(data))
		}

		return err
	}
}

func RequireJSON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodPost, fiber.MethodPut, fiber.MethodPatch:
			contentType := c.Get("Content-Type")

			if contentType == "" || !strings.HasPrefix(
				strings.ToLower(contentType),
				"application/json",
			) {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"success": false,
					"message": "Content-Type harus application/json",
				})
			}
		}

		return c.Next()
	}
}
