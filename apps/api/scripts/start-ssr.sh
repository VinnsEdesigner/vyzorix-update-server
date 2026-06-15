#!/bin/bash
# start-ssr.sh - Start SSR server with auto-recovery
# Usage: ./scripts/start-ssr.sh [OPTIONS]
#
# Options:
#   --port PORT       Set SSR port (default: 3001)
#   --mode MODE       Set mode: development, production (default: production)
#   --healthz URL    Set health check URL (default: /health)
#   --max-retries N  Max restart attempts (default: 3)
#   --retry-delay S  Delay between retries in seconds (default: 5)

set -e

# Defaults
SSR_PORT="${SSR_PORT:-3001}"
SSR_MODE="${SSR_MODE:-production}"
SSR_SERVER_URL="http://localhost:${SSR_PORT}"
MAX_RETRIES="${MAX_RETRIES:-3}"
RETRY_DELAY="${RETRY_DELAY:-5}"
SSR_SCRIPT="./ssr-server.js"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --port)
            SSR_PORT="$2"
            shift 2
            ;;
        --mode)
            SSR_MODE="$2"
            shift 2
            ;;
        --healthz)
            HEALTHZ_URL="$2"
            shift 2
            ;;
        --max-retries)
            MAX_RETRIES="$2"
            shift 2
            ;;
        --retry-delay)
            RETRY_DELAY="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  --port PORT       Set SSR port (default: $SSR_PORT)"
            echo "  --mode MODE       Set mode: development, production (default: $SSR_MODE)"
            echo "  --healthz URL     Set health check URL"
            echo "  --max-retries N   Max restart attempts (default: $MAX_RETRIES)"
            echo "  --retry-delay S   Delay between retries in seconds (default: $RETRY_DELAY)"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Health check function
check_health() {
    curl -sf "${SSR_SERVER_URL}/health" > /dev/null 2>&1
}

# Wait for SSR to be ready
wait_for_ready() {
    log_info "Waiting for SSR server to be ready..."
    local attempt=1
    while [ $attempt -le 30 ]; do
        if check_health; then
            log_info "SSR server is ready"
            return 0
        fi
        sleep 1
        ((attempt++))
    done
    log_error "SSR server failed to become ready"
    return 1
}

# Start SSR server
start_ssr() {
    export SSR_PORT="${SSR_PORT}"
    export SSR_MODE="${SSR_MODE}"
    export SSR_SERVER_URL="${SSR_SERVER_URL}"
    
    log_info "Starting SSR server..."
    log_info "  Port: ${SSR_PORT}"
    log_info "  Mode: ${SSR_MODE}"
    log_info "  Script: ${SSR_SCRIPT}"
    
    if [ ! -f "${SSR_SCRIPT}" ]; then
        log_error "SSR script not found: ${SSR_SCRIPT}"
        exit 1
    fi
    
    node "${SSR_SCRIPT}"
}

# Main loop with retry
main() {
    local retries=0
    
    while [ $retries -lt $MAX_RETRIES ]; do
        log_info "Starting SSR (attempt $((retries+1))/${MAX_RETRIES})"
        
        start_ssr &
        local pid=$!
        
        if wait_for_ready; then
            log_info "SSR server running with PID ${pid}"
            
            # Wait for process
            wait $pid
            local exit_code=$?
            
            if [ $exit_code -eq 0 ]; then
                log_info "SSR server stopped normally"
                exit 0
            else
                log_warn "SSR server crashed with exit code ${exit_code}"
            fi
        else
            log_warn "SSR server failed to start"
        fi
        
        ((retries++))
        
        if [ $retries -lt $MAX_RETRIES ]; then
            log_info "Restarting in ${RETRY_DELAY} seconds..."
            sleep $RETRY_DELAY
        fi
    done
    
    log_error "Max retries (${MAX_RETRIES}) reached, giving up"
    exit 1
}

main "$@"
