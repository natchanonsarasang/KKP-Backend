package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

// GatewayCallRecords registers the HTTP routes for Call Records and applies the JWT auth middleware to protect them.
func GatewayCallRecords(gateway HTTPGateway, app *fiber.App) {
	api := app.Group("/api/v1/call-records", middlewares.SetJWtHeaderHandler())

	api.Post("/", gateway.CreateCallRecord)
	api.Get("/:id", gateway.GetCallRecordByID)
	api.Get("/", gateway.GetAllCallRecords)
	api.Put("/:id", gateway.UpdateCallRecord)
	api.Delete("/:id", gateway.DeleteCallRecord)
}

func (h *HTTPGateway) CreateCallRecord(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	var bodyData entities.CallRecordDataModel
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	// Bind the call record to the authenticated user's ID
	bodyData.UserID = tokenDetails.UserID

	if err := h.CallRecordsService.CreateCallRecord(bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseModel{
			Message: "cannot create call record: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(entities.ResponseModel{
		Message: "success",
	})
}

func (h *HTTPGateway) GetCallRecordByID(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "id parameter is required",
		})
	}

	data, err := h.CallRecordsService.GetCallRecordByIDByUser(id, tokenDetails.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{
			Message: "cannot retrieve call record: " + err.Error(),
		})
	}
	if data == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseModel{
			Message: "call record not found",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    data,
	})
}

func (h *HTTPGateway) GetAllCallRecords(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	filter := entities.CallRecordFilter{
		Status:       ctx.Query("status"),
		WorkspaceID:  ctx.Query("workspace_id"),
		UserID:       ctx.Query("user_id"),
		BotnoiCallID: ctx.Query("botnoi_call_id"),
	}

	data, err := h.CallRecordsService.GetAllCallRecordsByUser(tokenDetails.UserID, filter)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{
			Message: "cannot retrieve call records: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
		Data:    data,
	})
}

func (h *HTTPGateway) UpdateCallRecord(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "id parameter is required",
		})
	}

	var bodyData entities.CallRecordDataModel
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	if err := h.CallRecordsService.UpdateCallRecordByUser(id, tokenDetails.UserID, bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseModel{
			Message: "cannot update call record: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
	})
}

func (h *HTTPGateway) DeleteCallRecord(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "id parameter is required",
		})
	}

	if err := h.CallRecordsService.DeleteCallRecordByUser(id, tokenDetails.UserID); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseModel{
			Message: "cannot delete call record: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{
		Message: "success",
	})
}
