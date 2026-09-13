package main

import (
	"log"

	"api-students/app/repository"
	"api-students/config"
	"api-students/database"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Println("warning: file .env tidak ditemukan")
	}

	logger, logFile, err := config.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	db, err := database.NewPostgresPool()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	studentRepo := repository.NewStudentRepository(db)

	app := config.NewApp(
		logger,
		db,
		studentRepo,
	)

	logger.Println("server berjalan pada port 3000")

	if err := app.Listen(":3000"); err != nil {
		logger.Fatal(err)
	}
}
