# Frontend Bug Fixes - Vyzorix Update Server Dashboard

**Date:** 2026-06-06  
**Branch:** `fix/websocket-origin-validation`  
**Status:** ✅ All 3 bugs fixed

---

## Summary

| # | Bug | File | Severity | Status |
|---|-----|------|----------|--------|
| FE-1 | Missing input validation on server URL | `_app.settings.connection.tsx` | Medium | ✅ Fixed |
| FE-2 | Race condition in operator name auto-save | `_app.settings.operator.tsx` | Low | ✅ Fixed |
| FE-3 | Hardcoded "Nokia C22" device name | `_app.dashboard.tsx` | Low | ✅ Fixed |

---

## Bug #FE-1: Missing Input Validation on Server URL ✅ FIXED

### Location
`src/routes/_app.settings.connection.tsx`

### Fix Applied
- Added `isValidServerUrl()` validation function
- Validates URL has proper protocol (`http://` or `https://`)
- Shows error message inline for invalid URLs
- Prevents saving with malformed URLs

### Implementation
```tsx
function isValidServerUrl(url: string): boolean {
  if (!url.trim()) return false;
  try {
    const u = new URL(url);
    return u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
}

const validateForm = (): boolean => {
  const trimmed = serverUrl.trim();
  if (!trimmed) {
    setServerUrlError("Server URL is required");
    return false;
  }
  if (!isValidServerUrl(trimmed)) {
    setServerUrlError("Invalid URL. Include http:// or https:// (e.g., http://localhost:3000)");
    return false;
  }
  return true;
};
```

---

## Bug #FE-2: Race Condition in Operator Name Auto-Save ✅ FIXED

### Location
`src/routes/_app.settings.operator.tsx`

### Fix Applied
- Added `lastSavedName` state to track the last saved value
- Added `nameRef` to access latest name in timeout callback
- Used `useCallback` for the save function to avoid stale closures
- Properly clears timer on cleanup

### Implementation
```tsx
const [lastSavedName, setLastSavedName] = useState(stored?.name ?? "");
const nameRef = useRef(name); // Keep ref in sync with latest name
nameRef.current = name;

// Debounced save function - reads fresh values from refs
const saveName = useCallback(async (nameToSave: string) => {
  setSavingName(true);
  try {
    await updateName(DEFAULT_SERVER_URL, nameToSave.trim());
    setLastSavedName(nameToSave.trim()); // Track saved value
    toast.success("Display name saved");
  } catch (e) {
    toast.error("Failed to save name", { description: e instanceof Error ? e.message : "try again" });
  } finally {
    setSavingName(false);
  }
}, []);
```

---

## Bug #FE-3: Hardcoded "Nokia C22" Device Name ✅ FIXED

### Location
`src/routes/_app.dashboard.tsx`

### Fix Applied
- Added `formatDeviceClass()` function
- Converts deviceClass (e.g., 'nokia_c22') to display name (e.g., 'Nokia C22')
- Uses deviceClass from API response

### Implementation
```tsx
// Format device class for display (e.g., "nokia_c22" -> "Nokia C22")
function formatDeviceClass(deviceClass: string | undefined): string {
  if (!deviceClass) return "Unknown Device";
  return deviceClass
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

// Usage in component:
<CardTitle className="flex items-center gap-2">
  {formatDeviceClass(status.data?.deviceClass)} <span className="text-xs font-normal text-muted-foreground">· {deviceId}</span>
</CardTitle>
```

---

## Additional Work

### Missing Log Components Created
Created `LogDock` and `LogConsole` components:
- `src/components/logs/log-dock.tsx` - Docked log panel
- `src/components/logs/log-console.tsx` - Full-page log viewer with filtering

### Tests Added
Added comprehensive test suite in `src/lib/settings.test.ts`:
- URL validation tests (21 test cases)
- Device class formatting tests (9 test cases)
- Config storage tests (4 test cases)
- Settings persistence tests (3 test cases)

**All 54 tests passing**

---

## Commit History

```
d21dd82 fix: implement frontend bug fixes
a3be224 docs: document frontend bug findings
f39750b revert: use strict password policy for user registration
1a8c309 fix: use user-friendly password policy for registration
6988c5d fix: implement bug fixes #13-15
3faa536 fix: implement bug fixes #7-12
642ba75 fix: add command_secrets_hash column with bcrypt hashing
edf4a8a fix: implement bug fixes #2-5
518628b fix: implement WebSocket origin validation
```

---

*Document updated after implementing fixes*