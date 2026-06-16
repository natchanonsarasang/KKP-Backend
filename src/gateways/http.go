package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	UserService         service.IUsersService
	DebtorService       service.IDebtorsService
	CallListItemService service.ICallListItemsService
	CallAttemptService  service.ICallAttemptsService
}

func NewHTTPGateway(
	app *fiber.App,
	users service.IUsersService,
	debtors service.IDebtorsService,
	items service.ICallListItemsService,
	attempts service.ICallAttemptsService,
) {
	gateway := &HTTPGateway{
		UserService:         users,
		DebtorService:       debtors,
		CallListItemService: items,
		CallAttemptService:  attempts,
	}

	GatewayUsers(*gateway, app)
	GatewayDebtors(*gateway, app)
	GatewayCallListItems(*gateway, app)
	GatewayCallAttempts(*gateway, app)
}
