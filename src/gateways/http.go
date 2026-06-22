package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	CallRecordsService      service.ICallRecordsService
	DebtorService           service.IDebtorsService
	CallListItemService     service.ICallListItemsService
	CallAttemptService      service.ICallAttemptsService
	CallSessionService      service.ICallSessionsService
	WorkspacesService       service.IWorkspacesService
	WebhookService          service.IWebhookService
	VoicebotMakeCallService service.IVoicebotMakeCallService
	CallProcessService  service.ICallProcessService
	UsersService            service.IUsersService
}

func NewHTTPGateway(
	app *fiber.App,
	workspaces service.IWorkspacesService,
	callRecords service.ICallRecordsService,
	debtors service.IDebtorsService,
	items service.ICallListItemsService,
	attempts service.ICallAttemptsService,
	sessions service.ICallSessionsService,
	webhook service.IWebhookService,
	voicebotMakeCall service.IVoicebotMakeCallService,
	callProcess service.ICallProcessService,
	users service.IUsersService,
) {
	gateway := &HTTPGateway{
		WorkspacesService:       workspaces,
		CallRecordsService:      callRecords,
		DebtorService:           debtors,
		CallListItemService:     items,
		CallAttemptService:      attempts,
		CallSessionService:      sessions,
		WebhookService:          webhook,
		VoicebotMakeCallService: voicebotMakeCall,
		CallProcessService:  callProcess,
		UsersService:            users,
	}

	GatewayAuth(*gateway, app)
	GatewayUsers(*gateway, app)
	GatewayWorkspaces(*gateway, app)
	GatewayCallRecords(*gateway, app)
	GatewayDebtors(*gateway, app)
	GatewayCallListItems(*gateway, app)
	GatewayCallAttempts(*gateway, app)
	GatewayCallSessions(*gateway, app)
	GatewayWebhooks(*gateway, app)
	GatewayVoicebotMakeCall(*gateway, app)
	GatewayProcessCallSession(*gateway, app)
}
