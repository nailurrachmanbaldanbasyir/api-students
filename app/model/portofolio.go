package model

import "time"

type Portfolio struct {
	StudentID    int                    `json:"student_id" bson:"student_id"`
	Headline     string                 `json:"headline" bson:"headline"`
	Bio          string                 `json:"bio" bson:"bio"`
	Projects     []PortfolioProject     `json:"projects" bson:"projects"`
	Certificates []PortfolioCertificate `json:"certificates" bson:"certificates"`
}

type PortfolioProject struct {
	ID           string   `json:"id" bson:"id"`
	Title        string   `json:"title" bson:"title"`
	Technologies []string `json:"technologies" bson:"technologies"`
	Link         string   `json:"link" bson:"link"`
	Year         int      `json:"year" bson:"year"`
}

type PortfolioCertificate struct {
	Name   string    `json:"name" bson:"name"`
	Issuer string    `json:"issuer" bson:"issuer"`
	Date   time.Time `json:"date" bson:"date"`
}
