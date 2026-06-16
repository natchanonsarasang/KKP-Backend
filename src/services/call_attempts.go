package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type callAttemptsService struct {
	CallAttemptsRepository  repositories.ICallAttemptsRepository
	CallListItemsRepository repositories.ICallListItemsRepository
	UsersService            IUsersService
}

type ICallAttemptsService interface {
	GetAttemptsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error)
	GetAttemptsByWorkspaceByUser(userID string, workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error)
	GetAttemptByID(id primitive.ObjectID) (*entities.CallAttemptModel, error)
	GetAttemptByIDByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) (*entities.CallAttemptModel, error)
	CreateAttempt(data entities.CallAttemptModel) error
	CreateAttemptByUser(userID string, data entities.CallAttemptModel) error
	// System Methods
	UpdateAttempt(id primitive.ObjectID, data entities.CallAttemptModel) error
	DeleteAttempt(id primitive.ObjectID) error
	// ByUser Methods
	UpdateAttemptByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID, data entities.CallAttemptModel) error
	DeleteAttemptByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) error
}

func NewCallAttemptsService(repo repositories.ICallAttemptsRepository, itemRepo repositories.ICallListItemsRepository, usersService IUsersService) ICallAttemptsService {
	return &callAttemptsService{
		CallAttemptsRepository:  repo,
		CallListItemsRepository: itemRepo,
		UsersService:            usersService,
	}
}

func (sv *callAttemptsService) GetAttemptsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *callAttemptsService) GetAttemptsByWorkspaceByUser(userID string, workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error) {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized access to this workspace")
	}
	return sv.CallAttemptsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *callAttemptsService) GetAttemptByID(id primitive.ObjectID) (*entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindByID(id)
}

func (sv *callAttemptsService) GetAttemptByIDByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) (*entities.CallAttemptModel, error) {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized access to this workspace")
	}
	return sv.CallAttemptsRepository.FindByIDByUser(id, workspaceID)
}

func (sv *callAttemptsService) CreateAttempt(data entities.CallAttemptModel) error {
	// Business Logic: ensure the CallListItemID exists
	if _, err := sv.CallListItemsRepository.FindByID(data.CallListItemID); err != nil {
		return errors.New("call list item not found")
	}

	data.CreatedAt = time.Now().Add(7 * time.Hour)
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallAttemptsRepository.Insert(data)
}

func (sv *callAttemptsService) CreateAttemptByUser(userID string, data entities.CallAttemptModel) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, data.WorkspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}
	return sv.CreateAttempt(data)
}

func (sv *callAttemptsService) UpdateAttempt(id primitive.ObjectID, data entities.CallAttemptModel) error {
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallAttemptsRepository.Update(id, data)
}

func (sv *callAttemptsService) DeleteAttempt(id primitive.ObjectID) error {
	return sv.CallAttemptsRepository.Delete(id)
}

func (sv *callAttemptsService) UpdateAttemptByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID, data entities.CallAttemptModel) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}

	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallAttemptsRepository.UpdateByUser(id, workspaceID, data)
}

func (sv *callAttemptsService) DeleteAttemptByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}

	return sv.CallAttemptsRepository.DeleteByUser(id, workspaceID)
}
