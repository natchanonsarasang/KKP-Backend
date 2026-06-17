package repositories

import (
	"context"
	. "go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"
	"os"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type callListItemsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallListItemsRepository interface {
	Insert(data entities.CallListItemModel) error
	FindAllByWorkspace(workspaceID string, userID string) (*[]entities.CallListItemModel, error)
	FindByID(id string) (*entities.CallListItemModel, error)
	FindByIDByUser(id string, workspaceID string) (*entities.CallListItemModel, error)
	FindByFilter(filter entities.CallListItemFilter) (*[]entities.CallListItemModel, error)
	// System Methods
	Update(id string, data entities.CallListItemModel) error
	Delete(id string) error
	// ByUser Methods
	UpdateByUser(id string, workspaceID string, userID string, data entities.CallListItemModel) error
	DeleteByUser(id string, workspaceID string, userID string) error
}

func NewCallListItemsRepository(db *MongoDB) ICallListItemsRepository {
	return &callListItemsRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("call_list_items"),
	}
}

func (repo *callListItemsRepository) Insert(data entities.CallListItemModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("CallListItems -> Insert: %s \n", err)
		return err
	}
	return nil
}

func (repo *callListItemsRepository) FindAllByWorkspace(workspaceID string, userID string) (*[]entities.CallListItemModel, error) {
	filter := bson.M{"workspace_id": workspaceID, "user_id": userID}
	var items []entities.CallListItemModel
	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallListItems -> FindAllByWorkspace: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)
	if err := cursor.All(repo.Context, &items); err != nil {
		fiberlog.Errorf("CallListItems -> FindAllByWorkspace: %s \n", err)
		return nil, err
	}
	return &items, nil
}

func (repo *callListItemsRepository) FindByID(id string) (*entities.CallListItemModel, error) {
	filter := bson.M{"id": id}
	var item entities.CallListItemModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&item); err != nil {
		fiberlog.Errorf("CallListItems -> FindByID: %s \n", err)
		return nil, err
	}
	return &item, nil
}

func (repo *callListItemsRepository) FindByIDByUser(id string, workspaceID string) (*entities.CallListItemModel, error) {
	filter := bson.M{"id": id, "workspace_id": workspaceID}
	var item entities.CallListItemModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&item); err != nil {
		fiberlog.Errorf("CallListItems -> FindByIDByUser: %s \n", err)
		return nil, err
	}
	return &item, nil
}

func (repo *callListItemsRepository) Update(id string, data entities.CallListItemModel) error {
	filter := bson.M{"id": id}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallListItems -> Update: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callListItemsRepository) Delete(id string) error {
	filter := bson.M{"id": id}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallListItems -> Delete: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callListItemsRepository) UpdateByUser(id string, workspaceID string, userID string, data entities.CallListItemModel) error {
	filter := bson.M{"id": id, "workspace_id": workspaceID, "user_id": userID}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallListItems -> UpdateByUser: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callListItemsRepository) DeleteByUser(id string, workspaceID string, userID string) error {
	filter := bson.M{"id": id, "workspace_id": workspaceID, "user_id": userID}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallListItems -> DeleteByUser: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callListItemsRepository) FindByFilter(filter entities.CallListItemFilter) (*[]entities.CallListItemModel, error) {
	queryFilter := bson.M{
		"workspace_id": filter.WorkspaceID,
	}
	if filter.UserID != "" {
		queryFilter["user_id"] = filter.UserID
	}
	if !filter.CalledAtGte.IsZero() {
		queryFilter["called_at"] = bson.M{"$gte": filter.CalledAtGte}
	}

	statusCond := bson.M{}
	if len(filter.StatusesIn) > 0 {
		statusCond["$in"] = filter.StatusesIn
	}
	if len(filter.StatusesNotIn) > 0 {
		statusCond["$nin"] = filter.StatusesNotIn
	}
	if len(statusCond) > 0 {
		queryFilter["status"] = statusCond
	}

	var items []entities.CallListItemModel
	cursor, err := repo.Collection.Find(repo.Context, queryFilter)
	if err != nil {
		fiberlog.Errorf("CallListItems -> FindByFilter: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)
	if err := cursor.All(repo.Context, &items); err != nil {
		fiberlog.Errorf("CallListItems -> FindByFilter: %s \n", err)
		return nil, err
	}
	return &items, nil
}
