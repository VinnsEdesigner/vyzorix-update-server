# Container Hardening Policy
# Implements PRD Section 4.14 - Container Hardening

## Overview

This document outlines the container security hardening measures implemented for the Vyzorix Update Server production deployment.

## 1. Non-Root User (Required)

All containers must run as a non-root user to prevent privilege escalation attacks.

```dockerfile
# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /home/appuser
```

## 2. Read-Only Root Filesystem (Required)

Containers should run with a read-only root filesystem to prevent unauthorized modifications.

```dockerfile
# Use read-only root filesystem
VOLUME ["/tmp", "/var/run"]
```

**Runtime enforcement:**
```yaml
securityContext:
  readOnlyRootFilesystem: true
  tmpOwnership: "0777"
```

## 3. No Privileged Mode (Required)

Containers must never run in privileged mode.

**Runtime enforcement:**
```yaml
securityContext:
  privileged: false
  allowPrivilegeEscalation: false
```

## 4. Resource Limits (Required)

All containers must have explicit resource limits to prevent denial-of-service attacks.

```yaml
resources:
  limits:
    cpu: "500m"
    memory: "512Mi"
  requests:
    cpu: "100m"
    memory: "128Mi"
```

## 5. Dropped Capabilities (Required)

Containers should drop all capabilities and only add those explicitly required.

```dockerfile
# Drop all capabilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# If using suid binaries, use --strip-components to remove them
RUN apt-get install -y --no-install-recommends \
    && dpkg --strip-components=1 /path/to/suid-binary
```

**Kubernetes enforcement:**
```yaml
securityContext:
  capabilities:
    drop:
      - ALL
```

## 6. Health Checks (Required)

All containers must implement health checks for orchestration readiness.

```dockerfile
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health/live || exit 1
```

## 7. Secret Management (Required)

Never bake secrets into images. Use Kubernetes secrets or external secret managers.

```yaml
# Environment variables from secrets (NOT recommended for highly sensitive data)
envFrom:
  - secretRef:
      name: app-secrets

# Better: Use mounted secrets
volumeMounts:
  - name: secrets
    mountPath: "/secrets"
    readOnly: true
```

## 8. Network Policies (Required)

Implement network segmentation to limit lateral movement.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-network-policy
spec:
  podSelector:
    matchLabels:
      app: vyzorix-api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: vyzorix-web
      ports:
        - port: 3000
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: vyzorix-db
      ports:
        - port: 5432
```

## 9. Image Scanning (Required)

All base images and dependencies must be scanned for vulnerabilities.

**CI/CD Integration:**
```yaml
- name: Trivy Vulnerability Scanner
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'vyzorix:latest'
    format: 'sarif'
    severity: 'CRITICAL,HIGH'
    exit-code: '1'
```

## 10. Secure Base Images

Use minimal, purpose-built base images with verified provenance.

**Recommended base images:**
- `gcr.io/distroless/static` - Minimal, no shell
- `cgr.dev/chainguard/static` - Built by Chainguard
- `redhat/ubi9-minimal` - Red Hat Universal Base Image

**NOT Recommended:**
- `ubuntu:latest` - Too large, frequent updates
- `debian:latest` - No vulnerability scanning by default
- `alpine:latest` - Use specific version tag

## Checklist

- [ ] Container runs as non-root user
- [ ] Root filesystem is read-only
- [ ] Privileged mode is disabled
- [ ] Resource limits are set
- [ ] Capabilities are dropped
- [ ] Health checks are implemented
- [ ] Secrets are externalized
- [ ] Network policies are configured
- [ ] Image scanning is enabled
- [ ] Base image is verified and minimal

## References

- [NIST Container Security Guide](https://nvd.nist.gov/800-53/Rev5/control/CV-4)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [Kubernetes Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
