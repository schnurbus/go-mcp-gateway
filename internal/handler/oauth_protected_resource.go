package handler

import "github.com/gofiber/fiber/v2"

func (h *Handler) HandleOAuthProtectedResourceMetadata(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"resource":                              h.baseURL,
		"issuer":                                h.baseURL,
		"authorization_servers":                 []string{h.baseURL},
		"token_endpoint_auth_methods_supported": h.auth.GetSupportTokenEndpointAuthMethods(),
	})
}
