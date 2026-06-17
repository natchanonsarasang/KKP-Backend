package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type callListItemsService struct {
	CallListItemsRepository repositories.ICallListItemsRepository
}

type ICallListItemsService interface {
	GetCallListItemsByWorkspace(workspaceID string) (*[]entities.CallListItemModel, error)
	GetCallListItemsByWorkspaceByUser(userID string, workspaceID string) (*[]entities.CallListItemModel, error)
	GetCallListItemByID(id string) (*entities.CallListItemModel, error)
	GetCallListItemByIDByUser(id string, userID string, workspaceID string) (*entities.CallListItemModel, error)
	GetCallListItemsByFilterByUser(userID string, filter entities.CallListItemFilter) (*[]entities.CallListItemModel, error)
	CreateCallListItem(data entities.CallListItemModel) error
	CreateCallListItemByUser(userID string, data entities.CallListItemModel) error
	// System Methods
	UpdateCallListItem(id string, data entities.CallListItemModel) error
	DeleteCallListItem(id string) error
	// ByUser Methods
	UpdateCallListItemByUser(id string, userID string, workspaceID string, data entities.CallListItemModel) error
	DeleteCallListItemByUser(id string, userID string, workspaceID string) error
}

func NewCallListItemsService(repo repositories.ICallListItemsRepository) ICallListItemsService {
	return &callListItemsService{
		CallListItemsRepository: repo,
	}
}

func (sv *callListItemsService) GetCallListItemsByWorkspace(workspaceID string) (*[]entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindAllByWorkspace(workspaceID, "")
}

func (sv *callListItemsService) GetCallListItemsByWorkspaceByUser(userID string, workspaceID string) (*[]entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindAllByWorkspace(workspaceID, userID)
}

func (sv *callListItemsService) GetCallListItemByID(id string) (*entities.CallListItemModel, error) {
	return sv.CallListItemsRepository.FindByID(id)
}

func (sv *callListItemsService) GetCallListItemByIDByUser(id string, userID string, workspaceID string) (*entities.CallListItemModel, error) {
	item, err := sv.CallListItemsRepository.FindByIDByUser(id, workspaceID)
	if err != nil {
		return nil, err
	}
	if item.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this call list item")
	}
	return item, nil
}

func (sv *callListItemsService) CreateCallListItem(data entities.CallListItemModel) error {
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return sv.CallListItemsRepository.Insert(data)
}

func (sv *callListItemsService) CreateCallListItemByUser(userID string, data entities.CallListItemModel) error {
	data.UserID = userID
	return sv.CreateCallListItem(data)
}

func (sv *callListItemsService) UpdateCallListItem(id string, data entities.CallListItemModel) error {
	data.UpdatedAt = time.Now()
	return sv.CallListItemsRepository.Update(id, data)
}

func (sv *callListItemsService) DeleteCallListItem(id string) error {
	return sv.CallListItemsRepository.Delete(id)
}

func (sv *callListItemsService) UpdateCallListItemByUser(id string, userID string, workspaceID string, data entities.CallListItemModel) error {
	// Ensure immutable fields are not modified
	data.ID = ""
	data.UserID = ""
	data.WorkspaceID = ""
	data.DebtorID = ""
	data.CreatedAt = time.Time{}

	data.UpdatedAt = time.Now()
	return sv.CallListItemsRepository.UpdateByUser(id, workspaceID, userID, data)
}

func (sv *callListItemsService) DeleteCallListItemByUser(id string, userID string, workspaceID string) error {
	return sv.CallListItemsRepository.DeleteByUser(id, workspaceID, userID)
}

func (sv *callListItemsService) GetCallListItemsByFilterByUser(userID string, filter entities.CallListItemFilter) (*[]entities.CallListItemModel, error) {
	if filter.WorkspaceID == "" {
		return nil, errors.New("workspace_id must not be empty")
	}
	filter.UserID = userID
	return sv.CallListItemsRepository.FindByFilter(filter)
}
