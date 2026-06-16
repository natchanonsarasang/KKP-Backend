package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	UserService        service.IUsersService
	CallRecordsService service.ICallRecordsService
}

func NewHTTPGateway(app *fiber.App, users service.IUsersService, callRecords service.ICallRecordsService) {
	gateway := &HTTPGateway{
		UserService:        users,
		CallRecordsService: callRecords,
	}

	GatewayUsers(*gateway, app)
	GatewayCallRecords(*gateway, app)
}
