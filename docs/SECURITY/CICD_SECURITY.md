# CI/CD Security Documentation
# Implements PRD Section 4.15 - CI/CD Security

## Overview

This document describes the CI/CD security measures implemented in the Vyzorix Update Server pipeline.

## Security Pipeline

The security pipeline runs on every push and PR, consisting of:

1. **Secret Scanning** - Detect committed secrets
2. **SAST** - Static Application Security Testing
3. **Dependency Scanning** - Check for vulnerable dependencies
4. **Container Scanning** - Scan container images
5. **Supply Chain Security** - Verify artifact integrity

## Secret Scanning

### GitHub Secret Scanning

GitHub automatically scans for secret patterns in commits.

**Supported Secret Types:**
- AWS credentials
- Azure credentials
- GCP credentials
- GitHub tokens
- Stripe tokens
- Slack tokens
- And many more...

### TruffleHog

Extended secret scanning with custom patterns:

```bash
trufflehog filesystem . --no-update
```

## Static Application Security Testing (SAST)

### GoSec

Security analysis for Go code:

```bash
gosec -no-fail -fmt json -out=gosec-results.json ./...
```

**Common Issues Detected:**
- Hardcoded credentials
- SQL injection vulnerabilities
- Path traversal
- Command injection
- Insecure random numbers

### golangci-lint

Code quality and security linting:

```bash
golangci-lint run --out-format json > results.json
```

## Dependency Scanning

### Go Vulnerability Database

```bash
govulncheck ./...
```

### npm Audit

```bash
pnpm audit --audit-level=high
```

### Dependabot

Automated dependency updates configured in `.github/dependabot.yml`.

**Update Schedule:** Weekly on Sundays

## Container Security

### Dockerfile Best Practices

```dockerfile
# Use specific version tags
FROM golang:1.22-alpine3.19

# Set working directory
WORKDIR /build

# Copy dependency files first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s"

# Use distroless or scratch base
FROM gcr.io/distroless/static:latest
COPY --from=0 /build/vyzorix-server /vyzorix-server

# Run as non-root
USER nonroot:nonroot

ENTRYPOINT ["/vyzorix-server"]
```

### Trivy Scanning

Container vulnerability scanning:

```bash
trivy image --severity CRITICAL,HIGH --exit-code 1 vyzorix:latest
```

### Container Security Checksum Verification

```bash
# Generate checksum
sha256sum vyzorix.tar > vyzorix.tar.sha256

# Verify before deployment
sha256sum -c vyzorix.tar.sha256
```

## Signing Artifacts

### Binary Signing

```bash
# Generate key pair
openssl genrsa -out signing_key.pem 4096
openssl rsa -in signing_key.pem -pubout -out public_key.pem

# Sign binary
openssl dgst -sha256 -sign signing_key.pem -out binary.sig binary

# Verify signature
openssl dgst -sha256 -verify public_key.pem -signature binary.sig binary
```

### Container Image Signing

```bash
# Sign image with Cosign
cosign sign --yes ghcr.io/vinnsedigner/vyzorix:v1.0.0

# Verify image
cosign verify ghcr.io/vinnsedigner/vyzorix:v1.0.0
```

## Supply Chain Security

### SLSA Compliance

Implement SLSA (Supply-chain Levels for Software Artifacts):

| Level | Requirements |
|-------|--------------|
| L1 | Build process documented, tamper evident |
| L2 | Hosted build service, tamper resistant |
| L3 | Hardened build service, reproducible |
| L4 | Two-party review, hermetic build |

### SBOM Generation

```bash
# Generate SPDX SBOM
syft . -o spdx-json > sbom.spdx.json

# Generate CycloneDX SBOM  
cyclonedx-gomod -json -output-file sbom.cdx.json
```

## Environment-Specific Pipelines

### Development

- Quick lint and test
- No container scanning
- No artifact signing

### Staging

- Full security scan
- Container scanning
- Signed artifacts

### Production

- Full security scan
- Container scanning with approval
- Signed artifacts
- Provenance attestation

## Security Gates

All gates must pass before deployment:

| Gate | Threshold |
|------|-----------|
| SAST | 0 Critical, 0 High |
| Dependencies | 0 Critical |
| Container | 0 Critical, 0 High |
| Secrets | 0 findings |
| Test Coverage | > 70% |

## Incident Response

### Pipeline Failure Response

1. **Critical/High SAST Finding:**
   - Block deployment
   - Create security ticket
   - Notify security team

2. **Vulnerable Dependency:**
   - Update to patched version
   - If no patch, create exception with mitigation

3. **Secret Exposed:**
   - Immediately rotate secret
   - Revoke all instances
   - Notify affected parties

## Best Practices

1. **Least Privilege:** Use minimal permissions for CI/CD service accounts
2. **Secret Rotation:** Rotate secrets regularly (90 days)
3. **Branch Protection:** Require PR reviews before merge
4. **Audit Logs:** Keep CI/CD logs for 90+ days
5. **Immutable Artifacts:** Never modify artifacts after build

## References

- [OWASP Top 10 CI/CD Security](https://owasp.org/www-project-top-10-ci-cd-security/)
- [SLSA Framework](https://slsa.dev/)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [GitHub Security Best Practices](https://docs.github.com/en/code-security)
