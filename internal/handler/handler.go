package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/middleware/session"
	fiberRedis "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
	"github.com/schnurbus/go-mcp-gateway/internal/config"
	"github.com/schnurbus/go-mcp-gateway/internal/provider/google"
)

type Handler struct {
	baseURL      string
	auth         *auth.Auth
	oauthGoogle  *google.GoogleProvider
	sessionStore *session.Store
}

func NewHandler(
	ctx context.Context,
	rdb *redis.Client,
	config *config.Config,
	auth *auth.Auth,
) (*Handler, error) {
	oauthGoogle, err := google.NewGoogleProvider(ctx, &google.GoogleConfig{
		GoogleClientID:     config.OAuthGoogleConfig.GoogleClientID,
		GoogleClientSecret: config.OAuthGoogleConfig.GoogleClientSecret,
		GoogleRedirectURI:  config.OAuthGoogleConfig.GoogleRedirectURI,
	}, rdb)
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth google provider: %w", err)
	}

	sessionStore := session.New(session.Config{
		Storage: fiberRedis.NewFromConnection(rdb),
	})

	return &Handler{
		baseURL:      config.BaseURL,
		auth:         auth,
		oauthGoogle:  oauthGoogle,
		sessionStore: sessionStore,
	}, nil
}
