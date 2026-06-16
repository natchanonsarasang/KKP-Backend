package gateways

import (
	"go-fiber-template/domain/entities"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) GetAllWorkspaceMembers(ctx *fiber.Ctx) error {

	data, err := h.WorkspaceMembersService.GetAllWorkspaceMembers()
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot get all workspace members data"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) GetWorkspaceMemberByID(ctx *fiber.Ctx) error {

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseModel{Message: "id is required"})
	}

	data, err := h.WorkspaceMembersService.GetWorkspaceMemberByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot get workspace member data"})
	}
	if data == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseModel{Message: "workspace member not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateWorkspaceMember(ctx *fiber.Ctx) error {

	bodyData := entities.WorkspaceMemberDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseModel{Message: "invalid json body"})
	}

	if bodyData.WorkspaceID == "" || bodyData.UserID == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseModel{Message: "invalid json body"})
	}

	if err := h.WorkspaceMembersService.InsertNewWorkspaceMember(bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot insert new workspace member."})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) UpdateWorkspaceMember(ctx *fiber.Ctx) error {

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseModel{Message: "id is required"})
	}

	bodyData := entities.WorkspaceMemberDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseModel{Message: "invalid json body"})
	}

	if err := h.WorkspaceMembersService.UpdateWorkspaceMember(id, bodyData); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot update workspace member."})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteWorkspaceMember(ctx *fiber.Ctx) error {

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseModel{Message: "id is required"})
	}

	if err := h.WorkspaceMembersService.DeleteWorkspaceMember(id); err != nil {
		return ctx.Status(fiber.StatusForbidden).JSON(entities.ResponseModel{Message: "cannot delete workspace member."})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}