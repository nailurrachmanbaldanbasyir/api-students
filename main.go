package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

const minSecretLength = 32

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Println("warning: file .env tidak ditemukan")
	}

	logger, logFile, err := config.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	jwtSecret := config.GetEnv("JWT_SECRET")
	if len(jwtSecret) < minSecretLength {
		logger.Fatalf("JWT_SECRET tidak diisi atau terlalu pendek (minimal %d karakter)", minSecretLength)
		slog.Error("JWT_SECRET tidak diisi atau terlalu pendek", slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	db, err := database.NewPostgresPool()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnvWithDefault("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)
	refreshTTL := time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	studentRepo := repository.NewStudentRepository(db)

	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Printf("gagal memuat permission: %v", err)
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Printf("permission dimuat, roles=%v", permissions.KnownRoles())

	authService := service.NewAuthService(userRepo, tokenRepo, jwtManager, permissions, refreshTTL)

	app := config.NewApp(logger, db, route.Dependencies{
		Pool:        db,
		JWT:         jwtManager,
		Permissions: permissions,
		StudentRepo: studentRepo,
		UserRepo:    userRepo,
		TokenRepo:   tokenRepo,
		AuthService: authService,
	})

	logger.Println("server berjalan pada port 3000")

	if err := app.Listen(":3000"); err != nil {
		logger.Fatal(err)
	}
}
