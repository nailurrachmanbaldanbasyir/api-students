package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	JWT         *helper.JWTManager
	StudentRepo *repository.StudentRepository
	UserRepo    *repository.UserRepository
	TokenRepo   *repository.TokenRepository
	AuthService *service.AuthService
	StudentService interface{} // placeholder jika ada
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	// --- publik ---
	api.Get("/health", healthCheck(deps.Pool))

	// --- autentikasi ---
	auth := api.Group("/auth", middleware.RequireJSON())
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// --- students: wajib membawa access token ---
	students := api.Group("/students", middleware.RequireJSON(), middleware.RequireAuth(deps.JWT))
	students.Get("/", listStudents(deps.StudentRepo))
	students.Get("/:id", getStudent(deps.StudentRepo))
	students.Post("/", createStudent(deps.StudentRepo))
	students.Put("/:id", replaceStudent(deps.StudentRepo))
	students.Patch("/:id", patchStudent(deps.StudentRepo))
	students.Delete("/:id", deleteStudent(deps.StudentRepo))
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"success": false,
				"message": "database tidak tersedia",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "service sehat",
		})
	}
}
