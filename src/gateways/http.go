package gateways

import (
	service "go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type HTTPGateway struct {
	UserService             service.IUsersService
	WorkspaceMembersService service.IWorkspaceMembersService
	WorkspacesService       service.IWorkspacesService
}

func NewHTTPGateway(app *fiber.App, users service.IUsersService, workspaceMembers service.IWorkspaceMembersService, workspaces service.IWorkspacesService) {
	gateway := &HTTPGateway{
		UserService:             users,
		WorkspaceMembersService: workspaceMembers,
		WorkspacesService:       workspaces,
	}

	GatewayUsers(*gateway, app)
	GatewayWorkspaceMembers(*gateway, app)
	GatewayWorkspaces(*gateway, app)
}