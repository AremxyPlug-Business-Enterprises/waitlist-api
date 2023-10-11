package db

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func ConnectDatabase() *mongo.Database {
	uri := os.Getenv("DB_URI")
	dbName := os.Getenv("DB_NAME")
	ctx := context.Background()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("error connecting to database: ", err)
	}

	db = client.Database(dbName)
	return db
}
