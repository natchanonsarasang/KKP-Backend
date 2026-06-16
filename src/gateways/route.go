package gateways

import "github.com/gofiber/fiber/v2"

func GatewayUsers(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/users")

	api.Post("/add_user", gateway.CreateUser)
	api.Get("/users", gateway.GetAllUserData)
}

func GatewayDebtors(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/debtors")

	api.Post("/", gateway.CreateDebtor)
	api.Get("/workspace/:workspace_id", gateway.GetDebtorsByWorkspace)
	api.Get("/:id", gateway.GetDebtorByID)
	api.Put("/:id", gateway.UpdateDebtor)
	api.Delete("/:id", gateway.DeleteDebtor)
}

func GatewayCallListItems(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-list-items")

	api.Post("/", gateway.CreateCallListItem)
	api.Get("/workspace/:workspace_id", gateway.GetCallListItemsByWorkspace)
	api.Get("/:id", gateway.GetCallListItemByID)
	api.Put("/:id", gateway.UpdateCallListItem)
	api.Delete("/:id", gateway.DeleteCallListItem)
}

func GatewayCallAttempts(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-attempts")

	api.Post("/", gateway.CreateCallAttempt)
	api.Get("/workspace/:workspace_id", gateway.GetCallAttemptsByWorkspace)
	api.Get("/:id", gateway.GetCallAttemptByID)
	api.Put("/:id", gateway.UpdateCallAttempt)
	api.Delete("/:id", gateway.DeleteCallAttempt)
}
