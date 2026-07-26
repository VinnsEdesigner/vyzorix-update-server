/**
 * SSL/TLS Certificate Pinning
 * 
 * NOTE: This implementation is a placeholder/layout for future use.
 * 
 * PURPOSE:
 * - Protects APK daemon server-to-server connections against MITM attacks
 * - APK daemon communicates directly with this API server (not browser-based)
 * - Will be activated once server-side SSL pin configuration is implemented
 * 
 * TO ACTIVATE: Set SSLPIN_ENABLED=true or call initSSLPinning({ enabled: true })
 * 
 * TODO (Server-side requirements before activation):
 * - Create SSL pin configuration endpoint: GET /.well-known/ssl-pins
 * - Add ssl_pin_configs table to database
 * - Implement SSLPinConfigService
 * - Add SSL pin fields to DeviceSettings for per-device pinning
 * - Server provides SHA-256 hashes of valid public keys
 */

import { createHash } from 'crypto';

export interface PinnedCertificate {
  subject: string;
  issuer: string;
  serialNumber: string;
  publicKeySha256: string;
  validFrom: Date;
  validTo: Date;
}

export interface SSLPinConfig {
  pins: string[];
  extraPins?: string[];
  allowSubdomains?: boolean;
  enforced?: boolean;
  reportUri?: string;
  enabled?: boolean;
}

const DEFAULT_CONFIG: SSLPinConfig = {
  pins: [],
  allowSubdomains: false,
  enforced: true,
  enabled: false,
};

export class SSLPinning {
  private config: SSLPinConfig;
  private connectionLog: ConnectionRecord[] = [];

  constructor(config: Partial<SSLPinConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  configure(config: Partial<SSLPinConfig>): void {
    this.config = { ...this.config, ...config };
  }

  addPin(sha256Hash: string): void {
    if (!this.config.pins.includes(sha256Hash)) {
      this.config.pins.push(sha256Hash);
    }
  }

  removePin(sha256Hash: string): void {
    this.config.pins = this.config.pins.filter(p => p !== sha256Hash);
  }

  getPins(): string[] {
    return [...this.config.pins];
  }

  isConfigured(): boolean {
    return this.config.pins.length > 0;
  }

  isEnabled(): boolean {
    return this.config.enabled ?? false;
  }

  enable(): void {
    this.config.enabled = true;
  }

  disable(): void {
    this.config.enabled = false;
  }

  validateCertificate(certificate: {
    publicKeySha256?: string;
    fingerprint256?: string;
    pemCertificate?: string;
  }): PinValidationResult {
    if (!this.config.enabled) {
      return {
        valid: true,
        action: 'bypassed',
      };
    }

    if (!this.isConfigured()) {
      return {
        valid: false,
        error: 'No pins configured',
        action: 'fail-open',
      };
    }

    const certificateHash = this.extractPublicKeyHash(certificate);

    if (!certificateHash) {
      return {
        valid: false,
        error: 'Could not extract certificate hash',
        action: this.config.enforced ? 'reject' : 'warn',
      };
    }

    const isValid = this.config.pins.some(pin => 
      constantTimeCompare(pin.toLowerCase(), certificateHash.toLowerCase())
    );

    if (isValid) {
      return {
        valid: true,
        matchedPin: certificateHash,
      };
    }

    const matchedExtraPin = this.config.extraPins?.some(pin =>
      constantTimeCompare(pin.toLowerCase(), certificateHash.toLowerCase())
    );

    if (matchedExtraPin) {
      return {
        valid: true,
        matchedPin: certificateHash,
        isBackupPin: true,
      };
    }

    return {
      valid: false,
      error: 'Certificate does not match any pinned hash',
      expectedPins: this.config.pins,
      actualPin: certificateHash,
      action: this.config.enforced ? 'reject' : 'warn',
    };
  }

  private extractPublicKeyHash(certificate: {
    publicKeySha256?: string;
    fingerprint256?: string;
    pemCertificate?: string;
  }): string | null {
    if (certificate.publicKeySha256) {
      return certificate.publicKeySha256.toLowerCase();
    }

    if (certificate.fingerprint256) {
      return certificate.fingerprint256.toLowerCase().replace(/:/g, '');
    }

    if (certificate.pemCertificate) {
      return this.calculateSha256(certificate.pemCertificate);
    }

    return null;
  }

  calculateSha256(data: string): string {
    return createHash('sha256').update(data).digest('hex').toLowerCase();
  }

  validateHostname(hostname: string, allowedDomains: string[]): boolean {
    if (!this.config.allowSubdomains) {
      return allowedDomains.includes(hostname);
    }

    return allowedDomains.some(domain => 
      hostname === domain || hostname.endsWith(`.${domain}`)
    );
  }

  recordConnection(record: ConnectionRecord): void {
    this.connectionLog.push({
      ...record,
      timestamp: Date.now(),
    });

    if (this.connectionLog.length > 1000) {
      this.connectionLog = this.connectionLog.slice(-500);
    }
  }

  getConnectionLog(): ConnectionRecord[] {
    return [...this.connectionLog];
  }

  clearConnectionLog(): void {
    this.connectionLog = [];
  }

  getFailedConnections(): ConnectionRecord[] {
    return this.connectionLog.filter(r => !r.success);
  }

  isEnforced(): boolean {
    return this.config.enforced ?? true;
  }

  setEnforced(enforced: boolean): void {
    this.config.enforced = enforced;
  }
}

export interface PinValidationResult {
  valid: boolean;
  matchedPin?: string;
  isBackupPin?: boolean;
  error?: string;
  expectedPins?: string[];
  actualPin?: string;
  action?: 'reject' | 'warn' | 'fail-open' | 'bypassed';
}

export interface ConnectionRecord {
  hostname: string;
  port?: number;
  timestamp?: number;
  success: boolean;
  error?: string;
  certificateHash?: string;
}

function constantTimeCompare(a: string, b: string): boolean {
  if (a.length !== b.length) {
    return false;
  }
  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return result === 0;
}

let sslPinningInstance: SSLPinning | null = null;

function isSSLEnabledFromEnv(): boolean {
  if (typeof import.meta !== 'undefined') {
    const meta = import.meta as { env?: Record<string, string | undefined> };
    const val = meta.env?.['SSLPIN_ENABLED'] || meta.env?.['VITE_SSLPIN_ENABLED'];
    if (val === 'true' || val === '1') return true;
  }
  
  if (typeof globalThis !== 'undefined') {
    const g = globalThis as Record<string, unknown>;
    if (g.process && typeof g.process === 'object') {
      const proc = g.process as Record<string, unknown>;
      const env = proc.env as Record<string, string | undefined> | undefined;
      const val = env?.['SSLPIN_ENABLED'] || env?.['VITE_SSLPIN_ENABLED'];
      if (val === 'true' || val === '1') return true;
    }
  }
  
  return false;
}

export function getSSLPinning(config?: Partial<SSLPinConfig>): SSLPinning {
  if (!sslPinningInstance) {
    const envEnabled = isSSLEnabledFromEnv();
    sslPinningInstance = new SSLPinning({ ...config, enabled: config?.enabled ?? envEnabled });
  } else if (config) {
    sslPinningInstance.configure(config);
  }
  return sslPinningInstance;
}

export function initSSLPinning(config: Partial<SSLPinConfig> = {}): SSLPinning {
  const envEnabled = isSSLEnabledFromEnv();
  sslPinningInstance = new SSLPinning({ ...config, enabled: config.enabled ?? envEnabled });
  return sslPinningInstance;
}

export function resetSSLPinning(): void {
  if (sslPinningInstance) {
    sslPinningInstance.clearConnectionLog();
  }
  sslPinningInstance = null;
}

export function validateServerCertificate(certificate: {
  publicKeySha256?: string;
  fingerprint256?: string;
  pemCertificate?: string;
}): PinValidationResult {
  const pinning = getSSLPinning();
  return pinning.validateCertificate(certificate);
}

export function isSSLPinningEnforced(): boolean {
  const pinning = getSSLPinning();
  return pinning.isEnforced();
}
