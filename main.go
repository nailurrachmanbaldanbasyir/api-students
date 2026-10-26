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

	// Koneksi PostgreSQL
	db, err := database.NewPostgresPool()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Koneksi MongoDB
	mongoCtx, mongoCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	mongoClient, err := database.NewMongoClient(mongoCtx)
	mongoCancel()

	if err != nil {
		logger.Fatalf("gagal terhubung ke MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Printf("gagal menutup koneksi MongoDB: %v", err)
		}
	}()

	mongoDB := mongoClient.Database(
		config.GetEnvWithDefault("MONGO_DATABASE", "api_students"),
	)

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

	// Repository portfolio MongoDB
	portfolioRepo := repository.NewPortfolioRepository(mongoDB)

	// Pastikan index student_id unik tersedia
	indexCtx, indexCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	err = portfolioRepo.EnsureIndexes(indexCtx)
	indexCancel()
	if err != nil {
		logger.Fatalf("gagal menyiapkan index portfolio: %v", err)
	}

	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Printf("gagal memuat permission: %v", err)
		os.Exit(1)
	}

	permissions := helper.NewPermissionSet(rawPermissions)
	logger.Printf("permission dimuat, roles=%v", permissions.KnownRoles())

	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		jwtManager,
		permissions,
		refreshTTL,
	)

	portfolioService := service.NewPortfolioService(
		studentRepo,
		portfolioRepo,
	)

	app := config.NewApp(logger, db, route.Dependencies{
		Pool:             db,
		JWT:              jwtManager,
		Permissions:      permissions,
		StudentRepo:      studentRepo,
		UserRepo:         userRepo,
		TokenRepo:        tokenRepo,
		AuthService:      authService,
		PortfolioService: portfolioService,
	})

	logger.Println("server berjalan pada port 3000")

	if err := app.Listen(":3000"); err != nil {
		logger.Fatal(err)
	}
}
