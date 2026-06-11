# Setup Guide for Vyzorix Update Server

## Quick Start

### 1. Install Dependencies Without Hanging

```bash
# Disable pnpm prompts to prevent hanging
export COREPACK_ENABLE_DOWNLOAD_PROMPT=0
export COREPACK_ENABLE_AUTOINSTALL=0
export PNPM_TELEMETRY=0

# Install pnpm and dependencies
cd /workspace/project/vyzorix-update-server
pnpm install
```

### 2. Install Go

```bash
# Download and install Go
export PATH=$PATH:/usr/local/go/bin

# Verify Go installation
go version
```

### 3. Build the Project

```bash
# Build config package
cd /workspace/project/vyzorix-update-server
pnpm --filter @vyzorix/config run build

# Build web app
pnpm --filter @vyzorix/web run build

# Build Go server
export PATH=$PATH:/usr/local/go/bin
cd apps/api
go build -o vyzorix-server .
```

### 4. Run the Server

```bash
# Start Go server
cd /workspace/project/vyzorix-update-server/apps/api
./vyzorix-server

# Server will be available at http://localhost:3000
```

## Development Mode

### Run Vite Dev Server

```bash
cd /workspace/project/vyzorix-update-server/apps/web
pnpm run dev -- --ssr

# Vite server will be available at http://localhost:5173
```

### Run Go Server

```bash
cd /workspace/project/vyzorix-update-server/apps/api
export SSR_ENABLED=false  # Disable SSR for development
export VYZORIX_PUBLIC_DIR="./public"
./vyzorix-server
```

## Production Mode

### Build for Production

```bash
# Build web app
cd /workspace/project/vyzorix-update-server/apps/web
pnpm run build

# Copy assets to Go public directory
cp -r dist/client/* ../api/public/

# Build Go server
export PATH=$PATH:/usr/local/go/bin
cd apps/api
go build -o vyzorix-server .
```

### Run Production Server

```bash
# Start Node.js SSR server (in separate terminal)
cd /workspace/project/vyzorix-update-server/apps/api
npm install
npm run prod

# Start Go server (in another terminal)
export SSR_ENABLED=true
export SSR_SERVER_URL=http://localhost:3001
./vyzorix-server
```

## Environment Variables

### Required Variables

```bash
# Go server
export VYZORIX_PUBLIC_DIR="./public"
export PORT="3000"

# SSR (optional)
export SSR_ENABLED=true
export SSR_SERVER_URL=http://localhost:3001
export SSR_PORT=3001
```

### Create .env File

```env
# .env file
VYZORIX_PUBLIC_DIR=./public
PORT=3000
SSR_ENABLED=true
SSR_SERVER_URL=http://localhost:3001
SSR_PORT=3001
```

## Troubleshooting

### pnpm Hanging

**Solution:** Set these environment variables before running pnpm:

```bash
export COREPACK_ENABLE_DOWNLOAD_PROMPT=0
export COREPACK_ENABLE_AUTOINSTALL=0
export PNPM_TELEMETRY=0
```

### Go Not Found

**Solution:** Install Go and add to PATH:

```bash
wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version
```

### Build Failures

**Solution:** Check for syntax errors and missing dependencies:

```bash
# Check Go syntax
cd apps/api
go build -o vyzorix-server .

# Check web app
cd apps/web
pnpm run build
```

## Committing Changes

### Commit to GitHub

```bash
# Add all changes
git add -A

# Commit with message
git commit -m "feat: add SSR support with Node.js and Go proxy"

# Push to branch
git push origin fix/vite-build-issues
```

### Create Pull Request

```bash
# Create PR from branch
gh pr create --base main --head fix/vite-build-issues --title "feat: SSR support" --body "Added SSR with Node.js server and Go proxy"
```

## Project Structure

```
vyzorix-update-server/
├── apps/
│   ├── api/              # Go backend
│   │   ├── ssr-server.js  # Node.js SSR server
│   │   └── main.go        # Go server
│   └── web/              # React frontend
│       ├── src/           # Source code
│       └── vite.config.ts # Vite config
├── packages/
│   └── config/           # Shared config
└── AGENTS/
    └── SETUP_GUIDE.md    # This file
```

## Key Files

- `apps/api/ssr-server.js` - Node.js SSR server
- `apps/api/pkg/config/ssr.go` - SSR configuration
- `apps/api/internal/api/middleware/ssr-proxy.go` - SSR proxy middleware
- `SSR-SETUP.md` - SSR setup documentation
- `AGENTS/SETUP_GUIDE.md` - This file

## Support

For issues:
1. Check logs first
2. Verify all servers are running
3. Test endpoints individually
4. Check browser console and network tab