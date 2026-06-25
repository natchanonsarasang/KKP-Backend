package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) GetCallTemplates(c *fiber.Ctx) error {
	id := c.Query("id")
	templateID := c.Query("template_id")

	templates, err := h.CallTemplatesService.GetTemplatesByFilter(c.Context(), id, templateID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(templates)
}

func (h *HTTPGateway) CreateCallTemplate(c *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	var bodyData entities.CallTemplateDataModel
	if err := c.BodyParser(&bodyData); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	// Bind the template to the authenticated user's ID
	bodyData.UserID = tokenDetails.UserID

	if err := h.CallTemplatesService.CreateCallTemplate(c.Context(), &bodyData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "cannot create call template: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(entities.ResponseMessage{
		Message: "success",
	})
}

func (h *HTTPGateway) UpdateCallTemplate(c *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "id parameter is required",
		})
	}

	var bodyData entities.CallTemplateDataModel
	if err := c.BodyParser(&bodyData); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	// Ensure user can only update their own templates
	bodyData.UserID = tokenDetails.UserID

	if err := h.CallTemplatesService.UpdateCallTemplate(c.Context(), id, &bodyData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "cannot update call template: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "success",
	})
}

func (h *HTTPGateway) DeleteCallTemplate(c *fiber.Ctx) error {
	_, err := middlewares.DecodeJWTToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "id parameter is required",
		})
	}

	if err := h.CallTemplatesService.DeleteCallTemplate(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "cannot delete call template: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "success",
	})
}
