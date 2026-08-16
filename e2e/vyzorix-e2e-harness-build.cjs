/**
 * vyzorix-e2e-harness-build.cjs
 *
 * Builds a self-contained browser bundle of the API Client for E2E testing.
 *
 * The API Client normally relies on Vite's `import.meta.env` for configuration.
 * In the headless-browser test harness (served by the proxy — no Vite), we
 * replace `import.meta.env` with a `window.__VYZORIX_ENV` global so the client
 * picks up VITE_API_URL="/api" etc. from the harness page.
 *
 * Output:
 *   e2e/harness/dist/vyzorix-api-client.browser.js  (ESM, self-contained)
 *
 * Usage:
 *   node e2e/vyzorix-e2e-harness-build.cjs
 */

'use strict';

const fs = require('node:fs');
const path = require('node:path');

const esbuild = require(
  path.resolve(__dirname, '..', 'node_modules', '.pnpm', 'esbuild@0.25.12', 'node_modules', 'esbuild')
);

const ROOT = path.resolve(__dirname, '..');
const API_CLIENT_SRC = path.join(ROOT, 'packages', 'API_Client', 'src');
const OUT_DIR = path.join(__dirname, 'harness', 'dist');

async function build() {
  fs.mkdirSync(OUT_DIR, { recursive: true });

  // Create a barrel entry that re-exports the public browser-safe surface.
  // The main index.ts only imports from crypto/browser-sign (Web Crypto API),
  // NOT from crypto/index.ts (node:crypto). The Node-only websocket-client
  // (which imports node crypto for signWebSocketConnect) is excluded — WS
  // tests use the browser's native WebSocket API directly.
  const entryContent = `
    export * from '${path.join(API_CLIENT_SRC, 'index.ts').replace(/\\/g, '/')}';
  `;
  const entryFile = path.join(OUT_DIR, '_entry.js');
  fs.writeFileSync(entryFile, entryContent);

  console.log('[harness-build] Bundling API Client for browser...');

  const result = await esbuild.build({
    entryPoints: [entryFile],
    bundle: true,
    format: 'esm',
    platform: 'browser',
    target: 'es2022',
    outfile: path.join(OUT_DIR, 'vyzorix-api-client.browser.js'),
    define: {
      // Replace import.meta.env with a window global. esbuild handles
      // import.meta.env.X as a property access on import.meta.env, so defining
      // import.meta.env as a JSON string makes all accesses resolve to the
      // window object's properties at runtime.
      'import.meta.env': 'window.__VYZORIX_ENV',
    },
    loader: { '.ts': 'ts' },
    sourcemap: 'inline',
    logLevel: 'info',
    external: [],
  });

  if (result.errors.length > 0) {
    console.error('[harness-build] Build failed with errors');
    process.exit(1);
  }

  // Clean up temp entry
  fs.unlinkSync(entryFile);

  const outPath = path.join(OUT_DIR, 'vyzorix-api-client.browser.js');
  const stats = fs.statSync(outPath);
  console.log(`[harness-build] Bundle: ${outPath} (${(stats.size / 1024).toFixed(1)} KB)`);

  // Copy harness files into the proxy's static serving directory so the proxy
  // serves them directly (no duplicate server). The proxy serves from
  // apps/VyzoriX_web/dist/client/ — we place the harness at /__e2e__/ and the
  // bundle at /__e2e__/assets/.
  const proxyStaticDir = path.join(ROOT, 'apps', 'VyzoriX_web', 'dist', 'client');
  const e2eStaticDir = path.join(proxyStaticDir, '__e2e__');
  const e2eAssetsDir = path.join(e2eStaticDir, 'assets');
  fs.mkdirSync(e2eAssetsDir, { recursive: true });

  const harnessSrc = path.join(__dirname, 'harness', 'index.html');
  const harnessDst = path.join(e2eStaticDir, 'index.html');
  fs.copyFileSync(harnessSrc, harnessDst);
  console.log(`[harness-build] Harness: ${harnessDst}`);

  const bundleDst = path.join(e2eAssetsDir, 'vyzorix-api-client.browser.js');
  fs.copyFileSync(outPath, bundleDst);
  console.log(`[harness-build] Installed: ${bundleDst}`);
  console.log('[harness-build] Access at: http://localhost:<proxy-port>/__e2e__/');
}

build().catch((err) => {
  console.error('[harness-build] Fatal error:', err);
  process.exit(1);
});
