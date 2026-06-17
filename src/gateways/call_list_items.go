package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) GetCallListItemsByWorkspace(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	workspaceID := ctx.Params("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	data, err := h.CallListItemService.GetCallListItemsByWorkspaceByUser(tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateCallListItem(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	bodyData := entities.CallListItemModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.WorkspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id is required"})
	}

	if err := h.CallListItemService.CreateCallListItemByUser(tokenData.UserID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) GetCallListItemByID(ctx *fiber.Ctx) error {
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

	data, err := h.CallListItemService.GetCallListItemByIDByUser(id, tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateCallListItem(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.CallListItemModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	if err := h.CallListItemService.UpdateCallListItemByUser(id, tokenData.UserID, workspaceID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteCallListItem(ctx *fiber.Ctx) error {
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

	if err := h.CallListItemService.DeleteCallListItemByUser(id, tokenData.UserID, workspaceID); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
