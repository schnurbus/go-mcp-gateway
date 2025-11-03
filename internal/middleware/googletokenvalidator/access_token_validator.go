package googletokenvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
	"github.com/schnurbus/go-mcp-gateway/pkg/jsonrpc"
)

type TokenInfoResponse struct {
	Azp           string        `json:"azp"`
	Aud           string        `json:"aud"`
	Sub           string        `json:"sub"`
	Scope         string        `json:"scope"`
	Exp           JsonTimestamp `json:"exp"`
	ExpiresIn     string        `json:"expires_in"`
	Email         string        `json:"email"`
	EmailVerified JsonBool      `json:"email_verified"`
	AccessTypeOff string        `json:"access_type"`
}

var (
	ErrDecode       = errors.New("cannot decode token info")
	ErrInvalidAud   = errors.New("invalid audience")
	ErrTokenExpired = errors.New("token is expired")
)

func New(googleClientId string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		log := logger.FromContext(c.Context()).With(
			slog.String("middleware", "googletokenvalidator"),
		)

		var req jsonrpc.JSONRPCRequest
		if err := c.BodyParser(&req); err != nil {
			log.Error("failed to parse request body", "error", err)
		}

		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			log.Error("missing authorization header", "request-id", req.ID, "method", req.Method)

			return c.Status(fiber.StatusOK).JSON(
				jsonrpc.NewErrorResponse(
					req.ID,
					"Authentication token missing",
					-32001,
					jsonrpc.AuthErrorData{
						Type:   "auth_error",
						Reason: "missing_token",
					}))
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		isValid, err := validateGoogleToken(accessToken, googleClientId)
		if err != nil || !isValid {
			log.Error("invalid authorization token", "error", err)
			if errors.Is(err, ErrInvalidAud) || errors.Is(err, ErrDecode) {
				return c.Status(fiber.StatusOK).JSON(
					jsonrpc.NewErrorResponse(
						req.ID,
						err.Error(),
						-32001,
						jsonrpc.AuthErrorData{
							Type:           "auth_error",
							Reason:         "invalid_token",
							RequiresReauth: true,
						}))
			}
			if errors.Is(err, ErrTokenExpired) {
				return c.Status(fiber.StatusOK).JSON(
					jsonrpc.NewErrorResponse(
						req.ID,
						err.Error(),
						-32001,
						jsonrpc.AuthErrorData{
							Type:           "auth_error",
							Reason:         "token_expired",
							RequiresReauth: true,
						}))
			}
			return c.Status(fiber.StatusOK).JSON(
				jsonrpc.NewErrorResponse(
					req.ID,
					err.Error(),
					-32001,
					jsonrpc.AuthErrorData{
						Type:           "auth_error",
						Reason:         "invalid_token",
						RequiresReauth: true,
					}))
		}
		return c.Next()
	}
}

func validateGoogleToken(token, googleClientId string) (bool, error) {
	const validationURL = "https://www.googleapis.com/oauth2/v3/tokeninfo?access_token="

	resp, err := http.Get(validationURL + token)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		tokenInfo := new(TokenInfoResponse)
		if err := json.NewDecoder(resp.Body).Decode(tokenInfo); err != nil {
			return false, ErrDecode
		}

		if tokenInfo.Aud != googleClientId {
			return false, ErrInvalidAud
		}

		if time.Now().After(time.Time(tokenInfo.Exp)) {
			return false, ErrTokenExpired
		}

		return true, nil
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status code from Google: %d, body: %s", resp.StatusCode, string(bodyBytes))
}
