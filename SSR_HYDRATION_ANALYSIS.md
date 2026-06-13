# SSR Hydration Analysis: Library vs vyzorix-update-server

## Executive Summary

| Aspect | Library (Designed) | vyzorix-update-server (Current) | Status |
|--------|-------------------|----------------------------------|--------|
| Server-side cookie reading | ✅ Implemented | ❌ Not implemented | 🔴 GAP |
| Prefetch operator data on server | ✅ Implemented | ❌ Not implemented | 🔴 GAP |
| `window.__VYZORIX_PREFETCHED_STATE__` | ✅ Implemented | ❌ Not implemented | 🔴 GAP |
| `hydrateRoot` (not `createRoot`) | ✅ Implemented | ✅ Implemented | ✅ OK |
| No whitespace in `<div id="root">` | ✅ Documented | ❓ Unknown | ⚠️ CHECK |
| Server-side auth check | ✅ Yes | ❌ No (client-side only) | 🔴 GAP |

---

## The Problem: Client-Side Only Auth Check

### Current vyzorix-update-server Flow (Client-Side)

```typescript
// login.tsx - Line 33-37
useEffect(() => {
  const token = localStorage.getItem("vyz.auth.token");
  if (token) navigate({ to: "/dashboard", replace: true });
}, [navigate]);
```

**Issues:**
1. ❌ Server doesn't know if user is authenticated
2. ❌ User sees login page briefly before redirect
3. ❌ Flicker/flash of unauthenticated content
4. ❌ SEO: Search engines see unauthenticated content

### Library's SSR Flow (Server-Side)

```typescript
// server.ts - Line 45-70
app.get('*', async (req, res) => {
  // 1. Server reads session cookie
  const sessionCookie = req.headers.cookie?.find(c => c.startsWith('vyz_session='));
  
  // 2. Server prefetches operator data
  const prefetchedState = sessionCookie 
    ? await fetchOperatorFromDB(sessionCookie)
    : { view: 'signup', profileData: null };

  // 3. Server injects state BEFORE React loads
  const stateScript = `<script>window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};</script>`;

  // 4. React hydrates with correct state immediately
  // No client-side auth check needed!
});
```

**Benefits:**
1. ✅ Server knows auth state immediately
2. ✅ User sees correct page on first paint
3. ✅ No flicker/flash
4. ✅ SEO: Search engines see correct content

---

## Detailed Comparison

### 1. Entry Points

#### Library (Proper SSR)
```typescript
// src/main.tsx
import { hydrateRoot } from 'react-dom/client';
hydrateRoot(container, <App />);  // ✅ Correct
```

#### vyzorix-update-server (Correct Entry)
```typescript
// src/entry-client.tsx
import { hydrateRoot } from 'react-dom/client';
hydrateRoot(document.body, <StartClient />);  // ✅ Correct
```

**Verdict:** Both use `hydrateRoot` correctly ✅

---

### 2. State Hydration Strategy

#### Library (Server-injected State)
```html
<!-- index.html -->
<div id="root"><!--app-html--></div>
<!--app-state-->
<script>window.__VYZORIX_PREFETCHED_STATE__ = {...};</script>
```

#### vyzorix-update-server (No Server Injection)

**Missing:**
- No `<!--app-state-->` placeholder
- No `window.__VYZORIX_PREFETCHED_STATE__` injection
- TanStack Start doesn't inject auth state

**Current approach:**
```typescript
// useAuth.ts - Client-side only
const [state, setState] = useState<AuthState>(() => {
  const token = getToken();  // localStorage access
  const operator = getStoredOperator();
  return { isAuthenticated: Boolean(token), ... };
});
```

**Verdict:** No SSR hydration for auth state 🔴

---

### 3. Login Page Auth Check

#### Library (Server-side redirect)
```typescript
// server.ts handles redirect BEFORE page renders
if (isAuthenticated) {
  res.redirect('/dashboard');
}
```

#### vyzorix-update-server (Client-side check)
```typescript
// login.tsx - Runs AFTER page renders
useEffect(() => {
  const token = localStorage.getItem("vyz.auth.token");
  if (token) navigate({ to: "/dashboard", replace: true });
}, [navigate]);
```

**Issues:**
1. User sees login page flash
2. Network request to `/dashboard` happens client-side
3. `useEffect` runs after hydration, causing delay

**Verdict:** Client-side only, not SSR 🔴

---

### 4. Cookie Reading

#### Library (Server reads cookies)
```typescript
// server.ts
const sessionCookie = req.headers.cookie
  ?.split('; ')
  .find(row => row.startsWith('vyz_session='));
```

#### vyzorix-update-server (No server cookie reading)

TanStack Start handles SSR but doesn't have built-in cookie reading for auth state.

**Verdict:** No server-side cookie reading 🔴

---

## What Needs to Be Implemented

### For True SSR Hydration

#### 1. Add Cookie Reader Middleware

```typescript
// apps/web/src/lib/server/cookie-reader.ts
export async function getSessionFromCookies(request: Request): Promise<Operator | null> {
  const cookieHeader = request.headers.get('cookie');
  if (!cookieHeader) return null;
  
  const cookies = Object.fromEntries(
    cookieHeader.split('; ').map(c => c.split('='))
  );
  
  const sessionCookie = cookies['vyz_session'];
  if (!sessionCookie) return null;
  
  // Decrypt operator ID from cookie
  const operatorId = decryptOperatorId(sessionCookie);
  if (!operatorId) return null;
  
  // Fetch operator from database/API
  try {
    const res = await fetch(`${process.env.API_URL}/v1/auth/me`, {
      headers: { Cookie: `vyz_session=${sessionCookie}` }
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}
```

#### 2. Update Route Loaders (TanStack Start)

```typescript
// apps/web/src/routes/_app.tsx
import { createFileRoute, redirect } from "@tanstack/react-router";

const AppLayout = () => { ... };

// Add server-side loader
export const Route = createFileRoute("/_app")({
  beforeLoad: async ({ context }) => {
    const operator = await getSessionFromCookies(context.request);
    if (!operator) {
      throw redirect({ to: "/login" });
    }
    return { operator };
  },
  component: AppLayout,
});
```

#### 3. Add State Injection to index.html

```html
<!-- apps/web/index.html -->
<!DOCTYPE html>
<html lang="en">
  <head>...</head>
  <body>
    <!-- NO WHITESPACE between these tags! -->
    <div id="root"><!--app-html--></div>
    
    <!--app-state-->
    
    <script type="module" src="/src/entry-client.tsx"></script>
  </body>
</html>
```

#### 4. Update Server Entry

```typescript
// apps/web/src/server.ts
// Already uses TanStack Start, but needs cookie context
```

---

## Migration Checklist

| Task | Priority | Status |
|------|----------|--------|
| Add cookie reading utility | P0 | ⬜ TODO |
| Create server-side loader for `/dashboard` routes | P0 | ⬜ TODO |
| Add `<!--app-state-->` to index.html | P1 | ⬜ TODO |
| Implement `window.__VYZORIX_PREFETCHED_STATE__` injection | P1 | ⬜ TODO |
| Update `_app.tsx` with `beforeLoad` auth check | P0 | ⬜ TODO |
| Remove client-side `useEffect` auth checks | P2 | ⬜ TODO |
| Test hydration without flicker | P1 | ⬜ TODO |

---

## Conclusion

**The Library's SSR hydration strategy is NOT implemented in vyzorix-update-server.**

The current implementation relies entirely on client-side JavaScript to:
1. Check localStorage for token
2. Redirect authenticated users
3. Load operator data

This causes:
- Flash of unauthenticated content
- Poor SEO (search engines see login page)
- Slower perceived performance
- Potential hydration mismatches

**Recommendation:** Implement proper SSR hydration before the cookie migration, or migrate both simultaneously to avoid double work.

---

## Next Steps

1. **Decide:** Migrate SSR first, then cookies? Or both together?
2. **Plan:** Add TanStack Start loader functions for auth
3. **Implement:** Server-side cookie reading + state injection
4. **Test:** Verify no hydration mismatches

---

*Document Version: 1.0*  
*Date: 2026-06-12*