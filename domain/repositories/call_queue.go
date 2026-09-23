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

type callQueueRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallQueueRepository interface {
	Insert(data entities.CallQueueModel) error
	// FindHead returns the most-recently-enqueued row (created_at desc). Newest
	// rather than oldest so that if a dead call ever leaves a row behind, the
	// freshly-placed call's row is still the one served.
	FindHead() (*entities.CallQueueModel, error)
	DeleteByOutboundID(outboundID string) error
	// DeleteByCallListItemID removes a queue row by its call_list_item_id. Used by
	// the stale-item reset, which knows the item id but not the Botnoi-assigned
	// batch_id (outbound_id).
	DeleteByCallListItemID(callListItemID string) error
	Count() (int64, error)
}

func NewCallQueueRepository(db *MongoDB) ICallQueueRepository {
	return &callQueueRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("call_queue"),
	}
}

func (repo *callQueueRepository) Insert(data entities.CallQueueModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("CallQueue -> Insert: %s \n", err)
		return err
	}
	return nil
}

func (repo *callQueueRepository) FindHead() (*entities.CallQueueModel, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var row entities.CallQueueModel
	err := repo.Collection.FindOne(repo.Context, bson.M{}, opts).Decode(&row)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("CallQueue -> FindHead: %s \n", err)
		return nil, err
	}
	return &row, nil
}

func (repo *callQueueRepository) DeleteByOutboundID(outboundID string) error {
	_, err := repo.Collection.DeleteOne(repo.Context, bson.M{"outbound_id": outboundID})
	if err != nil {
		fiberlog.Errorf("CallQueue -> DeleteByOutboundID: %s \n", err)
		return err
	}
	return nil
}

func (repo *callQueueRepository) DeleteByCallListItemID(callListItemID string) error {
	_, err := repo.Collection.DeleteOne(repo.Context, bson.M{"call_list_item_id": callListItemID})
	if err != nil {
		fiberlog.Errorf("CallQueue -> DeleteByCallListItemID: %s \n", err)
		return err
	}
	return nil
}

func (repo *callQueueRepository) Count() (int64, error) {
	count, err := repo.Collection.CountDocuments(repo.Context, bson.M{})
	if err != nil {
		fiberlog.Errorf("CallQueue -> Count: %s \n", err)
		return 0, err
	}
	return count, nil
}
