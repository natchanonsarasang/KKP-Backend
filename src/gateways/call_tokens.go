package gateways

import (
	"go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type CallTokensGateway struct {
	Service services.ICallTokensService
}

func NewCallTokensGateway(svc services.ICallTokensService) *CallTokensGateway {
	return &CallTokensGateway{Service: svc}
}

func (g *CallTokensGateway) GetTokens(c *fiber.Ctx) error {
	id := c.Query("id")
	userID := c.Query("user_id")

	tokens, err := g.Service.GetTokensByFilter(c.Context(), id, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tokens)
}
