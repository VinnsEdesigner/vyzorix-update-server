# SSR Implementation Status: ACTUAL vs GUIDE

## Current SSR Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CURRENT SSR FLOW                                 │
│                                                                          │
│  ┌──────────┐     ┌──────────────────┐     ┌────────────────────────┐   │
│  │ Browser  │────▶│ Go Server :3000  │────▶│ SSR Server :3001       │   │
│  └──────────┘     │ (API + Proxy)    │     │ (TanStack Start)       │   │
│                   │                  │     │                        │   │
│                   │ - Static assets  │     │ - React SSR rendering  │   │
│                   │ - API routes     │     │ - Route matching       │   │
│                   │ - JWT validation│     │ - beforeLoad hooks     │   │
│                   └──────────────────┘     └────────────────────────┘   │
│                           │                            │                 │
│                           │         NO AUTH STATE     │                 │
│                           │         PREFETCHING        │                 │
│                           ▼                            ▼                 │
│                   ┌─────────────────────────────────────────────┐       │
│                   │         What Actually Happens:              │       │
│                   │                                             │       │
│                   │  1. Browser requests page                  │       │
│                   │  2. Go proxies to SSR server                │       │
│                   │  3. SSR server renders React (NO DATA)      │       │
│                   │  4. HTML sent to browser with empty shell   │       │
│                   │  5. Client JS loads                         │       │
│                   │  6. useEffect runs → reads localStorage     │       │
│                   │  7. If logged in → redirect to dashboard   │       │
│                   │                                             │       │
│                   │  ⚠️ User sees login page briefly!          │       │
│                   └─────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## What the Guide Says vs What Actually Happens

### Guide's SSR Flow (Library)
```typescript
// 1. Server reads cookies BEFORE rendering
const sessionCookie = req.headers.cookie?.find(c => c.startsWith('vyz_session='));

// 2. Server fetches operator data from DB
const prefetchedState = await fetchOperatorFromDB(sessionCookie);

// 3. Server injects state INTO HTML
const stateScript = `<script>
  window.__VYZORIX_PREFETCHED_STATE__ = ${JSON.stringify(prefetchedState)};
</script>`;

// 4. React hydrates with correct state - NO FLASH
```

### Actual vyzorix-update-server Flow

```typescript
// _app.tsx - beforeLoad hook (CLIENT-SIDE ONLY!)
beforeLoad: () => {
  // ⚠️ This runs on CLIENT, not server!
  if (typeof window !== "undefined" && typeof localStorage !== "undefined") {
    const token = localStorage.getItem("vyz.auth.token");
    if (!token) {
      throw redirect({ to: "/login" });
    }
  }
}
```

**Result:** The beforeLoad hook checks localStorage AFTER hydration, not on the server.

---

## What's Actually Implemented

| Feature | Library (Guide) | vyzorix-update-server | Status |
|---------|----------------|----------------------|--------|
| TanStack Start SSR | ✅ Yes | ✅ Yes | ✅ DONE |
| Stream rendering | ✅ Yes | ✅ Yes (renderRouterToStream) | ✅ DONE |
| Server entry point | ✅ Yes | ✅ Yes (server.ts) | ✅ DONE |
| Route matching | ✅ Yes | ✅ Yes | ✅ DONE |
| **Cookie reading (server)** | ✅ Yes | ❌ No | 🔴 MISSING |
| **Data prefetching** | ✅ Yes | ❌ No | 🔴 MISSING |
| **State injection** | ✅ Yes (`__VYZORIX_PREFETCHED_STATE__`) | ❌ No | 🔴 MISSING |
| **Server-side auth check** | ✅ Yes | ❌ No (client-side only) | 🔴 MISSING |

---

## The Problem: No Server-Side Auth

### Current beforeLoad (Client-Side)
```typescript
// _app.tsx - Line 47-53
beforeLoad: () => {
  if (typeof window !== "undefined" && typeof localStorage !== "undefined") {
    const token = localStorage.getItem("vyz.auth.token"); // ❌ Client-only!
    if (!token) {
      throw redirect({ to: "/login" });
    }
  }
}
```

### What Should Happen (Server-Side)
```typescript
// What SHOULD be in _app.tsx
beforeLoad: async ({ request }) => {
  // ✅ Server-side cookie reading
  const cookieHeader = request.headers.get('cookie');
  const cookies = Object.fromEntries(
    cookieHeader.split('; ').map(c => c.split('='))
  );
  
  // ❌ Currently this doesn't exist
  const sessionCookie = cookies['vyz_session']; 
  
  if (!sessionCookie) {
    throw redirect({ to: "/login" });
  }
  
  // ✅ Server fetches operator data
  const operator = await fetchOperator(sessionCookie);
  
  return { operator };
}
```

---

## Why This Matters for Cookie Migration

When you migrate to HttpOnly cookies:

### Current (JWT in localStorage)
```typescript
// Works with client-side check (even if slow)
const token = localStorage.getItem("vyz.auth.token");
```

### Target (HttpOnly cookies)
```typescript
// ❌ BROKEN - JavaScript can't read HttpOnly cookies!
const token = localStorage.getItem("vyz.auth.token"); // Returns null!
```

**You MUST implement server-side cookie reading for HttpOnly cookies to work.**

---

## Infrastructure That Exists

### SSR Server Setup
```
apps/api/ssr-server.js     ✅ Exists - handles Vite SSR + production
apps/api/ssr-package.json ✅ Exists
apps/web/src/server.ts    ✅ Exists - TanStack Start entry
apps/web/src/entry-client.tsx ✅ Exists - hydrateRoot
```

### Missing Pieces
```
❌ No server-side cookie reader
❌ No operator data prefetch
❌ No state injection to window.__VYZORIX_PREFETCHED_STATE__
❌ No beforeLoad with server-side request context
```

---

## What Needs to Be Added

### 1. Create Server Cookie Reader
```typescript
// apps/web/src/lib/server/cookie-reader.ts
export async function getSessionOperator(request: Request): Promise<Operator | null> {
  const cookieHeader = request.headers.get('cookie');
  if (!cookieHeader) return null;
  
  const cookies = Object.fromEntries(
    cookieHeader.split('; ').map(c => c.split('='))
  );
  
  const sessionCookie = cookies['vyz_session'];
  if (!sessionCookie) return null;
  
  // Call Go API to validate and get operator
  const res = await fetch('http://localhost:3000/v1/auth/me', {
    headers: { Cookie: `vyz_session=${sessionCookie}` }
  });
  
  if (!res.ok) return null;
  return res.json();
}
```

### 2. Update beforeLoad to Use Server Context
```typescript
// apps/web/src/routes/_app.tsx
beforeLoad: async ({ request }) => {
  const operator = await getSessionOperator(request);
  if (!operator) {
    throw redirect({ to: "/login" });
  }
  return { operator };
}
```

### 3. Inject State for Hydration
```typescript
// TanStack Start handles this via context
// Pass operator through context so React can use it
```

---

## Summary

| Question | Answer |
|----------|--------|
| Is SSR implemented? | **Yes, but incomplete** |
| Does it match the guide? | **No** |
| Does it render HTML on server? | **Yes** |
| Does it read cookies on server? | **No** |
| Does it prefetch auth data? | **No** |
| Does it inject state? | **No** |
| Will HttpOnly cookies work? | **No** (must add server-side reading) |

---

## Recommendation

**Implement SSR cookie reading AS PART of the cookie migration** since:
1. HttpOnly cookies require server-side reading (can't use localStorage)
2. You need server-side auth anyway for proper SSR
3. The infrastructure is already in place (TanStack Start, SSR server)

---

*Document Version: 1.0*  
*Date: 2026-06-12*