package gateways

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type mockCallSessionsService struct {
	sessions map[string]entities.CallSessionDataModel
}

func (m *mockCallSessionsService) CreateCallSession(data entities.CallSessionDataModel) error {
	m.sessions[data.ID] = data
	return nil
}

func (m *mockCallSessionsService) GetCallSessionByID(id string) (*entities.CallSessionDataModel, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	return &s, nil
}

func (m *mockCallSessionsService) GetCallSessions(filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error) {
	var list []entities.CallSessionDataModel
	for _, s := range m.sessions {
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		if filter.WorkspaceID != "" && s.WorkspaceID != filter.WorkspaceID {
			continue
		}
		if filter.UserID != "" && s.UserID != filter.UserID {
			continue
		}
		list = append(list, s)
	}
	return &list, nil
}

func (m *mockCallSessionsService) UpdateCallSession(id string, data entities.CallSessionDataModel) error {
	m.sessions[id] = data
	return nil
}

func (m *mockCallSessionsService) DeleteCallSession(id string) error {
	delete(m.sessions, id)
	return nil
}

func (m *mockCallSessionsService) CreateCallSessionByUser(callerUserID string, data entities.CallSessionDataModel) error {
	if data.UserID == "" {
		data.UserID = callerUserID
	} else if data.UserID != callerUserID {
		return errors.New("unauthorized: user does not own this resource")
	}
	m.sessions[data.ID] = data
	return nil
}

func (m *mockCallSessionsService) GetCallSessionByIDByUser(callerUserID string, id string) (*entities.CallSessionDataModel, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	if s.UserID != callerUserID {
		return nil, errors.New("unauthorized: user does not own this resource")
	}
	return &s, nil
}

func (m *mockCallSessionsService) GetCallSessionsByUser(callerUserID string, filter entities.CallSessionFilter) (*[]entities.CallSessionDataModel, error) {
	if filter.UserID != "" && filter.UserID != callerUserID {
		return nil, errors.New("unauthorized: cannot filter by other user ID")
	}
	filter.UserID = callerUserID
	return m.GetCallSessions(filter)
}

func (m *mockCallSessionsService) UpdateCallSessionByUser(callerUserID string, id string, data entities.CallSessionDataModel) error {
	s, ok := m.sessions[id]
	if !ok {
		return errors.New("session not found")
	}
	if s.UserID != callerUserID {
		return errors.New("unauthorized: user does not own this resource")
	}
	m.sessions[id] = data
	return nil
}

func (m *mockCallSessionsService) DeleteCallSessionByUser(callerUserID string, id string) error {
	s, ok := m.sessions[id]
	if !ok {
		return errors.New("session not found")
	}
	if s.UserID != callerUserID {
		return errors.New("unauthorized: user does not own this resource")
	}
	delete(m.sessions, id)
	return nil
}

func TestCallSessionsGateway(t *testing.T) {
	// Setup JWT secret
	os.Setenv("JWT_SECRET_KEY", "test-secret-key-12345")

	// Generate tokens
	ownerTokenDetails, err := middlewares.GenerateJWTToken("user-owner", "uuid-owner")
	assert.NoError(t, err)

	otherTokenDetails, err := middlewares.GenerateJWTToken("user-other", "uuid-other")
	assert.NoError(t, err)

	// Setup App & Gateway
	app := fiber.New()
	mockSvc := &mockCallSessionsService{sessions: make(map[string]entities.CallSessionDataModel)}
	gateway := HTTPGateway{
		CallSessionService: mockSvc,
	}

	GatewayCallSessions(gateway, app)

	// Pre-populate mock database
	session1 := entities.CallSessionDataModel{
		ID:          "session-1",
		UserID:      "user-owner",
		WorkspaceID: "workspace-1",
		Status:      "pending",
	}
	mockSvc.sessions["session-1"] = session1

	// Test 1: Create Call Session (Authorized)
	t.Run("Create Call Session - Authorized", func(t *testing.T) {
		newSession := entities.CallSessionDataModel{
			ID:          "session-2",
			UserID:      "user-owner",
			WorkspaceID: "workspace-1",
			Status:      "pending",
		}
		bodyBytes, _ := json.Marshal(newSession)
		req := httptest.NewRequest("POST", "/api/v1/call-sessions/", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+*ownerTokenDetails.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		created, exists := mockSvc.sessions["session-2"]
		assert.True(t, exists)
		assert.Equal(t, "user-owner", created.UserID)
	})

	// Test 2: Create Call Session - Mismatched UserID -> Fail
	t.Run("Create Call Session - Mismatched UserID", func(t *testing.T) {
		newSession := entities.CallSessionDataModel{
			ID:          "session-3",
			UserID:      "user-owner", // Mismatch with "user-other" caller
			WorkspaceID: "workspace-1",
			Status:      "pending",
		}
		bodyBytes, _ := json.Marshal(newSession)
		req := httptest.NewRequest("POST", "/api/v1/call-sessions/", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+*otherTokenDetails.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 3: Get Call Sessions with query params
	t.Run("Get Call Sessions - Filtered", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/call-sessions/?status=pending&workspace_id=workspace-1", nil)
		req.Header.Set("Authorization", "Bearer "+*ownerTokenDetails.Token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res entities.ResponseModel
		body, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &res)
		assert.NoError(t, err)

		// Check output structure using maps
		dataList, ok := res.Data.([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(dataList), 1)
	})

	// Test 4: Get Call Sessions specifying other user ID -> Fail
	t.Run("Get Call Sessions - Attempt other user filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/call-sessions/?user_id=user-other", nil)
		req.Header.Set("Authorization", "Bearer "+*ownerTokenDetails.Token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// Test 5: Get Call Session by ID - Owner -> Success
	t.Run("Get Call Session by ID - Owner Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/call-sessions/session-1", nil)
		req.Header.Set("Authorization", "Bearer "+*ownerTokenDetails.Token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test 6: Get Call Session by ID - Non-owner -> Fail
	t.Run("Get Call Session by ID - Non-owner Fail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/call-sessions/session-1", nil)
		req.Header.Set("Authorization", "Bearer "+*otherTokenDetails.Token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Test 7: Update Call Session - Owner -> Success
	t.Run("Update Call Session - Owner Success", func(t *testing.T) {
		updatedSession := entities.CallSessionDataModel{
			ID:          "session-1",
			UserID:      "user-owner",
			WorkspaceID: "workspace-1",
			Status:      "running",
		}
		bodyBytes, _ := json.Marshal(updatedSession)
		req := httptest.NewRequest("PUT", "/api/v1/call-sessions/session-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+*ownerTokenDetails.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "running", mockSvc.sessions["session-1"].Status)
	})

	// Test 8: Update Call Session - Non-owner -> Fail
	t.Run("Update Call Session - Non-owner Fail", func(t *testing.T) {
		updatedSession := entities.CallSessionDataModel{
			ID:          "session-1",
			UserID:      "user-other",
			WorkspaceID: "workspace-1",
			Status:      "failed",
		}
		bodyBytes, _ := json.Marshal(updatedSession)
		req := httptest.NewRequest("PUT", "/api/v1/call-sessions/session-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+*otherTokenDetails.Token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.NotEqual(t, "failed", mockSvc.sessions["session-1"].Status)
	})

	// Test 9: Delete Call Session - Non-owner -> Fail
	t.Run("Delete Call Session - Non-owner Fail", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/call-sessions/session-1", nil)
		req.Header.Set("Authorization", "Bearer "+*otherTokenDetails.Token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		_, exists := mockSvc.sessions["session-1"]
		assert.True(t, exists)
	})

	// Test 10: Delete Call Session - Owner -> Success
	t.Run("Delete Call Session - Owner Success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/call-sessions/session-1", nil)
		req.Header.Set("Authorization", "Bearer "+*ownerTokenDetails.Token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_, exists := mockSvc.sessions["session-1"]
		assert.False(t, exists)
	})

	// Test 11: Unauthorized request (no token) -> Fail
	t.Run("Unauthorized Request - No Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/call-sessions/", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		// SetJWtHeaderHandler middleware returns 401 Unauthorized for invalid token
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
