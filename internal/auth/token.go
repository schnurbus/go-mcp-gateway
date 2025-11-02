package auth

import (
	"context"
	"encoding/base64"
	"slices"
	"strings"

	"github.com/schnurbus/go-mcp-gateway/internal/utils"
)

type TokenRequestParams struct {
	GrantType     string `form:"grant_type"`
	Code          string `form:"code"`
	RedirectURI   string `form:"redirect_uri"`
	ClientID      string `form:"client_id"`
	clientSecret  string `form:"client_secret"`
	CodeVerifier  string `form:"code_verifier"`
	Authorization string `header:"Authorization"`
	RefreshToken  string `form:"refresh_token"`
}

type RefreshTokenRequestParams struct {
	GrantType    string `form:"grant_type"`
	RefreshToken string `form:"refresh_token"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
}

func (a *Auth) TokenValidateParams(ctx context.Context, params *TokenRequestParams) *AuthError {
	if params.GrantType != "authorization_code" {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "grant_type must be authorization_code",
			},
		}
	}
	if params.Authorization != "" && (params.ClientID != "" || params.clientSecret != "") {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "When Authorization header is present, client_id and client_secret form parameters must not be sent",
			},
		}
	}
	if params.Authorization == "" && params.ClientID == "" {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "When Authorization header is not present, client_id form parameter is required",
			},
		}
	}
	if params.Code == "" || params.RedirectURI == "" || params.CodeVerifier == "" {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "code, redirect_uri and code_verifier are required",
			},
		}
	}
	// if params.Code == "" || params.RedirectURI == "" || params.ClientID == "" || params.CodeVerifier == "" {
	// 	return &AuthError{
	// 		AuthJsonError: AuthJsonError{
	// 			Code:        InvalidRequest,
	// 			Description: "code, redirect_uri, client_id, and code_verifier are required",
	// 		},
	// 	}
	// }
	if !utils.IsValidCodeChallengeOrVerifier(params.CodeVerifier) {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "code_verifier must be 43-128 chars and only [A-Z/a-z/0-9/-/./_/~] allowed",
			},
		}
	}
	return nil
}

func (a *Auth) TokenValidateClient(ctx context.Context, params *TokenRequestParams, client *Client) *AuthError {
	if !slices.Contains(client.RedirectURIs, params.RedirectURI) {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "redirect_uri is invalid",
			},
		}
	}

	if !slices.Contains(client.GrantTypes, params.GrantType) {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "response_type is invalid",
			},
		}
	}

	return nil
}

func (a *Auth) TokenValidateClientSecret(ctx context.Context, params *TokenRequestParams, client *Client) *AuthError {
	passed := false
	switch client.TokenEndpointAuthMethod {
	case "client_secret_post":
		if params.clientSecret != "" && params.clientSecret == client.ClientSecret {
			passed = true
		}
	case "client_secret_basic":
		header := params.Authorization
		if strings.HasPrefix(header, "Basic ") {
			payload, err := base64.StdEncoding.DecodeString(header[len("Basic "):])
			if err == nil {
				parts := strings.SplitN(string(payload), ":", 2)
				if len(parts) == 2 && parts[0] == params.ClientID && parts[1] == client.ClientSecret {
					passed = true
				}
			}
		}
	case "none":
		passed = true
	default:
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "unsupported token_endpoint_auth_method",
			},
		}
	}

	if !passed {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "invalid client_id and client_secret",
			},
		}
	}

	return nil
}


func (a *Auth) RefreshTokenValidateParams(ctx context.Context, params *RefreshTokenRequestParams) *AuthError {
	if params.GrantType != "refresh_token" {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "grant_type must be refresh_token",
			},
		}
	}
	if params.RefreshToken == "" {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "refresh_token is required",
			},
		}
	}
	if params.ClientID == "" || params.ClientSecret == "" {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "client_id and client_secret are required",
			},
		}
	}
	return nil
}

func (a *Auth) RefreshTokenValidateClient(ctx context.Context, params *RefreshTokenRequestParams, client *Client) *AuthError {
	if params.ClientID != client.ClientID {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "client_id is invalid",
			},
		}
	}
	if params.ClientSecret != client.ClientSecret {
		return &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "client_secret is invalid",
			},
		}
	}
	return nil
}
