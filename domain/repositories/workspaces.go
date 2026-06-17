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

type workspacesRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type IWorkspacesRepository interface {
	InsertWorkspace(data entities.WorkspaceDataModel) error
	FindAll() (*[]entities.WorkspaceDataModel, error)
	FindByID(id string) (*entities.WorkspaceDataModel, error)
	FindByFilter(filter entities.WorkspaceFilter) (*[]entities.WorkspaceDataModel, error)
	UpdateWorkspace(id string, data entities.WorkspaceDataModel) error
	DeleteWorkspace(id string) error
	UpdateWorkspaceByUser(id string, userID string, data entities.WorkspaceDataModel) error
	DeleteWorkspaceByUser(id string, userID string) error
}

func NewWorkspacesRepository(db *MongoDB) IWorkspacesRepository {
	return &workspacesRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("workspaces"),
	}
}

func (repo *workspacesRepository) InsertWorkspace(data entities.WorkspaceDataModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("Workspaces -> InsertWorkspace: %s \n", err)
		return err
	}
	return nil
}

func (repo *workspacesRepository) FindAll() (*[]entities.WorkspaceDataModel, error) {
	options := options.Find()
	filter := bson.M{}
	var workspaces []entities.WorkspaceDataModel

	cursor, err := repo.Collection.Find(repo.Context, filter, options)
	if err != nil {
		fiberlog.Errorf("Workspaces -> FindAll: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &workspaces)
	if err != nil {
		fiberlog.Errorf("Workspaces -> FindAll: %s \n", err)
		return nil, err
	}

	return &workspaces, nil
}

func (repo *workspacesRepository) FindByID(id string) (*entities.WorkspaceDataModel, error) {
	filter := bson.M{"id": id}
	var workspace entities.WorkspaceDataModel

	err := repo.Collection.FindOne(repo.Context, filter).Decode(&workspace)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("Workspaces -> FindByID: %s \n", err)
		return nil, err
	}

	return &workspace, nil
}

func (repo *workspacesRepository) FindByFilter(filterData entities.WorkspaceFilter) (*[]entities.WorkspaceDataModel, error) {
	options := options.Find()
	filter := bson.M{}

	if filterData.ID != "" {
		filter["id"] = filterData.ID
	}
	if filterData.Name != "" {
		filter["name"] = filterData.Name
	}
	if filterData.OwnerID != "" {
		filter["owner_id"] = filterData.OwnerID
	}

	var workspaces []entities.WorkspaceDataModel

	cursor, err := repo.Collection.Find(repo.Context, filter, options)
	if err != nil {
		fiberlog.Errorf("Workspaces -> FindByFilter: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &workspaces)
	if err != nil {
		fiberlog.Errorf("Workspaces -> FindByFilter: %s \n", err)
		return nil, err
	}

	return &workspaces, nil
}

func (repo *workspacesRepository) UpdateWorkspace(id string, data entities.WorkspaceDataModel) error {
	filter := bson.M{"id": id}
	update := bson.M{}

	if data.Name != "" {
		update["name"] = data.Name
	}
	if data.OwnerID != "" {
		update["owner_id"] = data.OwnerID
	}
	update["updated_at"] = data.UpdatedAt

	if len(update) == 0 {
		return nil
	}

	result, err := repo.Collection.UpdateOne(repo.Context, filter, bson.M{"$set": update})
	if err != nil {
		fiberlog.Errorf("Workspaces -> UpdateWorkspace: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (repo *workspacesRepository) DeleteWorkspace(id string) error {
	filter := bson.M{"id": id}

	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Workspaces -> DeleteWorkspace: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (repo *workspacesRepository) UpdateWorkspaceByUser(id string, userID string, data entities.WorkspaceDataModel) error {
	filter := bson.M{"id": id, "owner_id": userID}
	update := bson.M{}

	if data.Name != "" {
		update["name"] = data.Name
	}
	// OwnerID should ideally not be updatable, but if it is, we ensure we only update if we own it
	if data.OwnerID != "" {
		update["owner_id"] = data.OwnerID
	}
	update["updated_at"] = data.UpdatedAt

	if len(update) == 0 {
		return nil
	}

	result, err := repo.Collection.UpdateOne(repo.Context, filter, bson.M{"$set": update})
	if err != nil {
		fiberlog.Errorf("Workspaces -> UpdateWorkspaceByUser: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (repo *workspacesRepository) DeleteWorkspaceByUser(id string, userID string) error {
	filter := bson.M{"id": id, "owner_id": userID}

	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Workspaces -> DeleteWorkspaceByUser: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}