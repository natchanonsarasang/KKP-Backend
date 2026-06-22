package services

import (
	"context"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
)

type ICallTemplatesService interface {
	GetTemplatesByFilter(ctx context.Context, id string, templateID string) ([]*entities.CallTemplateDataModel, error)
}

type callTemplatesService struct {
	Repo repositories.ICallTemplatesRepository
}

func NewCallTemplatesService(repo repositories.ICallTemplatesRepository) ICallTemplatesService {
	return &callTemplatesService{Repo: repo}
}

func (s *callTemplatesService) GetTemplatesByFilter(ctx context.Context, id string, templateID string) ([]*entities.CallTemplateDataModel, error) {
	return s.Repo.FindByFilter(ctx, id, templateID)
}
