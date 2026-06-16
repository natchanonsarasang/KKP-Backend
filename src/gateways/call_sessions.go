package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) CreateCallSession(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	bodyData := entities.CallSessionDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.CallSessionService.CreateCallSessionByUser(tokenDetails.UserID, bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) GetCallSessions(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	filter := entities.CallSessionFilter{
		Status:      ctx.Query("status"),
		WorkspaceID: ctx.Query("workspace_id"),
		UserID:      ctx.Query("user_id"),
	}

	data, err := h.CallSessionService.GetCallSessionsByUser(tokenDetails.UserID, filter)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) GetCallSessionByID(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	data, err := h.CallSessionService.GetCallSessionByIDByUser(tokenDetails.UserID, id)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	if data == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "session not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateCallSession(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.CallSessionDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.CallSessionService.UpdateCallSessionByUser(tokenDetails.UserID, id, bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteCallSession(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	if err := h.CallSessionService.DeleteCallSessionByUser(tokenDetails.UserID, id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
