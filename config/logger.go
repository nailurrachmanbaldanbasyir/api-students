package config

import (
	"io"
	"log"
	"os"
)

func NewLogger() (*log.Logger, *os.File, error) {
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, nil, err
	}

	file, err := os.OpenFile(
		"logs/app.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)

	logger := log.New(
		multiWriter,
		"",
		log.Ldate|log.Ltime,
	)

	return logger, file, nil
}
