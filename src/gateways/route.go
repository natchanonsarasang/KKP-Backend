package gateways

import (
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func GatewayUsers(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/users")

	api.Post("/add_user", gateway.CreateUser)
	api.Get("/users", gateway.GetAllUserData)
}

func GatewayWorkspaceMembers(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/workspace-members")

	api.Post("/", gateway.CreateWorkspaceMember)
	api.Get("/", gateway.GetAllWorkspaceMembers)
	api.Get("/:id", gateway.GetWorkspaceMemberByID)
	api.Put("/:id", gateway.UpdateWorkspaceMember)
	api.Delete("/:id", gateway.DeleteWorkspaceMember)
}

func GatewayWorkspaces(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/workspaces", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateWorkspace)
	api.Get("/", gateway.GetWorkspaces)
	api.Get("/:id", gateway.GetWorkspaceByID)
	api.Put("/:id", gateway.UpdateWorkspace)
	api.Delete("/:id", gateway.DeleteWorkspace)
}