package handler

import "github.com/gofiber/fiber/v2"

func (h *Handler) HandleOAuthAuthorizationServerMetadata(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"issuer":                                h.baseURL,
		"authorization_endpoint":                h.auth.GetAuthorizationURL(),
		"token_endpoint":                        h.auth.GetTokenURL(),
		"registration_endpoint":                 h.auth.GetDynamicRegistrationURL(),
		"response_types_supported":              h.auth.GetSupportResponseTypes(),
		"grant_types_supported":                 h.auth.GetSupportGrantTypes(),
		"token_endpoint_auth_methods_supported": h.auth.GetSupportTokenEndpointAuthMethods(),
		"code_challenge_methods_supported":      h.auth.GetSupportCodeChallengeMethods(),
	})
}
