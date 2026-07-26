/**
 * SSL/TLS Certificate Pinning
 * 
 * Protects against man-in-the-middle attacks by validating server certificates
 * against pre-configured public key hashes.
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
}

const DEFAULT_CONFIG: SSLPinConfig = {
  pins: [],
  allowSubdomains: false,
  enforced: true,
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

  validateCertificate(certificate: {
    publicKeySha256?: string;
    fingerprint256?: string;
    pemCertificate?: string;
  }): PinValidationResult {
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
  action?: 'reject' | 'warn' | 'fail-open';
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

export function getSSLPinning(config?: Partial<SSLPinConfig>): SSLPinning {
  if (!sslPinningInstance) {
    sslPinningInstance = new SSLPinning(config);
  } else if (config) {
    sslPinningInstance.configure(config);
  }
  return sslPinningInstance;
}

export function initSSLPinning(config: Partial<SSLPinConfig>): SSLPinning {
  sslPinningInstance = new SSLPinning(config);
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
