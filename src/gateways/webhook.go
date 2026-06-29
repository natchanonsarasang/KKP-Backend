package gateways

import (
	"go-fiber-template/domain/entities"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func (h *HTTPGateway) Webhook(ctx *fiber.Ctx) error {
	// Log the raw inbound body first so we capture exactly what Botnoi sent,
	// including any fields not mapped onto WebhookPayload.
	log.Infof("[Webhook] inbound raw body: %s", string(ctx.Body()))

	var payload entities.WebhookPayload
	if err := ctx.BodyParser(&payload); err != nil {
		log.Errorf("Webhook Error: Failed to parse body: %v", err)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	// CRITICAL FIX: Ignore the initiation "Success" message
	if payload.Message != "" && strings.Contains(payload.Message, "Success Create Outbound call") {
		log.Info("Ignoring initiation acknowledgement message in webhook.")
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Initiation acknowledgement ignored",
		})
	}

	// Delegate processing to WebhookService
	if err := h.WebhookService.ProcessWebhook(payload); err != nil {
		log.Errorf("Webhook Service Error: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{
			Message: "failed to process webhook: " + err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}
