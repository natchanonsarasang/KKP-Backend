package repositories

import (
	"context"
	"go-fiber-template/domain/entities" // 👈 แก้ตรงนี้เป็น go-fiber-template เหมือนกัน

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ICallTokensRepository interface {
	FindByFilter(ctx context.Context, id string, userID string) ([]*entities.CallTokenDataModel, error)
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
