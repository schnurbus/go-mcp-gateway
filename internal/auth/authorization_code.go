package auth

import (
	"context"
	"encoding/json"

	"github.com/schnurbus/go-mcp-gateway/internal/utils"
)

type AuthorizationCodeParams struct {
	UID              string
	ClientID         string
	RedirectURI      string
	CodeChallenge    string
	GoogleAccessToken  string
	GoogleRefreshToken string
	GoogleExpiry       int64
}

type AuthorizationCodeResult struct {
	UID              string
	GoogleAccessToken  string
	GoogleRefreshToken string
	GoogleExpiry       int64
}

func (a *Auth) GenerateAuthorizationCode(ctx context.Context, params *AuthorizationCodeParams) (string, *AuthError) {
	code := utils.RandString(32)
	codeDataJSON, err := json.Marshal(params)
	if err != nil {
		return "", &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        ServerError,
				Description: "Failed to marshal authorization code",
			},
		}
	}
	if err := a.codeStore.Set(ctx, code, codeDataJSON); err != nil {
		return "", &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        ServerError,
				Description: "Failed to store authorization code",
			},
		}
	}

	return code, nil
}

func (a *Auth) VerifyAuthorizationCode(ctx context.Context, code, clientID, redirectURI, codeVerifier string) (*AuthorizationCodeResult, *AuthError) {
	storedCodeDataJSON, err := a.codeStore.GetDel(ctx, code)
	if err != nil {
		return nil, &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        ServerError,
				Description: "Failed to get authorization code",
			},
		}
	}

	var storedCodeData AuthorizationCodeParams
	if err := json.Unmarshal([]byte(storedCodeDataJSON), &storedCodeData); err != nil {
		return nil, &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        ServerError,
				Description: "Failed to unmarshal authorization code",
			},
		}
	}

	if storedCodeData.ClientID != clientID || storedCodeData.RedirectURI != redirectURI {
		return nil, &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "Invalid client or redirect URI",
			},
		}
	}

	hash := utils.S256(codeVerifier)
	if hash != storedCodeData.CodeChallenge {
		return nil, &AuthError{
			AuthJsonError: AuthJsonError{
				Code:        InvalidRequest,
				Description: "Invalid code challenge",
			},
		}
	}

	return &AuthorizationCodeResult{
		UID:              storedCodeData.UID,
		GoogleAccessToken:  storedCodeData.GoogleAccessToken,
		GoogleRefreshToken: storedCodeData.GoogleRefreshToken,
		GoogleExpiry:       storedCodeData.GoogleExpiry,
	}, nil
}
