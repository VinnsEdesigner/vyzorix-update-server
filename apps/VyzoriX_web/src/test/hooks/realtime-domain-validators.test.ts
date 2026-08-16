/**
 * Unit tests for the realtime domain validators (spec §6.3).
 *
 * These exercise the REAL validateTelemetry / validateEvent / validateCommand
 * type guards exported from @vyzorix/api-client. No mocks — pure functions.
 */
import { describe, it, expect } from 'vitest';
import {
  validateTelemetry,
  validateEvent,
  validateCommand,
} from '@vyzorix/api-client';

describe('validateTelemetry', () => {
  it('accepts a minimal valid frame (deviceId only)', () => {
    expect(validateTelemetry({ deviceId: 'dev-1' })).toBe(true);
  });

  it('accepts a fully populated frame with in-range metrics', () => {
    expect(
      validateTelemetry({
        deviceId: 'dev-1',
        riskScore: 75,
        bufferLevel: 50,
        thermalTemp: 42,
        uptime: 3600,
        audioMode: 1,
      }),
    ).toBe(true);
  });

  it('accepts boundary values 0 and 100 for riskScore + bufferLevel', () => {
    expect(validateTelemetry({ deviceId: 'd', riskScore: 0, bufferLevel: 0 })).toBe(true);
    expect(validateTelemetry({ deviceId: 'd', riskScore: 100, bufferLevel: 100 })).toBe(true);
  });

  it('rejects a riskScore above 100', () => {
    expect(validateTelemetry({ deviceId: 'd', riskScore: 101 })).toBe(false);
  });

  it('rejects a negative riskScore', () => {
    expect(validateTelemetry({ deviceId: 'd', riskScore: -1 })).toBe(false);
  });

  it('rejects a bufferLevel above 100', () => {
    expect(validateTelemetry({ deviceId: 'd', bufferLevel: 150 })).toBe(false);
  });

  it('rejects a NaN riskScore', () => {
    expect(validateTelemetry({ deviceId: 'd', riskScore: Number.NaN })).toBe(false);
  });

  it('rejects a non-string deviceId', () => {
    expect(validateTelemetry({ deviceId: 123 })).toBe(false);
  });

  it('rejects an empty deviceId', () => {
    expect(validateTelemetry({ deviceId: '' })).toBe(false);
  });

  it('rejects a non-numeric thermalTemp', () => {
    expect(validateTelemetry({ deviceId: 'd', thermalTemp: 'hot' })).toBe(false);
  });

  it('rejects null/undefined/primitives', () => {
    expect(validateTelemetry(null)).toBe(false);
    expect(validateTelemetry(undefined)).toBe(false);
    expect(validateTelemetry('frame')).toBe(false);
    expect(validateTelemetry(42)).toBe(false);
  });
});

describe('validateEvent', () => {
  it('accepts a valid event with a known type', () => {
    expect(
      validateEvent({
        id: 'evt-1',
        type: 'DEVICE_CONNECTED',
        deviceId: 'dev-1',
      }),
    ).toBe(true);
  });

  it('accepts every spec event type', () => {
    const types = [
      'DEVICE_CONNECTED',
      'DEVICE_DISCONNECTED',
      'THRESHOLD_BREACH',
      'COMMAND_DELIVERED',
      'COMMAND_FAILED',
      'ERROR',
    ];
    for (const type of types) {
      expect(validateEvent({ id: 'e', type, deviceId: 'd' })).toBe(true);
    }
  });

  it('rejects an unknown event type', () => {
    expect(
      validateEvent({ id: 'e', type: 'SOMETHING_ELSE', deviceId: 'd' }),
    ).toBe(false);
  });

  it('rejects an empty id', () => {
    expect(
      validateEvent({ id: '', type: 'DEVICE_CONNECTED', deviceId: 'd' }),
    ).toBe(false);
  });

  it('rejects a non-string type', () => {
    expect(
      validateEvent({ id: 'e', type: 42, deviceId: 'd' }),
    ).toBe(false);
  });
});

describe('validateCommand', () => {
  it('accepts a valid command with a known command type', () => {
    expect(
      validateCommand({
        dispatchId: 'disp-1',
        deviceImei: '356938035643809',
        command: 'FORCE_SPEAKER',
        parameters: { active: true },
        priority: 'high',
      }),
    ).toBe(true);
  });

  it('accepts every spec command type', () => {
    const types = [
      'FORCE_SPEAKER',
      'RESET_AUDIO_HAL',
      'TOGGLE_CAPTURE',
      'REINIT_PROJECTION',
      'DUMP_FLIGHT_DATA',
      'UPLOAD_CRASH_ZIP',
      'SET_LOG_LEVEL',
      'WAKE_UP_UPDATER',
    ];
    for (const command of types) {
      expect(
        validateCommand({
          dispatchId: 'd',
          deviceImei: 'imei',
          command,
          parameters: {},
          priority: 'normal',
        }),
      ).toBe(true);
    }
  });

  it('accepts every priority level', () => {
    for (const priority of ['high', 'normal', 'low'] as const) {
      expect(
        validateCommand({
          dispatchId: 'd',
          deviceImei: 'imei',
          command: 'FORCE_SPEAKER',
          parameters: {},
          priority,
        }),
      ).toBe(true);
    }
  });

  it('accepts a command without a priority (optional)', () => {
    expect(
      validateCommand({
        dispatchId: 'd',
        deviceImei: 'imei',
        command: 'FORCE_SPEAKER',
        parameters: {},
      }),
    ).toBe(true);
  });

  it('rejects an unknown command type', () => {
    expect(
      validateCommand({
        dispatchId: 'd',
        deviceImei: 'imei',
        command: 'SELF_DESTRUCT',
        parameters: {},
      }),
    ).toBe(false);
  });

  it('rejects an unknown priority', () => {
    expect(
      validateCommand({
        dispatchId: 'd',
        deviceImei: 'imei',
        command: 'FORCE_SPEAKER',
        parameters: {},
        priority: 'urgent',
      }),
    ).toBe(false);
  });

  it('rejects an empty dispatchId', () => {
    expect(
      validateCommand({
        dispatchId: '',
        deviceImei: 'imei',
        command: 'FORCE_SPEAKER',
        parameters: {},
      }),
    ).toBe(false);
  });

  it('rejects an empty deviceImei', () => {
    expect(
      validateCommand({
        dispatchId: 'd',
        deviceImei: '',
        command: 'FORCE_SPEAKER',
        parameters: {},
      }),
    ).toBe(false);
  });
});
