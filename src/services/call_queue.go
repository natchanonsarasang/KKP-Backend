package services

import (
	"strconv"
	"strings"
	"time"

	"go-fiber-template/domain/entities"
	"go-fiber-template/domain/repositories"

	"github.com/google/uuid"
)

type callQueueService struct {
	CallQueueRepository repositories.ICallQueueRepository
}

type ICallQueueService interface {
	// Enqueue snapshots the debtor's variables and pushes a queue row for a call.
	Enqueue(session *entities.CallSessionDataModel, item entities.CallListItemModel, debtor entities.DebtorModel, outboundID string) error
	// Head returns the variables of the current head-of-queue call, or nil when
	// the queue is empty.
	Head() (map[string]string, error)
	DeleteByOutboundID(outboundID string) error
	DeleteByCallListItemID(callListItemID string) error
}

func NewCallQueueService(repo repositories.ICallQueueRepository) ICallQueueService {
	return &callQueueService{CallQueueRepository: repo}
}

func (sv *callQueueService) Enqueue(
	session *entities.CallSessionDataModel,
	item entities.CallListItemModel,
	debtor entities.DebtorModel,
	outboundID string,
) error {
	row := entities.CallQueueModel{
		ID:             uuid.NewString(),
		OutboundID:     outboundID,
		CallListItemID: item.ID,
		DebtorID:       debtor.ID,
		SessionID:      session.ID,
		WorkspaceID:    session.WorkspaceID,
		UserID:         session.UserID,
		PhoneNumber:    debtor.PhoneNumber,
		Variables:      buildKKPVariables(debtor),
		CreatedAt:      time.Now().UTC(),
	}
	return sv.CallQueueRepository.Insert(row)
}

func (sv *callQueueService) Head() (map[string]string, error) {
	row, err := sv.CallQueueRepository.FindHead()
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return row.Variables, nil
}

func (sv *callQueueService) DeleteByOutboundID(outboundID string) error {
	return sv.CallQueueRepository.DeleteByOutboundID(outboundID)
}

func (sv *callQueueService) DeleteByCallListItemID(callListItemID string) error {
	return sv.CallQueueRepository.DeleteByCallListItemID(callListItemID)
}

// buildKKPVariables produces the flat variable map served to the V2 voicebot's
// KKP_Data tool. Values are kept RAW (numbers as digits, plate/province split)
// because the agent prompt formats them itself (reads amounts as integers, plate
// digit-by-digit, dates in Buddhist era). Any extra debtor variable keys not
// explicitly mapped here are passed through unchanged.
func buildKKPVariables(debtor entities.DebtorModel) map[string]string {
	vars := map[string]string{}
	for k, v := range debtor.Variables {
		vars[k] = v
	}

	// customer_name: prefer variables["name"], fall back to the debtor column.
	vars["customer_name"] = DebtorDisplayName(&debtor)

	// car_detail is uploaded as one combined "plate + province" string; the
	// prompt needs them as two variables.
	if carDetail, ok := vars["car_detail"]; ok {
		plate, province := splitCarDetail(carDetail)
		vars["car_detail"] = plate
		vars["province"] = province
	}

	// total_debt: when no total_debt variable was uploaded, fall back to the same
	// resolution the rest of the app uses (variables amount/outstanding_amount,
	// then the debtor's total_debt column).
	if strings.TrimSpace(vars["total_debt"]) == "" {
		if amount := DebtorDisplayAmount(&debtor); amount > 0 {
			vars["total_debt"] = strconv.FormatFloat(amount, 'f', -1, 64)
		}
	}

	// current_date is the reference date (Asia/Bangkok) the agent uses for all
	// relative-date math; ISO YYYY-MM-DD, Gregorian, as the prompt requires.
	loc, err := time.LoadLocation("Asia/Bangkok")
	now := time.Now()
	if err == nil {
		now = now.In(loc)
	}
	vars["current_date"] = now.Format("2006-01-02")

	return vars
}
