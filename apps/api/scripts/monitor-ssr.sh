#!/bin/bash
# monitor-ssr.sh - Monitor SSR server health
# Usage: ./scripts/monitor-ssr.sh [OPTIONS]
#
# Options:
#   --url URL       SSR health URL (default: http://localhost:3001/health)
#   --interval S    Check interval in seconds (default: 5)
#   --max-fail N    Max failures before alert (default: 3)
#   --on-fail CMD   Command to run on failure

set -e

# Defaults
SSR_URL="${SSR_URL:-http://localhost:3001/health}"
INTERVAL="${INTERVAL:-5}"
MAX_FAIL="${MAX_FAIL:-3}"
ON_FAIL="${ON_FAIL:-}"
LOG_FILE="${LOG_FILE:-}"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log() {
    local level="$1"
    local msg="$2"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local line="[${timestamp}] [${level}] ${msg}"
    
    if [ -n "$LOG_FILE" ]; then
        echo "$line" >> "$LOG_FILE"
    fi
    
    case "$level" in
        INFO)
            echo -e "${GREEN}${line}${NC}"
            ;;
        WARN)
            echo -e "${YELLOW}${line}${NC}"
            ;;
        ERROR)
            echo -e "${RED}${line}${NC}"
            ;;
        *)
            echo -e "${CYAN}${line}${NC}"
            ;;
    esac
}

log_info() { log INFO "$1"; }
log_warn() { log WARN "$1"; }
log_error() { log ERROR "$1"; }

# Check SSR health
check_health() {
    local response
    response=$(curl -sf -w "\n%{http_code}" "${SSR_URL}" 2>/dev/null) || return 1
    local body
    local status
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    if [ "$status" -eq 200 ]; then
        echo "$body" | grep -q '"status":"ok"' && return 0
    fi
    return 1
}

# Get SSR status
get_status() {
    curl -sf "${SSR_URL}" 2>/dev/null || echo '{"status":"unknown"}'
}

# Run failure command
run_on_fail() {
    if [ -n "$ON_FAIL" ]; then
        log_warn "Running failure command: $ON_FAIL"
        eval "$ON_FAIL"
    fi
}

# Main monitoring loop
main() {
    log_info "Starting SSR monitor"
    log_info "  URL: ${SSR_URL}"
    log_info "  Interval: ${INTERVAL}s"
    log_info "  Max failures: ${MAX_FAIL}"
    
    local failures=0
    local consecutive_failures=0
    
    while true; do
        if check_health; then
            if [ $consecutive_failures -gt 0 ]; then
                log_info "SSR recovered after ${consecutive_failures} failures"
            fi
            consecutive_failures=0
            failures=0
        else
            consecutive_failures=$((consecutive_failures + 1))
            failures=$((failures + 1))
            
            log_warn "SSR health check failed (${consecutive_failures}/${MAX_FAIL})"
            
            if [ $consecutive_failures -ge $MAX_FAIL ]; then
                log_error "SSR health check failed ${MAX_FAIL} times consecutively"
                run_on_fail
            fi
        fi
        
        # Show status periodically (every 60 seconds)
        if [ $((failures % 12)) -eq 0 ] && [ $failures -gt 0 ]; then
            log_info "SSR Status: $(get_status)"
        fi
        
        sleep "$INTERVAL"
    done
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --url)
            SSR_URL="$2"
            shift 2
            ;;
        --interval)
            INTERVAL="$2"
            shift 2
            ;;
        --max-fail)
            MAX_FAIL="$2"
            shift 2
            ;;
        --on-fail)
            ON_FAIL="$2"
            shift 2
            ;;
        --log)
            LOG_FILE="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --url URL        SSR health URL (default: $SSR_URL)"
            echo "  --interval S     Check interval in seconds (default: $INTERVAL)"
            echo "  --max-fail N     Max failures before alert (default: $MAX_FAIL)"
            echo "  --on-fail CMD    Command to run on failure"
            echo "  --log FILE       Log to file"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

main
