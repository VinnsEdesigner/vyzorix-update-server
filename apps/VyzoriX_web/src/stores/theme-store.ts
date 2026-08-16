import { createVyzorStore } from '@/lib/state';

export type ThemeMode = 'light' | 'dark' | 'system';

const STORAGE_KEY = 'vyzorix-theme';

function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function readStoredMode(): ThemeMode {
  if (typeof window === 'undefined') return 'system';
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    return stored === 'light' || stored === 'dark' || stored === 'system' ? stored : 'system';
  } catch {
    return 'system';
  }
}

function resolveTheme(mode: ThemeMode): 'light' | 'dark' {
  return mode === 'system' ? getSystemTheme() : mode;
}

function applyTheme(mode: ThemeMode): void {
  if (typeof document === 'undefined') return;
  const resolved = resolveTheme(mode);
  document.documentElement.classList.toggle('dark', resolved === 'dark');
}

export interface ThemeState {
  mode: ThemeMode;
  resolvedTheme: 'light' | 'dark';
  setMode: (mode: ThemeMode) => void;
  toggle: () => void;
  applyToDocument: () => void;
}

export const useThemeStore = createVyzorStore<ThemeState>('ThemeStore', (set, get) => ({
  mode: readStoredMode(),
  resolvedTheme: resolveTheme(readStoredMode()),
  setMode: (mode) => {
    try {
      window.localStorage.setItem(STORAGE_KEY, mode);
    } catch {
      void 0;
    }
    applyTheme(mode);
    set({ mode, resolvedTheme: resolveTheme(mode) });
  },
  toggle: () => {
    const current = get().resolvedTheme;
    get().setMode(current === 'dark' ? 'light' : 'dark');
  },
  applyToDocument: () => {
    applyTheme(get().mode);
  },
}));
