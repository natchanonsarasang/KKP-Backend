package repositories

import (
	"context"
	. "go-fiber-template/domain/datasources"
	"go-fiber-template/domain/entities"
	"os"

	fiberlog "github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type debtorsRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type IDebtorsRepository interface {
	Insert(data entities.DebtorModel) error
	FindAllByWorkspace(workspaceID string, userID string) (*[]entities.DebtorModel, error)
	FindByID(id string) (*entities.DebtorModel, error)
	FindByIDByUser(id string, workspaceID string) (*entities.DebtorModel, error)
	FindByPhoneNumber(phoneNumber string) (*entities.DebtorModel, error)
	// System Methods
	Update(id string, data entities.DebtorModel) error
	Delete(id string) error
	// ByUser Methods
	UpdateByUser(id string, workspaceID string, userID string, data entities.DebtorModel) error
	DeleteByUser(id string, workspaceID string, userID string) error
}

func NewDebtorsRepository(db *MongoDB) IDebtorsRepository {
	return &debtorsRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("debtors"),
	}
}

func (repo *debtorsRepository) Insert(data entities.DebtorModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("Debtors -> Insert: %s \n", err)
		return err
	}
	return nil
}

func (repo *debtorsRepository) FindAllByWorkspace(workspaceID string, userID string) (*[]entities.DebtorModel, error) {
	filter := bson.M{"workspace_id": workspaceID, "user_id": userID}
	var debtors []entities.DebtorModel
	cursor, err := repo.Collection.Find(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Debtors -> FindAllByWorkspace: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)
	if err := cursor.All(repo.Context, &debtors); err != nil {
		fiberlog.Errorf("Debtors -> FindAllByWorkspace: %s \n", err)
		return nil, err
	}
	return &debtors, nil
}

func (repo *debtorsRepository) FindByID(id string) (*entities.DebtorModel, error) {
	filter := bson.M{"id": id}
	var debtor entities.DebtorModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&debtor); err != nil {
		fiberlog.Errorf("Debtors -> FindByID: %s \n", err)
		return nil, err
	}
	return &debtor, nil
}

func (repo *debtorsRepository) FindByIDByUser(id string, workspaceID string) (*entities.DebtorModel, error) {
	filter := bson.M{"id": id, "workspace_id": workspaceID}
	var debtor entities.DebtorModel
	if err := repo.Collection.FindOne(repo.Context, filter).Decode(&debtor); err != nil {
		fiberlog.Errorf("Debtors -> FindByIDByUser: %s \n", err)
		return nil, err
	}
	return &debtor, nil
}

func (repo *debtorsRepository) FindByPhoneNumber(phoneNumber string) (*entities.DebtorModel, error) {
	filter := bson.M{"phone_number": phoneNumber}
	var debtor entities.DebtorModel
	err := repo.Collection.FindOne(repo.Context, filter).Decode(&debtor)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("Debtors -> FindByPhoneNumber: %s \n", err)
		return nil, err
	}
	return &debtor, nil
}

func (repo *debtorsRepository) Update(id string, data entities.DebtorModel) error {
	filter := bson.M{"id": id}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("Debtors -> Update: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *debtorsRepository) Delete(id string) error {
	filter := bson.M{"id": id}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Debtors -> Delete: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *debtorsRepository) UpdateByUser(id string, workspaceID string, userID string, data entities.DebtorModel) error {
	filter := bson.M{"id": id, "workspace_id": workspaceID, "user_id": userID}
	update := bson.M{"$set": data}
	result, err := repo.Collection.UpdateOne(repo.Context, filter, update)
	if err != nil {
		fiberlog.Errorf("Debtors -> UpdateByUser: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (repo *debtorsRepository) DeleteByUser(id string, workspaceID string, userID string) error {
	filter := bson.M{"id": id, "workspace_id": workspaceID, "user_id": userID}
	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Debtors -> DeleteByUser: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
