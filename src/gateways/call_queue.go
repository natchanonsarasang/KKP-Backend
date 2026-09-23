package gateways

import (
	"go-fiber-template/domain/entities"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// KKPData serves the head-of-queue debtor's variables to the V2 Botnoi voicebot's
// "Tools KKP_Data" call. That inbound call carries no identifier, so we rely on
// the queue: only one call is live at a time, and the head row is that debtor.
// The response is a flat JSON object of variable -> value.
func (h *HTTPGateway) KKPData(ctx *fiber.Ctx) error {
	vars, err := h.CallQueueService.Head()
	if err != nil {
		log.Errorf("[KKPData] failed to read queue head: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "failed to read call queue: " + err.Error(),
		})
	}
	if vars == nil {
		log.Warnf("[KKPData] queue is empty — no active call to serve")
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{
			Message: "no active call in queue",
		})
	}

	log.Infof("[KKPData] serving head-of-queue variables (%d fields)", len(vars))
	return ctx.Status(fiber.StatusOK).JSON(vars)
}
