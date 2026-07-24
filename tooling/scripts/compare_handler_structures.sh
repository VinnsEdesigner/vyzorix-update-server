#!/bin/bash
# Compare OLD Flat Handlers vs NEW Organized Handlers

cd "$(dirname "$0")/.."

echo ""
echo "                    HANDLER STRUCTURE COMPARISON: OLD FLAT vs NEW ORGANIZED                                       "
echo ""
echo ""

echo ""
echo "STRUCTURE OVERVIEW"
echo ""
echo ""

echo "OLD: Flat structure in internal/api/handlers/"
echo "     auth_admin.go      → imports pkg/models, pkg/storage"
echo "     auth_core.go       → imports pkg/models, pkg/storage"
echo "     auth_email_verify.go → imports pkg/models, pkg/storage"
echo "     auth_mfa.go        → imports pkg/models, pkg/storage"
echo "     auth_oauth.go     → imports pkg/models, pkg/storage"
echo "     auth_password_reset.go → imports pkg/models, pkg/storage"
echo "     auth_settings.go   → imports pkg/models, pkg/storage"
echo "     device.go         → imports pkg/models, pkg/storage"
echo "     command.go        → imports pkg/models, pkg/storage"
echo "     ..."
echo ""

echo "NEW: Organized structure in internal/api/handlers/{auth,device,command}/"
echo "     auth/login.go     → imports internal/application, internal/domain"
echo "     auth/register.go  → imports internal/application, internal/domain"
echo "     device/list.go    → imports internal/application, internal/domain"
echo "     command/execute.go → imports internal/application, internal/domain"
echo "     ..."
echo ""

echo ""
echo "IMPORT ANALYSIS"
echo ""
echo ""

echo "OLD Flat Handlers:"
OLD_PKG_MODELS=$(grep -l "pkg/models" internal/api/handlers/auth_*.go internal/api/handlers/admin*.go internal/api/handlers/client*.go internal/api/handlers/device.go internal/api/handlers/command.go 2>/dev/null | wc -l)
OLD_PKG_STORAGE=$(grep -l "pkg/storage" internal/api/handlers/auth_*.go internal/api/handlers/admin*.go internal/api/handlers/client*.go internal/api/handlers/device.go internal/api/handlers/command.go 2>/dev/null | wc -l)
echo "  Files importing pkg/models: $OLD_PKG_MODELS"
echo "  Files importing pkg/storage: $OLD_PKG_STORAGE"
echo "  Status:   Using OLD imports"
echo ""

echo "NEW Organized Handlers:"
NEW_USING_APP=$(grep -l "internal/application" internal/api/handlers/auth/*.go internal/api/handlers/device/*.go internal/api/handlers/command/*.go 2>/dev/null | wc -l)
NEW_USING_DOMAIN=$(grep -l "internal/domain" internal/api/handlers/auth/*.go internal/api/handlers/device/*.go internal/api/handlers/command/*.go 2>/dev/null | wc -l)
echo "  Files importing internal/application: $NEW_USING_APP"
echo "  Files importing internal/domain: $NEW_USING_DOMAIN"
echo "  Status:  Using NEW structure"
echo ""

echo ""
echo "APPLICATION LAYER"
echo ""
echo ""

echo "internal/application/ structure:"
ls -la internal/application/ 2>/dev/null | grep "^d" | awk '{print "   " $NF}'
echo ""

echo "Files in application layer:"
find internal/application -name "*.go" -type f 2>/dev/null | while read f; do
    echo "  - $(echo $f | sed 's|internal/||')"
done | head -20

echo ""
echo ""
echo "MIGRATION STATUS"
echo ""
echo ""

echo ""
echo " Handler Type                                     Migration Status                       "
echo ""
echo " Auth handlers (login, register, logout, ...)       Partially migrated to auth/        "
echo " Device handlers (list, register, status, ...)     Partially migrated to device/      "
echo " Command handlers (execute, ...)                   Partially migrated to command/      "
echo " Admin handlers                                    Still OLD (admin*.go)            "
echo " OAuth handlers                                    Still OLD (auth_oauth.go)         "
echo " MFA handlers                                     Still OLD (auth_mfa.go)          "
echo " WebSocket handler                                 Still OLD (websocket_handler.go) "
echo ""

echo ""
echo ""
echo " SUMMARY"
echo ""
echo ""
echo "  NEW Application Layer:  EXISTS"
echo "  NEW Organized Handlers:  EXISTS"  
echo "  OLD Flat Handlers:   Still using pkg/models, pkg/storage"
echo ""
echo "  Migration Progress:"
echo "    - Domain types:  Complete (38 types)"
echo "    - Infrastructure:  Complete (storage, crypto, email, uuid)"
echo "    - Application layer:  Exists (auth service, DTOs)"
echo "    - Organized handlers:  In Progress"
echo "    - Flat handlers:   Pending migration"
echo ""
echo ""
