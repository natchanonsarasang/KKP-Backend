package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type debtorsService struct {
	DebtorsRepository repositories.IDebtorsRepository
	UsersService      IUsersService
}

type IDebtorsService interface {
	GetDebtorsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.DebtorModel, error)
	GetDebtorsByWorkspaceByUser(userID string, workspaceID primitive.ObjectID) (*[]entities.DebtorModel, error)
	GetDebtorByID(id primitive.ObjectID) (*entities.DebtorModel, error)
	GetDebtorByIDByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) (*entities.DebtorModel, error)
	CreateDebtor(data entities.DebtorModel) error
	CreateDebtorByUser(userID string, data entities.DebtorModel) error
	// System Methods
	UpdateDebtor(id primitive.ObjectID, data entities.DebtorModel) error
	DeleteDebtor(id primitive.ObjectID) error
	// ByUser Methods
	UpdateDebtorByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID, data entities.DebtorModel) error
	DeleteDebtorByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) error
}

func NewDebtorsService(repo repositories.IDebtorsRepository, usersService IUsersService) IDebtorsService {
	return &debtorsService{
		DebtorsRepository: repo,
		UsersService:      usersService,
	}
}

func (sv *debtorsService) GetDebtorsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.DebtorModel, error) {
	return sv.DebtorsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *debtorsService) GetDebtorsByWorkspaceByUser(userID string, workspaceID primitive.ObjectID) (*[]entities.DebtorModel, error) {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized access to this workspace")
	}
	return sv.DebtorsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *debtorsService) GetDebtorByID(id primitive.ObjectID) (*entities.DebtorModel, error) {
	return sv.DebtorsRepository.FindByID(id)
}

func (sv *debtorsService) GetDebtorByIDByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) (*entities.DebtorModel, error) {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized access to this workspace")
	}
	return sv.DebtorsRepository.FindByIDByUser(id, workspaceID)
}

func (sv *debtorsService) CreateDebtor(data entities.DebtorModel) error {
	data.CreatedAt = time.Now().Add(7 * time.Hour)
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.DebtorsRepository.Insert(data)
}

func (sv *debtorsService) CreateDebtorByUser(userID string, data entities.DebtorModel) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, data.WorkspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}
	return sv.CreateDebtor(data)
}

func (sv *debtorsService) UpdateDebtor(id primitive.ObjectID, data entities.DebtorModel) error {
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.DebtorsRepository.Update(id, data)
}

func (sv *debtorsService) DeleteDebtor(id primitive.ObjectID) error {
	return sv.DebtorsRepository.Delete(id)
}

func (sv *debtorsService) UpdateDebtorByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID, data entities.DebtorModel) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}

	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.DebtorsRepository.UpdateByUser(id, workspaceID, data)
}

func (sv *debtorsService) DeleteDebtorByUser(id primitive.ObjectID, userID string, workspaceID primitive.ObjectID) error {
	isMember, err := sv.UsersService.VerifyUserInWorkspace(userID, workspaceID)
	if err != nil || !isMember {
		return errors.New("unauthorized access to this workspace")
	}

	return sv.DebtorsRepository.DeleteByUser(id, workspaceID)
}
