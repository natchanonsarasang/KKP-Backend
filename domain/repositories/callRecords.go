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
	FindAll() (*[]entities.CallRecordDataModel, error)
	UpdateCallRecord(id string, data entities.CallRecordDataModel) error
	DeleteCallRecord(id string) error
}

func NewCallRecordsRepository(db *MongoDB) ICallRecordsRepository {
	return &callRecordsRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("DATABASE_NAME")).Collection("call_records"),
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
	var records []entities.CallRecordDataModel

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
	var records []entities.CallRecordDataModel

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
