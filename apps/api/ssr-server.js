// @vyzorix/ssr-server - Node.js SSR Server for TanStack Start
// This server handles SSR rendering for the React app
// Go server can proxy requests to this server

import { createServer } from '@tanstack/react-start/server-entry'
import { createRequestHandler } from '@tanstack/react-start/server'
import express from 'express'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const PORT = process.env.SSR_PORT || 3001
const isProduction = process.env.NODE_ENV === 'production'

// Create TanStack Start SSR server
const startServer = createServer({
  // In production, we'll use the built files
  // In development, Vite handles this
  build: isProduction ? {
    client: path.join(__dirname, '../web/dist/client'),
    server: path.join(__dirname, '../web/dist/server'),
  } : undefined,
})

// Create Express app
const app = express()

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ ok: true, ssr: true, mode: isProduction ? 'production' : 'development' })
})

// Handle all other requests with TanStack Start SSR
app.all('*', createRequestHandler({ startServer }))

// Start server
app.listen(PORT, () => {
  console.log(`✅ SSR Server ready on http://localhost:${PORT}`)
  console.log(`📦 Mode: ${isProduction ? 'production' : 'development'}`)
  console.log(`🔄 Proxy from Go server to this SSR server`)
})

// Export for testing
export { app, startServer }