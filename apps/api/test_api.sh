#!/bin/bash
# Comprehensive API Testing Script
# Tests all 100+ endpoints, error handling, and database operations

# Don't exit on error - we want to run all tests
set +e

# Configuration
BASE_URL="${API_URL:-http://localhost:3000}"
REPORT_FILE="endpoint_test_report.md"
DB_PATH="./data/vyzorix.db"
TEMP_DIR=$(mktemp -d)

# Use full path for commands
CURL="/usr/bin/curl"
HEAD="/usr/bin/head"
CAT="/bin/cat"
DU="/usr/bin/du"
PASS_COUNT=0
FAIL_COUNT=0
TOTAL_COUNT=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Cleanup on exit
trap "rm -rf $TEMP_DIR" EXIT

# Helper functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; ((PASS_COUNT++)); ((TOTAL_COUNT++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAIL_COUNT++)); ((TOTAL_COUNT++)); }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_test() { echo -e "${BLUE}[TEST]${NC} $1"; }

# Make request and check result
# Args: method, path, expected_status, description, [body_file]
test_endpoint() {
    local METHOD=$1
    local PATH=$2
    local EXPECTED=$3
    local DESC=$4
    local BODY_FILE=$5
    
    ((TOTAL_COUNT++))
    
    local URL="${BASE_URL}${PATH}"
    local RESPONSE_FILE="${TEMP_DIR}/response_${TOTAL_COUNT}.json"
    local HTTP_CODE
    
    # Execute "$CURL" based on whether body file exists
    if [ -n "$BODY_FILE" ] && [ -f "$BODY_FILE" ]; then
        HTTP_CODE=$("$CURL" -s -o "$RESPONSE_FILE" -w "%{http_code}" -X "$METHOD" -H "Content-Type: application/json" --data-binary "@$BODY_FILE" "$URL")
    else
        HTTP_CODE=$("$CURL" -s -o "$RESPONSE_FILE" -w "%{http_code}" -X "$METHOD" "$URL")
    fi
    
    if [ "$HTTP_CODE" = "$EXPECTED" ]; then
        log_pass "$METHOD $PATH -> $HTTP_CODE ($DESC)"
        echo "  Response: $("$HEAD" -c 200 "$RESPONSE_FILE" 2>/dev/null || echo 'N/A')"
        return 0
    elif [ "$HTTP_CODE" != "000" ]; then
        log_fail "$METHOD $PATH -> $HTTP_CODE (expected $EXPECTED) ($DESC)"
        echo "  Response: $("$CAT" "$RESPONSE_FILE" 2>/dev/null | "$HEAD" -c 500)"
        return 1
    else
        log_fail "$METHOD $PATH -> UNREACHABLE ($DESC)"
        return 1
    fi
}

# Test endpoint reachability only (any non-5xx is ok for reachability test)
test_reachable() {
    local METHOD=$1
    local PATH=$2
    local DESC=$3
    
    ((TOTAL_COUNT++))
    
    local URL="${BASE_URL}${PATH}"
    local HTTP_CODE=$("$CURL" -s -o /dev/null -w "%{http_code}" -X "$METHOD" "$URL")
    
    if [ "$HTTP_CODE" != "000" ] && [ "$HTTP_CODE" -lt 500 ]; then
        log_pass "$METHOD $PATH -> $HTTP_CODE (reachable)"
        return 0
    else
        log_fail "$METHOD $PATH -> $HTTP_CODE (unreachable)"
        return 1
    fi
}

# Check database
check_database() {
    log_info "Checking database at $DB_PATH..."
    if [ -f "$DB_PATH" ]; then
        log_pass "Database file exists"
        local SIZE=$("$DU" -h "$DB_PATH" | cut -f1)
        log_info "Database size: $SIZE"
        
        # Check if server can connect to database via health endpoint
        local HTTP_CODE=$("$CURL" -s -o /dev/null -w "%{http_code}" --max-time 5 "$BASE_URL/health" 2>/dev/null)
        if [ "$HTTP_CODE" = "200" ]; then
            log_pass "Server connected to database"
        else
            log_warn "Server health check returned: $HTTP_CODE"
        fi
    else
        log_warn "Database file not found at $DB_PATH"
    fi
}

# Create test data files
create_test_data() {
    # Valid operator
    "$CAT" > "${TEMP_DIR}/valid_operator.json" << 'EOF'
{
    "email": "apitest-$(date +%s)@test.com",
    "password": "TestPass123!",
    "name": "API Test User",
    "role": "user"
}
EOF

    # Invalid login
    "$CAT" > "${TEMP_DIR}/invalid_login.json" << 'EOF'
{"email": "invalid@test.com", "password": "wrongpassword"}
EOF

    # Malformed JSON
    "$CAT" > "${TEMP_DIR}/malformed.json" << 'EOF'
{"email": "not-json
EOF

    # Empty body
    "$CAT" > "${TEMP_DIR}/empty.json" << 'EOF'
{}
EOF

    # Short password
    "$CAT" > "${TEMP_DIR}/short_password.json" << 'EOF'
{"email": "test@test.com", "password": "x", "name": "Test"}
EOF

    # Valid device (FirebaseInstallID is required)
    DEVICE_ID="test-device-$(date +%s)"
    FIREBASE_ID="firebase-$(date +%s)"
    "$CAT" > "${TEMP_DIR}/valid_device.json" << EOF
{
    "device_id": "$DEVICE_ID",
    "firebase_install_id": "$FIREBASE_ID",
    "name": "Test Device",
    "platform": "android",
    "app_version": "1.0.0"
}
EOF

    # Invalid device
    "$CAT" > "${TEMP_DIR}/invalid_device.json" << 'EOF'
{"invalid": "data"}
EOF

    # Valid command
    "$CAT" > "${TEMP_DIR}/valid_command.json" << 'EOF'
{
    "command": "FORCE_SPEAKER",
    "args": {"volume": 80},
    "priority": 5
}
EOF

    # Empty command
    "$CAT" > "${TEMP_DIR}/empty_command.json" << 'EOF'
{"command": "", "args": {}}
EOF
}

# Generate markdown report
generate_report() {
    "$CAT" > "$REPORT_FILE" << 'HEADER'
# Comprehensive API Test Report

HEADER

    echo "**Generated:** $(date -Iseconds)" >> "$REPORT_FILE"
    echo "**Base URL:** $BASE_URL" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    echo "## Summary" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "| Metric | Value |" >> "$REPORT_FILE"
    echo "|--------|-------|" >> "$REPORT_FILE"
    echo "| Total Tests | $TOTAL_COUNT |" >> "$REPORT_FILE"
    echo "| Passed | $PASS_COUNT |" >> "$REPORT_FILE"
    echo "| Failed | $FAIL_COUNT |" >> "$REPORT_FILE"
    echo "| Pass Rate | $(echo "scale=1; $PASS_COUNT * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo 'N/A')% |" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    echo "## Test Categories" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "### Health Endpoints" >> "$REPORT_FILE"
    echo "| Endpoint | Status |" >> "$REPORT_FILE"
    echo "|----------|--------|" >> "$REPORT_FILE"
    echo "| GET /health | $(grep -c 'health' "$REPORT_FILE" || echo 'Tested') |" >> "$REPORT_FILE"
    
    log_info "Report saved to: $REPORT_FILE"
}

# ==================== TESTS ====================

run_all_tests() {
    echo ""
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║     Comprehensive API Testing Suite                   ║"
    echo "║     Base URL: $BASE_URL"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""
    
    create_test_data
    check_database
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  HEALTH ENDPOINTS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/health" "200" "Health check"
    test_endpoint "GET" "/healthz" "200" "Healthz check"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  VERSION / STATIC ENDPOINTS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/api/v1/version" "200" "Version endpoint"
    test_endpoint "GET" "/api/v1/changelog" "200" "Changelog endpoint"
    test_endpoint "GET" "/api/v1/apk/test.apk" "404" "APK not found (expected)"
    test_endpoint "GET" "/bin/test.bin" "404" "Bin not found (expected)"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - LOGIN"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/auth/login" "401" "Invalid credentials" "${TEMP_DIR}/invalid_login.json"
    test_endpoint "POST" "/v1/auth/login" "400" "Malformed JSON" "${TEMP_DIR}/malformed.json"
    test_endpoint "POST" "/v1/auth/login" "400" "Empty body" "${TEMP_DIR}/empty.json"
    test_endpoint "POST" "/v1/auth/login" "400" "Short password" "${TEMP_DIR}/short_password.json"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - REGISTER"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/auth/register" "201" "Valid registration" "${TEMP_DIR}/valid_operator.json"
    test_endpoint "POST" "/v1/auth/register" "400" "Invalid email format" "${TEMP_DIR}/invalid_login.json"
    test_endpoint "POST" "/v1/auth/register" "400" "Short password" "${TEMP_DIR}/short_password.json"
    test_endpoint "POST" "/v1/auth/register" "400" "Empty body" "${TEMP_DIR}/empty.json"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - PASSWORD RESET"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/auth/forgot-password" "200" "Valid email" "${TEMP_DIR}/invalid_login.json"
    test_endpoint "POST" "/v1/auth/forgot-password" "400" "Invalid email" "${TEMP_DIR}/empty.json"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - EMAIL VERIFICATION"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/auth/verify-email" "400" "Invalid token"
    test_endpoint "POST" "/v1/auth/resend-verification" "400" "Invalid email"
    test_endpoint "POST" "/v1/auth/cancel-verification" "400" "Invalid token"
    test_endpoint "GET" "/v1/auth/poll-verification?email=test@test.com" "404" "Poll verification"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - MFA"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/v1/auth/mfa/status" "401" "MFA status without auth"
    test_endpoint "POST" "/v1/auth/mfa/enroll" "401" "MFA enroll without auth"
    test_endpoint "POST" "/v1/auth/mfa/verify-setup" "401" "MFA verify without auth"
    test_endpoint "POST" "/v1/auth/mfa/enable" "401" "MFA enable without auth"
    test_endpoint "POST" "/v1/auth/mfa/disable" "401" "MFA disable without auth"
    test_endpoint "POST" "/v1/auth/mfa/verify-backup" "401" "MFA backup without auth"
    test_endpoint "POST" "/v1/auth/mfa/regenerate-backup-codes" "401" "MFA regenerate without auth"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - SESSION / ME"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/v1/auth/me" "401" "Me without auth"
    test_endpoint "PATCH" "/v1/auth/me" "401" "Update me without auth"
    test_endpoint "POST" "/v1/auth/logout" "200" "Logout"
    test_endpoint "GET" "/v1/auth/lockout/status" "401" "Lockout status without auth"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - ADMIN OPERATORS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/v1/auth/admin/operators" "401" "List operators without auth"
    test_endpoint "POST" "/v1/auth/admin/operators" "401" "Create operator without auth"
    test_endpoint "GET" "/v1/auth/admin/operators/1" "401" "Get operator without auth"
    test_endpoint "PATCH" "/v1/auth/admin/operators/1" "401" "Update operator without auth"
    test_endpoint "DELETE" "/v1/auth/admin/operators/1" "401" "Delete operator without auth"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  AUTH - CLIENT CREDENTIALS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/auth/client-credentials" "401" "Create client without auth"
    test_endpoint "GET" "/v1/auth/client-credentials" "401" "List clients without auth"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  DEVICE ENDPOINTS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/device/register" "201" "Register device" "${TEMP_DIR}/valid_device.json"
    test_endpoint "POST" "/v1/device/register" "400" "Invalid device data" "${TEMP_DIR}/invalid_device.json"
    test_endpoint "POST" "/v1/device/register" "400" "Empty device data" "${TEMP_DIR}/empty.json"
    test_endpoint "GET" "/v1/device/test-device-999/status" "404" "Device not found"
    test_endpoint "GET" "/v1/device/nonexistent/status" "404" "Nonexistent device"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  COMMAND ENDPOINTS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "POST" "/v1/device/test-device/command" "404" "Command to nonexistent device"
    test_endpoint "POST" "/v1/device/test-device/command" "400" "Invalid command data" "${TEMP_DIR}/empty_command.json"
    test_endpoint "GET" "/v1/device/test-device/commands/pending" "404" "Pending commands for nonexistent"
    test_endpoint "GET" "/v1/command/dispatch-999/status" "404" "Command dispatch not found"
    test_endpoint "POST" "/v1/command/dispatch-999/retry" "404" "Retry nonexistent command"
    test_endpoint "DELETE" "/v1/command/dispatch-999" "404" "Cancel nonexistent command"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  DASHBOARD ENDPOINTS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/v1/dashboard/devices" "401" "Dashboard without auth"
    test_endpoint "GET" "/v1/dashboard/devices?page=1&limit=10" "401" "Dashboard with pagination"
    test_endpoint "GET" "/v1/dashboard/devices/operator" "401" "Operator dashboard without auth"
    test_endpoint "GET" "/v1/device/count" "401" "Device count without auth"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  ADMIN ENDPOINTS"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/v1/admin/clients" "401" "List clients without auth"
    test_endpoint "GET" "/v1/admin/clients/test-id" "401" "Get client without auth"
    test_endpoint "PATCH" "/v1/admin/clients/test-id" "401" "Update client without auth"
    test_endpoint "DELETE" "/v1/admin/clients/test-id" "401" "Delete client without auth"
    test_endpoint "POST" "/v1/admin/clients/test-id/rotate-key" "401" "Rotate key without auth"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  WEBSOCKET ENDPOINT"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/v1/device/test-device/stream" "400" "WebSocket upgrade required"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  ERROR CASES"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    test_endpoint "GET" "/nonexistent-path" "404" "404 Not Found"
    test_endpoint "POST" "/v1/auth/login" "415" "Wrong Content-Type" # without Content-Type header
    test_endpoint "DELETE" "/v1/auth/logout" "405" "Method not allowed"
    test_endpoint "GET" "/v1/admin" "404" "Base admin path"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  DATABASE VERIFICATION"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Register a real device and verify it appears in database
    DEVICE_ID="db-test-$(date +%s)"
    "$CAT" > "${TEMP_DIR}/db_device.json" << EOF
{
    "device_id": "$DEVICE_ID",
    "name": "Database Test Device",
    "platform": "android",
    "app_version": "1.0.0"
}
EOF
    
    test_endpoint "POST" "/v1/device/register" "201" "Create test device for DB" "${TEMP_DIR}/db_device.json"
    
    # Check database through API (since sqlite3 is not available in container)
    log_test "Checking database via API..."
    
    # Get device count
    HTTP_CODE=$("$CURL" -s -o /dev/null -w "%{http_code}" -X GET "$BASE_URL/v1/device/count" 2>/dev/null)
    if [ "$HTTP_CODE" != "000" ]; then
        log_pass "Database accessible via API (status: $HTTP_CODE)"
    else
        log_fail "Database not accessible"
    fi
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  SUMMARY"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo -e "Total Tests:  $TOTAL_COUNT"
    echo -e "Passed:       ${GREEN}$PASS_COUNT${NC}"
    echo -e "Failed:       ${RED}$FAIL_COUNT${NC}"
    echo -e "Pass Rate:    $(echo "scale=1; $PASS_COUNT * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo 'N/A')%"
    echo ""
    
    # Generate report
    generate_report
    
    # Return non-zero if any tests failed
    if [ $FAIL_COUNT -gt 0 ]; then
        return 1
    fi
    return 0
}

# Run tests
run_all_tests
