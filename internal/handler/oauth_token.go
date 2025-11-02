package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
)

func (h *Handler) HandleOauthToken(c *fiber.Ctx) error {
	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		requestId = uuid.New().String()
	}
	log := logger.FromContext(c.Context()).With(
		slog.String("handler", "HandleOauthToken"),
		slog.String("request_id", requestId),
	)
	ctx := logger.WithContext(c.Context(), log)

	contentType := c.Get(fiber.HeaderContentType)
	if contentType != fiber.MIMEApplicationForm {
		log.Warn("Invalid content type", "content-type", c.Get("content-type"))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "bad_request",
			"description": "Content-Type muste be application/x-www-form-urlencoded",
		})
	}

	params := new(auth.TokenRequestParams)
	if err := c.BodyParser(params); err != nil {
		log.Warn("Could not parse body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":       "bad_request",
			"description": "Invalid body",
		})
	}
	params.Authorization = c.Get(fiber.HeaderAuthorization)

	switch params.GrantType {
	case "authorization_code":
		googleTokens, authErr := h.generateAuthorizationCode(ctx, params)
		if authErr != nil {
			log.Error("cannot generate access token", "error", authErr)
			return HandleAuthError(c, authErr)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token_type":    "Bearer",
			"expires_in":    googleTokens.GoogleExpiry,
			"access_token":  googleTokens.GoogleAccessToken,
			"refresh_token": googleTokens.GoogleRefreshToken,
		})
	case "refresh_token":
		var refreshTokenParams auth.RefreshTokenRequestParams

		if err := c.BodyParser(&refreshTokenParams); err != nil {
			log.Warn("Could not parse body")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":       "bad_request",
				"description": "Invalid body",
			})

		}

		if params.Authorization != "" && refreshTokenParams.ClientID == "" {
			clientId, clientSecret, err := ExtractCredentials(params.Authorization)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "cannot extract authorization information",
				})
			}
			refreshTokenParams.ClientID = clientId
			refreshTokenParams.ClientSecret = clientSecret
		}

		newAccessToken, expiry, authErr := h.generateRefreshToken(ctx, &refreshTokenParams)
		if authErr != nil {
			log.Error("cannot generate refresh token", "error", authErr)
			return HandleAuthError(c, authErr)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"token_type":    "Bearer",
			"expires_in":    expiry,
			"access_token":  newAccessToken,
			"refresh_token": params.RefreshToken,
		})
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error":       "bad_request",
		"description": "grant_type must be authorization_code or refresh_token",
	})
}

func (h *Handler) generateAuthorizationCode(ctx context.Context, params *auth.TokenRequestParams) (*auth.AuthorizationCodeResult, *auth.AuthError) {
	if authErr := h.auth.TokenValidateParams(ctx, params); authErr != nil {
		return nil, authErr
	}

	// MCP Inspector workaround
	if params.Authorization != "" && params.ClientID == "" {
		clientId, _, err := ExtractCredentials(params.Authorization)
		if err != nil {
			return nil, &auth.AuthError{
				AuthJsonError: auth.AuthJsonError{
					Code:        "invalid_client_metadata",
					Description: "could not determine client id",
				},
			}
		}
		params.ClientID = clientId
	}

	client, authErr := h.auth.GetClient(ctx, params.ClientID)
	if authErr != nil {
		return nil, authErr
	}
	if authErr := h.auth.TokenValidateClient(ctx, params, client); authErr != nil {
		return nil, authErr
	}
	if authErr := h.auth.TokenValidateClientSecret(ctx, params, client); authErr != nil {
		return nil, authErr
	}
	result, authErr := h.auth.VerifyAuthorizationCode(ctx, params.Code, params.ClientID, params.RedirectURI, params.CodeVerifier)
	if authErr != nil {
		return nil, authErr
	}

	return result, nil
}

func (h *Handler) generateRefreshToken(ctx context.Context, params *auth.RefreshTokenRequestParams) (string, int64, *auth.AuthError) {
	log := logger.FromContext(ctx)

	if authErr := h.auth.RefreshTokenValidateParams(ctx, params); authErr != nil {
		return "", 0, authErr
	}

	client, authErr := h.auth.GetClient(ctx, params.ClientID)
	if authErr != nil {
		return "", 0, authErr
	}
	if authErr := h.auth.RefreshTokenValidateClient(ctx, params, client); authErr != nil {
		return "", 0, authErr
	}

	// Use Google's refresh token to get a new access token
	newAccessToken, expiry, err := h.oauthGoogle.RefreshToken(ctx, params.RefreshToken)
	if err != nil {
		log.Error("Failed to refresh Google token", "error", err)
		return "", 0, &auth.AuthError{
			AuthJsonError: auth.AuthJsonError{
				Code:        auth.ServerError,
				Description: "Failed to refresh token",
			},
		}
	}

	return newAccessToken, expiry, nil
}
