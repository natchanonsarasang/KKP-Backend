package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *HTTPGateway) GetCallAttemptsByWorkspace(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	workspaceIDStr := ctx.Params("workspace_id")
	workspaceID, err := primitive.ObjectIDFromHex(workspaceIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	data, err := h.CallAttemptService.GetAttemptsByWorkspaceByUser(tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateCallAttempt(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	bodyData := entities.CallAttemptModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.WorkspaceID.IsZero() {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id is required"})
	}

	if err := h.CallAttemptService.CreateAttemptByUser(tokenData.UserID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) GetCallAttemptByID(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	workspaceIDStr := ctx.Query("workspace_id")
	if workspaceIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	workspaceID, err := primitive.ObjectIDFromHex(workspaceIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	data, err := h.CallAttemptService.GetAttemptByIDByUser(id, tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateCallAttempt(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.CallAttemptModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.WorkspaceID.IsZero() {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id is required"})
	}

	if err := h.CallAttemptService.UpdateAttemptByUser(id, tokenData.UserID, bodyData.WorkspaceID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteCallAttempt(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	workspaceIDStr := ctx.Query("workspace_id")
	if workspaceIDStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	workspaceID, err := primitive.ObjectIDFromHex(workspaceIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	if err := h.CallAttemptService.DeleteAttemptByUser(id, tokenData.UserID, workspaceID); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
