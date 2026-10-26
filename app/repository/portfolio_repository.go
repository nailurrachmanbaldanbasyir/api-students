package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ErrPortfolioNotFound = errors.New("portfolio tidak ditemukan")
	ErrPortfolioLimit    = errors.New("maksimal 20 proyek per portfolio")
)

type PortfolioRepository struct {
	collection *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) *PortfolioRepository {
	return &PortfolioRepository{
		collection: db.Collection("student_portfolios"),
	}
}

func (r *PortfolioRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("membuat index portfolio: %w", err)
	}
	return nil
}

func (r *PortfolioRepository) Get(
	ctx context.Context,
	studentID int,
) (model.Portfolio, error) {
	var portfolio model.Portfolio

	err := r.collection.FindOne(
		ctx,
		bson.D{{Key: "student_id", Value: studentID}},
	).Decode(&portfolio)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, ErrPortfolioNotFound
	}
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("membaca portfolio: %w", err)
	}

	return portfolio, nil
}

func (r *PortfolioRepository) Upsert(
	ctx context.Context,
	portfolio model.Portfolio,
) error {
	filter := bson.D{{Key: "student_id", Value: portfolio.StudentID}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "headline", Value: portfolio.Headline},
		{Key: "bio", Value: portfolio.Bio},
		{Key: "projects", Value: portfolio.Projects},
		{Key: "certificates", Value: portfolio.Certificates},
	}}}

	_, err := r.collection.UpdateOne(
		ctx,
		filter,
		update,
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("menyimpan portfolio: %w", err)
	}

	return nil
}

func (r *PortfolioRepository) AddProject(
	ctx context.Context,
	studentID int,
	project model.PortfolioProject,
) error {
	filter := bson.D{
		{Key: "student_id", Value: studentID},
		{Key: "$expr", Value: bson.D{{Key: "$lt", Value: bson.A{
			bson.D{{Key: "$size", Value: bson.D{{Key: "$ifNull", Value: bson.A{"$projects", bson.A{}}}}}},
			20,
		}}}},
	}

	result, err := r.collection.UpdateOne(
		ctx,
		filter,
		bson.D{{Key: "$push", Value: bson.D{
			{Key: "projects", Value: project},
		}}},
	)
	if err != nil {
		return fmt.Errorf("menambahkan proyek: %w", err)
	}
	if result.MatchedCount == 0 {
		_, err := r.Get(ctx, studentID)
		if err != nil {
			return err
		}
		return ErrPortfolioLimit
	}

	return nil
}

func (r *PortfolioRepository) DeleteProject(
	ctx context.Context,
	studentID int,
	projectID string,
) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "student_id", Value: studentID}},
		bson.D{{Key: "$pull", Value: bson.D{
			{Key: "projects", Value: bson.D{
				{Key: "id", Value: projectID},
			}},
		}}},
	)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if result.MatchedCount == 0 {
		return ErrPortfolioNotFound
	}
	if result.ModifiedCount == 0 {
		return errors.New("proyek tidak ditemukan")
	}

	return nil
}

func (r *PortfolioRepository) TechnologyStats(
	ctx context.Context,
) ([]bson.M, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$projects.technologies"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("menghitung statistik teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("membaca statistik teknologi: %w", err)
	}
	if results == nil {
		results = []bson.M{}
	}

	return results, nil
}
