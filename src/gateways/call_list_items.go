package gateways

import (
	"go-fiber-template/domain/entities"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *HTTPGateway) GetCallListItemsByWorkspace(ctx *fiber.Ctx) error {
	workspaceIDStr := ctx.Params("workspace_id")
	workspaceID, err := primitive.ObjectIDFromHex(workspaceIDStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid workspace id"})
	}

	data, err := h.CallListItemService.GetItemsByWorkspace(workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot get call list items"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateCallListItem(ctx *fiber.Ctx) error {
	bodyData := entities.CallListItemModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.CallListItemService.CreateItem(bodyData); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot insert call list item"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) GetCallListItemByID(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	data, err := h.CallListItemService.GetItemByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseModel{Message: "item not found"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateCallListItem(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.CallListItemModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.CallListItemService.UpdateItem(id, bodyData); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot update call list item"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteCallListItem(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	if err := h.CallListItemService.DeleteItem(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseModel{Message: "cannot delete call list item"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
