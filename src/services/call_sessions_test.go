package services

import (
	"errors"
	"go-fiber-template/domain/entities"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockCallSessionsRepository struct {
	sessions map[string]entities.CallSessionDataModel
}

func (m *mockCallSessionsRepository) InsertCallSession(data entities.CallSessionDataModel) error {
	m.sessions[data.ID] = data
	return nil
}

func (m *mockCallSessionsRepository) FindByID(id string) (*entities.CallSessionDataModel, error) {
	s, exists := m.sessions[id]
	if !exists {
		return nil, nil
	}
	return &s, nil
}

func (m *mockCallSessionsRepository) FindOneByStatus(status string) (*entities.CallSessionDataModel, error) {
	for _, s := range m.sessions {
		if s.Status == status {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockCallSessionsRepository) FindByStatus(status string) (*[]entities.CallSessionDataModel, error) {
	var res []entities.CallSessionDataModel
	for _, s := range m.sessions {
		if s.Status == status {
			res = append(res, s)
		}
	}
	return &res, nil
}

func (m *mockCallSessionsRepository) FindByWorkspaceID(workspaceID string) (*[]entities.CallSessionDataModel, error) {
	var res []entities.CallSessionDataModel
	for _, s := range m.sessions {
		if s.WorkspaceID == workspaceID {
			res = append(res, s)
		}
	}
	return &res, nil
}

func (m *mockCallSessionsRepository) FindByUserID(userID string) (*[]entities.CallSessionDataModel, error) {
	var res []entities.CallSessionDataModel
	for _, s := range m.sessions {
		if s.UserID == userID {
			res = append(res, s)
		}
	}
	return &res, nil
}

func (m *mockCallSessionsRepository) FindByFilter(filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error) {
	var res []entities.CallSessionDataModel
	for _, s := range m.sessions {
		if filter.ID != "" && s.ID != filter.ID {
			continue
		}
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		if filter.WorkspaceID != "" && s.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.UserID != "" && s.UserID != filter.UserID {
			continue
		}
		res = append(res, s)
	}
	return &res, nil
}

func (m *mockCallSessionsRepository) UpdateCallSession(id string, data entities.CallSessionDataModel) error {
	m.sessions[id] = data
	return nil
}

func (m *mockCallSessionsRepository) DeleteCallSession(id string) error {
	delete(m.sessions, id)
	return nil
}

func TestCallSessionsService_Validation(t *testing.T) {
	repo := &mockCallSessionsRepository{sessions: make(map[string]entities.CallSessionDataModel)}
	service := NewCallSessionsService(repo)

	// Valid session base
	validSession := entities.CallSessionDataModel{
		ID:             "session-1",
		UserID:         "user-1",
		WorkspaceID:    "workspace-1",
		Status:         "pending",
		TotalCalls:     10,
		CompletedCalls: 5,
		FailedCalls:    3,
		ConfirmedCalls: 2,
		TokenUsed:      50,
	}

	// 1. Test empty userID
	s := validSession
	s.UserID = ""
	err := service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "userID must not be empty", err.Error())

	// 2. Test empty workspaceID
	s = validSession
	s.WorkspaceID = ""
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "workspaceID must not be empty", err.Error())

	// 3. Test negative totalCalls
	s = validSession
	s.TotalCalls = -1
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "totalCalls must not be negative", err.Error())

	// 4. Test negative completedCalls
	s = validSession
	s.CompletedCalls = -5
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "completedCalls must not be negative", err.Error())

	// 5. Test negative failedCalls
	s = validSession
	s.FailedCalls = -3
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "failedCalls must not be negative", err.Error())

	// 6. Test negative confirmedCalls
	s = validSession
	s.ConfirmedCalls = -2
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "confirmedCalls must not be negative", err.Error())

	// 7. Test negative tokenUsed
	s = validSession
	s.TokenUsed = -10
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "tokenUsed must not be negative", err.Error())

	// 8. Test invalid startedAt zero time
	s = validSession
	zeroTime := time.Time{}
	s.StartedAt = &zeroTime
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "invalid started_at datetime", err.Error())

	// 9. Test invalid completedAt zero time
	s = validSession
	s.CompletedAt = &zeroTime
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "invalid completed_at datetime", err.Error())

	// 10. Test completedAt before startedAt
	s = validSession
	started := time.Now().UTC()
	completed := started.Add(-10 * time.Minute)
	s.StartedAt = &started
	s.CompletedAt = &completed
	err = service.CreateCallSession(s)
	assert.Error(t, err)
	assert.Equal(t, "completed_at cannot be before started_at", err.Error())
}

func TestCallSessionsService_DirectSystemCRUD(t *testing.T) {
	repo := &mockCallSessionsRepository{sessions: make(map[string]entities.CallSessionDataModel)}
	service := NewCallSessionsService(repo)

	session := entities.CallSessionDataModel{
		ID:             "session-1",
		UserID:         "user-1",
		WorkspaceID:    "workspace-1",
		Status:         "pending",
		TotalCalls:     10,
		CompletedCalls: 5,
		FailedCalls:    3,
		ConfirmedCalls: 2,
		TokenUsed:      50,
	}

	// 1. Create
	err := service.CreateCallSession(session)
	assert.NoError(t, err)

	// 2. Get by ID
	found, err := service.GetCallSessionByID("session-1")
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "session-1", found.ID)

	// 3. Get with filter (Option A)
	sessions, err := service.GetCallSessions(entities.CallSessionFilter{
		Status: "pending",
	})
	assert.NoError(t, err)
	assert.Len(t, *sessions, 1)

	// 4. Update
	updatedSession := session
	updatedSession.CompletedCalls = 8
	err = service.UpdateCallSession("session-1", updatedSession)
	assert.NoError(t, err)

	foundUpdated, _ := service.GetCallSessionByID("session-1")
	assert.Equal(t, 8, foundUpdated.CompletedCalls)

	// 5. Delete
	err = service.DeleteCallSession("session-1")
	assert.NoError(t, err)

	deleted, _ := service.GetCallSessionByID("session-1")
	assert.Nil(t, deleted)
}

func TestCallSessionsService_UserCRUD(t *testing.T) {
	repo := &mockCallSessionsRepository{sessions: make(map[string]entities.CallSessionDataModel)}
	service := NewCallSessionsService(repo)

	session := entities.CallSessionDataModel{
		ID:             "session-1",
		UserID:         "user-owner",
		WorkspaceID:    "workspace-1",
		Status:         "pending",
		TotalCalls:     10,
		CompletedCalls: 5,
		FailedCalls:    3,
		ConfirmedCalls: 2,
		TokenUsed:      50,
	}

	// 1. Create by owner -> Success
	err := service.CreateCallSessionByUser("user-owner", session)
	assert.NoError(t, err)

	// 2. Create by non-owner -> Fail
	err = service.CreateCallSessionByUser("user-other", session)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("unauthorized: user does not own this resource")) || err.Error() == "unauthorized: user does not own this resource")

	// 3. Get by ID by owner -> Success
	found, err := service.GetCallSessionByIDByUser("user-owner", "session-1")
	assert.NoError(t, err)
	assert.NotNil(t, found)

	// 4. Get by ID by non-owner -> Fail
	_, err = service.GetCallSessionByIDByUser("user-other", "session-1")
	assert.Error(t, err)
	assert.Equal(t, "unauthorized: user does not own this resource", err.Error())

	// 5. Get sessions by owner -> Success, filters for owner's ID
	results, err := service.GetCallSessionsByUser("user-owner", entities.CallSessionFilter{})
	assert.NoError(t, err)
	assert.Len(t, *results, 1)

	// Get sessions specifying other user's ID -> Fail
	_, err = service.GetCallSessionsByUser("user-owner", entities.CallSessionFilter{UserID: "user-other"})
	assert.Error(t, err)
	assert.Equal(t, "unauthorized: cannot filter by other user ID", err.Error())

	// 6. Update by owner -> Success
	updated := session
	updated.CompletedCalls = 9
	err = service.UpdateCallSessionByUser("user-owner", "session-1", updated)
	assert.NoError(t, err)

	// Update by non-owner -> Fail
	err = service.UpdateCallSessionByUser("user-other", "session-1", updated)
	assert.Error(t, err)
	assert.Equal(t, "unauthorized: user does not own this resource", err.Error())

	// 7. Delete by non-owner -> Fail
	err = service.DeleteCallSessionByUser("user-other", "session-1")
	assert.Error(t, err)
	assert.Equal(t, "unauthorized: user does not own this resource", err.Error())

	// Delete by owner -> Success
	err = service.DeleteCallSessionByUser("user-owner", "session-1")
	assert.NoError(t, err)

	deleted, _ := service.GetCallSessionByID("session-1")
	assert.Nil(t, deleted)
}
