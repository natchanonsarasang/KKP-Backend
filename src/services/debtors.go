package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type debtorsService struct {
	DebtorsRepository repositories.IDebtorsRepository
}

type IDebtorsService interface {
	GetDebtorsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.DebtorModel, error)
	GetDebtorByID(id primitive.ObjectID) (*entities.DebtorModel, error)
	CreateDebtor(data entities.DebtorModel) error
	UpdateDebtor(id primitive.ObjectID, data entities.DebtorModel) error
	DeleteDebtor(id primitive.ObjectID) error
}

func NewDebtorsService(repo repositories.IDebtorsRepository) IDebtorsService {
	return &debtorsService{
		DebtorsRepository: repo,
	}
}

func (sv *debtorsService) GetDebtorsByWorkspace(workspaceID primitive.ObjectID) (*[]entities.DebtorModel, error) {
	return sv.DebtorsRepository.FindAllByWorkspace(workspaceID)
}

func (sv *debtorsService) GetDebtorByID(id primitive.ObjectID) (*entities.DebtorModel, error) {
	return sv.DebtorsRepository.FindByID(id)
}

func (sv *debtorsService) CreateDebtor(data entities.DebtorModel) error {
	data.CreatedAt = time.Now().Add(7 * time.Hour)
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.DebtorsRepository.Insert(data)
}

func (sv *debtorsService) UpdateDebtor(id primitive.ObjectID, data entities.DebtorModel) error {
	data.UpdatedAt = time.Now().Add(7 * time.Hour)
	return sv.DebtorsRepository.Update(id, data)
}

func (sv *debtorsService) DeleteDebtor(id primitive.ObjectID) error {
	return sv.DebtorsRepository.Delete(id)
}
