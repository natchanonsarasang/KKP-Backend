package repositories

import (
	"context"
	"os"
	"testing"
	"time"

	"go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestCallRecordsRepositoryCRUD(t *testing.T) {
	// Load environment variables from parent paths
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		t.Skip("Skipping integration test: MONGODB_URI not set")
	}

	db := datasources.NewMongoDB(10)
	
	// Test connectivity to MongoDB with a short timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()
	if err := db.MongoDB.Ping(pingCtx, nil); err != nil {
		t.Skipf("Skipping integration test: MongoDB connection unreachable: %v", err)
	}

	repo := NewCallRecordsRepository(db)

	testID := "test-record-uuid-1234"
	
	// Clean up any leftovers before test and after test
	_ = repo.DeleteCallRecord(testID)
	defer func() {
		_ = repo.DeleteCallRecord(testID)
	}()

	now := time.Now().UTC().Truncate(time.Second)
	dueDate := now.Add(24 * time.Hour)

	templateID := "test-tpl"
	var resultData interface{} = map[string]interface{}{"foo": "bar"}

	// 1. Insert
	record := entities.CallRecordDataModel{
		ID:              testID,
		TemplateID:      &templateID,
		PhoneNumber:     "+123456",
		AppointmentDate: "2026-06-16",
		AppointmentTime: "12:00:00",
		Status:          entities.StatusPending,
		BotnoiCallID:    "test-botnoi-id",
		ResultData:      &resultData,
		DueDate:         dueDate,
		Amount:          200.75,
		UserID:          "test-user",
		WorkspaceID:     "test-workspace",
		CallDuration:    45,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := repo.InsertCallRecord(record)
	assert.NoError(t, err)

	// 2. FindByID
	found, err := repo.FindByID(testID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testID, found.ID)
	assert.Equal(t, entities.StatusPending, found.Status)
	assert.Equal(t, 200.75, found.Amount)
	assert.True(t, found.DueDate.Equal(dueDate))
	assert.NotNil(t, found.TemplateID)
	assert.Equal(t, "test-tpl", *found.TemplateID)
	assert.NotNil(t, found.ResultData)
	resultDataMap, ok := (*found.ResultData).(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "bar", resultDataMap["foo"])

	// 3. Update
	newTemplateID := "test-tpl-updated"
	found.TemplateID = &newTemplateID
	found.Status = entities.StatusCompleted
	found.CallDuration = 90
	err = repo.UpdateCallRecord(testID, *found)
	assert.NoError(t, err)

	// Verify Update
	foundUpdated, err := repo.FindByID(testID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUpdated)
	assert.Equal(t, entities.StatusCompleted, foundUpdated.Status)
	assert.Equal(t, 90, foundUpdated.CallDuration)
	assert.NotNil(t, foundUpdated.TemplateID)
	assert.Equal(t, "test-tpl-updated", *foundUpdated.TemplateID)

	// 4. FindAll
	allRecords, err := repo.FindAll()
	assert.NoError(t, err)
	assert.NotNil(t, allRecords)
	
	foundInList := false
	for _, r := range *allRecords {
		if r.ID == testID {
			foundInList = true
			break
		}
	}
	assert.True(t, foundInList)

	// 4.5. FindByUserID
	userRecords, err := repo.FindByUserID("test-user")
	assert.NoError(t, err)
	assert.NotNil(t, userRecords)
	foundUserRec := false
	for _, r := range *userRecords {
		if r.ID == testID {
			foundUserRec = true
			break
		}
	}
	assert.True(t, foundUserRec)

	// 5. Delete
	err = repo.DeleteCallRecord(testID)
	assert.NoError(t, err)

	// Verify Delete
	deleted, err := repo.FindByID(testID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}
