package auth

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/schnurbus/go-mcp-gateway/internal/utils"
)

type ClientMetadata struct {
	ClientName              string   `json:"client_name,omitempty"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	JWKSURI                 string   `json:"jwks_uri,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
}

func (a *Auth) RegisterValidate(ctx context.Context, metadata *ClientMetadata) (*ClientMetadata, error) {
	// Validate redirect_uris
	if len(metadata.RedirectURIs) == 0 {
		return nil, fmt.Errorf("redirect_uris is required")
	}

	// Validate token_endpoint_auth_method
	if metadata.TokenEndpointAuthMethod == "" {
		metadata.TokenEndpointAuthMethod = "client_secret_post"
	}
	if !slices.Contains(a.supportedTokenEndpointAuthMethods, metadata.TokenEndpointAuthMethod) {
		return nil, fmt.Errorf("unsupported token_endpoint_auth_method")
	}

	// Validate grant_types
	if len(metadata.GrantTypes) == 0 {
		metadata.GrantTypes = []string{"authorization_code"}
	}
	for _, gt := range metadata.GrantTypes {
		if !slices.Contains(a.supportedGrantTypes, gt) {
			return nil, fmt.Errorf("unsupported grant_type")
		}
	}

	// Validate response_types
	if len(metadata.ResponseTypes) == 0 {
		metadata.ResponseTypes = []string{"code"}
	}
	for _, rt := range metadata.ResponseTypes {
		if !slices.Contains(a.supportedResponseTypes, rt) {
			return nil, fmt.Errorf("unsupported response_type")
		}
	}

	return metadata, nil
}

func (a *Auth) Register(ctx context.Context, metadata *ClientMetadata) *Client {
	clientID := utils.RandString(32)
	clientSecret := utils.RandString(32)

	var clientSecretExpiresAt int64
	if metadata.ClientSecretExpiresAt == 0 {
		clientSecretExpiresAt = 0
	} else {
		clientSecretExpiresAt = time.Now().Add(90 * 24 * time.Hour).Unix()
	}

	return &Client{
		ClientID:                clientID,
		ClientSecret:            clientSecret,
		ClientIDIssuedAt:        time.Now().Unix(),
		ClientSecretExpiresAt:   clientSecretExpiresAt,
		RedirectURIs:            metadata.RedirectURIs,
		GrantTypes:              metadata.GrantTypes,
		ResponseTypes:           metadata.ResponseTypes,
		TokenEndpointAuthMethod: metadata.TokenEndpointAuthMethod,
		JWKSURI:                 metadata.JWKSURI,
		LogoURI:                 metadata.LogoURI,
		RegistrationAccessToken: "",
		RegistrationClientURI:   "",
	}
}
