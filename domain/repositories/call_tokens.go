package repositories

import (
	"context"
	"go-fiber-template/domain/entities"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ICallTokensRepository interface {
	FindByFilter(ctx context.Context, id string, userID string) ([]*entities.CallTokenDataModel, error)
	InsertCallToken(ctx context.Context, data *entities.CallTokenDataModel) error
	UpdateCallToken(ctx context.Context, id string, data *entities.CallTokenDataModel) error
	DeleteCallToken(ctx context.Context, id string) error
}

type callTokensRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

func NewCallTokensRepository(ctx context.Context, db *mongo.Database) ICallTokensRepository {
	return &callTokensRepository{
		Context:    ctx,
		Collection: db.Collection("call_tokens"),
	}
}

func (r *callTokensRepository) FindByFilter(ctx context.Context, id string, userID string) ([]*entities.CallTokenDataModel, error) {
	filter := bson.M{}

	if id != "" {
		filter["_id"] = id
	}
	if userID != "" {
		filter["user_id"] = userID
	}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*entities.CallTokenDataModel
	if err := cursor.All(ctx, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *callTokensRepository) InsertCallToken(ctx context.Context, data *entities.CallTokenDataModel) error {
	_, err := r.Collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (r *callTokensRepository) UpdateCallToken(ctx context.Context, id string, data *entities.CallTokenDataModel) error {
	filter := bson.M{"id": id}
	if data.UpdatedAt.IsZero() {
		data.UpdatedAt = time.Now().UTC()
	}
	update := bson.M{"$set": data}
	_, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *callTokensRepository) DeleteCallToken(ctx context.Context, id string) error {
	filter := bson.M{"id": id}
	_, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
