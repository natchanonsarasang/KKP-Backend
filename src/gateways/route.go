package gateways

import (
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func GatewayHealth(gateway HTTPGateway, app *fiber.App) {
	// Liveness/health endpoint for platform health checks (e.g. Render).
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "callecto-api"})
	})
}

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

func GatewayCallStats(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-stats", middlewares.SetJWtHeaderHandler())

	api.Get("/by-debtor", gateway.GetCallStatsByDebtor)
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

func GatewayAuth(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/auth")

	api.Post("/register", gateway.Register)
	api.Post("/login", gateway.Login)
	api.Post("/google", gateway.GoogleSignIn)
	api.Post("/microsoft", gateway.MicrosoftSignIn)
}

func GatewayUsers(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/users", middlewares.SetJWtHeaderHandler())

	api.Get("/me", gateway.GetMe)
	api.Post("/", gateway.CreateUser)
	api.Get("/", gateway.GetUsers)
	api.Get("/:id", gateway.GetUserByID)
	api.Put("/:id", gateway.UpdateUser)
	api.Delete("/:id", gateway.DeleteUser)
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

func GatewayCallTemplates(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-templates", middlewares.SetJWtHeaderHandler())
	api.Post("/", gateway.CreateCallTemplate)
	api.Get("/", gateway.GetCallTemplates)
	api.Put("/:id", gateway.UpdateCallTemplate)
	api.Delete("/:id", gateway.DeleteCallTemplate)
}

func GatewayCallTokens(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-tokens", middlewares.SetJWtHeaderHandler())
	api.Post("/", gateway.CreateCallToken)
	api.Get("/", gateway.GetTokens)
	api.Put("/:id", gateway.UpdateCallToken)
	api.Delete("/:id", gateway.DeleteCallToken)
}

func GatewayAudioProxy(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/audio-proxy", middlewares.SetJWtHeaderHandler())
	api.Get("/", gateway.AudioProxy)
}
