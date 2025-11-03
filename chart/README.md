# go-mcp-gateway Helm Chart

This Helm chart deploys the go-mcp-gateway OAuth 2.0 authorization facilitator and reverse proxy gateway for MCP servers.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PV provisioner support in the underlying infrastructure (for Redis persistence)

## Installing the Chart

### From GitHub Container Registry (Recommended)

```bash
# Install latest version
helm install my-gateway oci://ghcr.io/schnurbus/go-mcp-gateway \
  --set config.oauth.google.clientId=YOUR_CLIENT_ID \
  --set config.oauth.google.clientSecret=YOUR_CLIENT_SECRET

# Install specific version
helm install my-gateway oci://ghcr.io/schnurbus/go-mcp-gateway \
  --version 0.1.0 \
  --set config.oauth.google.clientId=YOUR_CLIENT_ID \
  --set config.oauth.google.clientSecret=YOUR_CLIENT_SECRET
```

### From Local Chart (Development)

```bash
helm install my-gateway ./chart \
  --set config.oauth.google.clientId=YOUR_CLIENT_ID \
  --set config.oauth.google.clientSecret=YOUR_CLIENT_SECRET \
  --set config.oauth.google.redirectUri=http://localhost:8080/oauth/callback
```

### Production Installation

1. Create a `values-prod.yaml` file:

```yaml
# values-prod.yaml
replicaCount: 2

image:
  repository: ghcr.io/schnurbus/go-mcp-gateway
  tag: "v0.1.0"

config:
  baseUrl: "https://mcp-gateway.example.com"
  oauth:
    google:
      clientId: "YOUR_CLIENT_ID"
      clientSecret: "YOUR_CLIENT_SECRET"
      redirectUri: "https://mcp-gateway.example.com/oauth/callback"
  proxies:
    - pattern: "/calc/mcp"
      targetUrl: "http://calc-mcp-server:3000/mcp"

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: mcp-gateway.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: mcp-gateway-tls
      hosts:
        - mcp-gateway.example.com

redis:
  master:
    persistence:
      enabled: true
      size: 5Gi
      storageClass: "standard"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80
```

2. Install the chart:

```bash
helm install my-gateway ./chart -f values-prod.yaml
```

## Uninstalling the Chart

```bash
helm uninstall my-gateway
```

## Configuration

The following table lists the configurable parameters and their default values.

### Application Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `ghcr.io/schnurbus/go-mcp-gateway` |
| `image.tag` | Image tag | `""` (uses chart appVersion) |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `config.baseUrl` | Public base URL | `http://localhost:8080` |
| `config.port` | Server port | `8080` |
| `config.allowedOrigins` | Allowed CORS origins (comma-separated) | `"*"` |
| `config.oauth.google.clientId` | Google OAuth client ID | `""` |
| `config.oauth.google.clientSecret` | Google OAuth client secret | `""` |
| `config.oauth.google.redirectUri` | OAuth redirect URI | `http://localhost:8080/oauth/callback` |
| `config.oauth.google.scopes` | OAuth scopes (comma-separated) | `"openid,profile,email"` |
| `config.proxies` | Array of proxy configurations | See values.yaml |

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `service.annotations` | Service annotations | `{}` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | Ingress TLS configuration | `[]` |

### Redis Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `redis.enabled` | Enable built-in Redis | `true` |
| `redis.external.enabled` | Use external Redis | `false` |
| `redis.external.host` | External Redis host | `""` |
| `redis.external.port` | External Redis port | `6379` |
| `redis.external.password` | External Redis password | `""` |
| `redis.auth.enabled` | Enable Redis authentication | `true` |
| `redis.auth.password` | Redis password (auto-generated if empty) | `""` |
| `redis.master.persistence.enabled` | Enable Redis persistence | `true` |
| `redis.master.persistence.size` | Redis PVC size | `1Gi` |
| `redis.master.persistence.storageClass` | Redis storage class | `""` |
| `redis.resources` | Redis resource limits | See values.yaml |

### Autoscaling Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `autoscaling.enabled` | Enable HPA | `false` |
| `autoscaling.minReplicas` | Minimum replicas | `1` |
| `autoscaling.maxReplicas` | Maximum replicas | `10` |
| `autoscaling.targetCPUUtilizationPercentage` | Target CPU % | `80` |

## Examples

### Using External Redis

```yaml
redis:
  enabled: true
  external:
    enabled: true
    host: "redis.example.com"
    port: 6379
    password: "your-redis-password"
```

### Adding Multiple Proxy Routes

```yaml
config:
  proxies:
    - pattern: "/calc/mcp"
      targetUrl: "http://calc-server:3000/mcp"
    - pattern: "/files/mcp"
      targetUrl: "http://file-server:3001/mcp"
    - pattern: "/email/mcp"
      targetUrl: "http://email-server:3002/mcp"
```

### Enabling TLS with cert-manager

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
  hosts:
    - host: mcp-gateway.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: mcp-gateway-tls
      hosts:
        - mcp-gateway.example.com
```

### High Availability Setup

```yaml
replicaCount: 3

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app.kubernetes.io/name
                operator: In
                values:
                  - go-mcp-gateway
          topologyKey: kubernetes.io/hostname

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

redis:
  master:
    persistence:
      enabled: true
      size: 10Gi
      storageClass: "fast-ssd"
```

## Upgrading

### Upgrade with New Values

```bash
helm upgrade my-gateway ./chart -f values-prod.yaml
```

### Upgrade with CLI Parameters

```bash
helm upgrade my-gateway ./chart \
  --set image.tag=v0.2.0 \
  --set replicaCount=3
```

### Rollback

```bash
helm rollback my-gateway 1
```

## Security Considerations

### Secrets Management

For production, consider using external secret management:

1. **Sealed Secrets**:
```bash
kubectl create secret generic mcp-gateway-oauth \
  --from-literal=oauth-google-client-id=YOUR_ID \
  --from-literal=oauth-google-client-secret=YOUR_SECRET \
  --dry-run=client -o yaml | \
  kubeseal -o yaml > sealed-secret.yaml
```

2. **External Secrets Operator**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: go-mcp-gateway
spec:
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: go-mcp-gateway
  data:
    - secretKey: oauth-google-client-id
      remoteRef:
        key: mcp-gateway/oauth
        property: clientId
    - secretKey: oauth-google-client-secret
      remoteRef:
        key: mcp-gateway/oauth
        property: clientSecret
```

### Network Policies

Create a NetworkPolicy to restrict traffic:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: go-mcp-gateway
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: go-mcp-gateway
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app.kubernetes.io/component: redis
    ports:
    - protocol: TCP
      port: 6379
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # For Google OAuth
```

## Monitoring

### Prometheus ServiceMonitor

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: go-mcp-gateway
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: go-mcp-gateway
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=go-mcp-gateway
```

### View Logs

```bash
kubectl logs -f deployment/my-gateway-go-mcp-gateway
```

### Check Redis Connection

```bash
kubectl exec -it my-gateway-go-mcp-gateway-redis-0 -- redis-cli ping
```

### Debug OAuth Issues

```bash
# Port forward to test locally
kubectl port-forward svc/my-gateway-go-mcp-gateway 8080:8080

# Test OAuth metadata endpoint
curl http://localhost:8080/.well-known/oauth-authorization-server
```

### Check ConfigMap

```bash
kubectl get configmap my-gateway-go-mcp-gateway -o yaml
```

## Support

- GitHub: https://github.com/schnurbus/go-mcp-gateway
- Issues: https://github.com/schnurbus/go-mcp-gateway/issues
- Documentation: https://github.com/schnurbus/go-mcp-gateway/blob/main/README.md
