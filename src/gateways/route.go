package gateways

import (
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func GatewayDebtors(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/debtors", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateDebtor)
	api.Get("/workspace/:workspace_id", gateway.GetDebtorsByWorkspace)
	api.Get("/:id", gateway.GetDebtorByID)
	api.Put("/:id", gateway.UpdateDebtor)
	api.Delete("/:id", gateway.DeleteDebtor)
}

func GatewayCallListItems(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-list-items", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateCallListItem)
	api.Get("/workspace/:workspace_id", gateway.GetCallListItemsByWorkspace)
	api.Get("/:id", gateway.GetCallListItemByID)
	api.Put("/:id", gateway.UpdateCallListItem)
	api.Delete("/:id", gateway.DeleteCallListItem)
}

func GatewayCallAttempts(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-attempts", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateCallAttempt)
	api.Get("/workspace/:workspace_id", gateway.GetCallAttemptsByWorkspace)
	api.Get("/:id", gateway.GetCallAttemptByID)
	api.Put("/", gateway.UpdateMultipleCallAttempts)
	api.Put("/:id", gateway.UpdateCallAttempt)
	api.Delete("/:id", gateway.DeleteCallAttempt)
}

func GatewayCallSessions(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-sessions", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateCallSession)
	api.Get("/", gateway.GetCallSessions)
	api.Get("/:id", gateway.GetCallSessionByID)
	api.Put("/:id", gateway.UpdateCallSession)
	api.Delete("/:id", gateway.DeleteCallSession)
}

func GatewayCallRecords(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-records", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateCallRecord)
	api.Get("/:id", gateway.GetCallRecordByID)
	api.Get("/", gateway.GetAllCallRecords)
	api.Put("/:id", gateway.UpdateCallRecord)
	api.Delete("/:id", gateway.DeleteCallRecord)
}

func GatewayWorkspaces(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/workspaces", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateWorkspace)
	api.Get("/", gateway.GetWorkspaces)
	api.Get("/:id", gateway.GetWorkspaceByID)
	api.Put("/:id", gateway.UpdateWorkspace)
	api.Delete("/:id", gateway.DeleteWorkspace)
}

func GatewayWebhooks(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/webhooks")

	api.Post("/botnoi", gateway.Webhook)
}

func GatewayVoicebotMakeCall(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/voicebot/make-call", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.MakeCall)
}


func GatewayProcessCallSession(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-process", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.ProcessCallSession)
}