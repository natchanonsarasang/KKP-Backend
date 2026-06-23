package services

import (
	"context"
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type ICallTemplatesService interface {
	GetTemplatesByFilter(ctx context.Context, id string, templateID string) ([]*entities.CallTemplateDataModel, error)
	CreateCallTemplate(ctx context.Context, data *entities.CallTemplateDataModel) error
	UpdateCallTemplate(ctx context.Context, id string, data *entities.CallTemplateDataModel) error
	DeleteCallTemplate(ctx context.Context, id string) error
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

func (s *callTemplatesService) CreateCallTemplate(ctx context.Context, data *entities.CallTemplateDataModel) error {
	if data == nil {
		return errors.New("template data cannot be nil")
	}

	// Initialize default values for creation if not set
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	data.CreatedAt = now
	data.UpdatedAt = now

	return s.Repo.InsertCallTemplate(ctx, data)
}

func (s *callTemplatesService) UpdateCallTemplate(ctx context.Context, id string, data *entities.CallTemplateDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if data == nil {
		return errors.New("template data cannot be nil")
	}

	// Update timestamp
	data.UpdatedAt = time.Now().UTC()
	data.ID = id // Ensure ID cannot be changed

	return s.Repo.UpdateCallTemplate(ctx, id, data)
}

func (s *callTemplatesService) DeleteCallTemplate(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	return s.Repo.DeleteCallTemplate(ctx, id)
}
