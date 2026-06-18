package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

type processCallSessionRequest struct {
	SessionID string `json:"session_id"`
	Action    string `json:"action"`
}

// ProcessCallSession handles start/continue/pause/stop actions
// (ported from the Supabase edge function "process-call-session").
func (h *HTTPGateway) ProcessCallSession(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	body := processCallSessionRequest{}
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}
	if body.SessionID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "session_id is required"})
	}

	switch body.Action {
	case "pause":
		if err := h.CallProcessService.PauseSession(body.SessionID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "Session paused"})

	case "stop":
		if err := h.CallProcessService.StopSession(body.SessionID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "Session stopped"})

	case "start", "continue":
		// Run in background so the HTTP response returns immediately
		// (mirrors EdgeRuntime.waitUntil in the original edge function).
		go func() {
			_ = h.CallProcessService.ProcessSession(body.SessionID)
		}()
		msg := "Processing started"
		if body.Action == "continue" {
			msg = "Processing continued"
		}
		return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: msg})

	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "Invalid action"})
	}
}