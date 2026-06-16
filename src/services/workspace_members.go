package services

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"

	"github.com/google/uuid"
)

type workspaceMembersService struct {
	WorkspaceMembersRepository repositories.IWorkspaceMembersRepository
}

type IWorkspaceMembersService interface {
	GetAllWorkspaceMembers() (*[]entities.WorkspaceMemberDataModel, error)
	GetWorkspaceMemberByID(id string) (*entities.WorkspaceMemberDataModel, error)
	InsertNewWorkspaceMember(data entities.WorkspaceMemberDataModel) error
	UpdateWorkspaceMember(id string, data entities.WorkspaceMemberDataModel) error
	DeleteWorkspaceMember(id string) error
}

func NewWorkspaceMembersService(repo0 repositories.IWorkspaceMembersRepository) IWorkspaceMembersService {
	return &workspaceMembersService{
		WorkspaceMembersRepository: repo0,
	}
}

func (sv *workspaceMembersService) GetAllWorkspaceMembers() (*[]entities.WorkspaceMemberDataModel, error) {
	data, err := sv.WorkspaceMembersRepository.FindAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (sv *workspaceMembersService) GetWorkspaceMemberByID(id string) (*entities.WorkspaceMemberDataModel, error) {
	data, err := sv.WorkspaceMembersRepository.FindByID(id)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (sv *workspaceMembersService) InsertNewWorkspaceMember(data entities.WorkspaceMemberDataModel) error {
	data.ID = uuid.NewString()
	data.CreatedAt = time.Now().Add(7 * time.Hour)

	if data.Role == "" {
		data.Role = "member"
	}

	return sv.WorkspaceMembersRepository.InsertWorkspaceMember(data)
}

func (sv *workspaceMembersService) UpdateWorkspaceMember(id string, data entities.WorkspaceMemberDataModel) error {
	return sv.WorkspaceMembersRepository.UpdateByID(id, data)
}

func (sv *workspaceMembersService) DeleteWorkspaceMember(id string) error {
	return sv.WorkspaceMembersRepository.DeleteByID(id)
}