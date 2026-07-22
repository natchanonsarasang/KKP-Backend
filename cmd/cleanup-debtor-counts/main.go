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

// Removes the obsolete accept_count / reject_count / other_count fields from all
// debtor documents. These counters were dropped from the code (they came from the
// webhook `action` field, which is no longer used), but old documents still carry
// the stale values. This $unset is a one-off, re-runnable cleanup.
//
//	go run ./cmd/cleanup-debtor-counts
//
// It is a no-op on documents that no longer have the fields, so re-running is safe.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")
	if uri == "" || dbName == "" {
		log.Fatal("MONGODB_URI and MONGODB_NAME must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database(dbName).Collection("debtors")

	// Only touch documents that still have at least one of the fields, so the
	// reported "matched" count reflects real cleanup rather than every debtor.
	filter := bson.M{
		"$or": bson.A{
			bson.M{"accept_count": bson.M{"$exists": true}},
			bson.M{"reject_count": bson.M{"$exists": true}},
			bson.M{"other_count": bson.M{"$exists": true}},
		},
	}
	update := bson.M{
		"$unset": bson.M{
			"accept_count": "",
			"reject_count": "",
			"other_count":  "",
		},
	}

	res, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Fatalf("cleanup: %v", err)
	}

	log.Printf("debtors matched=%d modified=%d (removed accept_count/reject_count/other_count)", res.MatchedCount, res.ModifiedCount)
}
