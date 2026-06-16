package gateways

import (
	"go-fiber-template/domain/entities"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *HTTPGateway) GetDebtorsByWorkspace(ctx *fiber.Ctx) error {
	workspaceIDStr := ctx.Params("workspace_id")
	workspaceID, err := primitive.ObjectIDFromHex(workspaceIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	data, err := h.DebtorService.GetDebtorsByWorkspace(workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot get debtors"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateDebtor(ctx *fiber.Ctx) error {
	bodyData := entities.DebtorModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.DebtorService.CreateDebtor(bodyData); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot insert debtor"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) GetDebtorByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	data, err := h.DebtorService.GetDebtorByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseModel{Message: "debtor not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateDebtor(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.DebtorModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.DebtorService.UpdateDebtor(id, bodyData); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot update debtor"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteDebtor(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	if err := h.DebtorService.DeleteDebtor(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot delete debtor"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
