import { describe, it, expect } from 'vitest';
import {
  VyzorNoopAnalyticsAdapter,
  VYZOR_ANALYTICS_EVENTS,
} from '@/lib/analytics';

describe('VyzorNoopAnalyticsAdapter', () => {
  const adapter = new VyzorNoopAnalyticsAdapter();

  it('implements all adapter methods', () => {
    expect(typeof adapter.identify).toBe('function');
    expect(typeof adapter.track).toBe('function');
    expect(typeof adapter.page).toBe('function');
    expect(typeof adapter.flush).toBe('function');
    expect(typeof adapter.reset).toBe('function');
    expect(typeof adapter.setConsent).toBe('function');
  });

  it('all methods are no-ops (do not throw)', () => {
    expect(() => adapter.identify('user-1')).not.toThrow();
    expect(() => adapter.track('event', { foo: 1 })).not.toThrow();
    expect(() => adapter.page('home')).not.toThrow();
    expect(() => adapter.flush()).not.toThrow();
    expect(() => adapter.reset()).not.toThrow();
    expect(() => adapter.setConsent('granted')).not.toThrow();
  });
});

describe('VYZOR_ANALYTICS_EVENTS', () => {
  it('contains expected event names', () => {
    expect(VYZOR_ANALYTICS_EVENTS.appLoaded).toBe('app_loaded');
    expect(VYZOR_ANALYTICS_EVENTS.pageViewed).toBe('page_viewed');
    expect(VYZOR_ANALYTICS_EVENTS.errorReported).toBe('error_reported');
    expect(VYZOR_ANALYTICS_EVENTS.consentDecision).toBe('consent_decision');
    expect(VYZOR_ANALYTICS_EVENTS.updatePushed).toBe('update_pushed');
  });

  it('all values are snake_case strings', () => {
    for (const value of Object.values(VYZOR_ANALYTICS_EVENTS)) {
      expect(typeof value).toBe('string');
      expect(value).toMatch(/^[a-z_]+$/);
    }
  });
});
