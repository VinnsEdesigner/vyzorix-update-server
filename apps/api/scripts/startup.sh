#!/bin/bash
#===============================================================================
# Vyzorix Server Startup & Health Check Script
# 
# This script performs comprehensive checks and startup of the Vyzorix server stack:
# 1. Build verification (Go server, web app)
# 2. Asset validation and copying
# 3. SSR process management
# 4. Go server startup
# 5. Health verification
# 6. Browser auto-open (optional)
#
# Usage: ./startup.sh [--no-browser] [--dev] [--prod] [--check-only]
#===============================================================================

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_DIR="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$API_DIR/../web"
PUBLIC_DIR="$API_DIR/public"
DIST_CLIENT="$WEB_DIR/dist/client"
GO_BINARY="$API_DIR/vyzorix-server"
SSR_SCRIPT="$API_DIR/ssr-server.js"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
RESET='\033[0m'

# Flags
# AUTO_OPEN_BROWSER=true (default: true if not set, set to false to disable)
# BUILD_MODE=production|development (default: production)
if [[ "${AUTO_OPEN_BROWSER:-true}" == "false" ]] || [[ "${AUTO_OPEN_BROWSER:-true}" == "0" ]]; then
    OPEN_BROWSER=false
else
    OPEN_BROWSER=true
fi
BUILD_MODE="${BUILD_MODE:-production}"
CHECK_ONLY=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-browser)
            OPEN_BROWSER=false
            shift
            ;;
        --dev)
            BUILD_MODE="development"
            shift
            ;;
        --prod)
            BUILD_MODE="production"
            shift
            ;;
        --check-only)
            CHECK_ONLY=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --no-browser    Don't open browser"
            echo "  --dev           Development mode"
            echo "  --prod          Production mode (default)"
            echo "  --check-only     Only run checks, don't start servers"
            echo "  -h, --help      Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Print banner
print_banner() {
    echo -e "${MAGENTA}${BOLD}"
    echo "  +-------------------------------------------------------------+"
    echo "  |   _   _           _        ____                           |"
    echo "  |  |_| |_|   ___   | |__    |  _|  ___  ___                 |"
    echo "  |  | | | |  / _ \\  | '_ \\  | |_  / _ \\/ __|                |"
    echo "  |  | |_| | | (_) | | |_) | |  _|  __/\\__ \\                |"
    echo "  |  |___|_|  \\___/  |_.__/   |_|   \\___||___/               |"
    echo "  |                                                              |"
    echo "  |              SERVER STARTUP v1.0.0                          |"
    echo "  +-------------------------------------------------------------+${RESET}"
}

# Print section header
print_section() {
    echo -e "\n${CYAN}[$1]${RESET}"
}

# Print status
print_ok() {
    echo -e "  ${GREEN}✓${RESET} $1"
}

print_warn() {
    echo -e "  ${YELLOW}⚠${RESET} $1"
}

print_fail() {
    echo -e "  ${RED}✗${RESET} $1"
}

print_info() {
    echo -e "  ${DIM}>${RESET} $1"
}

# Count files by extension
count_ext() {
    local dir="$1"
    local ext="$2"
    find "$dir" -type f -name "*.$ext" 2>/dev/null | wc -l
}

# List assets by type
list_assets() {
    local dir="$1"
    local ext="$2"
    find "$dir" -type f -name "*.$ext" 2>/dev/null | sed "s|$dir/||" | head -5
    local count=$(count_ext "$dir" "$ext")
    if [ $count -gt 5 ]; then
        echo "  ... and $((count - 5)) more $ext files"
    fi
}

# Check if file exists
check_file() {
    if [ -f "$1" ]; then
        print_ok "$1 exists ($(wc -c < "$1" | numfmt --to=iec) bytes)"
        return 0
    else
        print_fail "$1 NOT FOUND"
        return 1
    fi
}

# Check if directory exists and has content
check_dir() {
    if [ -d "$1" ]; then
        local count=$(find "$1" -type f | wc -l)
        if [ $count -gt 0 ]; then
            print_ok "$1 exists ($count files)"
            return 0
        else
            print_warn "$1 is EMPTY"
            return 2
        fi
    else
        print_fail "$1 NOT FOUND"
        return 1
    fi
}

#===============================================================================
# MAIN STARTUP SEQUENCE
#===============================================================================

main() {
    print_banner
    
    local errors=0
    local warnings=0
    
    echo -e "${DIM}  Mode: $BUILD_MODE${RESET}"
    echo -e "${DIM}  Browser: $([ "$OPEN_BROWSER" = true ] && echo "Yes" || echo "No")${RESET}"
    echo ""
    
    #---------------------------------------------------------------------------
    # STEP 1: Check Go Server Binary
    #---------------------------------------------------------------------------
    print_section "STEP 1: Go Server Binary"
    
    if ! check_file "$GO_BINARY"; then
        print_info "Building Go server..."
        cd "$API_DIR"
        if go build -o vyzorix-server . 2>&1; then
            print_ok "Go server built successfully"
        else
            print_fail "Go server build FAILED"
            errors=$((errors + 1))
        fi
        cd - > /dev/null
    fi
    
    #---------------------------------------------------------------------------
    # STEP 2: Check Web App Build
    #---------------------------------------------------------------------------
    print_section "STEP 2: Web App Build"
    
    if ! check_dir "$DIST_CLIENT"; then
        print_warn "Web app not built. Building..."
        cd "$WEB_DIR"
        if pnpm run build 2>&1; then
            print_ok "Web app built successfully"
        else
            print_fail "Web app build FAILED"
            errors=$((errors + 1))
        fi
        cd - > /dev/null
    else
        # Count assets
        local js_count=$(count_ext "$DIST_CLIENT" "js")
        local css_count=$(count_ext "$DIST_CLIENT" "css")
        local html_count=$(count_ext "$DIST_CLIENT" "html")
        local img_count=$(find "$DIST_CLIENT" -type f \( -name "*.jpg" -o -name "*.png" -o -name "*.svg" -o -name "*.ico" \) | wc -l)
        
        print_info "Assets found:"
        echo "    JS:  $js_count files"
        echo "    CSS: $css_count files"
        echo "    HTML: $html_count files"
        echo "    Images: $img_count files"
        
        # Show sample files
        if [ $js_count -gt 0 ]; then
            echo -e "    ${DIM}Sample JS:${RESET}"
            list_assets "$DIST_CLIENT" "js" | sed 's/^/        /'
        fi
    fi
    
    #---------------------------------------------------------------------------
    # STEP 3: Validate & Copy Assets to Public
    #---------------------------------------------------------------------------
    print_section "STEP 3: Asset Validation"
    
    # Check what's in public
    if [ -d "$PUBLIC_DIR" ]; then
        local pub_count=$(find "$PUBLIC_DIR" -type f 2>/dev/null | wc -l)
        print_info "Public directory: $pub_count files"
    fi
    
    # Compare dist/client vs public
    local dist_js=$(count_ext "$DIST_CLIENT" "js")
    local dist_css=$(count_ext "$DIST_CLIENT" "css")
    local dist_html=$(count_ext "$DIST_CLIENT" "html")
    local pub_js=$(count_ext "$PUBLIC_DIR" "js")
    local pub_css=$(count_ext "$PUBLIC_DIR" "css")
    local pub_html=$(count_ext "$PUBLIC_DIR" "html")
    
    local sync_needed=false
    
    if [ "$dist_js" != "$pub_js" ] || [ "$dist_css" != "$pub_css" ] || [ "$dist_html" != "$pub_html" ]; then
        print_warn "Assets out of sync!"
        echo "    dist/client: $dist_js JS, $dist_css CSS, $dist_html HTML"
        echo "    public:     $pub_js JS, $pub_css CSS, $pub_html HTML"
        sync_needed=true
        warnings=$((warnings + 1))
    else
        print_ok "Assets in sync"
    fi
    
    if [ "$sync_needed" = true ]; then
        print_info "Syncing assets to public..."
        rm -rf "$PUBLIC_DIR"/*
        if cp -r "$DIST_CLIENT"/* "$PUBLIC_DIR/"; then
            print_ok "Assets copied to public"
        else
            print_fail "Asset copy FAILED"
            errors=$((errors + 1))
        fi
    fi
    
    # Check critical files
    print_info "Critical files check:"
    check_file "$PUBLIC_DIR/index.html" || errors=$((errors + 1))
    check_file "$PUBLIC_DIR/landing.html" || warnings=$((warnings + 1))
    check_file "$PUBLIC_DIR/manifest.json" || warnings=$((warnings + 1))
    
    #---------------------------------------------------------------------------
    # STEP 4: Check SSR Server
    #---------------------------------------------------------------------------
    print_section "STEP 4: SSR Server"
    
    if [ -f "$SSR_SCRIPT" ]; then
        print_ok "SSR script found"
    else
        print_fail "SSR script NOT FOUND at $SSR_SCRIPT"
        errors=$((errors + 1))
    fi
    
    if check_dir "$WEB_DIR/dist/server"; then
        local server_js=$(count_ext "$WEB_DIR/dist/server" "js")
        print_info "SSR server build: $server_js JS files"
    fi
    
    #---------------------------------------------------------------------------
    # STOP HERE IF CHECK ONLY MODE
    #---------------------------------------------------------------------------
    if [ "$CHECK_ONLY" = true ]; then
        print_section "CHECK SUMMARY"
        if [ $errors -eq 0 ] && [ $warnings -eq 0 ]; then
            echo -e "  ${GREEN}All checks passed!${RESET}"
        elif [ $errors -eq 0 ]; then
            echo -e "  ${YELLOW}$warnings warnings${RESET}"
        else
            echo -e "  ${RED}$errors errors, $warnings warnings${RESET}"
        fi
        exit $errors
    fi
    
    #---------------------------------------------------------------------------
    # STEP 5: Check for Running Servers
    #---------------------------------------------------------------------------
    print_section "STEP 5: Server Status"
    
    if pgrep -f "vyzorix-server" > /dev/null 2>&1; then
        print_warn "Go server already running"
        print_info "PID: $(pgrep -f vyzorix-server)"
    else
        print_ok "Go server not running"
    fi
    
    if pgrep -f "node.*ssr-server" > /dev/null 2>&1; then
        print_warn "SSR server already running"
        print_info "PID: $(pgrep -f 'node.*ssr-server')"
    else
        print_ok "SSR server not running"
    fi
    
    #---------------------------------------------------------------------------
    # STEP 6: Kill Existing Servers (if any)
    #---------------------------------------------------------------------------
    print_section "STEP 6: Cleanup"
    
    if pgrep -f "vyzorix-server" > /dev/null 2>&1 || pgrep -f "node.*ssr-server" > /dev/null 2>&1; then
        print_info "Stopping existing servers..."
        pkill -f "vyzorix-server" 2>/dev/null || true
        pkill -f "node.*ssr-server" 2>/dev/null || true
        sleep 2
        print_ok "Servers stopped"
    else
        print_ok "No servers to clean up"
    fi
    
    #---------------------------------------------------------------------------
    # STEP 7: Environment Setup
    #---------------------------------------------------------------------------
    print_section "STEP 7: Environment"
    
    export SSR_ENABLED=true
    export SSR_AUTO_START=true
    export SSR_AUTO_BUILD=false
    export NODE_ENV=production
    export TOKEN_SECRET="${TOKEN_SECRET:-test-token-secret-for-development-only-32chars}"
    export JWT_SECRET="${JWT_SECRET:-test-jwt-secret-for-development-only-32chars}"
    export DATABASE_URL="${DATABASE_URL:-./data/vyzorix.db}"
    export PUBLIC_DIR="${PUBLIC_DIR:-$PUBLIC_DIR}"
    export VYZORIX_API_DIR="${VYZORIX_API_DIR:-./data}"
    export PATH="$PATH:/usr/local/go/bin"
    
    print_info "Environment configured"
    print_info "SSR_ENABLED=$SSR_ENABLED"
    print_info "SSR_AUTO_START=$SSR_AUTO_START"
    print_info "NODE_ENV=$NODE_ENV"
    
    #---------------------------------------------------------------------------
    # STEP 8: Start Servers
    #---------------------------------------------------------------------------
    print_section "STEP 8: Starting Servers"
    
    # Start Go server in background
    print_info "Starting Go server..."
    cd "$API_DIR"
    ./vyzorix-server > server.log 2>&1 &
    local go_pid=$!
    echo $go_pid > .server.pid
    print_ok "Go server started (PID: $go_pid)"
    
    # Wait for startup
    sleep 3
    
    # Check if server started
    if ! kill -0 $go_pid 2>/dev/null; then
        print_fail "Go server failed to start!"
        cat server.log
        exit 1
    fi
    
    # Check health
    local health_status
    health_status=$(curl -s http://localhost:3000/health 2>/dev/null | jq -r '.ok' 2>/dev/null || echo "error")
    if [ "$health_status" = "true" ]; then
        print_ok "Go server health check passed"
    else
        print_warn "Go server health check: $health_status"
    fi
    
    # Check SSR
    local ssr_health
    ssr_health=$(curl -s http://localhost:3001/health 2>/dev/null | jq -r '.status' 2>/dev/null || echo "error")
    if [ "$ssr_health" = "ok" ]; then
        print_ok "SSR server health check passed"
    else
        print_warn "SSR server health check: $ssr_health"
    fi
    
    #---------------------------------------------------------------------------
    # STEP 9: Open Browser
    #---------------------------------------------------------------------------
    if [ "$OPEN_BROWSER" = true ]; then
        print_section "STEP 9: Opening Browser"
        
        sleep 2
        
        # Try multiple browser options in order of preference
        local browser_opened=false
        
        # Option 1: Use chromium/chromium-browser directly with no-sandbox (for containers)
        if command -v chromium > /dev/null 2>&1; then
            print_info "Using chromium..."
            chromium --no-sandbox --disable-setuid-sandbox --headless --disable-gpu --disable-software-rasterizer http://localhost:3000/ > /dev/null 2>&1 &
            browser_opened=true
        elif command -v chromium-browser > /dev/null 2>&1; then
            print_info "Using chromium-browser..."
            chromium-browser --no-sandbox --disable-setuid-sandbox --headless --disable-gpu http://localhost:3000/ > /dev/null 2>&1 &
            browser_opened=true
        elif command -v google-chrome > /dev/null 2>&1; then
            print_info "Using google-chrome..."
            google-chrome --no-sandbox --disable-setuid-sandbox --headless --disable-gpu http://localhost:3000/ > /dev/null 2>&1 &
            browser_opened=true
        # Option 2: Fallback to xdg-open for desktop environments
        elif command -v xdg-open > /dev/null 2>&1; then
            print_info "Using xdg-open..."
            xdg-open http://localhost:3000/ &
            browser_opened=true
        elif command -v open > /dev/null 2>&1; then
            print_info "Using open..."
            open http://localhost:3000/ &
            browser_opened=true
        fi
        
        if [ "$browser_opened" = true ]; then
            print_ok "Browser opened"
        else
            print_info "No browser found - please open manually: http://localhost:3000/"
        fi
    fi
    
    #---------------------------------------------------------------------------
    # STEP 10: Final Summary
    #---------------------------------------------------------------------------
    print_section "STARTUP COMPLETE"
    
    echo -e "  ${GREEN}Go Server:${RESET}  http://localhost:3000"
    echo -e "  ${GREEN}SSR Server:${RESET} http://localhost:3001"
    echo -e "  ${GREEN}Health:${RESET}    http://localhost:3000/health"
    echo -e "  ${GREEN}Logs:${RESET}      $API_DIR/server.log"
    echo ""
    echo -e "  ${DIM}Press Ctrl+C to stop servers${RESET}"
    echo ""
    
    # Wait for interrupt
    wait $go_pid
}

# Cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down...${RESET}"
    pkill -f "vyzorix-server" 2>/dev/null || true
    pkill -f "node.*ssr-server" 2>/dev/null || true
    echo -e "${GREEN}Done${RESET}"
}

trap cleanup EXIT INT TERM

# Run main
main "$@"
