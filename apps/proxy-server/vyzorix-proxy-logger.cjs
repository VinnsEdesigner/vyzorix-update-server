/**
 * vyzorix-proxy-logger.cjs
 *
 * Structured logging for the proxy server.
 * Colorized output for dev, plain for prod.
 */

'use strict';

let pc = null;
try {
  pc = require('picocolors');
} catch {
  // picocolors not available — create a no-op color shim.
  pc = function (s) { return s; };
  pc.dim = pc.cyan = pc.yellow = pc.red = pc.green = pc.magenta = pc.bold = pc.gray = pc.white = function (s) { return s; };
}

function ts() {
  return new Date().toISOString().slice(11, 23);
}

function log(level, message, data) {
  var color = { debug: pc.dim, info: pc.cyan, warn: pc.yellow, error: pc.red }[level] || pc.white;
  var sym = { debug: '·', info: '→', warn: '!', error: '✗' }[level] || '?';
  var prefix = pc.dim(ts()) + ' ' + color(sym) + ' ';
  if (data && Object.keys(data).length > 0) {
    var pairs = Object.entries(data)
      .map(function (entry) {
        var k = entry[0], v = entry[1];
        return pc.dim(k + '=') + (typeof v === 'string' ? v : JSON.stringify(v));
      })
      .join(' ');
    console.log(prefix + message + ' ' + pc.dim(pairs));
  } else {
    console.log(prefix + message);
  }
}

function info(message, data) { log('info', message, data); }
function warn(message, data) { log('warn', message, data); }
function error(message, data) { log('error', message, data); }
function debug(message, data) { log('debug', message, data); }

function logRequest(method, path, status, durationMs, extra) {
  var statusColor =
    status >= 500 ? pc.red : status >= 400 ? pc.yellow : status >= 300 ? pc.cyan : pc.green;
  var signed = extra && extra.signed ? pc.magenta('S') + ' ' : '';
  var intercepted = extra && extra.intercepted ? pc.blue('I') + ' ' : '';
  console.log(
    pc.dim(ts()) + ' ' + signed + intercepted +
    pc.bold(method.padEnd(6)) + ' ' + pc.dim(path) + ' → ' +
    statusColor(String(status)) + ' ' + pc.dim(durationMs + 'ms')
  );
}

function banner(lines) {
  console.log('');
  lines.forEach(function (line) { console.log(pc.magenta(line)); });
  console.log('');
}

function kv(key, value) {
  console.log('  ' + pc.bold(pc.cyan(key + ':')) + ' ' + pc.white(value));
}

function divider() {
  console.log('  ' + pc.gray('='.repeat(56)));
}

function success(message) {
  console.log('  ' + pc.green('*') + ' ' + message);
}

module.exports = { log, info, warn, error, debug, logRequest, banner, kv, divider, success, pc };
