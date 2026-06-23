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

type callRecordsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallRecordsRepository interface {
	InsertCallRecord(data entities.CallRecordDataModel) error
	FindByID(id string) (*entities.CallRecordDataModel, error)
	FindByUserID(userID string) (*[]entities.CallRecordDataModel, error)
	FindByFilter(filter entities.CallRecordFilter) (*[]entities.CallRecordDataModel, error)
	FindAll() (*[]entities.CallRecordDataModel, error)
	UpdateCallRecord(id string, data entities.CallRecordDataModel) error
	DeleteCallRecord(id string) error
	UpdateCallRecordByUser(id string, userID string, data entities.CallRecordDataModel) error
	DeleteCallRecordByUser(id string, userID string) error
}

func NewCallRecordsRepository(db *MongoDB) ICallRecordsRepository {
	return &callRecordsRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("call_records"),
	}
}

func (repo *callRecordsRepository) InsertCallRecord(data entities.CallRecordDataModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("CallRecords -> InsertCallRecord: %s \n", err)
		return err
	}
	return nil
}

func (repo *callRecordsRepository) FindByID(id string) (*entities.CallRecordDataModel, error) {
	filter := bson.M{"id": id}
	var record entities.CallRecordDataModel
	err := repo.Collection.FindOne(repo.Context, filter).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("CallRecords -> FindByID: %s \n", err)
		return nil, err
	}
	return &record, nil
}

func (repo *callRecordsRepository) FindAll() (*[]entities.CallRecordDataModel, error) {
	filter := bson.M{}
	records := []entities.CallRecordDataModel{}

	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallRecords -> FindAll: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &records)
	if err != nil {
		fiberlog.Errorf("CallRecords -> FindAll decoding: %s \n", err)
		return nil, err
	}

	return &records, nil
}

func (repo *callRecordsRepository) UpdateCallRecord(id string, data entities.CallRecordDataModel) error {
	filter := bson.M{"id": id}
	if data.UpdatedAt.IsZero() {
		data.UpdatedAt = time.Now().UTC()
	}
	update := bson.M{"$set": data}
	_, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallRecords -> UpdateCallRecord: %s \n", err)
		return err
	}
	return nil
}

func (repo *callRecordsRepository) DeleteCallRecord(id string) error {
	filter := bson.M{"id": id}
	_, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallRecords -> DeleteCallRecord: %s \n", err)
		return err
	}
	return nil
}

func (repo *callRecordsRepository) FindByUserID(userID string) (*[]entities.CallRecordDataModel, error) {
	filter := bson.M{"user_id": userID}
	records := []entities.CallRecordDataModel{}

	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallRecords -> FindByUserID: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &records)
	if err != nil {
		fiberlog.Errorf("CallRecords -> FindByUserID decoding: %s \n", err)
		return nil, err
	}

	return &records, nil
}

func (repo *callRecordsRepository) FindByFilter(filter entities.CallRecordFilter) (*[]entities.CallRecordDataModel, error) {
	query := bson.M{}
	if filter.UserID != "" {
		query["user_id"] = filter.UserID
	}
	if filter.WorkspaceID != "" {
		query["workspace_id"] = filter.WorkspaceID
	}
	if filter.Status != "" {
		query["status"] = filter.Status
	}
	if filter.BotnoiCallID != "" {
		query["botnoi_call_id"] = filter.BotnoiCallID
	}

	records := []entities.CallRecordDataModel{}
	cursor, err := repo.Collection.Find(repo.Context, query)
	if err != nil {
		fiberlog.Errorf("CallRecords -> FindByFilter: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &records)
	if err != nil {
		fiberlog.Errorf("CallRecords -> FindByFilter decoding: %s \n", err)
		return nil, err
	}

	return &records, nil
}

func (repo *callRecordsRepository) UpdateCallRecordByUser(id string, userID string, data entities.CallRecordDataModel) error {
	filter := bson.M{"id": id, "user_id": userID}
	if data.UpdatedAt.IsZero() {
		data.UpdatedAt = time.Now().UTC()
	}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("CallRecords -> UpdateCallRecordByUser: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *callRecordsRepository) DeleteCallRecordByUser(id string, userID string) error {
	filter := bson.M{"id": id, "user_id": userID}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("CallRecords -> DeleteCallRecordByUser: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
