#!/bin/bash
# Scan what imports OLD internal/ packages

cd "$(dirname "$0")/.."

echo "╔════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "║                    SCANNING: Files Using OLD internal/ Packages                                      ║"
echo "╚════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1. FILES IMPORTING OLD internal/auth (JWT, OAuth, etc)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/internal/auth:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/internal/auth" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2. FILES IMPORTING OLD internal/ws"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/internal/ws:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/internal/ws" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3. FILES IMPORTING OLD internal/audit"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/internal/audit:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/internal/audit" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4. FILES IMPORTING OLD internal/fcm"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/internal/fcm:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/internal/fcm" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5. FILES IMPORTING OLD internal/ssr"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/internal/ssr:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/internal/ssr" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6. FILES IMPORTING OLD internal/metrics"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/internal/metrics:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/internal/metrics" 2>/dev/null | grep -v "_test" | head -20

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "7. FILES IMPORTING OLD pkg/"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "Files importing vyzorix/apps/api/pkg:"
find . -name "*.go" -type f 2>/dev/null | xargs grep -l "vyzorix/apps/api/pkg" 2>/dev/null | grep -v "_test" | wc -l
echo "total files"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "8. MIGRATION MAP"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "┌────────────────────────────────────────────────┬────────────────────────────────────────────────┐"
echo "│ OLD Import                                          │ NEW Import                                     │"
echo "├────────────────────────────────────────────────┼────────────────────────────────────────────────┤"
echo "│ vyzorix/apps/api/internal/auth                   │ vyzorix/apps/api/internal/infrastructure/security │"
echo "│ vyzorix/apps/api/internal/ws                    │ vyzorix/apps/api/internal/infrastructure/websocket │"
echo "│ vyzorix/apps/api/internal/audit                  │ vyzorix/apps/api/internal/infrastructure/audit │"
echo "│ vyzorix/apps/api/internal/fcm                   │ vyzorix/apps/api/internal/infrastructure/fcm  │"
echo "│ vyzorix/apps/api/internal/ssr                   │ vyzorix/apps/api/internal/infrastructure/ssr  │"
echo "│ vyzorix/apps/api/internal/metrics               │ vyzorix/apps/api/internal/infrastructure/metrics │"
echo "│ vyzorix/apps/api/pkg/config                     │ vyzorix/apps/api/internal/infrastructure/config  │"
echo "│ vyzorix/apps/api/pkg/logging                    │ vyzorix/apps/api/internal/infrastructure/logging │"
echo "│ vyzorix/apps/api/pkg/models                     │ vyzorix/apps/api/internal/domain              │"
echo "│ vyzorix/apps/api/pkg/storage                    │ vyzorix/apps/api/internal/infrastructure/storage │"
echo "└────────────────────────────────────────────────┴────────────────────────────────────────────────┘"

echo ""
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
echo "ACTION: Update imports in affected files to use NEW infrastructure/ packages"
echo "════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
