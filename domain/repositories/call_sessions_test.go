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

func TestCallSessionsRepositoryCRUD(t *testing.T) {
	// Load environment variables from parent paths
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		t.Skip("Skipping integration test: MONGODB_URI not set")
	}

	db := datasources.NewMongoDB(10)
	
	// Test connectivity to MongoDB with a longer timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()
	if err := db.MongoDB.Ping(pingCtx, nil); err != nil {
		t.Skipf("Skipping integration test: MongoDB connection unreachable: %v", err)
	}

	repo := NewCallSessionsRepository(db)

	testID := "test-session-uuid-1234"
	
	// Clean up any leftovers before test and after test
	_ = repo.DeleteCallSession(testID)
	defer func() {
		_ = repo.DeleteCallSession(testID)
	}()

	now := time.Now().UTC().Truncate(time.Second)

	// 1. Insert
	session := entities.CallSessionDataModel{
		ID:             testID,
		UserID:         "test-session-user",
		WorkspaceID:    "test-session-workspace",
		Status:         "test-status-pending",
		TotalCalls:     10,
		CompletedCalls: 5,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.InsertCallSession(session)
	assert.NoError(t, err)

	// 2. FindByID
	found, err := repo.FindByID(testID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testID, found.ID)
	assert.Equal(t, "test-session-user", found.UserID)
	assert.Equal(t, "test-status-pending", found.Status)

	// 3. FindOneByStatus
	foundOne, err := repo.FindOneByStatus("test-status-pending")
	assert.NoError(t, err)
	assert.NotNil(t, foundOne)
	assert.Equal(t, "test-status-pending", foundOne.Status)

	// 4. FindByStatus
	allByStatus, err := repo.FindByStatus("test-status-pending")
	assert.NoError(t, err)
	assert.NotNil(t, allByStatus)
	foundInStatusList := false
	for _, s := range *allByStatus {
		if s.ID == testID {
			foundInStatusList = true
			break
		}
	}
	assert.True(t, foundInStatusList)

	// 5. FindByWorkspaceID
	allByWorkspace, err := repo.FindByWorkspaceID("test-session-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, allByWorkspace)
	foundInWorkspaceList := false
	for _, s := range *allByWorkspace {
		if s.ID == testID {
			foundInWorkspaceList = true
			break
		}
	}
	assert.True(t, foundInWorkspaceList)

	// 6. FindByUserID
	allByUser, err := repo.FindByUserID("test-session-user")
	assert.NoError(t, err)
	assert.NotNil(t, allByUser)
	foundInUserList := false
	for _, s := range *allByUser {
		if s.ID == testID {
			foundInUserList = true
			break
		}
	}
	assert.True(t, foundInUserList)

	// 7. Update
	found.Status = "test-status-running"
	found.CompletedCalls = 8
	err = repo.UpdateCallSession(testID, *found)
	assert.NoError(t, err)

	// Verify Update
	foundUpdated, err := repo.FindByID(testID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUpdated)
	assert.Equal(t, "test-status-running", foundUpdated.Status)
	assert.Equal(t, 8, foundUpdated.CompletedCalls)

	// 8. Delete
	err = repo.DeleteCallSession(testID)
	assert.NoError(t, err)

	// Verify Delete
	deleted, err := repo.FindByID(testID)
	assert.NoError(t, err)
	assert.Nil(t, deleted)
}
