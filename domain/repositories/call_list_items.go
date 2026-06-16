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

type callListItemsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type ICallListItemsRepository interface {
	Insert(data entities.CallListItemModel) error
	FindAllByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error)
	FindByID(id primitive.ObjectID) (*entities.CallListItemModel, error)
	Update(id primitive.ObjectID, data entities.CallListItemModel) error
	Delete(id primitive.ObjectID) error
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

func (repo *callListItemsRepository) FindAllByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error) {
	filter := bson.M{"workspace_id": workspaceID}
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

func (repo *callListItemsRepository) FindByID(id primitive.ObjectID) (*entities.CallListItemModel, error) {
	filter := bson.M{"_id": id}
	var item entities.CallListItemModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&item); err != nil {
		fiberlog.Errorf("CallListItems -> FindByID: %s \n", err)
		return nil, err
	}
	return &item, nil
}

func (repo *callListItemsRepository) Update(id primitive.ObjectID, data entities.CallListItemModel) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": data}
	if _, err := repo.Collection.UpdateOne(repo.Context, filter, update); err != nil {
		fiberlog.Errorf("CallListItems -> Update: %s \n", err)
		return err
	}
	return nil
}

func (repo *callListItemsRepository) Delete(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	if _, err := repo.Collection.DeleteOne(repo.Context, filter); err != nil {
		fiberlog.Errorf("CallListItems -> Delete: %s \n", err)
		return err
	}
	return nil
}
