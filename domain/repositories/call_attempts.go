package repositories

import (
	"context"
	. "go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"
	"os"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type callAttemptsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallAttemptsRepository interface {
	Insert(data entities.CallAttemptModel) error
	FindAllByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error)
	FindByID(id primitive.ObjectID) (*entities.CallAttemptModel, error)
	FindByIDByUser(id primitive.ObjectID, workspaceID primitive.ObjectID) (*entities.CallAttemptModel, error)
	// System Methods
	Update(id primitive.ObjectID, data entities.CallAttemptModel) error
	Delete(id primitive.ObjectID) error
	// ByUser Methods
	UpdateByUser(id primitive.ObjectID, workspaceID primitive.ObjectID, data entities.CallAttemptModel) error
	DeleteByUser(id primitive.ObjectID, workspaceID primitive.ObjectID) error
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

func (repo *callAttemptsRepository) FindAllByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error) {
	filter := bson.M{"workspace_id": workspaceID}
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

func (repo *callAttemptsRepository) FindByID(id primitive.ObjectID) (*entities.CallAttemptModel, error) {
	filter := bson.M{"_id": id}
	var attempt entities.CallAttemptModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&attempt); err != nil {
		fiberlog.Errorf("CallAttempts -> FindByID: %s \n", err)
		return nil, err
	}
	return &attempt, nil
}

func (repo *callAttemptsRepository) FindByIDByUser(id primitive.ObjectID, workspaceID primitive.ObjectID) (*entities.CallAttemptModel, error) {
	filter := bson.M{"_id": id, "workspace_id": workspaceID}
	var attempt entities.CallAttemptModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&attempt); err != nil {
		fiberlog.Errorf("CallAttempts -> FindByIDByUser: %s \n", err)
		return nil, err
	}
	return &attempt, nil
}

func (repo *callAttemptsRepository) Update(id primitive.ObjectID, data entities.CallAttemptModel) error {
	filter := bson.M{"_id": id}
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

func (repo *callAttemptsRepository) Delete(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
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

func (repo *callAttemptsRepository) UpdateByUser(id primitive.ObjectID, workspaceID primitive.ObjectID, data entities.CallAttemptModel) error {
	filter := bson.M{"_id": id, "workspace_id": workspaceID}
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

func (repo *callAttemptsRepository) DeleteByUser(id primitive.ObjectID, workspaceID primitive.ObjectID) error {
	filter := bson.M{"_id": id, "workspace_id": workspaceID}
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
