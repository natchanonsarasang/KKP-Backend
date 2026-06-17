package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type workspacesService struct {
	WorkspacesRepository repositories.IWorkspacesRepository
}

type IWorkspacesService interface {
	// User CRUD operations (with owner access check)
	CreateWorkspaceByUser(userID string, data entities.WorkspaceDataModel) error
	GetWorkspaceByIDByUser(id string, userID string) (*entities.WorkspaceDataModel, error)
	GetAllWorkspacesByUser(userID string, filter entities.WorkspaceFilter) (*[]entities.WorkspaceDataModel, error)
	UpdateWorkspaceByUser(id string, userID string, data entities.WorkspaceDataModel) error
	DeleteWorkspaceByUser(id string, userID string) error

	// Direct/System CRUD methods (no owner access check)
	CreateWorkspace(data entities.WorkspaceDataModel) error
	GetWorkspaceByID(id string) (*entities.WorkspaceDataModel, error)
	GetAllWorkspaces(filter entities.WorkspaceFilter) (*[]entities.WorkspaceDataModel, error)
	UpdateWorkspace(id string, data entities.WorkspaceDataModel) error
	DeleteWorkspace(id string) error
}

func NewWorkspacesService(repo repositories.IWorkspacesRepository) IWorkspacesService {
	return &workspacesService{
		WorkspacesRepository: repo,
	}
}

// validateWorkspace runs all business logic validations on a WorkspaceDataModel
func (sv *workspacesService) validateWorkspace(data *entities.WorkspaceDataModel) error {
	if data.Name == "" {
		return errors.New("name must not be empty")
	}
	if data.OwnerID == "" {
		return errors.New("owner_id must not be empty")
	}
	return nil
}

// ===== Direct/System CRUD =====

func (sv *workspacesService) CreateWorkspace(data entities.WorkspaceDataModel) error {
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	data.CreatedAt = now
	data.UpdatedAt = now

	if err := sv.validateWorkspace(&data); err != nil {
		return err
	}

	return sv.WorkspacesRepository.InsertWorkspace(data)
}

func (sv *workspacesService) GetWorkspaceByID(id string) (*entities.WorkspaceDataModel, error) {
	if id == "" {
		return nil, errors.New("id must not be empty")
	}
	return sv.WorkspacesRepository.FindByID(id)
}

func (sv *workspacesService) GetAllWorkspaces(filter entities.WorkspaceFilter) (*[]entities.WorkspaceDataModel, error) {
	return sv.WorkspacesRepository.FindByFilter(filter)
}

func (sv *workspacesService) UpdateWorkspace(id string, data entities.WorkspaceDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	existing, err := sv.WorkspacesRepository.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("workspace not found")
	}

	data.ID = id // Ensure ID cannot be changed
	data.UpdatedAt = time.Now().UTC()

	if err := sv.validateWorkspace(&data); err != nil {
		return err
	}

	return sv.WorkspacesRepository.UpdateWorkspace(id, data)
}

func (sv *workspacesService) DeleteWorkspace(id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	existing, err := sv.WorkspacesRepository.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("workspace not found")
	}

	return sv.WorkspacesRepository.DeleteWorkspace(id)
}

// ===== User access context CRUD =====

func (sv *workspacesService) CreateWorkspaceByUser(userID string, data entities.WorkspaceDataModel) error {
	if userID == "" {
		return errors.New("unauthorized: missing user id")
	}

	if data.OwnerID == "" {
		data.OwnerID = userID
	} else if data.OwnerID != userID {
		return errors.New("unauthorized: you cannot create a workspace owned by another user")
	}

	return sv.CreateWorkspace(data)
}

func (sv *workspacesService) GetWorkspaceByIDByUser(id string, userID string) (*entities.WorkspaceDataModel, error) {
	if id == "" {
		return nil, errors.New("id must not be empty")
	}
	if userID == "" {
		return nil, errors.New("unauthorized: missing user id")
	}

	workspace, err := sv.WorkspacesRepository.FindByID(id)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, nil
	}
	if workspace.OwnerID != userID {
		return nil, errors.New("unauthorized: you do not own this workspace")
	}
	return workspace, nil
}

func (sv *workspacesService) GetAllWorkspacesByUser(userID string, filter entities.WorkspaceFilter) (*[]entities.WorkspaceDataModel, error) {
	if userID == "" {
		return nil, errors.New("unauthorized: missing user id")
	}
	if filter.OwnerID != "" && filter.OwnerID != userID {
		return nil, errors.New("unauthorized: cannot filter by other user ID")
	}
	filter.OwnerID = userID
	return sv.WorkspacesRepository.FindByFilter(filter)
}

func (sv *workspacesService) UpdateWorkspaceByUser(id string, userID string, data entities.WorkspaceDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if userID == "" {
		return errors.New("unauthorized: missing user id")
	}

	// Ensure ID and OwnerID cannot be changed
	data.ID = id
	data.OwnerID = userID
	data.UpdatedAt = time.Now().UTC()

	if err := sv.validateWorkspace(&data); err != nil {
		return err
	}

	return sv.WorkspacesRepository.UpdateWorkspaceByUser(id, userID, data)
}

func (sv *workspacesService) DeleteWorkspaceByUser(id string, userID string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if userID == "" {
		return errors.New("unauthorized: missing user id")
	}

	return sv.WorkspacesRepository.DeleteWorkspaceByUser(id, userID)
}