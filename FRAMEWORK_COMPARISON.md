# Frontend Framework Comparison: Library vs vyzorix-update-server

## Summary

| Aspect | Library (Design) | vyzorix-update-server (TanStack) | Recommendation |
|--------|------------------|----------------------------------|-----------------|
| **Routing** | React state machine | TanStack Start (file-based) | **TanStack** ✅ |
| **Type Safety** | TypeScript (basic) | TypeScript + Zod schemas | **TanStack** ✅ |
| **Forms** | Manual state | react-hook-form + Zod | **TanStack** ✅ |
| **UI Components** | Custom (inline) | shadcn/ui (Radix-based) | **TanStack** ✅ |
| **Dark Theme** | Custom CSS | oklch + Tailwind v4 | **TanStack** ✅ |
| **SSR/Hydration** | Node server.ts | Nitro (TanStack Start) | **TanStack** ✅ |
| **Dev Experience** | Vite + React | Vite + TanStack | **Equal** |
| **Bundle Size** | Smaller | Larger (more dependencies) | **Library** ✅ |
| **Auth UI Design** | Polished, custom | Basic card | **Library** ✅ |

## Verdict

**Use TanStack Start as the base, migrate Library's auth UI components into it.**

### Rationale

1. **Library's strength is the visual design**, not the architecture
2. **TanStack Start has better DX**: file-based routing, type-safe routes, SSR
3. **Migration is straightforward**: Copy components, adapt API calls
4. **Keep the shadcn/ui foundation**: Better accessibility, consistent styling

## Migration Strategy

### Keep from TanStack Start
- `apps/web/src/routes/` structure (file-based routing)
- `apps/web/src/components/ui/` (shadcn components)
- `apps/web/src/lib/api/` (type-safe API layer)
- `apps/web/src/styles.css` (oklch design tokens)
- `apps/web/vite.config.ts` (build configuration)

### Copy from Library
- `src/components/SignUpForm.tsx` → `apps/web/src/components/auth/SignUpForm.tsx`
- `src/components/LoginForm.tsx` → `apps/web/src/components/auth/LoginForm.tsx`
- `src/components/ForgotPasswordForm.tsx` → `apps/web/src/components/auth/ForgotPasswordForm.tsx`
- `src/components/WaitingVerification.tsx` → `apps/web/src/components/auth/WaitingVerification.tsx`
- `src/components/SuccessView.tsx` → `apps/web/src/components/auth/SuccessView.tsx`
- Dark background image + gradient styles

### Adapt for TanStack Start
- Replace Library's `ARCHITECTURE_CONFIG` with TanStack's `useAuth` hook
- Replace `triggerToast()` with `sonner` (already in project)
- Replace inline Tailwind with existing component patterns
- Update API calls to use `credentials: "include"` for cookies

## Implementation Notes

### API Client Adaptation

```typescript
// Library pattern (needs adaptation)
import { ARCHITECTURE_CONFIG } from '../config';
const IS_SIMULATED = ARCHITECTURE_CONFIG.IS_SIMULATED;

// TanStack pattern (keep this)
import { useAuth } from '@/hooks/use-auth';
const { operator, isAuthenticated } = useAuth();
```

### Cookie-Aware Fetch

```typescript
// New wrapper needed
export const cookieFetch = async (url: string, options: RequestInit = {}) => {
  return fetch(url, {
    ...options,
    credentials: 'include', // Critical for cookie auth
  });
};
```

### Component Adaptation Example

```tsx
// Library: src/components/LoginForm.tsx
const handleSubmit = (e: React.FormEvent) => {
  e.preventDefault();
  onLogin(loginIdent, loginPass); // Calls parent's handler
};

// TanStack: apps/web/src/components/auth/LoginForm.tsx
const handleSubmit = async (e: React.FormEvent) => {
  e.preventDefault();
  setLoading(true);
  try {
    await login(serverUrl, loginIdent, loginPass);
    navigate({ to: '/dashboard' });
  } catch (err) {
    toast.error(err.message);
  } finally {
    setLoading(false);
  }
};
```