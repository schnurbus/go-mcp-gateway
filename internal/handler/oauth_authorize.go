package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
)

func (h *Handler) HandleOAuthAuthorize(c *fiber.Ctx) error {
	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		requestId = uuid.New().String()
	}
	log := logger.FromContext(c.Context()).With(
		slog.String("handler", "HandleOAuthAuthorize"),
		slog.String("request_id", requestId),
	)
	ctx := logger.WithContext(c.Context(), log)

	sess, err := h.sessionStore.Get(c)
	if err != nil {
		log.Error("Could not get session", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":       "internal_server_error",
			"description": "Could not get session",
		})
	}

	sessionId := sess.ID()
	if err := sess.Save(); err != nil {
		log.Error("Could not save session", "sid", sessionId)
	}

	params := new(auth.AuthorizationParams)
	if err := c.QueryParser(params); err != nil {
		log.Warn("Could not parse query", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "bad_request",
			"description": "Could not parse query",
		})
	}

	client, authErr := h.auth.GetClient(ctx, params.ClientID)
	if authErr != nil {
		log.Error("Failed to get client", "error", err)
		return HandleAuthError(c, authErr)
	}

	if authErr := h.auth.ValidateAuthorizationClient(ctx, params, client); authErr != nil {
		log.Error("Failed to validate authorization client", "error", authErr)
		return HandleAuthError(c, authErr)
	}

	if authErr := h.auth.ValidateAuthorizationParams(ctx, params); authErr != nil {
		log.Warn("Failed to validate authorization params", "error", authErr)
		return HandleAuthError(c, authErr)
	}

	if authErr := h.auth.StoreAuthorization(ctx, sessionId, params); authErr != nil {
		log.Warn("Failed to store authorization", "error", authErr)
		return HandleAuthError(c, authErr)
	}

	authCode, err := h.oauthGoogle.GetAuthCodeURL(ctx, sessionId)
	if err != nil {
		log.Warn("Failed to get auth code URL", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get auth code URL",
		})
	}

	return c.Redirect(authCode, fiber.StatusFound)
}
