package googletokenvalidator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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

func New(googleClientId string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header missing or invalid format (Expected: Bearer <token>)",
			})
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")

		isValid, err := validateGoogleToken(accessToken, googleClientId)
		if err != nil {
			fmt.Println("Google validation error:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to validate token with Google",
			})
		}

		if !isValid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired Google access token",
			})
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
		// Hier könnten Sie zusätzlich die JSON-Antwort von Google lesen und prüfen,
		// ob die 'aud' (Audience) mit Ihrer Client ID übereinstimmt,
		// um Man-in-the-Middle-Angriffe zu verhindern (sehr empfohlen!).

		tokenInfo := new(TokenInfoResponse)
		if err := json.NewDecoder(resp.Body).Decode(tokenInfo); err != nil {
			return false, fmt.Errorf("could not decode token info")
		}

		if tokenInfo.Aud != googleClientId {
			return false, fmt.Errorf("token has invalid audience")
		}

		if time.Now().After(time.Time(tokenInfo.Exp)) {
			return false, fmt.Errorf("token is expired")
		}

		return true, nil
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status code from Google: %d, body: %s", resp.StatusCode, string(bodyBytes))
}
