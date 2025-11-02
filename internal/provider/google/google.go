package google

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/redis/go-redis/v9"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
	"github.com/schnurbus/go-mcp-gateway/internal/store"
	"github.com/schnurbus/go-mcp-gateway/internal/utils"
	"golang.org/x/oauth2"
)

type GoogleConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
}

type GoogleProvider struct {
	googleStateStore *store.Store // key: sid, value: google state
	googleNonceStore *store.Store // key: sid, value: google nonce
	googleCodeStore  *store.Store // key: sid, value: google code
	oidcConfig       *oauth2.Config
	verifier         *oidc.IDTokenVerifier
}

type GoogleClaims struct {
	Sub               string   `json:"sub"`
	Name              string   `json:"name"`
	Email             string   `json:"email"`
	Ver               int      `json:"ver"`
	Iss               string   `json:"iss"`
	Aud               string   `json:"aud"`
	Iat               int      `json:"iat"`
	Exp               int      `json:"exp"`
	Jti               string   `json:"jti"`
	Amr               []string `json:"amr"`
	Idp               string   `json:"idp"`
	Nonce             string   `json:"nonce"`
	PreferredUsername string   `json:"preferred_username"`
	AuthTime          int      `json:"auth_time"`
	AtHash            string   `json:"at_hash"`
}

type GoogleAuthResult struct {
	Claims       *GoogleClaims
	AccessToken  string
	RefreshToken string
	Expiry       int64 // Unix timestamp
}

func NewGoogleProvider(ctx context.Context, config *GoogleConfig, rdb *redis.Client) (*GoogleProvider, error) {
	googleStateStore := store.NewStore(rdb, "google_state", store.OAuthStateTTL)
	googleNonceStore := store.NewStore(rdb, "google_nonce", store.OAuthStateTTL)
	googleCodeStore := store.NewStore(rdb, "google_code", store.OAuthStateTTL)

	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to create oidc provider: %w", err)
	}

	oidcConfig := &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{"openid", "profile", "email"},
		RedirectURL:  config.GoogleRedirectURI,
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: config.GoogleClientID})

	return &GoogleProvider{
		googleStateStore: googleStateStore,
		googleNonceStore: googleNonceStore,
		googleCodeStore:  googleCodeStore,
		oidcConfig:       oidcConfig,
		verifier:         verifier,
	}, nil
}

func (p *GoogleProvider) GetAuthCodeURL(ctx context.Context, sid string) (string, error) {
	log := logger.FromContext(ctx)
	state := utils.RandString(16)
	nonce := utils.RandString(16)
	codeVerifier := utils.RandString(96)
	hashedCodeVerifier := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hashedCodeVerifier[:])

	log.Info("Saving state", "sid", sid, "state", state)
	if err := p.googleStateStore.Set(ctx, sid, state); err != nil {
		return "", err
	}

	if err := p.googleNonceStore.Set(ctx, sid, nonce); err != nil {
		return "", err
	}

	if err := p.googleCodeStore.Set(ctx, sid, codeVerifier); err != nil {
		return "", err
	}

	return p.oidcConfig.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.AccessTypeOffline,
		// oauth2.ApprovalForce,
	), nil
}

func (p *GoogleProvider) Callback(ctx context.Context, sid, state, code string) (*GoogleAuthResult, error) {
	log := logger.FromContext(ctx)

	savedState, err := p.googleStateStore.GetDel(ctx, sid)
	if err != nil {
		log.Error("could not get state from store", "sid", sid)
		return nil, err
	}

	codeVerifier, err := p.googleCodeStore.GetDel(ctx, sid)
	if err != nil {
		log.Error("could not get code from store", "sid", sid)
		return nil, err
	}

	if savedState != state {
		return nil, fmt.Errorf("invalid state: %s", savedState)
	}

	oauth2Tok, err := p.oidcConfig.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	rawIDToken, ok := oauth2Tok.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("failed to get id_token")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify id_token: %w", err)
	}

	var claims GoogleClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	nonce, err := p.googleNonceStore.GetDel(ctx, sid)
	if err != nil {
		return nil, err
	}

	if claims.Nonce != nonce {
		return nil, fmt.Errorf("invalid nonce: %s", claims.Nonce)
	}

	return &GoogleAuthResult{
		Claims:       &claims,
		AccessToken:  oauth2Tok.AccessToken,
		RefreshToken: oauth2Tok.RefreshToken,
		Expiry:       oauth2Tok.Expiry.Unix(),
	}, nil
}

func (p *GoogleProvider) RefreshToken(ctx context.Context, refreshToken string) (string, int64, error) {
	log := logger.FromContext(ctx)

	tokenSource := p.oidcConfig.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	newToken, err := tokenSource.Token()
	if err != nil {
		log.Error("failed to refresh token", "error", err)
		return "", 0, fmt.Errorf("failed to refresh token: %w", err)
	}

	log.Info("Successfully refreshed Google token")
	return newToken.AccessToken, newToken.Expiry.Unix(), nil
}
