package repositories

import (
	"context"
	. "go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"
	"os"

	fiberlog "github.com/gofiber/fiber/v2/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type usersRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type IUsersRepository interface {
	InsertUser(data entities.UserDataModel) error
	FindAll() (*[]entities.UserDataModel, error)
	VerifyUserInWorkspace(userID string, workspaceID primitive.ObjectID) (bool, error)
}

func NewUsersRepository(db *MongoDB) IUsersRepository {
	return &usersRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("DATABASE_NAME")).Collection("users"),
	}
}

func (repo *usersRepository) InsertUser(data entities.UserDataModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("Users -> InsertNewUser: %s \n", err)
		return err
	}
	return nil
}

func (repo *usersRepository) FindAll() (*[]entities.UserDataModel, error) {
	options := options.Find()
	filter := bson.M{}
	var users []entities.UserDataModel

	cursor, err := repo.Collection.Find(repo.Context, filter, options)
	if err != nil {
		fiberlog.Errorf("Users -> FindAll: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &users)
	if err != nil {
		fiberlog.Errorf("Users -> FindAll: %s \n", err)
		return nil, err
	}

	return &users, nil
}

func (repo *usersRepository) VerifyUserInWorkspace(userID string, workspaceID primitive.ObjectID) (bool, error) {
	// Assuming users have a workspace_id field or a many-to-many relationship
	// For now, let's check if the user exists and has this workspace_id
	filter := bson.M{"user_id": userID, "workspace_id": workspaceID}
	count, err := repo.Collection.CountDocuments(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Users -> VerifyUserInWorkspace: %s \n", err)
		return false, err
	}
	return count > 0, nil
}
