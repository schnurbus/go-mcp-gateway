package auth

import (
	"github.com/redis/go-redis/v9"
	"github.com/schnurbus/go-mcp-gateway/internal/store"
)

type Auth struct {
	baseURL                           string
	clientStore                       *store.Store // key: client_id, value: client
	codeStore                         *store.Store // key: code, value: code
	authorizationStore                *store.Store // key: sid, value: authorization param
	registerPath                      string
	authorizePath                     string
	callbackPath                      string
	tokenPath                         string
	supportedTokenEndpointAuthMethods []string
	supportedGrantTypes               []string
	supportedResponseTypes            []string
	supportedCodeChallengeMethods     []string
}

func NewAuth(baseURL string, rdb *redis.Client) *Auth {
	clientStore := store.NewStore(rdb, "client", store.OAuthClientTTL)
	codeStore := store.NewStore(rdb, "code", store.OAuthStateTTL)
	authorizationStore := store.NewStore(rdb, "authorization", store.OAuthStateTTL)

	return &Auth{
		baseURL:                           baseURL,
		clientStore:                       clientStore,
		codeStore:                         codeStore,
		authorizationStore:                authorizationStore,
		registerPath:                      "/oauth/register",
		authorizePath:                     "/oauth/authorize",
		callbackPath:                      "/oauth/callback",
		tokenPath:                         "/oauth/token",
		supportedTokenEndpointAuthMethods: []string{"client_secret_basic", "client_secret_post", "none"},
		supportedGrantTypes:               []string{"authorization_code", "refresh_token"},
		supportedResponseTypes:            []string{"code"},
		supportedCodeChallengeMethods:     []string{"S256"},
	}
}
