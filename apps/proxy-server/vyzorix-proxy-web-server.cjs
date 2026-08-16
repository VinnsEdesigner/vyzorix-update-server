/**
 * vyzorix-proxy-web-server.cjs
 *
 * Static file serving for the web app.
 *
 * Serves pre-built assets from apps/VyzoriX_web/dist/client/ with SPA fallback
 * (index.html for client-side routing). No vite involvement — the web app must
 * be built first (cd apps/VyzoriX_web && pnpm build).
 *
 * Works identically in dev and prod: the only difference is cache headers
 * and verbosity. In dev, no-cache is used everywhere; in prod, hashed assets
 * get immutable caching.
 */

'use strict';

const fs = require('node:fs');
const path = require('node:path');
const stream = require('node:stream');
var cfg = require('./vyzorix-proxy-config.cjs');
var logger = require('./vyzorix-proxy-logger.cjs');
var config = cfg.config;
var debug = logger.debug;
var warn = logger.warn;
var error = logger.error;

var MIME_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.mjs': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.ttf': 'font/ttf',
  '.eot': 'application/vnd.ms-fontobject',
  '.webp': 'image/webp',
  '.webmanifest': 'application/manifest+json',
  '.map': 'application/json; charset=utf-8',
  '.txt': 'text/plain; charset=utf-8',
  '.xml': 'application/xml; charset=utf-8',
  '.wasm': 'application/wasm',
};

/**
 * Serve a static file or SPA fallback.
 * @param {string} pathname - URL pathname
 * @param {object} res - http.ServerResponse
 */
async function serveStaticFile(pathname, res) {
  var staticDir = config.webStaticDir;

  // Check if dist/client exists.
  try {
    await fs.promises.stat(staticDir);
  } catch {
    var msg =
      '<!DOCTYPE html><html><body>' +
      '<h1>Vyzorix Proxy Server</h1>' +
      '<p>Web app not built. Run:</p>' +
      '<pre>cd apps/VyzoriX_web && pnpm build</pre>' +
      '<p>Then restart the proxy.</p>' +
      '</body></html>';
    warn('Web app dist not found, serving placeholder', { dir: staticDir });
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(msg);
    return;
  }

  // Normalize and prevent path traversal.
  var safePath = path.normalize(pathname).replace(/^(\.\.[/\\])+/, '');
  var filePath = path.join(staticDir, safePath);
  var ext = path.extname(safePath);

  // For extensionless paths, try the directory's index.html first.
  if (!ext) {
    var dirIndex = path.join(filePath, 'index.html');
    try {
      var dirStats = await fs.promises.stat(dirIndex);
      if (dirStats.isFile()) {
        filePath = dirIndex;
        ext = '.html';
      }
    } catch {
      // Not a directory with index.html — fall through to SPA root index.
      filePath = path.join(staticDir, 'index.html');
      ext = '.html';
    }
  }

  try {
    var stats = await fs.promises.stat(filePath);
    if (!stats.isFile()) {
      return serveSpaIndex(staticDir, res);
    }

    var contentType = MIME_TYPES[ext] || 'application/octet-stream';
    var cacheControl;
    if (config.mode === 'production' && safePath.includes('/assets/')) {
      cacheControl = 'public, max-age=31536000, immutable';
    } else {
      cacheControl = 'no-cache';
    }

    res.writeHead(200, {
      'Content-Type': contentType,
      'Cache-Control': cacheControl,
      'Content-Length': stats.size,
    });

    // Stream the file.
    var fileStream = fs.createReadStream(filePath);
    stream.pipeline(fileStream, res, function (err) {
      if (err) {
        error('File stream error', { path: filePath, error: err.message });
        if (!res.writableEnded) res.end();
      }
    });
    debug('Served static file', { path: safePath, type: contentType });
  } catch {
    // SPA fallback: serve index.html for client-side routing.
    return serveSpaIndex(staticDir, res);
  }
}

async function serveSpaIndex(staticDir, res) {
  var indexPath = path.join(staticDir, 'index.html');
  try {
    var stats = await fs.promises.stat(indexPath);
    res.writeHead(200, {
      'Content-Type': 'text/html; charset=utf-8',
      'Cache-Control': 'no-cache',
      'Content-Length': stats.size,
    });
    var fileStream = fs.createReadStream(indexPath);
    stream.pipeline(fileStream, res, function (err) {
      if (err) {
        error('SPA index stream error', { error: err.message });
        if (!res.writableEnded) res.end();
      }
    });
  } catch {
    warn('index.html not found — web app may not be built');
    res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
    res.end(
      '<!DOCTYPE html><html><body>' +
      '<h1>Vyzorix Proxy Server</h1>' +
      '<p>Web app not built. Run: cd apps/VyzoriX_web && pnpm build</p>' +
      '</body></html>'
    );
  }
}

module.exports = { serveStaticFile, MIME_TYPES };
