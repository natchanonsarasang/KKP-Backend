package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type callListItemsService struct {
	CallListItemsRepository repositories.ICallListItemsRepository
	DebtorsRepository       repositories.IDebtorsRepository
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

func NewCallListItemsService(repo repositories.ICallListItemsRepository, debtorsRepo repositories.IDebtorsRepository) ICallListItemsService {
	return &callListItemsService{
		CallListItemsRepository: repo,
		DebtorsRepository:       debtorsRepo,
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
	// Snapshot debtor phone/name/amount onto the item so completed history stays
	// readable after the debtor row is deleted. Only look up when the caller didn't
	// already supply a snapshot and a debtor is referenced.
	if data.DebtorPhone == "" && data.DebtorID != "" && sv.DebtorsRepository != nil {
		if debtor, err := sv.DebtorsRepository.FindByID(data.DebtorID); err == nil && debtor != nil {
			data.DebtorPhone = debtor.PhoneNumber
			data.DebtorName = DebtorDisplayName(debtor)
			data.DebtorAmount = DebtorDisplayAmount(debtor)
		}
	}
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return sv.CallListItemsRepository.Insert(data)
}

// DebtorDisplayName mirrors the frontend's name resolution: prefer the
// variables["name"] field, falling back to the debtor's Name column.
func DebtorDisplayName(d *entities.DebtorModel) string {
	if d.Variables != nil {
		if n, ok := d.Variables["name"]; ok && strings.TrimSpace(n) != "" {
			return n
		}
	}
	return d.Name
}

// DebtorDisplayAmount mirrors the frontend's amount resolution: prefer
// variables["amount"]/["outstanding_amount"] (comma-stripped), else TotalDebt.
func DebtorDisplayAmount(d *entities.DebtorModel) float64 {
	if d.Variables != nil {
		for _, key := range []string{"amount", "outstanding_amount"} {
			raw := strings.TrimSpace(d.Variables[key])
			if raw == "" {
				continue
			}
			if v, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", ""), 64); err == nil {
				return v
			}
		}
	}
	return d.TotalDebt
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
