package database

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func NewMongoClient(ctx context.Context) (*mongo.Client, error) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	maxPool := uint64(20)
	if value := os.Getenv("MONGO_MAX_POOL"); value != "" {
		n, err := strconv.ParseUint(value, 10, 64)
		if err != nil || n == 0 {
			return nil, fmt.Errorf("MONGO_MAX_POOL harus berupa angka positif")
		}
		maxPool = n
	}

	opts := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(maxPool)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat client MongoDB: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("gagal menghubungi MongoDB: %w", err)
	}

	return client, nil
}
