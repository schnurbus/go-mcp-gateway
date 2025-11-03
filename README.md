# go-mcp-gateway

An OAuth 2.0 authorization facilitator and reverse proxy gateway for MCP (Model Context Protocol) servers. The gateway implements RFC 6749 OAuth 2.0 authorization code flow with PKCE, integrates with Google OIDC for user authentication, and forwards Google access tokens to clients for accessing protected MCP resources.

## Features

- **OAuth 2.0 Authorization Server**: Full implementation of RFC 6749 with PKCE (RFC 7636)
- **Dynamic Client Registration**: RFC 7591 compliant client registration
- **Google OIDC Integration**: Authenticates users via Google and forwards Google tokens
- **Token Validation**: Validates Google access tokens before proxying requests
- **Reverse Proxy**: Routes authenticated requests to configured MCP servers
- **Metadata Discovery**: OAuth 2.0 Authorization Server Metadata (RFC 8414)
- **Redis-backed Storage**: Session management and OAuth state storage
- **Security**: No long-term storage of Google credentials, PKCE enforcement

## Architecture

### Key Components

- **Auth Layer** (`internal/auth/`): OAuth 2.0 authorization server with client registration, authorization code, and token management
- **Handler Layer** (`internal/handler/`): HTTP handlers for OAuth endpoints and metadata discovery
- **Google Provider** (`internal/provider/google/`): Google OIDC integration with state/nonce validation
- **Store Layer** (`internal/store/`): Redis abstraction with namespacing and TTL management
- **Middleware** (`internal/middleware/googletokenvalidator/`): Google access token validation

### How It Works

1. **Client Registration**: Clients register with the gateway to receive OAuth credentials
2. **Authorization**: Clients initiate OAuth flow, gateway redirects to Google for authentication
3. **Token Exchange**: After Google authentication, gateway returns Google's access/refresh tokens to the client
4. **Protected Access**: Clients use Google tokens to access proxied MCP endpoints
5. **Token Refresh**: Clients can refresh expired tokens using Google refresh tokens

The gateway acts as an **OAuth facilitator** - it does not issue its own tokens but forwards Google's tokens to clients. This ensures MCP servers receive valid Google credentials to access Google resources on behalf of authenticated users.

## Prerequisites

- Go 1.21 or higher
- Redis server
- Google OAuth 2.0 credentials

## Getting Started

### 1. Set up Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API
4. Create OAuth 2.0 credentials (OAuth client ID)
5. Add authorized redirect URI: `http://localhost:8080/oauth/callback`

### 2. Configure Environment Variables

Create a `.env` file in the project root:

```env
ALLOWED_ORIGINS=*
BASE_URL=http://localhost:8080
PORT=8080
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
OAUTH_GOOGLE_CLIENT_ID=your-google-client-id
OAUTH_GOOGLE_CLIENT_SECRET=your-google-client-secret
OAUTH_GOOGLE_REDIRECT_URI=http://localhost:8080/oauth/callback
OAUTH_GOOGLE_SCOPES=openid,profile,email
```

### 3. Configure Proxy Routes

Edit `config.yaml` to define your MCP server endpoints:

```yaml
proxies:
  - pattern: "/calc/mcp"
    target_url: "http://localhost:3000/mcp"
  - pattern: "/files/mcp"
    target_url: "http://localhost:3001/mcp"
```

### 4. Start Redis

```bash
redis-server
```

### 5. Run the Gateway

Development mode:
```bash
go run cmd/server/main.go
```

Or build and run:
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

The server will start on `http://localhost:8080` (or your configured port).

## Deployment Options

### Kubernetes with Helm (Recommended for Production)

Deploy to Kubernetes using the included Helm chart:

```bash
# Install from GitHub Container Registry (recommended)
helm install my-gateway oci://ghcr.io/schnurbus/go-mcp-gateway \
  --version 0.1.0 \
  --set config.oauth.google.clientId=YOUR_CLIENT_ID \
  --set config.oauth.google.clientSecret=YOUR_CLIENT_SECRET

# Or install from local chart
helm install my-gateway ./chart \
  --set config.oauth.google.clientId=YOUR_CLIENT_ID \
  --set config.oauth.google.clientSecret=YOUR_CLIENT_SECRET

# Production install with custom values
helm install my-gateway oci://ghcr.io/schnurbus/go-mcp-gateway \
  --version 0.1.0 \
  -f values-prod.yaml
```

See the [Helm Chart README](chart/README.md) for detailed configuration options.

### Docker Deployment

#### Using Pre-built Images from GHCR

Pull the latest image from GitHub Container Registry:

```bash
docker pull ghcr.io/schnurbus/go-mcp-gateway:latest
```

Or pull a specific version:

```bash
docker pull ghcr.io/schnurbus/go-mcp-gateway:v1.0.0
```

Run with Docker Compose using the pre-built image:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Using Docker Compose for Development (Recommended)

The easiest way to run the gateway with Redis:

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f gateway

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

Make sure your `.env` file is configured before running docker-compose.

### Using Docker

Build the image:
```bash
docker build -t go-mcp-gateway .
```

Run with an external Redis instance:
```bash
docker run -d \
  --name go-mcp-gateway \
  -p 8080:8080 \
  -e BASE_URL=http://localhost:8080 \
  -e REDIS_ADDR=host.docker.internal:6379 \
  -e OAUTH_GOOGLE_CLIENT_ID=your-client-id \
  -e OAUTH_GOOGLE_CLIENT_SECRET=your-client-secret \
  -e OAUTH_GOOGLE_REDIRECT_URI=http://localhost:8080/oauth/callback \
  go-mcp-gateway
```

Run with Docker network and Redis container:
```bash
# Create network
docker network create mcp-network

# Start Redis
docker run -d \
  --name redis \
  --network mcp-network \
  redis:7-alpine

# Start gateway
docker run -d \
  --name go-mcp-gateway \
  --network mcp-network \
  -p 8080:8080 \
  -e BASE_URL=http://localhost:8080 \
  -e REDIS_ADDR=redis:6379 \
  -e OAUTH_GOOGLE_CLIENT_ID=your-client-id \
  -e OAUTH_GOOGLE_CLIENT_SECRET=your-client-secret \
  -e OAUTH_GOOGLE_REDIRECT_URI=http://localhost:8080/oauth/callback \
  go-mcp-gateway
```

## Usage

### Client Registration

Register your client application to receive OAuth credentials:

```bash
curl -X POST http://localhost:8080/oauth/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "My MCP Client",
    "redirect_uris": ["http://localhost:5000/callback"],
    "grant_types": ["authorization_code", "refresh_token"],
    "response_types": ["code"],
    "token_endpoint_auth_method": "client_secret_basic"
  }'
```

Response:
```json
{
  "client_id": "550e8400-e29b-41d4-a716-446655440000",
  "client_secret": "generated-secret",
  "client_name": "My MCP Client",
  "redirect_uris": ["http://localhost:5000/callback"],
  ...
}
```

### Authorization Flow

#### Step 1: Generate PKCE Code Verifier and Challenge

```python
import secrets
import hashlib
import base64

# Generate code verifier
code_verifier = base64.urlsafe_b64encode(secrets.token_bytes(32)).decode('utf-8').rstrip('=')

# Generate code challenge
code_challenge = base64.urlsafe_b64encode(
    hashlib.sha256(code_verifier.encode('utf-8')).digest()
).decode('utf-8').rstrip('=')
```

#### Step 2: Redirect User to Authorization Endpoint

```
http://localhost:8080/oauth/authorize?
  client_id=YOUR_CLIENT_ID&
  redirect_uri=http://localhost:5000/callback&
  response_type=code&
  scope=openid email profile&
  code_challenge=CODE_CHALLENGE&
  code_challenge_method=S256&
  state=random-state-value
```

#### Step 3: Exchange Authorization Code for Tokens

After the user authorizes, they'll be redirected to your `redirect_uri` with a code parameter:

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "CLIENT_ID:CLIENT_SECRET" \
  -d "grant_type=authorization_code&code=AUTHORIZATION_CODE&redirect_uri=http://localhost:5000/callback&code_verifier=CODE_VERIFIER"
```

Response:
```json
{
  "access_token": "ya29.a0AfH6SMBx...",
  "refresh_token": "1//0gw...",
  "token_type": "Bearer",
  "expires_in": 3599,
  "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6..."
}
```

#### Step 4: Access Protected MCP Resources

Use the Google access token to make requests to proxied MCP endpoints:

```bash
curl http://localhost:8080/calc/mcp \
  -H "Authorization: Bearer ya29.a0AfH6SMBx..." \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "add", "params": [1, 2], "id": 1}'
```

#### Step 5: Refresh Expired Tokens

When the access token expires, use the refresh token:

```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "CLIENT_ID:CLIENT_SECRET" \
  -d "grant_type=refresh_token&refresh_token=REFRESH_TOKEN"
```

## API Endpoints

### OAuth 2.0 Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/oauth/register` | POST | Dynamic client registration (RFC 7591) |
| `/oauth/authorize` | GET | Authorization endpoint - initiates OAuth flow |
| `/oauth/callback` | GET | Google OIDC callback handler |
| `/oauth/token` | POST | Token endpoint - exchange code for tokens or refresh tokens |

### Metadata Discovery Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/.well-known/oauth-authorization-server` | GET | Authorization server metadata (RFC 8414) |
| `/.well-known/oauth-protected-resource` | GET | Protected resource metadata |

### Proxied Routes

Routes are dynamically registered based on `config.yaml`. All proxied routes require a valid Google access token in the `Authorization` header.

## Development

### Running Tests

Run all tests:
```bash
go test ./...
```

Run tests for a specific package:
```bash
go test ./internal/auth
```

Run tests with verbose output:
```bash
go test -v ./...
```

### Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/                    # OAuth 2.0 authorization server
│   │   ├── auth.go
│   │   ├── client.go            # Dynamic client registration
│   │   ├── authorization_code.go
│   │   ├── token.go
│   │   └── ...
│   ├── handler/                 # HTTP request handlers
│   │   ├── oauth_authorize.go
│   │   ├── oauth_callback.go
│   │   ├── oauth_token.go
│   │   └── ...
│   ├── provider/google/         # Google OIDC integration
│   │   └── google.go
│   ├── middleware/              # HTTP middleware
│   │   └── googletokenvalidator/
│   ├── store/                   # Redis storage abstraction
│   ├── config/                  # Configuration management
│   ├── logger/                  # Structured logging
│   └── utils/                   # Utility functions
├── config.yaml                  # Proxy configuration
├── .env                         # Environment variables
└── go.mod
```

## Configuration Reference

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BASE_URL` | No | `http://localhost:8080` | Base URL of the server |
| `PORT` | No | `8080` | Server port |
| `REDIS_ADDR` | No | `localhost:6379` | Redis server address |
| `REDIS_PASSWORD` | No | | Redis password |
| `OAUTH_GOOGLE_CLIENT_ID` | Yes | | Google OAuth client ID |
| `OAUTH_GOOGLE_CLIENT_SECRET` | Yes | | Google OAuth client secret |
| `OAUTH_GOOGLE_REDIRECT_URI` | Yes | | OAuth callback URL |

### Proxy Configuration (config.yaml)

```yaml
proxies:
  - pattern: "/endpoint/path"    # URL pattern to match
    target_url: "http://host:port/path"  # Target MCP server URL
```

Multiple proxy routes can be defined. Each route will require Google token validation.

## Security Considerations

- **PKCE Required**: All authorization code flows must use PKCE with S256 method
- **No Long-term Storage**: Google tokens are only stored temporarily (5 minutes) during the OAuth exchange
- **Token Validation**: All proxied requests validate Google tokens with Google's tokeninfo endpoint
- **Redis Security**: Use strong Redis passwords in production and enable TLS
- **HTTPS**: Use HTTPS in production environments
- **Client Secrets**: Store client secrets securely, never commit to version control

## Storage TTLs

| Store Type | TTL | Purpose |
|------------|-----|---------|
| Authorization codes | 5 minutes | OAuth code exchange with embedded Google tokens |
| OAuth state/nonce | 5 minutes | Google OIDC flow validation |
| Client registrations | 90 days | Registered OAuth clients |
| Sessions | 7 days | User session management |

## Troubleshooting

### Redis Connection Error
Ensure Redis is running and accessible at the configured address:
```bash
redis-cli ping
```

### Google OAuth Error
- Verify your Google OAuth credentials are correct
- Ensure the redirect URI matches exactly what's configured in Google Cloud Console
- Check that required Google APIs are enabled

### Token Validation Failures
- Ensure the access token is not expired
- Verify the token was issued by Google for the correct client
- Check that the token has the required scopes

## Security

Security is a top priority for this project. Please review our security documentation:

- **[Security Policy](SECURITY.md)**: Comprehensive security guide and best practices
- **[Security Checklist](.github/SECURITY_CHECKLIST.md)**: Pre-publication security checklist
- **Reporting Vulnerabilities**: Use GitHub's private vulnerability reporting or contact the maintainers directly

### Key Security Features

- No long-term storage of Google credentials
- PKCE enforcement for OAuth flows
- Google token validation on all proxied requests
- Redis-backed session management with TTLs
- Automated dependency scanning via Dependabot
- CodeQL and Trivy security scanning

## Contributing

Contributions are welcome. Please follow the existing code style:
- Error handling with `fmt.Errorf` for context
- Structured logging via `internal/logger`
- PKCE S256 method exclusively
- Review [SECURITY.md](SECURITY.md) before contributing

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

### What does GPL-3.0 mean?

- ✅ You can use this software for any purpose
- ✅ You can modify the software
- ✅ You can distribute the software
- ✅ You can distribute your modifications
- ⚠️ If you distribute modified versions, you must:
  - Make the source code available
  - License it under GPL-3.0
  - State changes made to the code
  - Include the original copyright notice

This ensures the software remains free and open source.

## References

- [RFC 6749 - OAuth 2.0 Authorization Framework](https://tools.ietf.org/html/rfc6749)
- [RFC 7636 - PKCE](https://tools.ietf.org/html/rfc7636)
- [RFC 7591 - Dynamic Client Registration](https://tools.ietf.org/html/rfc7591)
- [RFC 8414 - Authorization Server Metadata](https://tools.ietf.org/html/rfc8414)
- [Model Context Protocol (MCP)](https://modelcontextprotocol.io/)

## Acknowledgments

Special thanks to:

- **[SecureMCP](https://github.com/securemcp)** - For inspiration and pioneering work in securing MCP implementations
- **[Claude Code](https://claude.ai/code)** - For invaluable assistance in development, documentation, and security best practices
