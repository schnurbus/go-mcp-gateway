package googletokenvalidator

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GoogleAccessTokenValidator ist die Middleware-Funktion
func New() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header missing or invalid format (Expected: Bearer <token>)",
			})
		}

		accessToken := strings.TrimPrefix(authHeader, "Bearer ")

		isValid, err := validateGoogleToken(accessToken)
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

func validateGoogleToken(token string) (bool, error) {
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

		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println(string(bodyBytes))

		return true, nil
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return false, nil
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status code from Google: %d, body: %s", resp.StatusCode, string(bodyBytes))
}
