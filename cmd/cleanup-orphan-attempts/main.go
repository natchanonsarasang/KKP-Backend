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

// Removes orphaned call_attempts (and their call_records) whose parent
// call_list_item no longer exists. These get left behind when a call_list_item
// is deleted (e.g. an older "Clear Completed" that only removed the item) — the
// attempts still surface as "-" rows in the Analytics "Recent Calls" history.
//
//	go run ./cmd/cleanup-orphan-attempts
//
// It is re-runnable: it only touches attempts whose call_list_item_id points at
// a now-missing item, so a second run is a no-op.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")
	if uri == "" || dbName == "" {
		log.Fatal("MONGODB_URI and MONGODB_NAME must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)
	items := db.Collection("call_list_items")
	attempts := db.Collection("call_attempts")
	records := db.Collection("call_records")

	// Build the set of existing call_list_item ids.
	itemIDs := map[string]struct{}{}
	cur, err := items.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"id": 1}))
	if err != nil {
		log.Fatalf("load items: %v", err)
	}
	for cur.Next(ctx) {
		var row struct {
			ID string `bson:"id"`
		}
		if err := cur.Decode(&row); err == nil && row.ID != "" {
			itemIDs[row.ID] = struct{}{}
		}
	}
	cur.Close(ctx)
	log.Printf("existing call_list_items: %d", len(itemIDs))

	// Scan attempts; an attempt is orphaned when its call_list_item_id is set but
	// no longer maps to an existing item.
	cur, err = attempts.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"id": 1, "call_list_item_id": 1, "call_record_id": 1}))
	if err != nil {
		log.Fatalf("load attempts: %v", err)
	}
	orphanAttemptIDs := []string{}
	orphanRecordIDs := map[string]struct{}{}
	for cur.Next(ctx) {
		var row struct {
			ID           string `bson:"id"`
			ItemID       string `bson:"call_list_item_id"`
			CallRecordID string `bson:"call_record_id"`
		}
		if err := cur.Decode(&row); err != nil {
			continue
		}
		if row.ItemID == "" {
			continue // no parent link — leave it alone
		}
		if _, ok := itemIDs[row.ItemID]; ok {
			continue // parent still exists
		}
		orphanAttemptIDs = append(orphanAttemptIDs, row.ID)
		if row.CallRecordID != "" {
			orphanRecordIDs[row.CallRecordID] = struct{}{}
		}
	}
	cur.Close(ctx)

	if len(orphanAttemptIDs) == 0 {
		log.Println("no orphaned call_attempts found — nothing to do")
		return
	}

	delAttempts, err := attempts.DeleteMany(ctx, bson.M{"id": bson.M{"$in": orphanAttemptIDs}})
	if err != nil {
		log.Fatalf("delete attempts: %v", err)
	}

	recordIDs := make([]string, 0, len(orphanRecordIDs))
	for rid := range orphanRecordIDs {
		recordIDs = append(recordIDs, rid)
	}
	var deletedRecords int64
	if len(recordIDs) > 0 {
		delRecords, err := records.DeleteMany(ctx, bson.M{"id": bson.M{"$in": recordIDs}})
		if err != nil {
			log.Fatalf("delete records: %v", err)
		}
		deletedRecords = delRecords.DeletedCount
	}

	log.Printf("removed orphaned call_attempts=%d call_records=%d", delAttempts.DeletedCount, deletedRecords)
}
