package services

import (
	"errors"
	"regexp"
	"time"

	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"

	"github.com/google/uuid"
)

type callRecordsService struct {
	Repo repositories.ICallRecordsRepository
}

type ICallRecordsService interface {
	CreateCallRecord(data entities.CallRecordDataModel) error
	GetCallRecordByIDByUser(id string, userID string) (*entities.CallRecordDataModel, error)
	GetAllCallRecordsByUser(userID string) (*[]entities.CallRecordDataModel, error)
	UpdateCallRecordByUser(id string, userID string, data entities.CallRecordDataModel) error
	DeleteCallRecordByUser(id string, userID string) error

	// Direct/System CRUD methods (e.g. for voicebot webhook)
	GetCallRecordByID(id string) (*entities.CallRecordDataModel, error)
	GetAllCallRecords() (*[]entities.CallRecordDataModel, error)
	UpdateCallRecord(id string, data entities.CallRecordDataModel) error
	DeleteCallRecord(id string) error
}

func NewCallRecordsService(repo repositories.ICallRecordsRepository) ICallRecordsService {
	return &callRecordsService{
		Repo: repo,
	}
}

func (sv *callRecordsService) CreateCallRecord(data entities.CallRecordDataModel) error {
	// Initialize default values for creation if not set
	if data.ID == "" {
		data.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	data.CreatedAt = now
	data.UpdatedAt = now

	// Run business logic validations
	if err := sv.validateCallRecord(&data); err != nil {
		return err
	}

	return sv.Repo.InsertCallRecord(data)
}

func (sv *callRecordsService) GetCallRecordByID(id string) (*entities.CallRecordDataModel, error) {
	if id == "" {
		return nil, errors.New("id must not be empty")
	}
	return sv.Repo.FindByID(id)
}

func (sv *callRecordsService) GetAllCallRecords() (*[]entities.CallRecordDataModel, error) {
	return sv.Repo.FindAll()
}

func (sv *callRecordsService) UpdateCallRecord(id string, data entities.CallRecordDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	// Fetch existing record to ensure it exists
	existing, err := sv.Repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("call record not found")
	}

	// Update timestamp
	data.UpdatedAt = time.Now().UTC()
	data.ID = id // Ensure ID cannot be changed

	// Validate the updated record
	if err := sv.validateCallRecord(&data); err != nil {
		return err
	}

	return sv.Repo.UpdateCallRecord(id, data)
}

func (sv *callRecordsService) DeleteCallRecord(id string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}

	// Fetch existing record to ensure it exists
	existing, err := sv.Repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("call record not found")
	}

	return sv.Repo.DeleteCallRecord(id)
}

func (sv *callRecordsService) GetCallRecordByIDByUser(id string, userID string) (*entities.CallRecordDataModel, error) {
	if id == "" {
		return nil, errors.New("id must not be empty")
	}
	if userID == "" {
		return nil, errors.New("unauthorized: missing user id")
	}

	record, err := sv.Repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	if record.UserID != userID {
		return nil, errors.New("unauthorized: you do not own this call record")
	}
	return record, nil
}

func (sv *callRecordsService) GetAllCallRecordsByUser(userID string) (*[]entities.CallRecordDataModel, error) {
	if userID == "" {
		return nil, errors.New("unauthorized: missing user id")
	}
	return sv.Repo.FindByUserID(userID)
}

func (sv *callRecordsService) UpdateCallRecordByUser(id string, userID string, data entities.CallRecordDataModel) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if userID == "" {
		return errors.New("unauthorized: missing user id")
	}

	// Fetch existing record to verify ownership
	existing, err := sv.Repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("call record not found")
	}
	if existing.UserID != userID {
		return errors.New("unauthorized: you do not own this call record")
	}

	// Ensure ID and UserID cannot be changed
	data.ID = id
	data.UserID = userID
	data.UpdatedAt = time.Now().UTC()

	// Validate the updated record
	if err := sv.validateCallRecord(&data); err != nil {
		return err
	}

	return sv.Repo.UpdateCallRecord(id, data)
}

func (sv *callRecordsService) DeleteCallRecordByUser(id string, userID string) error {
	if id == "" {
		return errors.New("id must not be empty")
	}
	if userID == "" {
		return errors.New("unauthorized: missing user id")
	}

	// Fetch existing record to verify ownership
	existing, err := sv.Repo.FindByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("call record not found")
	}
	if existing.UserID != userID {
		return errors.New("unauthorized: you do not own this call record")
	}

	return sv.Repo.DeleteCallRecord(id)
}

// validateCallRecord runs all business logic validations on a CallRecordDataModel
func (sv *callRecordsService) validateCallRecord(data *entities.CallRecordDataModel) error {
	// 1. Phone number must not be empty and must match format 0909722021 (10 digits starting with 0)
	if data.PhoneNumber == "" {
		return errors.New("phone_number must not be empty")
	}
	matched, _ := regexp.MatchString(`^0[0-9]{9}$`, data.PhoneNumber)
	if !matched {
		return errors.New("phone_number must be in format 0909722021 (10 digits starting with 0)")
	}

	// 2. Status empty -> pending
	if data.Status == "" {
		data.Status = entities.StatusPending
	} else {
		// Validate enum values
		switch data.Status {
		case entities.StatusConfirmed, entities.StatusDeclined, entities.StatusNoResponse,
			entities.StatusNoAnswer, entities.StatusHangedUp, entities.StatusPending,
			entities.StatusCompleted, entities.StatusBusy, entities.StatusFailed,
			entities.StatusRejected, entities.StatusVoicemail:
			// Valid
		default:
			return errors.New("invalid status value")
		}
	}

	// 3. Amount must not be negative
	if data.Amount < 0 {
		return errors.New("amount must not be negative")
	}

	// 4. Appointment date format validation (YYYY-MM-DD) - empty is allowed
	if data.AppointmentDate != "" {
		if _, err := time.Parse("2006-01-02", data.AppointmentDate); err != nil {
			return errors.New("appointment_date must be in YYYY-MM-DD format")
		}
	}

	// 5. User ID and Workspace ID must not be empty
	if data.UserID == "" {
		return errors.New("user_id must not be empty")
	}
	if data.WorkspaceID == "" {
		return errors.New("workspace_id must not be empty")
	}

	// 6. Call duration must not be negative
	if data.CallDuration < 0 {
		return errors.New("call_duration must not be negative")
	}

	return nil
}
