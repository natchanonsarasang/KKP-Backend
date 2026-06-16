package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	UserService         service.IUsersService
	CallRecordsService  service.ICallRecordsService
	DebtorService       service.IDebtorsService
	CallListItemService service.ICallListItemsService
	CallAttemptService  service.ICallAttemptsService
	CallSessionService  service.ICallSessionsService
}

func NewHTTPGateway(
	app *fiber.App,
	users service.IUsersService, callRecords service.ICallRecordsService,
	debtors service.IDebtorsService,
	items service.ICallListItemsService,
	attempts service.ICallAttemptsService,
	sessions service.ICallSessionsService,
) {
	gateway := &HTTPGateway{
		UserService:         users,
		DebtorService:       debtors,
		CallListItemService: items,
		CallAttemptService:  attempts,
		CallSessionService:  sessions,
		CallRecordsService:  callRecords,
	}

	GatewayUsers(*gateway, app)
	GatewayDebtors(*gateway, app)
	GatewayCallListItems(*gateway, app)
	GatewayCallAttempts(*gateway, app)
	GatewayCallSessions(*gateway, app)
	GatewayCallRecords(*gateway, app)
}
