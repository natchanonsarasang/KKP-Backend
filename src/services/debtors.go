package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type debtorsService struct {
	DebtorsRepository repositories.IDebtorsRepository
}

type IDebtorsService interface {
	GetDebtorsByWorkspace(workspaceID string) (*[]entities.DebtorModel, error)
	GetDebtorsByWorkspaceByUser(userID string, workspaceID string) (*[]entities.DebtorModel, error)
	GetDebtorByID(id string) (*entities.DebtorModel, error)
	GetDebtorByIDByUser(id string, userID string, workspaceID string) (*entities.DebtorModel, error)
	CreateDebtor(data entities.DebtorModel) error
	CreateDebtorByUser(userID string, data entities.DebtorModel) error
	// System Methods
	UpdateDebtor(id string, data entities.DebtorModel) error
	DeleteDebtor(id string) error
	// ByUser Methods
	UpdateDebtorByUser(id string, userID string, workspaceID string, data entities.DebtorModel) error
	DeleteDebtorByUser(id string, userID string, workspaceID string) error
}

func NewDebtorsService(repo repositories.IDebtorsRepository) IDebtorsService {
	return &debtorsService{
		DebtorsRepository: repo,
	}
}

func (sv *debtorsService) GetDebtorsByWorkspace(workspaceID string) (*[]entities.DebtorModel, error) {
	// For system use, we might pass empty userID to get all (if repository allowed it).
	// To maintain the interface, we'll pass empty string. The repo will filter by it, so this might return none if user_id is empty.
	// Actually, FindAllByWorkspace now requires userID in repo. Let's pass empty for system (you might need a separate repo method for true system Get All).
	return sv.DebtorsRepository.FindAllByWorkspace(workspaceID, "")
}

func (sv *debtorsService) GetDebtorsByWorkspaceByUser(userID string, workspaceID string) (*[]entities.DebtorModel, error) {
	return sv.DebtorsRepository.FindAllByWorkspace(workspaceID, userID)
}

func (sv *debtorsService) GetDebtorByID(id string) (*entities.DebtorModel, error) {
	return sv.DebtorsRepository.FindByID(id)
}

func (sv *debtorsService) GetDebtorByIDByUser(id string, userID string, workspaceID string) (*entities.DebtorModel, error) {
	debtor, err := sv.DebtorsRepository.FindByIDByUser(id, workspaceID)
	if err != nil {
		return nil, err
	}
	if debtor.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this debtor record")
	}
	return debtor, nil
}

func (sv *debtorsService) CreateDebtor(data entities.DebtorModel) error {
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	return sv.DebtorsRepository.Insert(data)
}

func (sv *debtorsService) CreateDebtorByUser(userID string, data entities.DebtorModel) error {
	data.UserID = userID
	return sv.CreateDebtor(data)
}

func (sv *debtorsService) UpdateDebtor(id string, data entities.DebtorModel) error {
	data.UpdatedAt = time.Now()
	return sv.DebtorsRepository.Update(id, data)
}

func (sv *debtorsService) DeleteDebtor(id string) error {
	return sv.DebtorsRepository.Delete(id)
}

func (sv *debtorsService) UpdateDebtorByUser(id string, userID string, workspaceID string, data entities.DebtorModel) error {
	data.ID = id
	data.UserID = userID
	data.UpdatedAt = time.Now()
	return sv.DebtorsRepository.UpdateByUser(id, workspaceID, userID, data)
}

func (sv *debtorsService) DeleteDebtorByUser(id string, userID string, workspaceID string) error {
	return sv.DebtorsRepository.DeleteByUser(id, workspaceID, userID)
}
