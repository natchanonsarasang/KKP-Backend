package gateways

import (
	"encoding/json"
	"errors"
	"strconv"

	"go-fiber-template/domain/entities"
	"go-fiber-template/src/client"
	"go-fiber-template/src/middlewares"
	"go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

// ListBotnoiLogFiles handles GET /api/v1/botnoi-logs/files?start_date=&end_date=
// (dates as YYYY/MM/DD). The Botnoi response is passed through as `data`.
func (h *HTTPGateway) ListBotnoiLogFiles(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	body, err := h.BotnoiLogsService.ListFiles(ctx.Query("start_date"), ctx.Query("end_date"))
	if err != nil {
		return botnoiLogsError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: rawJSONOrString(body)})
}

// ReadBotnoiLog handles GET /api/v1/botnoi-logs/log?file_path=
func (h *HTTPGateway) ReadBotnoiLog(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	body, err := h.BotnoiLogsService.ReadLog(ctx.Query("file_path"))
	if err != nil {
		return botnoiLogsError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: rawJSONOrString(body)})
}

// GetBotnoiAudio handles GET /api/v1/botnoi-logs/audio?file_path= and streams
// the recording back (audio/wav).
func (h *HTTPGateway) GetBotnoiAudio(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	body, contentType, contentDisposition, err := h.BotnoiLogsService.GetAudio(ctx.Query("file_path"))
	if err != nil {
		return botnoiLogsError(ctx, err)
	}

	if contentType == "" {
		contentType = "audio/wav"
	}
	ctx.Set(fiber.HeaderContentType, contentType)
	ctx.Set(fiber.HeaderContentLength, strconv.Itoa(len(body)))
	ctx.Set(fiber.HeaderCacheControl, "private, max-age=3600")
	if contentDisposition != "" {
		ctx.Set(fiber.HeaderContentDisposition, contentDisposition)
	}
	return ctx.Status(fiber.StatusOK).Send(body)
}

func botnoiLogsError(ctx *fiber.Ctx, err error) error {
	var upstream *client.BotnoiUpstreamError
	switch {
	case errors.Is(err, services.ErrBotnoiInvalidFilePath):
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: err.Error()})
	case errors.Is(err, services.ErrBotnoiAgentNotConfigured):
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: err.Error()})
	case errors.As(err, &upstream):
		return ctx.Status(fiber.StatusBadGateway).JSON(entities.ResponseModel{
			Message: "Botnoi API error",
			Status:  upstream.StatusCode,
		})
	default:
		return ctx.Status(fiber.StatusBadGateway).JSON(entities.ResponseMessage{Message: "Botnoi API request failed"})
	}
}

// rawJSONOrString embeds a JSON body as-is, or falls back to a string when the
// upstream answered with something that is not JSON.
func rawJSONOrString(body []byte) interface{} {
	if json.Valid(body) {
		return json.RawMessage(body)
	}
	return string(body)
}
