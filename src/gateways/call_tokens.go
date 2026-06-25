package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

func (h *HTTPGateway) GetTokens(c *fiber.Ctx) error {
	id := c.Query("id")
	userID := c.Query("user_id")

	tokens, err := h.CallTokensService.GetTokensByFilter(c.Context(), id, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tokens)
}

func (h *HTTPGateway) CreateCallToken(c *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{
			Message: "Unauthorized: " + err.Error(),
		})
	}

	var bodyData entities.CallTokenDataModel
	if err := c.BodyParser(&bodyData); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	// Bind the token to the authenticated user's ID
	bodyData.UserID = tokenDetails.UserID

	if err := h.CallTokensService.CreateCallToken(c.Context(), &bodyData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "cannot create call token: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(entities.ResponseMessage{
		Message: "success",
	})
}

func (h *HTTPGateway) UpdateCallToken(c *fiber.Ctx) error {
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

	var bodyData entities.CallTokenDataModel
	if err := c.BodyParser(&bodyData); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{
			Message: "invalid json body: " + err.Error(),
		})
	}

	// Ensure user can only update their own tokens
	bodyData.UserID = tokenDetails.UserID

	if err := h.CallTokensService.UpdateCallToken(c.Context(), id, &bodyData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "cannot update call token: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "success",
	})
}

func (h *HTTPGateway) DeleteCallToken(c *fiber.Ctx) error {
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

	if err := h.CallTokensService.DeleteCallToken(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{
			Message: "cannot delete call token: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(entities.ResponseMessage{
		Message: "success",
	})
}
