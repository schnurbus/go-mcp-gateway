package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
)

func (h *Handler) HandleOAuthRegister(c *fiber.Ctx) error {
	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		requestId = uuid.New().String()
	}
	log := logger.FromContext(c.Context()).With(
		slog.String("handler", "HandleOAuthRegister"),
		slog.String("request_id", requestId),
	)
	ctx := logger.WithContext(c.Context(), log)

	contentType := c.Get(fiber.HeaderContentType)
	if contentType != fiber.MIMEApplicationJSON {
		log.Warn("Invalid content type", "content-type", c.Get("content-type"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "bad_request",
			"description": "Content-Type muste be application/json",
		})
	}

	metadata := new(auth.ClientMetadata)
	if err := c.BodyParser(metadata); err != nil {
		log.Warn("Failed to decode body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "bad_request",
			"description": "Failed to decode body",
		})
	}

	validatedMetadata, err := h.auth.RegisterValidate(ctx, metadata)
	if err != nil {
		log.Warn("Failed to validate client metadata", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "bad_request",
			"description": "Failed to validate metadata",
		})
	}

	client := h.auth.Register(ctx, validatedMetadata)

	if err := h.auth.SaveClient(ctx, client.ClientID, client); err != nil {
		log.Error("Failed to save client", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":       "internal_server_error",
			"description": "Failed to save client",
		})
	}

	return c.Status(fiber.StatusOK).JSON(client)
}
