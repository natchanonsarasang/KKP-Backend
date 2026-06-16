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
}

type ICallAttemptsService interface {
	GetAttemptsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error)
	GetAttemptByID(id primitive.ObjectID) (*entities.CallAttemptModel, error)
	CreateAttempt(data entities.CallAttemptModel) error
	UpdateAttempt(id primitive.ObjectID, data entities.CallAttemptModel) error
	DeleteAttempt(id primitive.ObjectID) error
}

func NewCallAttemptsService(repo repositories.ICallAttemptsRepository, itemRepo repositories.ICallListItemsRepository) ICallAttemptsService {
	return &callAttemptsService{
		CallAttemptsRepository:  repo,
		CallListItemsRepository: itemRepo,
	}
}

func (sv *callAttemptsService) GetAttemptsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *callAttemptsService) GetAttemptByID(id primitive.ObjectID) (*entities.CallAttemptModel, error) {
	return sv.CallAttemptsRepository.FindByID(id)
}

func (sv *callAttemptsService) CreateAttempt(data entities.CallAttemptModel) error {
	// Business Logic: ensure the CallListItemID exists
	if _, err := sv.CallListItemsRepository.FindByID(data.CallListItemID); err != nil {
		return errors.New("call list item not found")
	}

	data.CreatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallAttemptsRepository.Insert(data)
}

func (sv *callAttemptsService) UpdateAttempt(id primitive.ObjectID, data entities.CallAttemptModel) error {
	return sv.CallAttemptsRepository.Update(id, data)
}

func (sv *callAttemptsService) DeleteAttempt(id primitive.ObjectID) error {
	return sv.CallAttemptsRepository.Delete(id)
}
