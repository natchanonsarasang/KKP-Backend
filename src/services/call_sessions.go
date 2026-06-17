package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"
	"time"
)

type callSessionsService struct {
	CallSessionsRepository repositories.ICallSessionsRepository
}

type ICallSessionsService interface {
	// Direct system CRUD operations (no owner access check)
	CreateCallSession(data entities.CallSessionDataModel) error
	GetCallSessionByID(id string) (*entities.CallSessionDataModel, error)
	GetCallSessions(filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error)
	UpdateCallSession(id string, data entities.CallSessionDataModel) error
	DeleteCallSession(id string) error

	// User CRUD operations (with owner access check)
	CreateCallSessionByUser(callerUserID string, data entities.CallSessionDataModel) error
	GetCallSessionByIDByUser(callerUserID string, id string) (*entities.CallSessionDataModel, error)
	GetCallSessionsByUser(callerUserID string, filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error)
	UpdateCallSessionByUser(callerUserID string, id string, data entities.CallSessionDataModel) error
	DeleteCallSessionByUser(callerUserID string, id string) error
}

func NewCallSessionsService(repo repositories.ICallSessionsRepository) ICallSessionsService {
	return &callSessionsService{
		CallSessionsRepository: repo,
	}
}

// validateCallSession validates the call session data according to business rules
func validateCallSession(data entities.CallSessionDataModel) error {
	if data.UserID == "" {
		return errors.New("userID must not be empty")
	}
	if data.WorkspaceID == "" {
		return errors.New("workspaceID must not be empty")
	}
	if data.TotalCalls < 0 {
		return errors.New("totalCalls must not be negative")
	}
	if data.CompletedCalls < 0 {
		return errors.New("completedCalls must not be negative")
	}
	if data.FailedCalls < 0 {
		return errors.New("failedCalls must not be negative")
	}
	if data.ConfirmedCalls < 0 {
		return errors.New("confirmedCalls must not be negative")
	}
	if data.TokenUsed < 0 {
		return errors.New("tokenUsed must not be negative")
	}
	if data.StartedAt != nil {
		if data.StartedAt.IsZero() {
			return errors.New("invalid started_at datetime")
		}
	}
	if data.CompletedAt != nil {
		if data.CompletedAt.IsZero() {
			return errors.New("invalid completed_at datetime")
		}
	}
	if data.StartedAt != nil && data.CompletedAt != nil {
		if data.CompletedAt.Before(*data.StartedAt) {
			return errors.New("completed_at cannot be before started_at")
		}
	}
	return nil
}

// CreateCallSession (Direct System)
func (sv *callSessionsService) CreateCallSession(data entities.CallSessionDataModel) error {
	if err := validateCallSession(data); err != nil {
		return err
	}

	if data.ID == "" {
		session := entities.NewCallSession()
		data.ID = session.ID
		if data.Status == "" {
			data.Status = session.Status
		}
		data.CreatedAt = session.CreatedAt
		data.UpdatedAt = session.UpdatedAt
	} else {
		now := time.Now().UTC()
		if data.CreatedAt.IsZero() {
			data.CreatedAt = now
		}
		if data.UpdatedAt.IsZero() {
			data.UpdatedAt = now
		}
	}

	return sv.CallSessionsRepository.InsertCallSession(data)
}

// GetCallSessionByID (Direct System)
func (sv *callSessionsService) GetCallSessionByID(id string) (*entities.CallSessionDataModel, error) {
	return sv.CallSessionsRepository.FindByID(id)
}

// GetCallSessions (Direct System)
func (sv *callSessionsService) GetCallSessions(filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error) {
	return sv.CallSessionsRepository.FindByFilter(filter)
}

// UpdateCallSession (Direct System)
func (sv *callSessionsService) UpdateCallSession(id string, data entities.CallSessionDataModel) error {
	if err := validateCallSession(data); err != nil {
		return err
	}
	return sv.CallSessionsRepository.UpdateCallSession(id, data)
}

// DeleteCallSession (Direct System)
func (sv *callSessionsService) DeleteCallSession(id string) error {
	return sv.CallSessionsRepository.DeleteCallSession(id)
}

// CreateCallSessionByUser (User access context)
func (sv *callSessionsService) CreateCallSessionByUser(callerUserID string, data entities.CallSessionDataModel) error {
	if callerUserID == "" {
		return errors.New("unauthorized: callerUserID must not be empty")
	}

	if data.UserID == "" {
		data.UserID = callerUserID
	} else if data.UserID != callerUserID {
		return errors.New("unauthorized: user does not own this resource")
	}

	return sv.CreateCallSession(data)
}

// GetCallSessionByIDByUser (User access context)
func (sv *callSessionsService) GetCallSessionByIDByUser(callerUserID string, id string) (*entities.CallSessionDataModel, error) {
	if callerUserID == "" {
		return nil, errors.New("unauthorized: callerUserID must not be empty")
	}

	session, err := sv.CallSessionsRepository.FindByID(id)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}

	if session.UserID != callerUserID {
		return nil, errors.New("unauthorized: user does not own this resource")
	}

	return session, nil
}

// GetCallSessionsByUser (User access context)
func (sv *callSessionsService) GetCallSessionsByUser(callerUserID string, filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error) {
	if callerUserID == "" {
		return nil, errors.New("unauthorized: callerUserID must not be empty")
	}

	if filter.UserID != "" && filter.UserID != callerUserID {
		return nil, errors.New("unauthorized: cannot filter by other user ID")
	}
	filter.UserID = callerUserID

	return sv.GetCallSessions(filter)
}

// UpdateCallSessionByUser (User access context)
func (sv *callSessionsService) UpdateCallSessionByUser(callerUserID string, id string, data entities.CallSessionDataModel) error {
	if callerUserID == "" {
		return errors.New("unauthorized: callerUserID must not be empty")
	}

	existing, err := sv.CallSessionsRepository.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("session not found")
	}

	if existing.UserID != callerUserID {
		return errors.New("unauthorized: user does not own this resource")
	}

	// Ensure ID, UserID, and WorkspaceID cannot be changed
	data.ID = id
	data.UserID = callerUserID
	data.WorkspaceID = existing.WorkspaceID
	data.UpdatedAt = time.Now()

	return sv.CallSessionsRepository.UpdateCallSessionByUser(id, callerUserID, data)
}

// DeleteCallSessionByUser (User access context)
func (sv *callSessionsService) DeleteCallSessionByUser(callerUserID string, id string) error {
	if callerUserID == "" {
		return errors.New("unauthorized: callerUserID must not be empty")
	}

	return sv.CallSessionsRepository.DeleteCallSessionByUser(id, callerUserID)
}
