# SSR Module Architecture

The SSR (Server-Side Rendering) system is now modular, with separated concerns for configuration, process management, monitoring, and building.

## Architecture Overview

```

                         SSR Manager                               
                 
     Config         Process        Monitor              
                    Manager                             
   - Enable                       - Health              
   - URL          - Start/Stop      checks              
   - Port         - Retry        - Crash               
   - Timeout      - Crash          detection           
   - Retries        detection    - Logging             
                 
                                                               
                                                
    Builder                                                 
                                                            
   - Build                                                  
   - Copy                                 
     Assets                                                  
                                                 

```

## Module Structure

### `internal/ssr/config.go`
Handles all SSR configuration loading from environment variables with sensible defaults.

```go
cfg := ssr.LoadConfig()
// or
cfg := ssr.DefaultConfig()
```

**Environment Variables:**
- `SSR_ENABLED` - Enable SSR (default: true)
- `SSR_SERVER_URL` - SSR server URL (default: http://localhost:3001)
- `SSR_PORT` - SSR port (default: 3001)
- `SSR_AUTO_START` - Auto-start SSR process (default: true)
- `SSR_AUTO_BUILD` - Auto-build web app (default: true)
- `SSR_BUILD_TIMEOUT` - Build timeout in seconds (default: 60)
- `SSR_HEALTH_CHECK_INTERVAL` - Health check interval (default: 5)
- `SSR_RETRY_ATTEMPTS` - Number of retry attempts (default: 3)
- `SSR_RETRY_BACKOFF` - Retry backoff multiplier (default: 2)

### `internal/ssr/process.go`
Manages the SSR subprocess lifecycle with resilience features.

```go
pm := ssr.NewProcessManager(cfg, logger)
pm.Start("/path/to/ssr-server.js")
pm.Stop()
pm.HealthCheck() // Returns *HealthStatus
```

**Features:**
- Automatic retry with exponential backoff
- Process state tracking (Stopped, Starting, Running, Stopping, Crashed)
- Graceful shutdown with SIGTERM
- Force kill after timeout
- Health check support

### `internal/ssr/monitor.go`
Background monitoring of the SSR process health.

```go
monitor := ssr.NewMonitor(pm, logger, 5*time.Second)
monitor.Start(ctx)
defer monitor.Stop()
```

**Features:**
- Periodic health checks
- State transition logging
- Crash detection
- Configurable check interval

### `internal/ssr/builder.go`
Handles building the web app and copying assets.

```go
builder := ssr.NewBuilder(logger, "/path/to/web", "/path/to/public")
builder.BuildIfNeeded() // Builds only if needed
builder.Clean()        // Remove build artifacts
```

**Features:**
- Check if build is needed
- Automatic build with pnpm
- Asset copying to public directory
- Clean build artifacts

### `internal/ssr/manager.go`
Orchestrates all SSR components.

```go
manager := ssr.NewManager(logger, "", "") // Auto-discovers paths
manager.Start()
defer manager.Stop()

manager.IsReady()     // Check if SSR is ready
manager.HealthCheck() // Get health status
manager.String()      // String representation
```

**Features:**
- Auto-discovers web and public directories
- Coordinates all components
- Clean Start/Stop lifecycle

## Scripts

### `scripts/start-ssr.sh`
Starts the SSR server with auto-recovery.

```bash
./scripts/start-ssr.sh --port 3001 --mode production --max-retries 3
```

**Options:**
- `--port PORT` - Set SSR port
- `--mode MODE` - development or production
- `--healthz URL` - Health check URL
- `--max-retries N` - Max restart attempts
- `--retry-delay S` - Delay between retries

### `scripts/monitor-ssr.sh`
Monitors SSR health and can trigger alerts.

```bash
./scripts/monitor-ssr.sh --url http://localhost:3001/health --interval 5
```

**Options:**
- `--url URL` - Health check URL
- `--interval S` - Check interval
- `--max-fail N` - Failures before alert
- `--on-fail CMD` - Command on failure
- `--log FILE` - Log to file

## Usage in main.go

```go
import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ssr"

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    // Create SSR manager (auto-discovers paths)
    ssrManager := ssr.NewManager(logger, "", "")
    defer ssrManager.Stop()
    
    // Start SSR
    if err := ssrManager.Start(); err != nil {
        logger.Warn("SSR start failed, using SPA fallback", "err", err)
    }
    
    // Check status
    if ssrManager.IsReady() {
        logger.Info("SSR is ready")
    }
    
    // Access health status
    health := ssrManager.HealthCheck()
    logger.Info("SSR status", "health", health)
}
```

## Process States

| State | Description |
|-------|-------------|
| `Stopped` | Process is not running |
| `Starting` | Process is starting up |
| `Running` | Process is running and healthy |
| `Stopping` | Process is shutting down |
| `Crashed` | Process has crashed |

## Health Status

```go
type HealthStatus struct {
    State      string        // Current state
    Ready      bool          // Is ready to serve
    Healthy    bool          // Is healthy (passes HTTP check)
    Uptime     time.Duration // How long running
    PID        int           // Process ID
    HTTPStatus int           // Last HTTP status code
    Error      string        // Error message if unhealthy
}
```

## Environment Variables

All configuration can be done via environment variables:

```bash
export SSR_ENABLED=true
export SSR_SERVER_URL=http://localhost:3001
export SSR_PORT=3001
export SSR_AUTO_START=true
export SSR_AUTO_BUILD=true
export SSR_BUILD_TIMEOUT=60
export SSR_HEALTH_CHECK_INTERVAL=5
export SSR_RETRY_ATTEMPTS=3
export SSR_RETRY_BACKOFF=2
```
