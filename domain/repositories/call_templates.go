package repositories

import (
	"context"
	"go-fiber-template/domain/entities"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ICallTemplatesRepository interface {
	FindByFilter(ctx context.Context, id string, templateID string) ([]*entities.CallTemplateDataModel, error)
	InsertCallTemplate(ctx context.Context, data *entities.CallTemplateDataModel) error
	UpdateCallTemplate(ctx context.Context, id string, data *entities.CallTemplateDataModel) error
	DeleteCallTemplate(ctx context.Context, id string) error
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

func (r *callTemplatesRepository) InsertCallTemplate(ctx context.Context, data *entities.CallTemplateDataModel) error {
	_, err := r.Collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (r *callTemplatesRepository) UpdateCallTemplate(ctx context.Context, id string, data *entities.CallTemplateDataModel) error {
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

func (r *callTemplatesRepository) DeleteCallTemplate(ctx context.Context, id string) error {
	filter := bson.M{"id": id}
	_, err := r.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
