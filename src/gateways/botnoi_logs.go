package gateways

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

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

// maxConversationLookup caps how many ids one lookup may ask for.
const maxConversationLookup = 500

// LookupBotnoiConversations handles GET /api/v1/botnoi-logs/conversations?ids=a,b,c
// and returns the debtor/call details for the ids that belong to the caller's
// own call records. Ids from calls not placed through this system are omitted.
func (h *HTTPGateway) LookupBotnoiConversations(ctx *fiber.Ctx) error {
	tokenData, err := middlewares.DecodeJWTToken(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	ids := []string{}
	for _, id := range strings.Split(ctx.Query("ids"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) > maxConversationLookup {
		return ctx.Status(fiber.StatusBadRequest).JSON(entities.ResponseMessage{Message: "too many ids (max " + strconv.Itoa(maxConversationLookup) + ")"})
	}

	links, err := h.BotnoiLogsService.LookupConversationsByUser(tokenData.UserID, ids)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(entities.ResponseMessage{Message: "failed to look up conversations"})
	}
	return ctx.Status(fiber.StatusOK).JSON(entities.ResponseModel{Message: "success", Data: links})
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
	case errors.Is(err, services.ErrBotnoiAgentIDNotConfigured):
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
