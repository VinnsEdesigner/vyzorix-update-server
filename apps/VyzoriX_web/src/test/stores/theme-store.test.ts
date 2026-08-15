import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useThemeStore } from '@/stores/theme-store';

describe('useThemeStore', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
    useThemeStore.setState({
      mode: 'system',
      resolvedTheme: 'light',
    });
  });

  it('starts with system mode by default', () => {
    const fresh = useThemeStore.getState();
    expect(fresh.mode).toBe('system');
  });

  it('setMode updates mode and resolvedTheme', () => {
    useThemeStore.getState().setMode('dark');
    const state = useThemeStore.getState();
    expect(state.mode).toBe('dark');
    expect(state.resolvedTheme).toBe('dark');
  });

  it('setMode persists to localStorage', () => {
    useThemeStore.getState().setMode('dark');
    expect(localStorage.getItem('vyzorix-theme')).toBe('dark');
  });

  it('setMode applies dark class to document', () => {
    useThemeStore.getState().setMode('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('setMode to light removes dark class', () => {
    useThemeStore.getState().setMode('dark');
    useThemeStore.getState().setMode('light');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('toggle switches between light and dark', () => {
    useThemeStore.getState().setMode('light');
    useThemeStore.getState().toggle();
    expect(useThemeStore.getState().resolvedTheme).toBe('dark');
    useThemeStore.getState().toggle();
    expect(useThemeStore.getState().resolvedTheme).toBe('light');
  });

  it('applyToDocument syncs the document class with current mode', () => {
    useThemeStore.setState({ mode: 'dark', resolvedTheme: 'light' });
    useThemeStore.getState().applyToDocument();
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('system mode resolves via matchMedia', () => {
    const matchMediaSpy = vi.fn().mockReturnValue({ matches: true });
    vi.stubGlobal('matchMedia', matchMediaSpy);
    useThemeStore.getState().setMode('system');
    expect(useThemeStore.getState().resolvedTheme).toBe('dark');
    vi.unstubAllGlobals();
  });
});
