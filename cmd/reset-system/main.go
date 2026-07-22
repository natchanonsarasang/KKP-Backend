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

// Wipes all operational/transactional data so the system starts fresh, while
// keeping everything needed to log in and keep workspaces/config intact.
//
//	go run ./cmd/reset-system            # dry run: prints per-collection counts
//	go run ./cmd/reset-system --yes      # actually deletes
//
// WIPED    : debtors, call_list_items, call_attempts, call_records, call_sessions
// PRESERVED: users, workspaces, call_templates, call_tokens (login + config)
//
// This is destructive and irreversible — the --yes guard is required so a bare
// run never deletes anything.
var wipeCollections = []string{
	"debtors",
	"call_list_items",
	"call_attempts",
	"call_records",
	"call_sessions",
}

var preservedCollections = []string{
	"users",
	"workspaces",
	"call_templates",
	"call_tokens",
}

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

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)
	log.Printf("target database: %q", dbName)
	log.Printf("preserved (kept): %v", preservedCollections)

	if !confirmed {
		log.Println("DRY RUN — no data will be deleted. Re-run with --yes to proceed.")
		for _, name := range wipeCollections {
			count, err := db.Collection(name).CountDocuments(ctx, bson.M{})
			if err != nil {
				log.Printf("  %-16s count error: %v", name, err)
				continue
			}
			log.Printf("  would delete %-16s docs=%d", name, count)
		}
		return
	}

	log.Println("WIPING operational data (--yes given)...")
	var total int64
	for _, name := range wipeCollections {
		res, err := db.Collection(name).DeleteMany(ctx, bson.M{})
		if err != nil {
			log.Fatalf("delete %s: %v", name, err)
		}
		total += res.DeletedCount
		log.Printf("  cleared %-16s deleted=%d", name, res.DeletedCount)
	}
	log.Printf("done — total documents deleted=%d (login + config preserved)", total)
}
