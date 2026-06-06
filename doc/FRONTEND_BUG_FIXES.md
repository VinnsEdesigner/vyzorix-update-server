# Frontend Bug Fixes - Vyzorix Update Server Dashboard

**Date:** 2026-06-06  
**Branch:** `fix/websocket-origin-validation`  
**Status:** 3 bugs found and documented

---

## Summary

| # | Bug | File | Severity | Status |
|---|-----|------|----------|--------|
| FE-1 | Missing input validation on server URL | `vyzorix-config.tsx` | Medium | Documented |
| FE-2 | Race condition in operator name auto-save | `_app.settings.operator.tsx` | Low | Documented |
| FE-3 | Hardcoded "Nokia C22" device name | `_app.dashboard.tsx` | Low | Documented |

---

## Bug #FE-1: Missing Input Validation on Server URL

### Location
`src/lib/vyzorix-config.tsx` (line 32)

### Description
The server URL input field in settings allows users to enter any arbitrary string without validation. Invalid URLs (e.g., empty strings, malformed URLs, URLs without protocol) are stored and used for API calls, causing silent failures.

### Current Code
```tsx
const save = () => {
  cfg.update({
    serverUrl: serverUrl.trim() || DEFAULT_SERVER_URL,  // Only checks for empty
    // ...
  });
};
```

### Impact
- User enters `localhost:3000` (missing `http://`)
- WebSocket connects to `ws://localhost:3000/...` which fails silently
- Health checks fail with unclear error messages
- User doesn't understand why dashboard isn't loading

### Suggested Fix
```tsx
const isValidUrl = (url: string): boolean => {
  try {
    const u = new URL(url);
    return u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
};

const save = () => {
  const trimmed = serverUrl.trim();
  if (!trimmed) {
    toast.error("Server URL is required");
    return;
  }
  if (!isValidUrl(trimmed)) {
    toast.error("Invalid URL format. Include http:// or https://");
    return;
  }
  cfg.update({ serverUrl: trimmed });
  toast.success("Connection settings saved");
};
```

### Files to Modify
- `src/routes/_app.settings.connection.tsx` - Add URL validation

---

## Bug #FE-2: Race Condition in Operator Name Auto-Save

### Location
`src/routes/_app.settings.operator.tsx` (lines 47-69)

### Description
The auto-save timer for the operator name has a race condition. If the user types rapidly, the cleanup effect may not properly clear the timer, and the saved state check compares against stale `stored` value.

### Current Code
```tsx
useEffect(() => {
  if (saveTimer.current) clearTimeout(saveTimer.current);
  if (!stored) return;

  // Only save if name actually changed from what we have stored
  if (name.trim() && name.trim() !== stored.name) {  // BUG: stored is captured in closure
    saveTimer.current = setTimeout(async () => {
      // ...
    }, 1000);
  }
  // ...
}, [name]);  // stored is NOT in deps
```

### Impact
- If `stored` changes (e.g., after login), the effect may not re-run
- The comparison `name.trim() !== stored.name` uses stale closure value
- Multiple rapid saves may trigger unnecessary API calls

### Suggested Fix
```tsx
useEffect(() => {
  if (saveTimer.current) clearTimeout(saveTimer.current);
  
  const currentStored = getStoredOperator();  // Fresh read
  if (!currentStored) return;

  const trimmedName = name.trim();
  if (!trimmedName) return;
  
  if (trimmedName !== currentStored.name) {
    saveTimer.current = setTimeout(async () => {
      setSavingName(true);
      try {
        await updateName(DEFAULT_SERVER_URL, trimmedName);
        toast.success("Display name saved");
      } catch (e) {
        toast.error("Failed to save name", { description: e instanceof Error ? e.message : "try again" });
      } finally {
        setSavingName(false);
      }
    }, 1000);
  }

  return () => {
    if (saveTimer.current) clearTimeout(saveTimer.current);
  };
}, [name]);
```

### Files to Modify
- `src/routes/_app.settings.operator.tsx`

---

## Bug #FE-3: Hardcoded "Nokia C22" Device Name

### Location
`src/routes/_app.dashboard.tsx` (line 110)

### Description
The dashboard displays "Nokia C22" as a hardcoded string instead of using the actual `deviceClass` from the device status. This will show incorrect information for devices that aren't Nokia C22.

### Current Code
```tsx
<CardTitle className="flex items-center gap-2">
  Nokia C22 <span className="text-xs font-normal text-muted-foreground">· {deviceId}</span>
</CardTitle>
```

### Impact
- Device type is always displayed as "Nokia C22"
- User cannot distinguish between different device types in fleet view
- Misleading when testing with mock devices

### Suggested Fix
```tsx
const deviceDisplayName = status.data?.deviceClass 
  ? status.data.deviceClass.replace(/_/g, " ").replace(/\b\w/g, c => c.toUpperCase())
  : "Unknown Device";

<CardTitle className="flex items-center gap-2">
  {deviceDisplayName} <span className="text-xs font-normal text-muted-foreground">· {deviceId}</span>
</CardTitle>
```

### Files to Modify
- `src/routes/_app.dashboard.tsx`

---

## Additional Observations

### 1. Auth Disabled Comment
**File:** `src/routes/_app.tsx` (lines 16-18)
```tsx
// Auth temporarily disabled for local exploration. Re-enable by restoring the
// beforeLoad guard below once Google sign-in is configured.
```
This should be addressed before production deployment.

### 2. Missing Error Boundaries
The app lacks React error boundaries. If any component throws an error during rendering, the entire app crashes without user-friendly feedback.

### 3. Dashboard Token Stored in Plain Text
The `dashboardToken` is stored in localStorage without encryption. While this is acceptable for a development/demo environment, production should use more secure storage (e.g., httpOnly cookies).

---

## Testing Recommendations

1. **FE-1 Test:** Enter various invalid URLs and verify proper error messages
2. **FE-2 Test:** Rapidly type in the name field and check network tab for duplicate save calls
3. **FE-3 Test:** Register a device with a different `deviceClass` and verify display

---

## Commit History

```
f39750b revert: use strict password policy for user registration
1a8c309 fix: use user-friendly password policy for registration
6988c5d fix: implement bug fixes #13-15
3faa536 fix: implement bug fixes #7-12
642ba75 fix: add command_secrets_hash column with bcrypt hashing
edf4a8a fix: implement bug fixes #2-5
518628b fix: implement WebSocket origin validation
```

---

*Document generated during frontend bug hunting session*