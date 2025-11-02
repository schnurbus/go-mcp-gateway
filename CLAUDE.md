# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

go-mcp-gateway is an OAuth 2.0 authorization facilitator and reverse proxy gateway for MCP (Model Context Protocol) servers. It implements RFC 6749 OAuth 2.0 authorization code flow with PKCE, integrates with Google OIDC for user authentication, and forwards Google access tokens to clients. The gateway validates Google tokens before proxying requests to configured MCP servers, ensuring MCP servers can access Google resources on behalf of the user.

## Build and Development Commands

### Running the Application

```bash
go run cmd/server/main.go
```

The server listens on port 8080 by default.

### Building

```bash
go build -o bin/server cmd/server/main.go
```

### Running Tests

```bash
go test ./...
```

To run tests for a specific package:

```bash
go test ./internal/auth
```

To run tests with verbose output:

```bash
go test -v ./...
```

## Configuration

The application uses environment variables for configuration, loaded from a `.env` file:

- `BASE_URL`: Base URL of the server (default: http://localhost:8080)
- `PORT`: Server port (default: 8080)
- `REDIS_ADDR`: Redis address (default: localhost:6379)
- `REDIS_PASSWORD`: Redis password (optional)
- `OAUTH_GOOGLE_CLIENT_ID`: Google OAuth client ID (required)
- `OAUTH_GOOGLE_CLIENT_SECRET`: Google OAuth client secret (required)
- `OAUTH_GOOGLE_REDIRECT_URI`: Google OAuth redirect URI (required)

Proxy configuration is defined in `config.yaml`:

```yaml
proxies:
  - pattern: "/calc/mcp"
    target_url: "http://localhost:3000/mcp"
```

## Architecture

### Core Components

**Auth Layer** (`internal/auth/`): Implements OAuth 2.0 authorization server functionality. The `Auth` struct manages five Redis-backed stores for different OAuth entities (clients, codes, authorizations, access tokens, refresh tokens). Key files:
- `auth.go`: Main Auth struct and initialization
- `client.go`: Dynamic client registration (RFC 7591)
- `authorization_code.go`: Authorization code generation and validation
- `token.go`: Access and refresh token management
- `authorization_validation.go`: Request parameter validation

**Handler Layer** (`internal/handler/`): HTTP request handlers for OAuth endpoints and metadata discovery. Integrates Auth with GoogleProvider for user authentication flow:
- `handler.go`: Handler struct initialization with session management
- `oauth_authorization_server.go`: Authorization server metadata endpoint
- `oauth_protected_resource.go`: Protected resource metadata endpoint
- `oauth_register.go`: Dynamic client registration endpoint
- `oauth_authorize.go`: Authorization endpoint (initiates Google OIDC flow)
- `oauth_callback.go`: Google OIDC callback handler
- `oauth_token.go`: Token endpoint (exchanges codes for tokens)

**Google Provider** (`internal/provider/google/`): Manages Google OIDC integration with state/nonce validation and PKCE. Maintains three Redis stores for temporary OAuth state, nonce, and code verifier. The `Callback` method exchanges Google authorization codes for Google access/refresh tokens and ID tokens, validates nonces, and returns a `GoogleAuthResult` containing all tokens. The `RefreshToken` method uses Google's refresh token to obtain a new access token via `oauth2.TokenSource`.

**Store Layer** (`internal/store/`): Redis abstraction with namespacing and TTL management. Each store has a prefix and TTL constant:
- OAuth authorization codes (with embedded Google tokens): 5 minutes
- OAuth state/nonce for Google OIDC: 5 minutes
- Client registrations: 90 days
- Sessions: 7 days

Note: The gateway no longer stores Google access/refresh tokens long-term. Google tokens are only stored temporarily (5 min) within the authorization code during the OAuth exchange flow.

**Middleware** (`internal/middleware/googletokenvalidator/`): Validates Google access tokens by calling Google's tokeninfo API endpoint (`https://www.googleapis.com/oauth2/v3/tokeninfo`). Returns 401 if the token is missing, invalid, or expired. Applied to all proxied routes in `cmd/server/main.go:92`.

### Request Flow

1. **Client Registration**: Client POSTs to `/oauth/register` with metadata, receives client_id and client_secret
2. **Authorization Request**: Client initiates OAuth flow at `/oauth/authorize`, gateway stores authorization params in Redis and redirects to Google OIDC
3. **Google Callback**: Google redirects to `/oauth/callback`, gateway validates state/nonce, exchanges Google authorization code for Google access/refresh tokens and ID token, stores Google tokens with authorization code (5 min TTL)
4. **Token Exchange**: Client POSTs authorization code to `/oauth/token`, gateway validates PKCE challenge and returns **Google's access/refresh tokens** to the client
5. **Protected Resource Access**: Client sends Google access token to proxied MCP endpoints (e.g., `/calc/mcp`), middleware validates Google token with Google's tokeninfo endpoint, request proxies to configured target with the same token
6. **Token Refresh**: Client uses Google refresh token at `/oauth/token` with grant_type=refresh_token, gateway validates client credentials and uses GoogleProvider to exchange refresh token for new Google access token

### OAuth Flow Details

The gateway acts as an OAuth authorization facilitator and Google OIDC client. It does NOT generate its own access tokens; instead, it forwards Google's tokens to clients. Authorization parameters are stored in Redis keyed by session ID (sid). After successful Google authentication, the authorization code temporarily stores the Google tokens (access, refresh, expiry) with a 5-minute TTL. When the client exchanges the authorization code for tokens, it receives Google's original tokens. The gateway validates Google tokens via the `googletokenvalidator` middleware which calls Google's tokeninfo endpoint before proxying requests to backend MCP servers. This architecture ensures MCP servers receive valid Google tokens to access Google resources on behalf of the user, while the gateway maintains minimal security risk by not storing long-term Google credentials.

### Session Management

Fiber sessions are backed by Redis using `gofiber/storage/redis`. Session IDs (sid) are used as keys to correlate Google OAuth state, authorization parameters, and user identity across the multi-step OAuth flow.

### Dependencies

- Fiber v2: HTTP web framework
- go-redis v9: Redis client
- golang.org/x3/oauth2: OAuth2 client for Google integration
- coreos/go-oidc: OpenID Connect client library
- google/uuid: UUID generation for OAuth codes/tokens

## Important Paths and Routes

OAuth 2.0 endpoints (defined in `internal/auth/auth.go:39-42`):
- `/oauth/register`: Dynamic client registration
- `/oauth/authorize`: Authorization endpoint
- `/oauth/callback`: Google OIDC callback
- `/oauth/token`: Token endpoint

Metadata discovery (OAuth 2.0 RFC 8414):
- `/.well-known/oauth-authorization-server`: Server metadata
- `/.well-known/oauth-protected-resource`: Resource server metadata

Proxied routes are dynamically registered from `config.yaml` at `cmd/server/main.go:96-100`.

## Code Style Notes

- Error handling: Errors are wrapped with `fmt.Errorf` to provide context
- Logging: Uses structured logging via `internal/logger` with context propagation
- URLs: Configuration validation ensures URLs don't end with slashes
- PKCE: Uses S256 code challenge method exclusively
