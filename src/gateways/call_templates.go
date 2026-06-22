package gateways

import (
	"go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

type CallTemplatesGateway struct {
	Service services.ICallTemplatesService
}

func NewCallTemplatesGateway(svc services.ICallTemplatesService) *CallTemplatesGateway {
	return &CallTemplatesGateway{Service: svc}
}

func (g *CallTemplatesGateway) GetTemplates(c *fiber.Ctx) error {
	id := c.Query("id")
	templateID := c.Query("template_id")

	templates, err := g.Service.GetTemplatesByFilter(c.Context(), id, templateID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(templates)
}
