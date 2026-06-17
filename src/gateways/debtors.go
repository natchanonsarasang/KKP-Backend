package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) GetDebtorsByWorkspace(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	workspaceID := ctx.Params("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	data, err := h.DebtorService.GetDebtorsByWorkspaceByUser(tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateDebtor(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	bodyData := entities.DebtorModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.WorkspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id is required"})
	}

	if err := h.DebtorService.CreateDebtorByUser(tokenData.UserID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) GetDebtorByID(ctx *fiber.Ctx) error {
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

	data, err := h.DebtorService.GetDebtorByIDByUser(id, tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateDebtor(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.DebtorModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id query parameter is required"})
	}

	if err := h.DebtorService.UpdateDebtorByUser(id, tokenData.UserID, workspaceID, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteDebtor(ctx *fiber.Ctx) error {
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

	if err := h.DebtorService.DeleteDebtorByUser(id, tokenData.UserID, workspaceID); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
