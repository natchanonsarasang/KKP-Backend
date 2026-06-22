package gateways

import (
	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"

	"github.com/gofiber/fiber/v2"
)

// Register handles POST /api/v1/auth/register (public).
// It creates an email/password account and returns an application JWT.
func (h *HTTPGateway) Register(ctx *fiber.Ctx) error {
	bodyData := entities.SignUpRequest{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	auth, err := h.UsersService.Register(bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: auth})
}

// Login handles POST /api/v1/auth/login (public).
// It verifies email/password credentials and returns an application JWT.
func (h *HTTPGateway) Login(ctx *fiber.Ctx) error {
	bodyData := entities.SignInRequest{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	auth, err := h.UsersService.Login(bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: auth})
}

// GoogleSignIn handles POST /api/v1/auth/google (public).
// It verifies the Google ID token, provisions/links the user, and returns an
// application JWT plus the user profile.
func (h *HTTPGateway) GoogleSignIn(ctx *fiber.Ctx) error {
	bodyData := entities.GoogleSignInRequest{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if bodyData.IDToken == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "id_token is required"})
	}

	auth, err := h.UsersService.GoogleSignIn(bodyData.IDToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: auth})
}

// GetMe handles GET /api/v1/users/me — the authenticated user's own profile.
func (h *HTTPGateway) GetMe(ctx *fiber.Ctx) error {
	tokenDetails, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	data, err := h.UsersService.GetUserByID(tokenDetails.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	if data == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "user not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) CreateUser(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	bodyData := entities.UserDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	created, err := h.UsersService.CreateUser(bodyData)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: created})
}

func (h *HTTPGateway) GetUsers(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	filter := entities.UserFilter{
		Email:    ctx.Query("email"),
		Name:     ctx.Query("name"),
		Provider: ctx.Query("provider"),
	}

	data, err := h.UsersService.GetAllUsers(filter)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) GetUserByID(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	data, err := h.UsersService.GetUserByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}
	if data == nil {
		return ctx.Status(fiber.StatusNotFound).JSON(entities.ResponseMessage{Message: "user not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: data})
}

func (h *HTTPGateway) UpdateUser(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	bodyData := entities.UserDataModel{}
	if err := ctx.BodyParser(&bodyData); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(entities.ResponseMessage{Message: "invalid json body"})
	}

	if err := h.UsersService.UpdateUser(id, bodyData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}

func (h *HTTPGateway) DeleteUser(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "invalid id"})
	}

	if err := h.UsersService.DeleteUser(id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success"})
}
