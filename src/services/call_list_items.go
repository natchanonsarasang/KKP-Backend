package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type callListItemsService struct {
	CallListItemsRepository repositories.ICallListItemsRepository
	UsersService            IUsersService
}

type ICallListItemsService interface {
	GetCallListItemsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error)
	GetCallListItemsByWorkspaceByUser(userID string, workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error)
	GetCallListItemByID(id primitive.ObjectID) (*entities.CallListItemModel, error)
	GetCallListItemByIDByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) (*entities.CallListItemModel, error)
	CreateCallListItem(data entities.CallListItemModel) error
	CreateCallListItemByUser(userID string, data entities.CallListItemModel) error
	// System Methods
	UpdateCallListItem(id primitive.ObjectID, data entities.CallListItemModel) error
	DeleteCallListItem(id primitive.ObjectID) error
	// ByUser Methods
	UpdateCallListItemByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID, data entities.CallListItemModel) error
	DeleteCallListItemByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) error
}

func NewCallListItemsService(repo repositories.ICallListItemsRepository, usersService IUsersService) ICallListItemsService {
	return &callListItemsService{
		CallListItemsRepository: repo,
		UsersService:            usersService,
	}
}

func (sv *callListItemsService) GetCallListItemsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *callListItemsService) GetCallListItemsByWorkspaceByUser(userID string, workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error) {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized access to this workspace")
	}
	return sv.CallListItemsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *callListItemsService) GetCallListItemByID(id primitive.ObjectID) (*entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindByID(id)
}

func (sv *callListItemsService) GetCallListItemByIDByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) (*entities.CallListItemModel, error) {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized access to this workspace")
	}
	return sv.CallListItemsRepository.FindByIDByUser(id, workspaceID)
}

func (sv *callListItemsService) CreateCallListItem(data entities.CallListItemModel) error {
	data.CreatedAt = time.Now().Add(7 * time.Hour)
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallListItemsRepository.Insert(data)
}

func (sv *callListItemsService) CreateCallListItemByUser(userID string, data entities.CallListItemModel) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, data.WorkspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}
	return sv.CreateCallListItem(data)
}

func (sv *callListItemsService) UpdateCallListItem(id primitive.ObjectID, data entities.CallListItemModel) error {
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallListItemsRepository.Update(id, data)
}

func (sv *callListItemsService) DeleteCallListItem(id primitive.ObjectID) error {
	return sv.CallListItemsRepository.Delete(id)
}

func (sv *callListItemsService) UpdateCallListItemByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID, data entities.CallListItemModel) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}

	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallListItemsRepository.UpdateByUser(id, workspaceID, data)
}

func (sv *callListItemsService) DeleteCallListItemByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}

	return sv.CallListItemsRepository.DeleteByUser(id, workspaceID)
}
