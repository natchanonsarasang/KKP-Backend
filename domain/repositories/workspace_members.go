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

type workspaceMembersRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type IWorkspaceMembersRepository interface {
	InsertWorkspaceMember(data entities.WorkspaceMemberDataModel) error
	FindAll() (*[]entities.WorkspaceMemberDataModel, error)
	FindByID(id string) (*entities.WorkspaceMemberDataModel, error)
	UpdateByID(id string, data entities.WorkspaceMemberDataModel) error
	DeleteByID(id string) error
}

func NewWorkspaceMembersRepository(db *MongoDB) IWorkspaceMembersRepository {
	return &workspaceMembersRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("DATABASE_NAME")).Collection("workspace_members"),
	}
}

func (repo *workspaceMembersRepository) InsertWorkspaceMember(data entities.WorkspaceMemberDataModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("WorkspaceMembers -> InsertWorkspaceMember: %s \n", err)
		return err
	}
	return nil
}

func (repo *workspaceMembersRepository) FindAll() (*[]entities.WorkspaceMemberDataModel, error) {
	options := options.Find()
	filter := bson.M{}
	var members []entities.WorkspaceMemberDataModel

	cursor, err := repo.Collection.Find(repo.Context, filter, options)
	if err != nil {
		fiberlog.Errorf("WorkspaceMembers -> FindAll: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &members)
	if err != nil {
		fiberlog.Errorf("WorkspaceMembers -> FindAll: %s \n", err)
		return nil, err
	}

	return &members, nil
}

func (repo *workspaceMembersRepository) FindByID(id string) (*entities.WorkspaceMemberDataModel, error) {
	filter := bson.M{"id": id}
	var member entities.WorkspaceMemberDataModel

	err := repo.Collection.FindOne(repo.Context, filter).Decode(&member)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("WorkspaceMembers -> FindByID: %s \n", err)
		return nil, err
	}

	return &member, nil
}

func (repo *workspaceMembersRepository) UpdateByID(id string, data entities.WorkspaceMemberDataModel) error {
	filter := bson.M{"id": id}
	update := bson.M{}

	if data.WorkspaceID != "" {
		update["workspace_id"] = data.WorkspaceID
	}
	if data.UserID != "" {
		update["user_id"] = data.UserID
	}
	if data.Role != "" {
		update["role"] = data.Role
	}

	if len(update) == 0 {
		return nil
	}

	result, err := repo.Collection.UpdateOne(repo.Context, filter, bson.M{"$set": update})
	if err != nil {
		fiberlog.Errorf("WorkspaceMembers -> UpdateByID: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (repo *workspaceMembersRepository) DeleteByID(id string) error {
	filter := bson.M{"id": id}

	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("WorkspaceMembers -> DeleteByID: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}