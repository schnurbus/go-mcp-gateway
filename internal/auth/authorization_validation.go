package auth

import (
	"context"
	"slices"
	"strings"

	"github.com/schnurbus/go-mcp-gateway/internal/utils"
)

type AuthorizationParams struct {
	ClientID            string `query:"client_id"`
	RedirectURI         string `query:"redirect_uri"`
	ResponseType        string `query:"response_type"`
	State               string `query:"state"`
	CodeChallenge       string `query:"code_challenge"`
	CodeChallengeMethod string `query:"code_challenge_method"`
}

func (a *Auth) ValidateAuthorizationClient(ctx context.Context, params *AuthorizationParams, client *Client) *AuthError {
	if !slices.Contains(client.RedirectURIs, params.RedirectURI) {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "redirect_uri is invalid",
			},
		}
	}

	if !slices.Contains(client.ResponseTypes, params.ResponseType) {
		return &AuthError{
			AuthRedirectError: AuthRedirectError{
				RedirectURI:      params.RedirectURI,
				ErrorCode:        InvalidRequest,
				ErrorDescription: "unsupported response_type",
				State:            params.State,
			},
		}
	}

	return nil
}

func (a *Auth) ValidateAuthorizationParams(ctx context.Context, params *AuthorizationParams) *AuthError {
	if params.State == "" && params.CodeChallenge == "" {
		return &AuthError{
			AuthRedirectError: AuthRedirectError{
				RedirectURI:      params.RedirectURI,
				ErrorCode:        InvalidRequest,
				ErrorDescription: "state or code_challenge is required",
				State:            params.State,
			},
		}
	}

	if !slices.Contains(a.supportedCodeChallengeMethods, params.CodeChallengeMethod) {
		return &AuthError{
			AuthRedirectError: AuthRedirectError{
				RedirectURI:      params.RedirectURI,
				ErrorCode:        InvalidRequest,
				ErrorDescription: "code_challenge_method must be " + strings.Join(a.supportedCodeChallengeMethods, ", "),
				State:            params.State,
			},
		}
	}

	if !utils.IsValidCodeChallengeOrVerifier(params.CodeChallenge) {
		return &AuthError{
			AuthRedirectError: AuthRedirectError{
				RedirectURI:      params.RedirectURI,
				ErrorCode:        InvalidRequest,
				ErrorDescription: "code_challenge must be 43-128 chars and only [A-Z/a-z/0-9/-/./_/~] allowed",
				State:            params.State,
			},
		}
	}

	return nil
}
