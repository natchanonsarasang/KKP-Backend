package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

// GetCallStatsByDebtor returns per-debtor call summaries (total / picked_up /
// not_picked_up / confirmed / declined / no_response) for a workspace, computed
// server-side from call_records. The debtor list uses this instead of pulling
// every call_record to the browser and counting client-side.
func (h *HTTPGateway) GetCallStatsByDebtor(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return err
	}

	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		workspaceID = ctx.Params("workspace_id")
	}
	if workspaceID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "workspace_id is required"})
	}

	data, err := h.CallListItemService.GetCallStatsByDebtor(tokenData.UserID, workspaceID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}
