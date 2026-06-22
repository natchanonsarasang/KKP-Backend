package repositories

import (
	"context"
	"go-fiber-template/domain/entities" // 👈 แก้ตรงนี้จาก callecto-api เป็น go-fiber-template

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ICallTemplatesRepository interface {
	FindByFilter(ctx context.Context, id string, templateID string) ([]*entities.CallTemplateDataModel, error)
}

type callTemplatesRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

func NewCallTemplatesRepository(ctx context.Context, db *mongo.Database) ICallTemplatesRepository {
	return &callTemplatesRepository{
		Context:    ctx,
		Collection: db.Collection("call_templates"),
	}
}

func (r *callTemplatesRepository) FindByFilter(ctx context.Context, id string, templateID string) ([]*entities.CallTemplateDataModel, error) {
	filter := bson.M{}

	if id != "" {
		filter["_id"] = id
	}
	if templateID != "" {
		filter["template_id"] = templateID
	}

	cursor, err := r.Collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var templates []*entities.CallTemplateDataModel
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}
