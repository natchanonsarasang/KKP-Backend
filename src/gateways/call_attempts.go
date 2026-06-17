package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) GetCallAttemptsByWorkspace(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	workspaceID := ctx.Params("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	limit := ctx.QueryInt("limit")

	filter := entities.CallAttemptFilter{
		WorkspaceID:    workspaceID,
		CallListItemID: ctx.Query("call_list_item_id"),
		Status:         ctx.Query("status"),
		Limit:          int64(limit),
	}

	data, err := h.CallAttemptService.GetAttemptsByFilterByUser(tokenData.UserID, filter)
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

	if bodyData.WorkspaceID == "" {
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

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
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

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.CallAttemptModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	if err := h.CallAttemptService.UpdateAttemptByUser(id, tokenData.UserID, workspaceID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteCallAttempt(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	if err := h.CallAttemptService.DeleteAttemptByUser(id, tokenData.UserID, workspaceID); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}



func (h *HTTPGateway) UpdateMultipleCallAttempts(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	filter := entities.CallAttemptFilter{
		WorkspaceID:    workspaceID,
		CallListItemID: ctx.Query("call_list_item_id"),
		Status:         ctx.Query("status"),
	}

	bodyData := entities.CallAttemptModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	modifiedCount, err := h.CallAttemptService.UpdateMultipleAttemptsByUser(tokenData.UserID, filter, bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":        "success",
		"modified_count": modifiedCount,
	})
}
