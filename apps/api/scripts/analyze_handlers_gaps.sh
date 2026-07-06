#!/bin/bash
# Gap Analysis: internal/api/handlers (OLD) vs what needs NEW structure

cd "$(dirname "$0")/.."

echo ""
echo "                    HANDLERS GAP ANALYSIS: OLD vs NEW                                         "
echo ""
echo ""

echo ""
echo "1. OLD Handler Files (need migration to NEW structure)"
echo ""
echo ""

echo "Top-level handlers:"
for f in internal/api/handlers/*.go; do
    if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
        echo "  - $(basename $f)"
    fi
done

echo ""
echo "Subdirectories:"
for dir in internal/api/handlers/*/; do
    if [ -d "$dir" ]; then
        echo "   $(basename $dir)/"
        for f in "$dir"*.go; do
            if [ -f "$f" ] && ! echo "$f" | grep -q "_test"; then
                echo "      - $(basename $f)"
            fi
        done
    fi
done

echo ""
echo ""
echo "2. Handler Functions (need NEW domain handlers)"
echo ""
echo ""

echo "Auth handlers:"
grep "^func (ac \*AuthController)" internal/api/handlers/auth_*.go 2>/dev/null | sed 's/func (ac \*AuthController) /  /' | head -20

echo ""
echo "Device handlers:"
grep "^func (dc \*DeviceController)" internal/api/handlers/device*.go 2>/dev/null | sed 's/func (dc \*DeviceController) /  /'

echo ""
echo "Command handlers:"
grep "^func (cc \*CommandController)" internal/api/handlers/command*.go 2>/dev/null | sed 's/func (cc \*CommandController) /  /'

echo ""
echo "Admin handlers:"
grep "^func " internal/api/handlers/admin*.go 2>/dev/null | grep -v "^func init" | sed 's/func /  /' | head -15

echo ""
echo ""
echo "3. What Handlers Import from OLD"
echo ""

echo "Handlers importing pkg/models:"
grep -l "pkg/models" internal/api/handlers/*.go internal/api/handlers/*/*.go 2>/dev/null | wc -l
echo "handlers"

echo "Handlers importing pkg/storage:"
grep -l "pkg/storage" internal/api/handlers/*.go internal/api/handlers/*/*.go 2>/dev/null | wc -l
echo "handlers"

echo ""
echo ""
echo "4. Summary: Handler → NEW Structure Mapping Needed"
echo ""

echo ""
echo ""
echo " OLD Handler                                      NEW Structure Needed                   "
echo ""
echo " internal/api/handlers/auth_admin.go              internal/application/auth/admin.go     "
echo " internal/api/handlers/auth_core.go               internal/application/auth/login.go     "
echo " internal/api/handlers/auth_email_verify.go        internal/application/auth/email.go     "
echo " internal/api/handlers/auth_mfa.go                internal/application/auth/mfa.go      "
echo " internal/api/handlers/auth_oauth.go              internal/application/auth/oauth.go    "
echo " internal/api/handlers/auth_password_reset.go      internal/application/auth/password.go  "
echo " internal/api/handlers/auth_settings.go           internal/application/auth/settings.go  "
echo " internal/api/handlers/device.go                 internal/application/device/*.go       "
echo " internal/api/handlers/command.go                internal/application/command/*.go       "
echo " internal/api/handlers/admin_clients.go          internal/application/admin/client.go   "
echo " internal/api/handlers/client_credentials.go     internal/application/auth/client.go    "
echo " internal/api/handlers/updater.go                internal/application/updater/*.go      "
echo " internal/api/handlers/websocket_handler.go      internal/ws/hub.go, client.go         "
echo ""

echo ""
echo ""
echo "STATUS:   Handlers are OLD - need migration to internal/application/ handlers"
echo ""
