package handler

import (
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
)

func HandleAuthError(c *fiber.Ctx, err *auth.AuthError) error {
	if err.AuthJsonError.Code != "" {
		status := fiber.StatusOK
		if err.AuthJsonError.Code == auth.InvalidClientMetadata {
			status = fiber.StatusBadRequest
		}
		if err.AuthJsonError.Code == auth.InvalidRequest {
			status = fiber.StatusBadRequest
		}
		if err.AuthJsonError.Code == auth.UnauthorizedClient {
			status = fiber.StatusUnauthorized
		}
		if err.AuthJsonError.Code == auth.ServerError {
			status = fiber.StatusInternalServerError
		}

		return c.Status(status).JSON(err.AuthJsonError)
	} else {
		params := "?error=" + url.QueryEscape(err.AuthRedirectError.ErrorCode) + "&error_description=" + url.QueryEscape(err.AuthRedirectError.ErrorDescription)
		if err.AuthRedirectError.State != "" {
			params += "&state=" + url.QueryEscape(err.AuthRedirectError.State)
		}

		return c.Redirect(err.AuthRedirectError.RedirectURI+params, fiber.StatusFound)
	}
}

func ExtractCredentials(authHeader string) (string, string, error) {
	if authHeader == "" {
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "Authorization header missing")
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "Authorization scheme must be Basic")
	}

	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")

	decoded, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "Invalid Base64 encoding")
	}

	credentials := string(decoded)

	parts := strings.SplitN(credentials, ":", 2)
	if len(parts) != 2 {
		return "", "", fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials format (client_id:client_secret required)")
	}

	client_id := parts[0]
	client_secret := parts[1]

	return client_id, client_secret, nil
}
