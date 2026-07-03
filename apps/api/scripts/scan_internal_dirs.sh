#!/bin/bash
# Scan internal/ directories for logic to migrate

cd "$(dirname "$0")/.."

echo "╔══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    SCANNING internal/ DIRECTORIES FOR MIGRATION                                       ║"
echo "╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. internal/audit/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
if [ -d "internal/audit" ]; then
    echo "Files:"
    for f in internal/audit/*.go; do
        if [ -f "$f" ]; then
            funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
            types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
            echo "  $(basename $f): $types types, $funcs funcs"
        fi
    done
    echo ""
    echo "Logic:"
    grep -E "^func |^type " internal/audit/*.go 2>/dev/null | grep -v "_test" | head -20
else
    echo "  Directory not found"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. internal/fcm/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
if [ -d "internal/fcm" ]; then
    echo "Files:"
    for f in internal/fcm/*.go; do
        if [ -f "$f" ]; then
            funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
            types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
            echo "  $(basename $f): $types types, $funcs funcs"
        fi
    done
    echo ""
    echo "Logic:"
    grep -E "^func |^type " internal/fcm/*.go 2>/dev/null | grep -v "_test" | head -20
else
    echo "  Directory not found"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. internal/ssr/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
if [ -d "internal/ssr" ]; then
    echo "Files:"
    for f in internal/ssr/*.go; do
        if [ -f "$f" ]; then
            funcs=$(grep -c "^func " "$f" 2>/dev/null || echo "0")
            types=$(grep -c "^type " "$f" 2>/dev/null || echo "0")
            echo "  $(basename $f): $types types, $funcs funcs"
        fi
    done
    echo ""
    echo "Logic:"
    grep -E "^func |^type " internal/ssr/*.go 2>/dev/null | grep -v "_test" | head -20
else
    echo "  Directory not found"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. internal/api/server.go"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
if [ -f "internal/api/server.go" ]; then
    echo "Logic:"
    grep -E "^func |^type " internal/api/server.go 2>/dev/null | head -20
else
    echo "  File not found"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. internal/email.go & internal/command_signer.go"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "internal/email.go:"
if [ -f "internal/email.go" ]; then
    funcs=$(grep -c "^func " internal/email.go 2>/dev/null || echo "0")
    types=$(grep -c "^type " internal/email.go 2>/dev/null || echo "0")
    echo "  $types types, $funcs funcs"
fi

echo ""
echo "internal/command_signer.go:"
if [ -f "internal/command_signer.go" ]; then
    funcs=$(grep -c "^func " internal/command_signer.go 2>/dev/null || echo "0")
    types=$(grep -c "^type " internal/command_signer.go 2>/dev/null || echo "0")
    echo "  $types types, $funcs funcs"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. MIGRATION STATUS"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "┌────────────────────────────────────────────────┬────────────────────────────────────────┐"
echo "│ Directory                                        │ Status                                  │"
echo "├────────────────────────────────────────────────┼────────────────────────────────────────┤"

# Check audit
if [ -d "internal/infrastructure/audit" ]; then
    echo "│ internal/audit/                                │ ✅ Migrated to infrastructure/audit/ │"
else
    echo "│ internal/audit/                                │ ⚠️  Needs migration                   │"
fi

# Check fcm
if [ -d "internal/infrastructure/fcm" ]; then
    echo "│ internal/fcm/                                  │ ✅ Migrated to infrastructure/fcm/    │"
else
    echo "│ internal/fcm/                                  │ ⚠️  Needs migration                   │"
fi

# Check ssr
if [ -d "internal/infrastructure/ssr" ]; then
    echo "│ internal/ssr/                                  │ ✅ Migrated to infrastructure/ssr/    │"
else
    echo "│ internal/ssr/                                  │ ⚠️  Needs migration                   │"
fi

# Check email
if [ -d "internal/infrastructure/email" ]; then
    echo "│ internal/email.go                              │ ✅ Migrated to infrastructure/email/ │"
else
    echo "│ internal/email.go                              │ ⚠️  Needs migration                   │"
fi

# Check command_signer
if grep -rq "CommandSigner" internal/infrastructure/crypto/ 2>/dev/null; then
    echo "│ internal/command_signer.go                     │ ✅ Migrated to infrastructure/crypto/ │"
else
    echo "│ internal/command_signer.go                     │ ⚠️  Needs migration                   │"
fi

echo "└────────────────────────────────────────────────┴────────────────────────────────────────┘"

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
