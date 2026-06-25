package gateways

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"go-fiber-template/domain/entities"
	"go-fiber-template/src/middlewares"
	"go-fiber-template/src/services"

	"github.com/gofiber/fiber/v2"
)

var audioProxyFilenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// AudioProxy handles GET /api/v1/audio-proxy, replacing the Supabase Edge
// Function the frontend used to proxy/download call recording audio. It
// fetches the upstream url server-side (avoiding CORS/expiry issues on the
// client) and streams back the full audio body.
func (h *HTTPGateway) AudioProxy(ctx *fiber.Ctx) error {
	if _, err := middlewares.DecodeJWTToken(ctx); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(entities.ResponseMessage{Message: "Unauthorized token"})
	}

	url := ctx.Query("url")
	if url == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing url parameter"})
	}

	body, upstreamContentType, err := h.AudioProxyService.FetchAudio(url)
	if err != nil {
		var fetchErr *services.AudioFetchError
		if errors.As(err, &fetchErr) {
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error":  "Audio source is not accessible",
				"status": fetchErr.StatusCode,
				"detail": "The upstream audio file is private or expired and cannot be downloaded.",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	contentType := upstreamContentType
	if contentType == "" {
		if strings.Contains(strings.ToLower(url), ".wav") {
			contentType = "audio/wav"
		} else {
			contentType = "audio/mpeg"
		}
	}
	isWav := strings.Contains(strings.ToLower(contentType), "wav")

	ctx.Set(fiber.HeaderContentType, contentType)
	ctx.Set(fiber.HeaderCacheControl, "public, max-age=3600")
	ctx.Set(fiber.HeaderContentLength, strconv.Itoa(len(body)))
	ctx.Set("Cross-Origin-Resource-Policy", "cross-origin")

	if ctx.Query("download") == "1" {
		filename := ctx.Query("filename", "call_audio.mp3")
		ctx.Set(fiber.HeaderContentDisposition, `attachment; filename="`+sanitizeAudioFilename(filename, isWav)+`"`)
	}

	return ctx.Status(fiber.StatusOK).Send(body)
}

// sanitizeAudioFilename strips characters outside [a-zA-Z0-9._-] and forces
// the extension to match the detected audio format, dropping any existing
// audio extension first so it isn't duplicated (e.g. "x.mp3.wav").
func sanitizeAudioFilename(filename string, isWav bool) string {
	sanitized := audioProxyFilenameSanitizer.ReplaceAllString(filename, "_")

	ext := ".mp3"
	if isWav {
		ext = ".wav"
	}

	lower := strings.ToLower(sanitized)
	for _, known := range []string{".mp3", ".wav", ".m4a", ".ogg"} {
		if strings.HasSuffix(lower, known) {
			sanitized = sanitized[:len(sanitized)-len(known)]
			break
		}
	}

	return sanitized + ext
}
