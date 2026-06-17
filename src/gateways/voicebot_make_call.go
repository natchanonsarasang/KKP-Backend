package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) MakeCall(ctx *fiber.Ctx) error {
	_, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(map[string]string{"message": "Unauthorized token"})
	}

	bodyData := entities.VoicebotMakeCallDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(map[string]string{"message": "invalid json body"})
	}

	// Call the service to make the call
	if err := h.VoicebotMakeCallService.MakeCall(bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(map[string]string{"message": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(map[string]string{"message": "success"})
}
