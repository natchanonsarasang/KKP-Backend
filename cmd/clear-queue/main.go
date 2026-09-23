package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Empties the call_queue collection (the KKP_Data queue the voicebot fetches
// mid-call). Queue rows are transient — one is pushed per placed call and deleted
// when the call finishes — so clearing leftovers (e.g. the seed row) is safe.
//
//	go run ./cmd/clear-queue          # dry run: prints how many rows exist
//	go run ./cmd/clear-queue --yes    # actually deletes all rows
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")
	if uri == "" || dbName == "" {
		log.Fatal("MONGODB_URI and MONGODB_NAME must be set")
	}

	confirmed := false
	for _, arg := range os.Args[1:] {
		if arg == "--yes" || arg == "-y" {
			confirmed = true
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database(dbName).Collection("call_queue")
	log.Printf("target database: %q, collection: call_queue", dbName)

	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Fatalf("count call_queue: %v", err)
	}

	if !confirmed {
		log.Printf("DRY RUN — would delete %d call_queue rows. Re-run with --yes to proceed.", count)
		return
	}

	res, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatalf("delete call_queue: %v", err)
	}
	log.Printf("done — cleared call_queue, deleted=%d", res.DeletedCount)
}
