package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/schnurbus/go-mcp-gateway/internal/auth"
	"github.com/schnurbus/go-mcp-gateway/internal/config"
	"github.com/schnurbus/go-mcp-gateway/internal/handler"
	"github.com/schnurbus/go-mcp-gateway/internal/logger"
	"github.com/schnurbus/go-mcp-gateway/internal/middleware/googletokenvalidator"
)

func main() {
	mainLogger := logger.NewLogger()
	ctx := logger.WithContext(context.Background(), mainLogger)

	// Load config
	cfg, proxies, err := config.NewConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// Create redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
	})

	// Create Auth
	auth := auth.NewAuth(cfg.BaseURL, rdb)

	// Create Handler
	handler, err := handler.NewHandler(ctx, rdb, cfg, auth)
	if err != nil {
		log.Fatalf("failed to create handler: %v", err)
	}

	// Create proxy
	proxy.WithTlsConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	// Fiber App
	app := fiber.New()

	// Fiber Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:6274",
		AllowMethods:     "GET, POST, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, MCP-Protocol-Version",
		AllowCredentials: true,
	}))
	app.Use(healthcheck.New())
	app.Use(limiter.New(limiter.Config{
		Max:               100,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	app.Use(fiberLogger.New())
	app.Use(recover.New())
	app.Use(requestid.New())

	// Routes
	app.Get("/.well-known/oauth-protected-resource", handler.HandleOAuthProtectedResourceMetadata)
	app.Get("/.well-known/oauth-authorization-server", handler.HandleOAuthAuthorizationServerMetadata)
	app.Post(auth.GetDynamicRegistrationPath(), handler.HandleOAuthRegister)
	app.Get(auth.GetAuthorizationPath(), handler.HandleOAuthAuthorize)
	app.Get(auth.GetCallbackPath(), handler.HandleOAuthCallback)
	app.Post(auth.GetTokenPath(), handler.HandleOauthToken)

	// Validate Google access tokens for all proxied requests
	app.Use(googletokenvalidator.New())

	// Proxies
	for _, p := range proxies {
		mainLogger.Info("Register proxy", "pattern", p.Pattern, "target", p.TargetURL.String())
		app.Get(p.Pattern, proxy.Forward(p.TargetURL.String()))
		app.Post(p.Pattern, proxy.Forward(p.TargetURL.String()))
	}

	// Server

	log.Fatal(app.Listen(":8080"))
}
