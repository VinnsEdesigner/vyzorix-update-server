import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useVyzorErrorStore } from '@/stores/vyzor-error-store';
import type { VyzorErrorClassification } from '@/lib/error';

const networkClassification: VyzorErrorClassification = {
  category: 'network',
  recoverable: true,
  retryable: true,
};

describe('vyzor-error-store', () => {
  beforeEach(() => {
    useVyzorErrorStore.getState().dismiss();
    vi.restoreAllMocks();
  });

  it('starts with no error', () => {
    expect(useVyzorErrorStore.getState().error).toBeNull();
  });

  it('reports an error with classification', () => {
    const error = new Error('Connection failed');
    useVyzorErrorStore.getState().report(error, networkClassification, 'fetch-versions');
    const state = useVyzorErrorStore.getState();
    expect(state.error).not.toBeNull();
    expect(state.error?.message).toBe('Connection failed');
    expect(state.error?.classification).toEqual(networkClassification);
    expect(state.error?.context).toBe('fetch-versions');
    expect(state.error?.retryCount).toBe(0);
  });

  it('generates unique error IDs', () => {
    useVyzorErrorStore.getState().report(new Error('first'), networkClassification);
    const firstId = useVyzorErrorStore.getState().error?.id;
    useVyzorErrorStore.getState().report(new Error('second'), networkClassification);
    const secondId = useVyzorErrorStore.getState().error?.id;
    expect(firstId).not.toBe(secondId);
  });

  it('dismisses the error', () => {
    useVyzorErrorStore.getState().report(new Error('test'), networkClassification);
    expect(useVyzorErrorStore.getState().error).not.toBeNull();
    useVyzorErrorStore.getState().dismiss();
    expect(useVyzorErrorStore.getState().error).toBeNull();
  });

  it('increments retry count', () => {
    useVyzorErrorStore.getState().report(new Error('test'), networkClassification);
    expect(useVyzorErrorStore.getState().error?.retryCount).toBe(0);
    useVyzorErrorStore.getState().incrementRetry();
    expect(useVyzorErrorStore.getState().error?.retryCount).toBe(1);
    useVyzorErrorStore.getState().incrementRetry();
    expect(useVyzorErrorStore.getState().error?.retryCount).toBe(2);
  });

  it('does nothing when incrementing retry with no error', () => {
    useVyzorErrorStore.getState().incrementRetry();
    expect(useVyzorErrorStore.getState().error).toBeNull();
  });

  it('calls window.location.reload on retry', () => {
    const reloadSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      value: { reload: reloadSpy },
      writable: true,
    });
    useVyzorErrorStore.getState().report(new Error('test'), networkClassification);
    useVyzorErrorStore.getState().retry();
    expect(reloadSpy).toHaveBeenCalledTimes(1);
  });

  it('calls window.location.reload on reload', () => {
    const reloadSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      value: { reload: reloadSpy },
      writable: true,
    });
    useVyzorErrorStore.getState().reload();
    expect(reloadSpy).toHaveBeenCalledTimes(1);
  });

  it('navigates home on goHome', () => {
    const hrefSpy = vi.fn();
    Object.defineProperty(window, 'location', {
      value: { href: '' },
      writable: true,
    });
    Object.defineProperty(window.location, 'href', {
      set: hrefSpy,
      get: () => '/',
    });
    useVyzorErrorStore.getState().goHome();
    expect(hrefSpy).toHaveBeenCalledWith('/');
  });

  it('handles non-Error error values', () => {
    useVyzorErrorStore.getState().report('string error', networkClassification);
    const state = useVyzorErrorStore.getState();
    expect(state.error?.message).toBe('string error');
    expect(state.error?.error).toBeNull();
  });

  it('handles null error values', () => {
    useVyzorErrorStore.getState().report(null, networkClassification);
    const state = useVyzorErrorStore.getState();
    expect(state.error?.message).toBe('Unknown error');
    expect(state.error?.error).toBeNull();
  });
});
