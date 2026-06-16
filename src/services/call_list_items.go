package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type callListItemsService struct {
	CallListItemsRepository repositories.ICallListItemsRepository
}

type ICallListItemsService interface {
	GetItemsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error)
	GetItemByID(id primitive.ObjectID) (*entities.CallListItemModel, error)
	CreateItem(data entities.CallListItemModel) error
	UpdateItem(id primitive.ObjectID, data entities.CallListItemModel) error
	DeleteItem(id primitive.ObjectID) error
}

func NewCallListItemsService(repo repositories.ICallListItemsRepository) ICallListItemsService {
	return &callListItemsService{
		CallListItemsRepository: repo,
	}
}

func (sv *callListItemsService) GetItemsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *callListItemsService) GetItemByID(id primitive.ObjectID) (*entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindByID(id)
}

func (sv *callListItemsService) CreateItem(data entities.CallListItemModel) error {
	data.CreatedAt = time.Now().Add(7 * time.Hour)
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallListItemsRepository.Insert(data)
}

func (sv *callListItemsService) UpdateItem(id primitive.ObjectID, data entities.CallListItemModel) error {
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.CallListItemsRepository.Update(id, data)
}

func (sv *callListItemsService) DeleteItem(id primitive.ObjectID) error {
	return sv.CallListItemsRepository.Delete(id)
}
