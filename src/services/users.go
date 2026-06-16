package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"go-fiber-template/httpclient"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type usersService struct {
	UsersRepository repositories.IUsersRepository
}

type IUsersService interface {
	GetAllUsers() (*[]entities.UserDataModel, error)
	InsertNewUser(data entities.UserDataModel) error
	VerifyUserInWorkspace(userID string, workspaceID primitive.ObjectID) (bool, error)
}

func NewUsersService(repo0 repositories.IUsersRepository) IUsersService {
	return &usersService{
		UsersRepository: repo0,
	}
}

func (sv *usersService) GetAllUsers() (*[]entities.UserDataModel, error) {
	data, err := sv.UsersRepository.FindAll()
	if err != nil {
		return nil, err
	}

	return data, nil

}

func (sv *usersService) InsertNewUser(data entities.UserDataModel) error {
	data.CreatedAt = time.Now().Add(7 * time.Hour)
	dataIp, err := httpclient.GetIP()
	if err != nil {
		return err
	}
	data.Ip = dataIp

	return sv.UsersRepository.InsertUser(data)
}

func (sv *usersService) VerifyUserInWorkspace(userID string, workspaceID primitive.ObjectID) (bool, error) {
	return sv.UsersRepository.VerifyUserInWorkspace(userID, workspaceID)
}
