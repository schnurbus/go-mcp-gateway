package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
)

func (h *Handler) HandleOAuthCallback(c *fiber.Ctx) error {
	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		requestId = uuid.New().String()
	}
	log := logger.FromContext(c.Context()).With(
		slog.String("handler", "HandleOAuthCallback"),
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
	state := c.Query("state")
	code := c.Query("code")

	googleAuthResult, err := h.oauthGoogle.Callback(ctx, sessionId, state, code)
	if err != nil {
		log.Error("Failed to get user ID", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":       "internal_server_error",
			"description": "Failed to get user ID",
		})
	}

	authParams, authErr := h.auth.GetAuthorization(ctx, sessionId)
	if authErr != nil {
		log.Error("Failed to get authorization", "error", authErr)
		return HandleAuthError(c, authErr)
	}

	authCode, authErr := h.auth.GenerateAuthorizationCode(ctx, &auth.AuthorizationCodeParams{
		UID:              googleAuthResult.Claims.Sub,
		ClientID:         authParams.ClientID,
		RedirectURI:      authParams.RedirectURI,
		CodeChallenge:    authParams.CodeChallenge,
		GoogleAccessToken:  googleAuthResult.AccessToken,
		GoogleRefreshToken: googleAuthResult.RefreshToken,
		GoogleExpiry:       googleAuthResult.Expiry,
	})
	if authErr != nil {
		log.Error("Failed to generate authorization code", "error", authErr)
		return HandleAuthError(c, authErr)
	}

	urlParams := "?code=" + authCode
	if authParams.State != "" {
		urlParams += "&state=" + authParams.State
	}

	return c.Redirect(authParams.RedirectURI+urlParams, fiber.StatusFound)
}
