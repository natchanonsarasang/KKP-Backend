package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type callAttemptsService struct {
	CallAttemptsRepository  repositories.ICallAttemptsRepository
	CallListItemsRepository repositories.ICallListItemsRepository
}

type ICallAttemptsService interface {
	GetAttemptsByWorkspace(workspaceID string) (*[]entities.CallAttemptModel, error)
	GetAttemptsByWorkspaceByUser(userID string, workspaceID string) (*[]entities.CallAttemptModel, error)
	GetAttemptByID(id string) (*entities.CallAttemptModel, error)
	GetAttemptByIDByUser(id string, userID string, workspaceID string) (*entities.CallAttemptModel, error)
	CreateAttempt(data entities.CallAttemptModel) error
	CreateAttemptByUser(userID string, data entities.CallAttemptModel) error
	// System Methods
	UpdateAttempt(id string, data entities.CallAttemptModel) error
	DeleteAttempt(id string) error
	// ByUser Methods
	UpdateAttemptByUser(id string, userID string, workspaceID string, data entities.CallAttemptModel) error
	DeleteAttemptByUser(id string, userID string, workspaceID string) error
}

func NewCallAttemptsService(repo repositories.ICallAttemptsRepository, itemRepo repositories.ICallListItemsRepository) ICallAttemptsService {
	return &callAttemptsService{
		CallAttemptsRepository:  repo,
		CallListItemsRepository: itemRepo,
	}
}

func (sv *callAttemptsService) GetAttemptsByWorkspace(workspaceID string) (*[]entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindAllByWorkspace(workspaceID, "")
}

func (sv *callAttemptsService) GetAttemptsByWorkspaceByUser(userID string, workspaceID string) (*[]entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindAllByWorkspace(workspaceID, userID)
}

func (sv *callAttemptsService) GetAttemptByID(id string) (*entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindByID(id)
}

func (sv *callAttemptsService) GetAttemptByIDByUser(id string, userID string, workspaceID string) (*entities.CallAttemptModel, error) {
	attempt, err := sv.CallAttemptsRepository.FindByIDByUser(id, workspaceID)
	if err != nil {
		return nil, err
	}
	if attempt.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this call attempt")
	}
	return attempt, nil
}

func (sv *callAttemptsService) CreateAttempt(data entities.CallAttemptModel) error {
	// Business Logic: ensure the CallListItemID exists
	if _, err := sv.CallListItemsRepository.FindByID(data.CallListItemID); err != nil {
		return errors.New("call list item not found")
	}

	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return sv.CallAttemptsRepository.Insert(data)
}

func (sv *callAttemptsService) CreateAttemptByUser(userID string, data entities.CallAttemptModel) error {
	data.UserID = userID
	return sv.CreateAttempt(data)
}

func (sv *callAttemptsService) UpdateAttempt(id string, data entities.CallAttemptModel) error {
	data.UpdatedAt = time.Now()
	return sv.CallAttemptsRepository.Update(id, data)
}

func (sv *callAttemptsService) DeleteAttempt(id string) error {
	return sv.CallAttemptsRepository.Delete(id)
}

func (sv *callAttemptsService) UpdateAttemptByUser(id string, userID string, workspaceID string, data entities.CallAttemptModel) error {
	data.ID = id
	data.UserID = userID
	data.UpdatedAt = time.Now()
	return sv.CallAttemptsRepository.UpdateByUser(id, workspaceID, userID, data)
}

func (sv *callAttemptsService) DeleteAttemptByUser(id string, userID string, workspaceID string) error {
	return sv.CallAttemptsRepository.DeleteByUser(id, workspaceID, userID)
}
