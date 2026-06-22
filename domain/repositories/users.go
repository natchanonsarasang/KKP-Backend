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

type usersRepository struct {
	Context    context.Context
	Collection *mongo.Collection
}

type IUsersRepository interface {
	InsertUser(data entities.UserDataModel) error
	FindAll() (*[]entities.UserDataModel, error)
	FindByID(id string) (*entities.UserDataModel, error)
	FindByEmail(email string) (*entities.UserDataModel, error)
	FindByGoogleID(googleID string) (*entities.UserDataModel, error)
	FindByFilter(filter entities.UserFilter) (*[]entities.UserDataModel, error)
	UpdateUser(id string, data entities.UserDataModel) error
	DeleteUser(id string) error
}

func NewUsersRepository(db *MongoDB) IUsersRepository {
	return &usersRepository{
		Context:    db.Context,
		Collection: db.MongoDB.Database(os.Getenv("MONGODB_NAME")).Collection("users"),
	}
}

func (repo *usersRepository) InsertUser(data entities.UserDataModel) error {
	if _, err := repo.Collection.InsertOne(repo.Context, data); err != nil {
		fiberlog.Errorf("Users -> InsertUser: %s \n", err)
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

func (repo *usersRepository) FindByID(id string) (*entities.UserDataModel, error) {
	filter := bson.M{"id": id}
	var user entities.UserDataModel

	err := repo.Collection.FindOne(repo.Context, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("Users -> FindByID: %s \n", err)
		return nil, err
	}

	return &user, nil
}

func (repo *usersRepository) FindByEmail(email string) (*entities.UserDataModel, error) {
	filter := bson.M{"email": email}
	var user entities.UserDataModel

	err := repo.Collection.FindOne(repo.Context, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("Users -> FindByEmail: %s \n", err)
		return nil, err
	}

	return &user, nil
}

func (repo *usersRepository) FindByGoogleID(googleID string) (*entities.UserDataModel, error) {
	filter := bson.M{"google_id": googleID}
	var user entities.UserDataModel

	err := repo.Collection.FindOne(repo.Context, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		fiberlog.Errorf("Users -> FindByGoogleID: %s \n", err)
		return nil, err
	}

	return &user, nil
}

func (repo *usersRepository) FindByFilter(filterData entities.UserFilter) (*[]entities.UserDataModel, error) {
	options := options.Find()
	filter := bson.M{}

	if filterData.ID != "" {
		filter["id"] = filterData.ID
	}
	if filterData.Email != "" {
		filter["email"] = filterData.Email
	}
	if filterData.Name != "" {
		filter["name"] = filterData.Name
	}
	if filterData.GoogleID != "" {
		filter["google_id"] = filterData.GoogleID
	}
	if filterData.Provider != "" {
		filter["provider"] = filterData.Provider
	}

	var users []entities.UserDataModel

	cursor, err := repo.Collection.Find(repo.Context, filter, options)
	if err != nil {
		fiberlog.Errorf("Users -> FindByFilter: %s \n", err)
		return nil, err
	}
	defer cursor.Close(repo.Context)

	err = cursor.All(repo.Context, &users)
	if err != nil {
		fiberlog.Errorf("Users -> FindByFilter: %s \n", err)
		return nil, err
	}

	return &users, nil
}

func (repo *usersRepository) UpdateUser(id string, data entities.UserDataModel) error {
	filter := bson.M{"id": id}
	update := bson.M{}

	if data.Email != "" {
		update["email"] = data.Email
	}
	if data.Name != "" {
		update["name"] = data.Name
	}
	if data.Picture != "" {
		update["picture"] = data.Picture
	}
	if data.GoogleID != "" {
		update["google_id"] = data.GoogleID
	}
	if data.Provider != "" {
		update["provider"] = data.Provider
	}
	if data.EmailVerified {
		update["email_verified"] = data.EmailVerified
	}
	if !data.LastLoginAt.IsZero() {
		update["last_login_at"] = data.LastLoginAt
	}
	update["updated_at"] = data.UpdatedAt

	if len(update) == 0 {
		return nil
	}

	result, err := repo.Collection.UpdateOne(repo.Context, filter, bson.M{"$set": update})
	if err != nil {
		fiberlog.Errorf("Users -> UpdateUser: %s \n", err)
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (repo *usersRepository) DeleteUser(id string) error {
	filter := bson.M{"id": id}

	result, err := repo.Collection.DeleteOne(repo.Context, filter)
	if err != nil {
		fiberlog.Errorf("Users -> DeleteUser: %s \n", err)
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
