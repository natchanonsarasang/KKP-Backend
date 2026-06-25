package main

import (
	"context"
	"log"
	"os"
	"time"

	"go-fiber-template/domain/entities"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Seed data for a real "process call session" test.
//
//	go run ./cmd/seed
//
// It inserts (upsert by id, so it's safe to re-run):
//   - one debtor
//   - one pending call_list_item
//   - one running call_session
//
// After seeding, trigger processing with the printed session id:
//
//	POST /api/v1/call-process  { "session_id": "<id>", "action": "start" }
const (
	userID      = "1895de87-a230-4435-a5de-f8ade5c5b219"
	workspaceID = "e0f85108-036d-4a34-aab4-d4841b49fe7c"
	phoneNumber = "0968871055"
	debtorID    = "66038941-cdc0-4b1e-ab2e-44f57667a188"

	// Fixed ids so re-running the seed updates the same docs instead of duplicating.
	sessionID = "11111111-1111-1111-1111-111111111111"
	itemID    = "22222222-2222-2222-2222-222222222222"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")
	if uri == "" || dbName == "" {
		log.Fatal("MONGODB_URI and MONGODB_NAME must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)
	db := client.Database(dbName)

	now := time.Now().UTC()

	debtor := entities.DebtorModel{
		ID:          debtorID,
		PhoneNumber: phoneNumber,
		Name:        "ทดสอบ",
		LastName:    "ระบบ",
		Status:      "active",
		Variables: map[string]string{
			"name":                "คุณทดสอบ",
			"outstanding_amount":  "1500",
			"overdue_installment": "2",
			"due_date":            "25 มิถุนายน 2569",
			"policy_no":           "1234567890",
		},
		UserID:      userID,
		WorkspaceID: workspaceID,
		IsBlocked:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	item := entities.CallListItemModel{
		ID:          itemID,
		UserID:      userID,
		DebtorID:    debtorID,
		WorkspaceID: workspaceID,
		Status:      "pending",
		RetryCount:  0,
		ScheduledAt: &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	session := entities.CallSessionDataModel{
		ID:          sessionID,
		UserID:      userID,
		WorkspaceID: workspaceID,
		Status:      "running",
		TotalCalls:  1,
		Settings: entities.CallSessionSettings{
			MaxRetries:        1,
			ConcurrentCalls:   5,
			BusinessHoursOnly: false,
			TestMode:          false,
			Interruptible:     true,
		},
		StartedAt: &now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	upsert(ctx, db.Collection("debtors"), debtorID, debtor)
	upsert(ctx, db.Collection("call_list_items"), itemID, item)
	upsert(ctx, db.Collection("call_sessions"), sessionID, session)

	log.Println("Seed complete.")
	log.Printf("  debtor_id  = %s (phone %s)", debtorID, phoneNumber)
	log.Printf("  item_id    = %s (status=pending)", itemID)
	log.Printf("  session_id = %s (status=running)", sessionID)
	log.Println("Trigger: POST /api/v1/call-process { \"session_id\": \"" + sessionID + "\", \"action\": \"start\" }")
}

func upsert(ctx context.Context, col *mongo.Collection, id string, doc any) {
	_, err := col.ReplaceOne(ctx, bson.M{"id": id}, doc, options.Replace().SetUpsert(true))
	if err != nil {
		log.Fatalf("upsert into %s (%s): %v", col.Name(), id, err)
	}
	log.Printf("upserted %s id=%s", col.Name(), id)
}
