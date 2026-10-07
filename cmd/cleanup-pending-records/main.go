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

// Removes stale "pending" call_records — ones the dialer created but whose
// webhook never landed (e.g. early calls where the webhook arrived without a
// callerNumber and the handler bailed). These pile up and, before CreatedAt was
// stamped on insert, confused the webhook's "most recent pending" correlation.
//
//	go run ./cmd/cleanup-pending-records
//	go run ./cmd/cleanup-pending-records 0659368915      # only this phone
//
// "Stale" means status=="pending" and older than staleAfter (records younger
// than that may be a live call still awaiting its webhook). It is re-runnable.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")
	if uri == "" || dbName == "" {
		log.Fatal("MONGODB_URI and MONGODB_NAME must be set")
	}

	// Only delete pending records older than this, so a live call still waiting
	// for its webhook is never swept out from under the handler.
	const staleAfter = 10 * time.Minute
	cutoff := time.Now().UTC().Add(-staleAfter)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	records := client.Database(dbName).Collection("call_records")

	filter := bson.M{
		"status": "pending",
		// Legacy records inserted before CreatedAt was stamped have no created_at
		// field ($lt won't match a missing field), so match both: too old, or
		// never stamped at all.
		"$or": []bson.M{
			{"created_at": bson.M{"$lt": cutoff}},
			{"created_at": bson.M{"$exists": false}},
		},
	}
	// Optional phone-number argument scopes the cleanup to one test number.
	if len(os.Args) > 1 && os.Args[1] != "" {
		filter["phone_number"] = os.Args[1]
		log.Printf("scoping to phone_number=%q", os.Args[1])
	}

	matched, err := records.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatalf("count: %v", err)
	}
	if matched == 0 {
		log.Println("no stale pending call_records found — nothing to do")
		return
	}

	res, err := records.DeleteMany(ctx, filter)
	if err != nil {
		log.Fatalf("delete: %v", err)
	}
	log.Printf("removed stale pending call_records=%d (older than %s)", res.DeletedCount, staleAfter)
}
