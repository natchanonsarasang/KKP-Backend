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

// Seed a single call_queue row so GET /api/v1/kkp-data returns it, letting you
// test the voicebot's KKP_Data fetch without running the whole call flow.
//
//	go run ./cmd/seed-queue
//
// Re-runnable (upsert by fixed id). The Variables map is exactly what the
// KKP_Data endpoint serves back, so edit it to match the prompt's variables.
const (
	queueID     = "99999999-9999-9999-9999-999999999999"
	outboundID  = "outbound_test-queue-0957380848"
	phoneNumber = "0957380848"
	userID      = "1895de87-a230-4435-a5de-f8ade5c5b219"
	workspaceID = "e0f85108-036d-4a34-aab4-d4841b49fe7c"
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

	// Reference date (Asia/Bangkok) the agent uses for relative-date math.
	loc, _ := time.LoadLocation("Asia/Bangkok")
	today := time.Now()
	if loc != nil {
		today = today.In(loc)
	}

	row := entities.CallQueueModel{
		ID:             queueID,
		OutboundID:     outboundID,
		CallListItemID: "test-item-0957380848",
		DebtorID:       "test-debtor-0957380848",
		SessionID:      "test-session",
		WorkspaceID:    workspaceID,
		UserID:         userID,
		PhoneNumber:    phoneNumber,
		// Exactly what GET /api/v1/kkp-data returns. Raw values — the agent prompt
		// formats numbers/plate/dates itself.
		Variables: map[string]string{
			"customer_name":       "คุณสมชาย ใจดี",
			"car_detail":          "ฅฆ 9091",
			"province":            "ประจวบคีรีขันธ์",
			"product_type":        "ค่างวดรถยนต์",
			"overdue_installment": "2",
			"total_debt":          "15000",
			"total_interest":      "500",
			"total_fine":          "200",
			"other_expense":       "0",
			"account_number":      "1234567890",
			"current_date":        today.Format("2006-01-02"),
		},
		CreatedAt: time.Now().UTC(),
	}

	_, err = db.Collection("call_queue").ReplaceOne(
		ctx, bson.M{"id": queueID}, row, options.Replace().SetUpsert(true),
	)
	if err != nil {
		log.Fatalf("upsert call_queue: %v", err)
	}

	log.Println("Seed complete.")
	log.Printf("  queue id     = %s", queueID)
	log.Printf("  outbound_id  = %s (phone %s)", outboundID, phoneNumber)
	log.Println("Fetch it:   GET  /api/v1/kkp-data")
	log.Printf("Remove it:  POST /api/v1/webhooks/botnoi { \"outbound_id\": \"%s\", \"status\": \"completed\" }", outboundID)
}
