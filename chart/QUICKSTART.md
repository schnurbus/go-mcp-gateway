# Helm Chart Quick Start

## Prerequisites

- Kubernetes cluster (1.19+)
- Helm 3.0+
- kubectl configured
- Google OAuth credentials

## 1. Quick Install (Development)

```bash
# Install with minimal configuration
helm install my-gateway ./chart \
  --set config.oauth.google.clientId=YOUR_CLIENT_ID \
  --set config.oauth.google.clientSecret=YOUR_CLIENT_SECRET

# Check status
helm status my-gateway

# Get pods
kubectl get pods -l app.kubernetes.io/name=go-mcp-gateway

# View logs
kubectl logs -f -l app.kubernetes.io/name=go-mcp-gateway

# Port forward to test locally
kubectl port-forward svc/my-gateway-go-mcp-gateway 8080:8080
```

Access the application at http://localhost:8080

## 2. Production Install

### Step 1: Create Values File

```bash
cp chart/values-example.yaml values-prod.yaml
```

Edit `values-prod.yaml` and set:
- Your domain name
- Allowed CORS origins
- Google OAuth credentials
- MCP server proxy routes
- Resource limits
- Storage settings

### Step 2: Install

```bash
helm install my-gateway ./chart \
  -f values-prod.yaml \
  --namespace mcp-gateway \
  --create-namespace
```

### Step 3: Verify

```bash
# Check all resources
kubectl get all -n mcp-gateway

# Check ingress
kubectl get ingress -n mcp-gateway

# Check secrets
kubectl get secrets -n mcp-gateway

# Test OAuth endpoint
curl https://your-domain.com/.well-known/oauth-authorization-server
```

## 3. Common Operations

### Upgrade

```bash
# Upgrade with new values
helm upgrade my-gateway ./chart -f values-prod.yaml

# Upgrade to new image version
helm upgrade my-gateway ./chart \
  --set image.tag=v0.2.0 \
  --reuse-values
```

### Rollback

```bash
# List revisions
helm history my-gateway

# Rollback to previous version
helm rollback my-gateway

# Rollback to specific revision
helm rollback my-gateway 2
```

### Uninstall

```bash
# Uninstall (keeps PVCs)
helm uninstall my-gateway

# Delete PVCs
kubectl delete pvc -l app.kubernetes.io/name=go-mcp-gateway
```

## 4. Configuration Examples

### Enable Ingress with TLS

```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
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

### Use External Redis

```yaml
redis:
  enabled: true
  external:
    enabled: true
    host: "redis.example.com"
    port: 6379
    password: "your-password"
```

### Enable Autoscaling

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 75
```

### Add Multiple Proxy Routes

```yaml
config:
  proxies:
    - pattern: "/calc/mcp"
      targetUrl: "http://calc-server:3000/mcp"
    - pattern: "/files/mcp"
      targetUrl: "http://files-server:3001/mcp"
    - pattern: "/email/mcp"
      targetUrl: "http://email-server:3002/mcp"
```

## 5. Troubleshooting

### Pod Not Starting

```bash
# Check pod status
kubectl get pods -l app.kubernetes.io/name=go-mcp-gateway

# Describe pod
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>

# Check events
kubectl get events --sort-by='.lastTimestamp'
```

### OAuth Errors

```bash
# Check secret
kubectl get secret my-gateway-go-mcp-gateway -o yaml

# Update OAuth credentials
kubectl create secret generic my-gateway-go-mcp-gateway \
  --from-literal=oauth-google-client-id=NEW_ID \
  --from-literal=oauth-google-client-secret=NEW_SECRET \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods
kubectl rollout restart deployment my-gateway-go-mcp-gateway
```

### Redis Connection Issues

```bash
# Check Redis pod
kubectl get pods -l app.kubernetes.io/component=redis

# Test Redis connection
kubectl exec -it my-gateway-go-mcp-gateway-redis-0 -- redis-cli ping

# Check Redis password
kubectl get secret my-gateway-go-mcp-gateway-redis -o jsonpath='{.data.redis-password}' | base64 -d
```

### Ingress Not Working

```bash
# Check ingress
kubectl get ingress

# Describe ingress
kubectl describe ingress my-gateway-go-mcp-gateway

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx

# Test without ingress
kubectl port-forward svc/my-gateway-go-mcp-gateway 8080:8080
curl http://localhost:8080/.well-known/oauth-authorization-server
```

## 6. Monitoring

### Check Application Health

```bash
# Check liveness/readiness probes
kubectl describe pod <pod-name> | grep -A 10 "Liveness\|Readiness"

# Test health endpoint directly
kubectl exec <pod-name> -- wget -qO- http://localhost:8080/.well-known/oauth-authorization-server
```

### View Metrics

```bash
# If Prometheus is installed
kubectl port-forward svc/prometheus 9090:9090

# View metrics in browser: http://localhost:9090
```

### Check Resource Usage

```bash
# Top pods
kubectl top pods -l app.kubernetes.io/name=go-mcp-gateway

# Top nodes
kubectl top nodes

# Describe HPA (if enabled)
kubectl describe hpa my-gateway-go-mcp-gateway
```

## 7. Security

### Update Secrets

```bash
# Create new secret from file
kubectl create secret generic my-gateway-go-mcp-gateway \
  --from-file=oauth-google-client-id=./client-id.txt \
  --from-file=oauth-google-client-secret=./client-secret.txt \
  --dry-run=client -o yaml | kubectl apply -f -

# Or from literals
kubectl create secret generic my-gateway-go-mcp-gateway \
  --from-literal=oauth-google-client-id=NEW_ID \
  --from-literal=oauth-google-client-secret=NEW_SECRET \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Apply Network Policy

```bash
# Restrict pod network access
kubectl apply -f - <<EOF
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
  - ports:
    - protocol: TCP
      port: 443  # For Google OAuth
EOF
```

## 8. Backup and Restore

### Backup Redis Data

```bash
# Create Redis backup
kubectl exec my-gateway-go-mcp-gateway-redis-0 -- redis-cli BGSAVE

# Copy backup file
kubectl cp my-gateway-go-mcp-gateway-redis-0:/data/dump.rdb ./redis-backup.rdb
```

### Restore Redis Data

```bash
# Copy backup to pod
kubectl cp ./redis-backup.rdb my-gateway-go-mcp-gateway-redis-0:/data/dump.rdb

# Restart Redis
kubectl delete pod my-gateway-go-mcp-gateway-redis-0
```

## 9. Useful Commands Reference

```bash
# Get all resources
helm list
kubectl get all -l app.kubernetes.io/name=go-mcp-gateway

# Check configuration
helm get values my-gateway
helm get manifest my-gateway

# Watch pods
kubectl get pods -w -l app.kubernetes.io/name=go-mcp-gateway

# Stream logs from all pods
kubectl logs -f -l app.kubernetes.io/name=go-mcp-gateway --all-containers=true

# Execute command in pod
kubectl exec -it <pod-name> -- /bin/sh

# Copy files from pod
kubectl cp <pod-name>:/app/config.yaml ./config.yaml

# Test OAuth flow
curl http://localhost:8080/.well-known/oauth-authorization-server
curl http://localhost:8080/oauth/register -X POST -H "Content-Type: application/json" -d '{...}'
```

## Need More Help?

- Full documentation: [chart/README.md](README.md)
- Main README: [../README.md](../README.md)
- Security guide: [../SECURITY.md](../SECURITY.md)
- Issues: https://github.com/schnurbus/go-mcp-gateway/issues
