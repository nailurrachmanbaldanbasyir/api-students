
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
	Pool             *pgxpool.Pool
	JWT              *helper.JWTManager
	Permissions      *helper.PermissionSet
	StudentRepo      *repository.StudentRepository
	UserRepo         *repository.UserRepository
	TokenRepo        *repository.TokenRepository
	AuthService      *service.AuthService
	StudentService   any
	PortfolioService *service.PortfolioService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON())
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	students := api.Group(
		"/students",
		middleware.RequireJSON(),
		middleware.RequireAuth(deps.JWT),
	)

	students.Get(
		"/",
		middleware.RequirePermission(deps.Permissions, "student:list"),
		listStudents(deps.StudentRepo),
	)
	students.Get("/:id", getStudent(deps.StudentRepo, deps.Permissions))
	students.Post(
		"/",
		middleware.RequirePermission(deps.Permissions, "student:create"),
		createStudent(deps.StudentRepo),
	)
	students.Put("/:id", replaceStudent(deps.StudentRepo, deps.Permissions))
	students.Patch("/:id", patchStudent(deps.StudentRepo, deps.Permissions))
	students.Delete(
		"/:id",
		middleware.RequirePermission(deps.Permissions, "student:delete"),
		deleteStudent(deps.StudentRepo),
	)

	// Portfolio mahasiswa
	students.Get("/:id/portfolio", deps.PortfolioService.Get)
	students.Put("/:id/portfolio", deps.PortfolioService.Put)
	students.Post("/:id/portfolio/projects", deps.PortfolioService.AddProject)
	students.Delete(
		"/:id/portfolio/projects/:projectId",
		deps.PortfolioService.DeleteProject,
	)

	// Statistik teknologi portfolio
	api.Get(
		"/portfolios/stats/technologies",
		middleware.RequireAuth(deps.JWT),
		deps.PortfolioService.TechnologyStats,
	)
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
