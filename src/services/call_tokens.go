package services

import (
	"context"
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type ICallTokensService interface {
	GetTokensByFilter(ctx context.Context, id string, userID string) ([]*entities.CallTokenDataModel, error)
	CreateCallToken(ctx context.Context, data *entities.CallTokenDataModel) error
	UpdateCallToken(ctx context.Context, id string, data *entities.CallTokenDataModel) error
	DeleteCallToken(ctx context.Context, id string) error
}

type callTokensService struct {
	Repo repositories.ICallTokensRepository
}

func NewCallTokensService(repo repositories.ICallTokensRepository) ICallTokensService {
	return &callTokensService{Repo: repo}
}

func (s *callTokensService) GetTokensByFilter(ctx context.Context, id string, userID string) ([]*entities.CallTokenDataModel, error) {
	return s.Repo.FindByFilter(ctx, id, userID)
}

func (s *callTokensService) CreateCallToken(ctx context.Context, data *entities.CallTokenDataModel) error {
	if data == nil {
		return errors.New("token data cannot be nil")
	}

	// Initialize default values for creation if not set
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	data.CreatedAt = now
	data.UpdatedAt = now

	return s.Repo.InsertCallToken(ctx, data)
}

func (s *callTokensService) UpdateCallToken(ctx context.Context, id string, data *entities.CallTokenDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if data == nil {
		return errors.New("token data cannot be nil")
	}

	// Update timestamp
	data.UpdatedAt = time.Now().UTC()
	data.ID = id // Ensure ID cannot be changed

	return s.Repo.UpdateCallToken(ctx, id, data)
}

func (s *callTokensService) DeleteCallToken(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	return s.Repo.DeleteCallToken(ctx, id)
}
