package services

import (
	"errors"
	"testing"
	"time"

	"go-fiber-template/domain/entities"

	"github.com/stretchr/testify/assert"
)

// mockCallRecordsRepository implements repositories.ICallRecordsRepository for unit testing services.
type mockCallRecordsRepository struct {
	InsertFunc       func(data entities.CallRecordDataModel) error
	FindByIDFunc     func(id string) (*entities.CallRecordDataModel, error)
	FindByUserIDFunc func(userID string) (*[]entities.CallRecordDataModel, error)
	FindByFilterFunc func(filter entities.CallRecordFilter) (*[]entities.CallRecordDataModel, error)
	FindAllFunc      func() (*[]entities.CallRecordDataModel, error)
	UpdateFunc       func(id string, data entities.CallRecordDataModel) error
	DeleteFunc       func(id string) error
	UpdateByUserFunc func(id string, userID string, data entities.CallRecordDataModel) error
	DeleteByUserFunc func(id string, userID string) error
}

func (m *mockCallRecordsRepository) InsertCallRecord(data entities.CallRecordDataModel) error {
	return m.InsertFunc(data)
}

func (m *mockCallRecordsRepository) FindByID(id string) (*entities.CallRecordDataModel, error) {
	return m.FindByIDFunc(id)
}

func (m *mockCallRecordsRepository) FindByUserID(userID string) (*[]entities.CallRecordDataModel, error) {
	return m.FindByUserIDFunc(userID)
}

func (m *mockCallRecordsRepository) FindByFilter(filter entities.CallRecordFilter) (*[]entities.CallRecordDataModel, error) {
	return m.FindByFilterFunc(filter)
}

func (m *mockCallRecordsRepository) FindAll() (*[]entities.CallRecordDataModel, error) {
	return m.FindAllFunc()
}

func (m *mockCallRecordsRepository) UpdateCallRecord(id string, data entities.CallRecordDataModel) error {
	return m.UpdateFunc(id, data)
}

func (m *mockCallRecordsRepository) DeleteCallRecord(id string) error {
	return m.DeleteFunc(id)
}

func (m *mockCallRecordsRepository) UpdateCallRecordByUser(id string, userID string, data entities.CallRecordDataModel) error {
	if m.UpdateByUserFunc != nil {
		return m.UpdateByUserFunc(id, userID, data)
	}
	if m.FindByIDFunc != nil {
		rec, err := m.FindByIDFunc(id)
		if err != nil {
			return err
		}
		if rec == nil {
			return errors.New("call record not found")
		}
		if rec.UserID != userID {
			return errors.New("unauthorized: you do not own this call record")
		}
	}
	return nil
}

func (m *mockCallRecordsRepository) DeleteCallRecordByUser(id string, userID string) error {
	if m.DeleteByUserFunc != nil {
		return m.DeleteByUserFunc(id, userID)
	}
	if m.FindByIDFunc != nil {
		rec, err := m.FindByIDFunc(id)
		if err != nil {
			return err
		}
		if rec == nil {
			return errors.New("call record not found")
		}
		if rec.UserID != userID {
			return errors.New("unauthorized: you do not own this call record")
		}
	}
	return nil
}

func TestCreateCallRecordValidation(t *testing.T) {
	now := time.Now().UTC()
	nowPtr := &now
	templateID := "tpl-1"

	tests := []struct {
		name      string
		setupData func() entities.CallRecordDataModel
		wantErr   string
	}{
		{
			name: "valid record",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "",
		},
		{
			name: "empty phone number",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "phone_number must not be empty",
		},
		{
			name: "invalid phone number format (wrong prefix)",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "1909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "phone_number must be in format 0909722021",
		},
		{
			name: "invalid phone number format (too short)",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "090972202",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "phone_number must be in format 0909722021",
		},
		{
			name: "invalid phone number format (too long)",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "09097220210",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "phone_number must be in format 0909722021",
		},
		{
			name: "negative amount",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
					Amount:          -10.5,
				}
			},
			wantErr: "amount must not be negative",
		},
		{
			name: "invalid appointment date format",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "16-06-2026",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "appointment_date must be in YYYY-MM-DD format",
		},
		{
			name: "empty due date",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nil,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "",
		},
		{
			name: "empty appointment date",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "",
		},
		{
			name: "empty user id",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "",
					WorkspaceID:     "workspace-1",
				}
			},
			wantErr: "user_id must not be empty",
		},
		{
			name: "empty workspace id",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "",
				}
			},
			wantErr: "workspace_id must not be empty",
		},
		{
			name: "negative call duration",
			setupData: func() entities.CallRecordDataModel {
				return entities.CallRecordDataModel{
					PhoneNumber:     "0909722021",
					AppointmentDate: "2026-06-16",
					DueDate:         nowPtr,
					UserID:          "user-1",
					WorkspaceID:     "workspace-1",
					CallDuration:    -5,
				}
			},
			wantErr: "call_duration must not be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCallRecordsRepository{
				InsertFunc: func(data entities.CallRecordDataModel) error {
					// Check default status behavior
					if tt.name == "valid record" {
						assert.Equal(t, entities.StatusPending, data.Status)
					}
					return nil
				},
			}
			sv := NewCallRecordsService(mockRepo)
			data := tt.setupData()
			data.TemplateID = &templateID

			err := sv.CreateCallRecord(data)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCallRecordOwnershipValidation(t *testing.T) {
	templateID := "tpl-123"
	dueTime := time.Now()
	existingRecord := &entities.CallRecordDataModel{
		ID:              "rec-123",
		UserID:          "owner-user",
		PhoneNumber:     "0909722021",
		AppointmentDate: "2026-06-16",
		DueDate:         &dueTime,
		WorkspaceID:     "ws-123",
		TemplateID:      &templateID,
	}

	mockRepo := &mockCallRecordsRepository{
		FindByIDFunc: func(id string) (*entities.CallRecordDataModel, error) {
			if id == "rec-123" {
				return existingRecord, nil
			}
			return nil, nil
		},
		UpdateFunc: func(id string, data entities.CallRecordDataModel) error {
			return nil
		},
		DeleteFunc: func(id string) error {
			return nil
		},
	}

	sv := NewCallRecordsService(mockRepo)

	// 1. Get record by correct user
	rec, err := sv.GetCallRecordByIDByUser("rec-123", "owner-user")
	assert.NoError(t, err)
	assert.NotNil(t, rec)
	assert.Equal(t, "owner-user", rec.UserID)

	// 2. Get record by incorrect user (unauthorized)
	rec, err = sv.GetCallRecordByIDByUser("rec-123", "unauthorized-user")
	assert.Error(t, err)
	assert.Nil(t, rec)
	assert.Contains(t, err.Error(), "unauthorized")

	// 3. Update record by correct user
	err = sv.UpdateCallRecordByUser("rec-123", "owner-user", *existingRecord)
	assert.NoError(t, err)

	// 4. Update record by incorrect user (unauthorized)
	err = sv.UpdateCallRecordByUser("rec-123", "unauthorized-user", *existingRecord)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")

	// 5. Delete record by correct user
	err = sv.DeleteCallRecordByUser("rec-123", "owner-user")
	assert.NoError(t, err)

	// 6. Delete record by incorrect user (unauthorized)
	err = sv.DeleteCallRecordByUser("rec-123", "unauthorized-user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestCallRecordsServiceFilter(t *testing.T) {
	mockRepo := &mockCallRecordsRepository{
		FindByFilterFunc: func(filter entities.CallRecordFilter) (*[]entities.CallRecordDataModel, error) {
			assert.Equal(t, "user-owner", filter.UserID)
			assert.Equal(t, "pending", string(filter.Status))
			assert.Equal(t, "botnoi-123", filter.BotnoiCallID)
			return &[]entities.CallRecordDataModel{
				{ID: "rec-1", UserID: "user-owner", Status: entities.StatusPending, BotnoiCallID: "botnoi-123"},
			}, nil
		},
	}

	sv := NewCallRecordsService(mockRepo)

	// 1. Test GetAllCallRecordsByUser success with filter
	results, err := sv.GetAllCallRecordsByUser("user-owner", entities.CallRecordFilter{
		Status:       "pending",
		BotnoiCallID: "botnoi-123",
	})
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Len(t, *results, 1)

	// 2. Test GetAllCallRecordsByUser mismatched filter user ID -> error
	_, err = sv.GetAllCallRecordsByUser("user-owner", entities.CallRecordFilter{
		UserID: "user-other",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized: cannot filter by other user ID")
}

