package repositories

import (
	"context"
	. "go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"
	"os"
	"time"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type callSessionsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallSessionsRepository interface {
	InsertCallSession(data entities.CallSessionDataModel) error
	FindByID(id string) (*entities.CallSessionDataModel, error)
	FindOneByStatus(status string) (*entities.CallSessionDataModel, error)
	FindByStatus(status string) (*[]entities.CallSessionDataModel, error)
	FindByWorkspaceID(workspaceID string) (*[]entities.CallSessionDataModel, error)
	FindByUserID(userID string) (*[]entities.CallSessionDataModel, error)
	FindByFilter(filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error)
	UpdateCallSession(id string, data entities.CallSessionDataModel) error
	DeleteCallSession(id string) error
}

func NewCallSessionsRepository(db *MongoDB) ICallSessionsRepository {
	return &callSessionsRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("DATABASE_NAME")).Collection("call_sessions"),
	}
}

func (repo *callSessionsRepository) InsertCallSession(data entities.CallSessionDataModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("CallSessions -> InsertCallSession: %s \n", err)
		return err
	}
	return nil
}

func (repo *callSessionsRepository) FindByID(id string) (*entities.CallSessionDataModel, error) {
	filter := bson.M{"id": id}
	var session entities.CallSessionDataModel
	err := repo.Collection.FindOne(repo.Context, filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("CallSessions -> FindByID: %s \n", err)
		return nil, err
	}
	return &session, nil
}

func (repo *callSessionsRepository) FindOneByStatus(status string) (*entities.CallSessionDataModel, error) {
	filter := bson.M{"status": status}
	var session entities.CallSessionDataModel
	err := repo.Collection.FindOne(repo.Context, filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("CallSessions -> FindOneByStatus: %s \n", err)
		return nil, err
	}
	return &session, nil
}

func (repo *callSessionsRepository) FindByStatus(status string) (*[]entities.CallSessionDataModel, error) {
	filter := bson.M{"status": status}
	sessions := []entities.CallSessionDataModel{}

	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByStatus: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &sessions)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByStatus decoding: %s \n", err)
		return nil, err
	}

	return &sessions, nil
}

func (repo *callSessionsRepository) FindByWorkspaceID(workspaceID string) (*[]entities.CallSessionDataModel, error) {
	filter := bson.M{"workspace_id": workspaceID}
	sessions := []entities.CallSessionDataModel{}

	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByWorkspaceID: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &sessions)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByWorkspaceID decoding: %s \n", err)
		return nil, err
	}

	return &sessions, nil
}

func (repo *callSessionsRepository) FindByUserID(userID string) (*[]entities.CallSessionDataModel, error) {
	filter := bson.M{"user_id": userID}
	sessions := []entities.CallSessionDataModel{}

	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByUserID: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &sessions)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByUserID decoding: %s \n", err)
		return nil, err
	}

	return &sessions, nil
}

func (repo *callSessionsRepository) UpdateCallSession(id string, data entities.CallSessionDataModel) error {
	filter := bson.M{"id": id}
	if data.UpdatedAt.IsZero() {
		data.UpdatedAt = time.Now().UTC()
	}
	update := bson.M{"$set": data}
	_, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallSessions -> UpdateCallSession: %s \n", err)
		return err
	}
	return nil
}

func (repo *callSessionsRepository) DeleteCallSession(id string) error {
	filter := bson.M{"id": id}
	_, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallSessions -> DeleteCallSession: %s \n", err)
		return err
	}
	return nil
}

func (repo *callSessionsRepository) FindByFilter(filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error) {
	query := bson.M{}
	if filter.ID != "" {
		query["id"] = filter.ID
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.WorkspaceID != "" {
		query["workspace_id"] = filter.WorkspaceID
	}
	if filter.UserID != "" {
		query["user_id"] = filter.UserID
	}

	sessions := []entities.CallSessionDataModel{}
	cursor, err := repo.Collection.Find(repo.Context, query)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByFilter: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &sessions)
	if err != nil {
		fiberlog.Errorf("CallSessions -> FindByFilter decoding: %s \n", err)
		return nil, err
	}

	return &sessions, nil
}
