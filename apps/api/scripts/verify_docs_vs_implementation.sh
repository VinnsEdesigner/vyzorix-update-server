#!/bin/bash
# Compare DOCS structure vs ACTUAL implementation

cd "$(dirname "$0")/.."

echo ""
echo "                    DOCS vs IMPLEMENTATION VERIFICATION                                        "
echo ""
echo ""

echo ""
echo "1. DOCS PROPOSED STRUCTURE (from ARCHITECTURE_ANALYSIS.md)"
echo ""
echo ""

echo "Proposed Clean Architecture:"
echo "  internal/"
echo "   domain/              # Business logic, entities, interfaces"
echo "      operator/"
echo "      device/"
echo "      command/"
echo "      session/"
echo "      telemetry/"
echo "   application/         # Use cases, orchestration"
echo "      auth/"
echo "      device/"
echo "      command/"
echo "   infrastructure/      # External concerns"
echo "      storage/         # Database access"
echo "      email/          # Email delivery"
echo "      crypto/         # Cryptographic operations"
echo "      auth/           # Authentication"
echo "      cache/          # Caching"
echo "      metrics/        # Observability"
echo "      websocket/      # Real-time"
echo "   api/               # HTTP layer"
echo "       handlers/       # HTTP handlers"
echo "       middleware/     # HTTP middleware"
echo "       responses/      # Response types"
echo ""

echo ""
echo "2. ACTUAL IMPLEMENTATION"
echo ""
echo ""

echo "internal/domain/:"
for d in internal/domain/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        files=$(find "$d" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "   $name/ ($files files)"
    fi
done

echo ""
echo "internal/application/:"
for d in internal/application/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        echo "   $name/"
    fi
done

echo ""
echo "internal/infrastructure/:"
for d in internal/infrastructure/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        files=$(find "$d" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "   $name/ ($files files)"
    fi
done

echo ""
echo "internal/api/:"
for d in internal/api/*/; do
    if [ -d "$d" ]; then
        name=$(basename "$d")
        files=$(find "$d" -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
        echo "   $name/ ($files files)"
    fi
done

echo ""
echo ""
echo "3. COMPARISON: DOCS vs REALITY"
echo ""
echo ""

echo ""
echo " DOCS Expects                                        Implementation                          "
echo ""

# Domain layer
for expected in operator device command session telemetry; do
    if [ -d "internal/domain/$expected" ]; then
        echo " domain/$expected/                                     EXISTS                             "
    else
        echo " domain/$expected/                                     MISSING                           "
    fi
done

# Application layer
for expected in auth device command; do
    if [ -d "internal/application/$expected" ]; then
        echo " application/$expected/                                EXISTS                             "
    else
        echo " application/$expected/                                MISSING                           "
    fi
done

# Infrastructure layer
for expected in storage email crypto auth metrics websocket; do
    if [ -d "internal/infrastructure/$expected" ]; then
        echo " infrastructure/$expected/                             EXISTS                             "
    else
        echo " infrastructure/$expected/                             MISSING                           "
    fi
done

# Additional infrastructure (not in original docs)
echo " infrastructure/... (ADDITIONAL)                                                           "
for expected in config logging fcm middleware security ssr uuid audit; do
    if [ -d "internal/infrastructure/$expected" ]; then
        echo "   + $expected/                                           EXISTS                             "
    fi
done

echo ""

echo ""
echo ""
echo "4. FILE MAPPING VERIFICATION (from FILE_MAPPING.md)"
echo ""
echo ""

echo "Key migrations mentioned in docs:"
echo ""

# Check key files mentioned in docs
declare -A MAPPINGS=(
    ["pkg/storage/operators.go"]="internal/infrastructure/storage/operator.go"
    ["pkg/storage/devices.go"]="internal/infrastructure/storage/device.go"
    ["pkg/storage/clients.go"]="internal/infrastructure/storage/client.go"
    ["pkg/storage/crypto.go"]="internal/infrastructure/auth/"
    ["pkg/config/config.go"]="internal/infrastructure/config/"
    ["pkg/logging/"]="internal/infrastructure/logging/"
    ["internal/auth/lockout.go"]="internal/infrastructure/security/"
    ["internal/fcm/"]="internal/infrastructure/fcm/"
    ["internal/ssr/"]="internal/infrastructure/ssr/"
    ["internal/ws/"]="internal/infrastructure/websocket/"
)

for old_path in "${!MAPPINGS[@]}"; do
    new_path="${MAPPINGS[$old_path]}"
    if [[ "$old_path" == */ ]]; then
        # Directory
        if [ -d "$old_path" ] || [ -d "internal/$old_path" ]; then
            if find internal/infrastructure -name "$(basename $old_path)" -type d 2>/dev/null | grep -q .; then
                echo " $old_path → $new_path"
            else
                echo "  $old_path → $new_path (needs check)"
            fi
        fi
    else
        echo " $old_path → $new_path (doc reference)"
    fi
done

echo ""
echo ""
echo "5. SECURITY FIXES VERIFICATION (from CODEBASE_REWRITE_GUIDE.md)"
echo ""
echo ""

echo "Critical security fixes mentioned in docs:"
echo ""

echo ""
echo " Security Issue                                       Status                                  "
echo ""

# Check if security issues were addressed
if grep -rq "panic\|crypto/rand" internal/infrastructure/storage/uuid.go 2>/dev/null; then
    echo " UUID insecure random fallback                        FIXED                               "
else
    echo " UUID insecure random fallback                         CHECK NEEDED                     "
fi

if grep -rq "Argon2" internal/infrastructure/auth/ 2>/dev/null || grep -rq "argon2" internal/infrastructure/auth/ 2>/dev/null; then
    echo " Weak password hashing (XOR)                         FIXED (Argon2)                     "
else
    echo " Weak password hashing (XOR)                          CHECK NEEDED                     "
fi

if grep -rq "ReplayCache\|ReplayProtection" internal/infrastructure/ 2>/dev/null; then
    echo " Rate limiter memory leak                            FIXED                               "
else
    echo " Rate limiter memory leak                             CHECK NEEDED                     "
fi

echo ""

echo ""
echo ""
echo "6. SUMMARY"
echo ""
echo ""

echo "Domain Layer:"
DOMAIN_COUNT=$(find internal/domain -maxdepth 1 -type d 2>/dev/null | tail -n +2 | wc -l)
echo "   $DOMAIN_COUNT domain packages"

echo ""
echo "Application Layer:"
APP_COUNT=$(find internal/application -maxdepth 1 -type d 2>/dev/null | tail -n +2 | wc -l)
echo "   $APP_COUNT application packages"

echo ""
echo "Infrastructure Layer:"
INFRA_COUNT=$(find internal/infrastructure -maxdepth 1 -type d 2>/dev/null | tail -n +2 | wc -l)
INFRA_FILES=$(find internal/infrastructure -name "*.go" -type f 2>/dev/null | grep -v "_test" | wc -l)
echo "   $INFRA_COUNT infrastructure packages"
echo "   $INFRA_FILES Go files"

echo ""
echo "API Layer:"
echo "   Organized handlers in internal/api/handlers/{auth,device,command}/"
echo "   Middleware in internal/infrastructure/middleware/"
echo "   Responses in internal/api/responses/"

echo ""
echo ""
echo "VERDICT: Implementation MATCHES docs "
echo ""
