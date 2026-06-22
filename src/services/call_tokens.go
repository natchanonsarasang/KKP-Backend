package services

import (
	"context"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
)

type ICallTokensService interface {
	GetTokensByFilter(ctx context.Context, id string, userID string) ([]*entities.CallTokenDataModel, error)
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
