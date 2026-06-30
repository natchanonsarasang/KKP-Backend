package repositories

import (
	"context"
	. "go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"
	"os"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type callAttemptsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallAttemptsRepository interface {
	Insert(data entities.CallAttemptModel) error
	FindAllByWorkspace(workspaceID string, userID string) (*[]entities.CallAttemptModel, error)
	FindByID(id string) (*entities.CallAttemptModel, error)
	FindByIDByUser(id string, workspaceID string) (*entities.CallAttemptModel, error)
	FindByFilter(filter entities.CallAttemptFilter) (*[]entities.CallAttemptModel, error)
	// System Methods
	Update(id string, data entities.CallAttemptModel) error
	Delete(id string) error
	// ByUser Methods
	UpdateByUser(id string, workspaceID string, userID string, data entities.CallAttemptModel) error
	DeleteByUser(id string, workspaceID string, userID string) error
	UpdateMultipleByUser(filter entities.CallAttemptFilter, data entities.CallAttemptModel) (int64, error)
	// Process session Methods
	UpdateStatusByListItemID(listItemID, fromStatus, toStatus, outcome string, pickedUp bool, errorReason string) error
}

func NewCallAttemptsRepository(db *MongoDB) ICallAttemptsRepository {
	return &callAttemptsRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("call_attempts"),
	}
}

func (repo *callAttemptsRepository) Insert(data entities.CallAttemptModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("CallAttempts -> Insert: %s \n", err)
		return err
	}
	return nil
}

func (repo *callAttemptsRepository) FindAllByWorkspace(workspaceID string, userID string) (*[]entities.CallAttemptModel, error) {
	// userID is empty for system/webhook lookups (workspace-wide). Only scope by
	// user when a userID is actually provided, otherwise we'd filter on
	// user_id == "" and match nothing.
	filter := bson.M{"workspace_id": workspaceID}
	if userID != "" {
		filter["user_id"] = userID
	}
	var attempts []entities.CallAttemptModel
	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> FindAllByWorkspace: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)
	if err := cursor.All(repo.Context, &attempts); err != nil {
		fiberlog.Errorf("CallAttempts -> FindAllByWorkspace: %s \n", err)
		return nil, err
	}
	return &attempts, nil
}

func (repo *callAttemptsRepository) FindByID(id string) (*entities.CallAttemptModel, error) {
	filter := bson.M{"id": id}
	var attempt entities.CallAttemptModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&attempt); err != nil {
		fiberlog.Errorf("CallAttempts -> FindByID: %s \n", err)
		return nil, err
	}
	return &attempt, nil
}

func (repo *callAttemptsRepository) FindByIDByUser(id string, workspaceID string) (*entities.CallAttemptModel, error) {
	filter := bson.M{"id": id, "workspace_id": workspaceID}
	var attempt entities.CallAttemptModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&attempt); err != nil {
		fiberlog.Errorf("CallAttempts -> FindByIDByUser: %s \n", err)
		return nil, err
	}
	return &attempt, nil
}

func (repo *callAttemptsRepository) Update(id string, data entities.CallAttemptModel) error {
	filter := bson.M{"id": id}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> Update: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callAttemptsRepository) Delete(id string) error {
	filter := bson.M{"id": id}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> Delete: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callAttemptsRepository) UpdateByUser(id string, workspaceID string, userID string, data entities.CallAttemptModel) error {
	filter := bson.M{"id": id, "workspace_id": workspaceID, "user_id": userID}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> UpdateByUser: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callAttemptsRepository) DeleteByUser(id string, workspaceID string, userID string) error {
	filter := bson.M{"id": id, "workspace_id": workspaceID, "user_id": userID}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> DeleteByUser: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callAttemptsRepository) FindByFilter(filter entities.CallAttemptFilter) (*[]entities.CallAttemptModel, error) {
	queryFilter := bson.M{
		"workspace_id": filter.WorkspaceID,
	}
	if filter.UserID != "" {
		queryFilter["user_id"] = filter.UserID
	}
	if filter.CallListItemID != "" {
		queryFilter["call_list_item_id"] = filter.CallListItemID
	}
	if filter.Status != "" {
		queryFilter["status"] = filter.Status
	}

	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(filter.Limit)
	}

	var attempts []entities.CallAttemptModel
	cursor, err := repo.Collection.Find(repo.Context, queryFilter, findOptions)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> FindByFilter: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)
	if err := cursor.All(repo.Context, &attempts); err != nil {
		fiberlog.Errorf("CallAttempts -> FindByFilter: %s \n", err)
		return nil, err
	}
	return &attempts, nil
}



func (repo *callAttemptsRepository) UpdateStatusByListItemID(listItemID, fromStatus, toStatus, outcome string, pickedUp bool, errorReason string) error {
	filter := bson.M{"call_list_item_id": listItemID, "status": fromStatus}
	update := bson.M{"$set": bson.M{
		"status":       toStatus,
		"call_outcome": outcome,
		"picked_up":    pickedUp,
		"error_reason": errorReason,
	}}
	_, err := repo.Collection.UpdateMany(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> UpdateStatusByListItemID: %s \n", err)
	}
	return err
}

func (repo *callAttemptsRepository) UpdateMultipleByUser(filter entities.CallAttemptFilter, data entities.CallAttemptModel) (int64, error) {
	queryFilter := bson.M{
		"workspace_id": filter.WorkspaceID,
		"user_id":      filter.UserID,
	}
	if filter.CallListItemID != "" {
		queryFilter["call_list_item_id"] = filter.CallListItemID
	}
	if filter.Status != "" {
		queryFilter["status"] = filter.Status
	}

	// Prevent updating immutable fields
	data.ID = ""
	data.UserID = ""
	data.WorkspaceID = ""

	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateMany(repo.Context, queryFilter, update)
	if err != nil {
		fiberlog.Errorf("CallAttempts -> UpdateMultipleByUser: %s \n", err)
		return 0, err
	}
	return result.ModifiedCount, nil
}
