package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	UserService                     service.IUsersService
	CallRecordsService  service.ICallRecordsService
	DebtorService       service.IDebtorsService
	CallListItemService service.ICallListItemsService
	CallAttemptService  service.ICallAttemptsService
	CallSessionService  service.ICallSessionsService
	WorkspaceMembersService service.IWorkspaceMembersService
	WorkspacesService       service.IWorkspacesService
}

func NewHTTPGateway(
	app *fiber.App,
	users service.IUsersService, workspaceMembers service.IWorkspaceMembersService, workspaces service.IWorkspacesService, callRecords service.ICallRecordsService,
	debtors service.IDebtorsService,
	items service.ICallListItemsService,
	attempts service.ICallAttemptsService,
	sessions service.ICallSessionsService,
) {
	gateway := &HTTPGateway{
		UserService: users,
	}

	GatewayUsers(*gateway, app)
}