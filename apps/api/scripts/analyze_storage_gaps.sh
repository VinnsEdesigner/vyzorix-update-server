#!/bin/bash
# Gap Analysis: pkg/storage vs internal/infrastructure/storage
# Shows which Store methods are MISSING in the new Repository pattern

set -e
cd "$(dirname "$0")/.."

echo "╔══════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    GAP ANALYSIS: pkg/storage vs internal/infrastructure/storage                    ║"
echo "╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 MISSING Store Methods (OLD → NEW comparison)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

MISSING_COUNT=0

# Function to check if a method exists in NEW
check_missing() {
    local old_method="$1"
    local new_file="$2"
    local method_name=$(echo "$old_method" | cut -d"(" -f1 | tr -d " ")
    
    if ! grep -q "$method_name" "$new_file" 2>/dev/null; then
        echo "  ❌ (s *Store) $old_method"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
}

# Check commands.go
echo "📦 commands.go:"
echo "  OLD methods:"
for method in $(grep "^func (s \*Store)" pkg/storage/commands.go | sed 's/func (s \*Store)//' | sed 's/(ctx/\n(/g' | grep -v "^$"); do
    if [ -n "$method" ]; then
        check_missing "$method" internal/infrastructure/storage/command.go
    fi
done
grep "^func (s \*Store)" pkg/storage/commands.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/command.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

# Check devices.go  
echo "📦 devices.go:"
grep "^func (s \*Store)" pkg/storage/devices.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/device.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

# Check operators.go
echo "📦 operators.go:"
grep "^func (s \*Store)" pkg/storage/operators.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/operator.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

# Check sessions.go
echo "📦 sessions.go:"
grep "^func (s \*Store)" pkg/storage/sessions.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/session.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

# Check settings.go
echo "📦 settings.go:"
grep "^func (s \*Store)" pkg/storage/settings.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/operator.go 2>/dev/null; then
        # Check if it belongs to email_verification or password_reset
        if echo "$line" | grep -qi "email\|verification"; then
            if ! grep -q "$method" internal/infrastructure/storage/email_verification.go 2>/dev/null; then
                echo "  ❌ $line"
                MISSING_COUNT=$((MISSING_COUNT + 1))
            fi
        elif echo "$line" | grep -qi "password\|reset\|resend"; then
            if ! grep -q "$method" internal/infrastructure/storage/password_reset.go 2>/dev/null; then
                echo "  ❌ $line"
                MISSING_COUNT=$((MISSING_COUNT + 1))
            fi
        else
            echo "  ❌ $line"
            MISSING_COUNT=$((MISSING_COUNT + 1))
        fi
    fi
done
echo ""

# Check telemetry.go
echo "📦 telemetry.go:"
grep "^func (s \*Store)" pkg/storage/telemetry.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/telemetry.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

# Check clients.go
echo "📦 clients.go:"
grep "^func (s \*Store)" pkg/storage/clients.go | sed 's/func (s \*Store) /  /' | while read line; do
    method=$(echo "$line" | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/client.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 MISSING Standalone Functions"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Check standalone functions
echo "📦 crypto.go - Standalone functions:"
grep "^func Hash\|^func Verify\|^func Generate\|^func Decode" pkg/storage/crypto.go | while read line; do
    method=$(echo "$line" | sed 's/func //' | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/auth/argon2_hasher.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

echo "📦 clients.go - Standalone functions:"
grep "^func Hash\|^func Derive\|^func Verify" pkg/storage/clients.go | while read line; do
    method=$(echo "$line" | sed 's/func //' | cut -d"(" -f1 | tr -d " ")
    if ! grep -q "$method" internal/infrastructure/storage/client.go 2>/dev/null; then
        echo "  ❌ $line"
        MISSING_COUNT=$((MISSING_COUNT + 1))
    fi
done
echo ""

echo "📦 uuid.go - Functions (already migrated to internal/infrastructure/uuid/):"
echo "  ✅ NewUUIDv7 → internal/infrastructure/uuid/New()"
echo "  ✅ ParseUUIDv7 → internal/infrastructure/uuid/Parse()"
echo "  ✅ IsUUIDv7 → internal/infrastructure/uuid/IsValid()"
echo "  ✅ UUIDv7ToTime → internal/infrastructure/uuid/ExtractTime()"
echo ""

echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
echo "📊 SUMMARY"
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
echo ""
echo "  Missing Store methods/functions: ~$MISSING_COUNT"
echo ""
echo "  NOTE: This shows logic in OLD pkg/storage that is NOT in NEW internal/infrastructure/storage"
echo "  The NEW infrastructure uses a Repository pattern (interface-based) instead of Store methods."
echo ""
if [ $MISSING_COUNT -gt 20 ]; then
    echo "  ⚠️  SIGNIFICANT GAPS - Major migration work needed!"
elif [ $MISSING_COUNT -gt 5 ]; then
    echo "  ⚠️  MODERATE GAPS - Some migration work needed"
else
    echo "  ✅ MINOR GAPS - Mostly migrated"
fi
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
