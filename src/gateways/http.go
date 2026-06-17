package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	CallRecordsService  service.ICallRecordsService
	DebtorService       service.IDebtorsService
	CallListItemService service.ICallListItemsService
	CallAttemptService  service.ICallAttemptsService
	CallSessionService  service.ICallSessionsService
	WorkspacesService   service.IWorkspacesService
}

func NewHTTPGateway(
	app *fiber.App,
	workspaces service.IWorkspacesService,
	callRecords service.ICallRecordsService,
	debtors service.IDebtorsService,
	items service.ICallListItemsService,
	attempts service.ICallAttemptsService,
	sessions service.ICallSessionsService,
) {
	gateway := &HTTPGateway{
		WorkspacesService:   workspaces,
		CallRecordsService:  callRecords,
		DebtorService:       debtors,
		CallListItemService: items,
		CallAttemptService:  attempts,
		CallSessionService:  sessions,
	}

	GatewayWorkspaces(*gateway, app)
	GatewayCallRecords(*gateway, app)
	GatewayDebtors(*gateway, app)
	GatewayCallListItems(*gateway, app)
	GatewayCallAttempts(*gateway, app)
	GatewayCallSessions(*gateway, app)
}
