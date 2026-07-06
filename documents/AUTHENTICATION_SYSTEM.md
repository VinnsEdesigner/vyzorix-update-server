# Authentication System - Enterprise Requirements Specification

> **Version:** 1.0  
> **Status:** Draft  
> **Created:** 2026-06-25  
> **Target:** Production MVP  
> **Architecture:** Layered (Following `FRONTEND_ARCHITECTURE.md`)

---

## Table of Contents

1. [Overview](#1-overview)
2. [Architecture](#2-architecture)
3. [Target File Structure](#3-target-file-structure)
4. [Page Structure](#4-page-structure)
5. [Domain Layer](#5-domain-layer)
6. [Data Layer](#6-data-layer)
7. [Presentation Layer - Hooks](#7-presentation-layer---hooks)
8. [UI Layer - Components](#8-ui-layer---components)
9. [File Changes Summary](#9-file-changes-summary)
10. [Implementation Order](#10-implementation-order)

---

>  **Architecture Alignment Note (v1.0)**
> 
> This document follows the **4-layer architecture** defined in `FRONTEND_ARCHITECTURE.md`:
> - **UI Layer** (`src/components/`) - Pure UI rendering, imports only from hooks
> - **Presentation Layer** (`src/hooks/`) - UI logic, state management, imports from domain & data
> - **Domain Layer** (`src/domain/` - NEW) - Types, transforms, validation (NO external imports)
> - **Data Layer** (`src/lib/api/` - GraphQL/REST) - API clients, imports only domain types
>
> **Dependency Rule:** UI → Hooks → Domain → API (flow inward only)

---

## 1. Overview

### 1.1 Purpose

The Authentication System handles:
- **Login** - Credential-based and OAuth authentication
- **Multi-Factor Authentication (MFA)** - TOTP-based 2FA
- **Password Reset** - Forgot password and reset flows
- **Session Management** - JWT token handling, refresh, logout
- **Auth State** - Global auth context, protected routes

### 1.2 Auth Flows

| Flow | Trigger | Steps |
|------|---------|-------|
| **Credential Login** | User submits email + password | Validate → JWT → Redirect |
| **OAuth Login** | User clicks Google button | Redirect → Callback → JWT |
| **MFA Login** | After credential verification | TOTP code → Verify → JWT |
| **Forgot Password** | User clicks "Forgot password" | Email input → Send reset link |
| **Reset Password** | User clicks email reset link | New password → Confirm → Login |
| **Logout** | User clicks logout | Clear tokens → Redirect to login |

### 1.3 Design Principles

- **Zero plaintext passwords** - Never log or store plaintext
- **Secure token storage** - HttpOnly cookies or secure storage
- **MFA required** - All accounts must enroll in 2FA
- **Session timeout** - Auto-logout after inactivity
- **Graceful degradation** - Work offline with cached auth state

---

## 2. Architecture

### 2.1 Layered Architecture Overview

```

                        FRONTEND ARCHITECTURE                        

                                                                     
     
                        UI LAYER                                  
                     (src/components/)                           
                                                                  
      Pages, Components, Shared UI                               
      ONLY renders UI. Uses hooks for everything.                 
      NEVER imports from Data or Domain.                          
     
                                                                     
                               uses                                  
                                                                     
     
                     PRESENTATION LAYER                           
                        (src/hooks/)                             
                                                                  
      Custom hooks that:                                         
      - Handle UI logic                                           
      - Transform data for UI                                     
      - Manage state                                              
      NEVER renders UI. NEVER imports from UI layer.             
     
                                                                     
                               uses                                  
                                                                     
     
                        DOMAIN LAYER                             
                       (src/domain/)                             
                                                                  
      Pure functions that:                                       
      - Define types and interfaces                             
      - Transform data (no side effects)                        
      - Validate input                                           
      NEVER imports from UI, Presentation, or Data.            
     
                                                                     
                               uses                                  
                                                                     
     
                         DATA LAYER                              
                     (src/lib/api/)                              
                                                                  
      API clients that:                                          
      - Make HTTP requests                                       
      - Handle authentication                                     
      - Parse responses                                          
      NEVER imports from UI or Presentation.                     
     
                                                                     

```

### 2.2 Auth State Flow

```

                         AUTH STATE FLOW                             

                                                                     
  USER ACTION                                                       
                                                                   
                                                                   
     
    UI LAYER                                                        
    - LoginForm component calls useLogin() hook                     
     
                                                                   
                                                                   
     
    PRESENTATION LAYER (Hooks)                                     
    - useLogin() validates input, calls authApi.login()           
    - Returns { user, isLoading, error }                          
     
                                                                   
                                                                   
     
    DOMAIN LAYER                                                  
    - loginFromRaw() transforms API response                     
    - validateCredentials() checks input                          
     
                                                                   
                                                                   
     
    DATA LAYER (API)                                              
    - POST /v1/auth/login                                         
    - Handles JWT storage                                          
     
                                                                   
                                                                   
     
    AUTH CONTEXT                                                  
    - Stores user, tokens, auth state                             
    - Provides useAuth() hook globally                             
     
                                                                     

```

---

## 3. Target File Structure

### 3.1 Complete Directory Tree

```
apps/web/src/

 domain/                          # DOMAIN LAYER (NEW)
    common/
       pagination.ts            # Pagination types (reused)
       error.ts                 # Domain error types
       types.ts                 # Shared domain types
   
    auth/
        auth-types.ts            # AuthUser, LoginCredentials, MFAState
        auth-transforms.ts       # userFromRaw(), tokenFromRaw()
        auth-validation.ts       # validateEmail(), validatePassword()

 lib/
    api/
        graphql/
           queries/
              auth-queries.ts  # GET_ME, GET_MFA_STATUS
           mutations/
              auth-mutations.ts # LOGIN, REGISTER, LOGOUT, etc.
           fragments/
               user.fragment.ts  # User fragment
       
        rest/
            auth-rest.ts         # REST fallback endpoints

 hooks/                           # PRESENTATION LAYER
    auth/
       use-login.ts           # Login flow
       use-register.ts        # Registration flow
       use-logout.ts          # Logout flow
       use-mfa.ts             # MFA verification
       use-password-reset.ts # Forgot/reset password
       use-session.ts         # Session management
       use-auth-callback.ts   # OAuth callback
       index.ts               # Barrel export
   
    shared/
        use-debounce.ts        # Debounce utility

 components/                      # UI LAYER
    shared/                    # Shared auth components
       auth-card.tsx          # Centered card wrapper
       auth-input.tsx         # Email/password input
       auth-button.tsx        # Primary submit button
       auth-error.tsx         # Error message display
       auth-links.tsx         # Forgot password, register links
       mfa-input.tsx          # 6-digit TOTP input
       password-strength.tsx  # Password strength indicator
       index.ts               # Barrel export
   
    auth/
        login-form.tsx         # Login form
        register-form.tsx       # Registration form
        mfa-form.tsx           # MFA verification form
        forgot-password-form.tsx # Forgot password form
        reset-password-form.tsx # Reset password form
        auth-layout.tsx        # Auth page layout
        protected-route.tsx   # Route wrapper
        index.ts               # Barrel export

 context/
    auth-context.tsx           # Auth provider (NEW - global state)

 routes/
     auth.tsx                   # Auth layout (EXISTS - modify)
     auth.login.tsx            # /auth/login (EXISTS - modify)
     auth.register.tsx         # /auth/register (NEW)
     auth.forgot.tsx           # /auth/forgot (NEW)
     auth.reset.tsx            # /auth/reset (NEW)
     auth.callback.tsx         # /auth/callback (NEW - OAuth)
```

---

## 4. Page Structure

### 4.1 Routes

| Route | File | Purpose | Status |
|-------|------|---------|--------|
| `/auth/login` | `auth.login.tsx` | Login page | **MODIFY** |
| `/auth/register` | `auth.register.tsx` | Registration page | **NEW** |
| `/auth/forgot` | `auth.forgot.tsx` | Forgot password | **NEW** |
| `/auth/reset` | `auth.reset.tsx` | Reset password (from email link) | **NEW** |
| `/auth/callback` | `auth.callback.tsx` | OAuth callback handler | **NEW** |

### 4.2 Page Layouts

#### Login Page (`/auth/login`)
```

                                                                     
                                          
                           [Vyzorix Logo]                         
                                                                   
                        Welcome back                              
                                                                   
                        [Email Input]                             
                        [Password Input]                          
                                                                   
                        [    Login    ]                           
                                                                   
                        Forgot password?                           
                                                 
                        Don't have account?                       
                                                                   
                        [  Sign up with Google  ]                 
                                                                   
                                          
                                                                     

```

#### Registration Page (`/auth/register`)
```

                                                                     
                                          
                           [Vyzorix Logo]                         
                                                                   
                        Create account                            
                                                                   
                        [Email Input]                             
                        [Password Input]                          
                        [Confirm Password]                         
                                                                   
                        [  Create Account  ]                      
                                                                   
                        Already have account?                      
                                                                   
                                          
                                                                     

```

#### MFA Verification Page (`/auth/mfa`)
```

                                                                     
                                          
                           [Vyzorix Logo]                         
                                                                   
                        Two-factor auth                           
                                                                   
                        Enter 6-digit code                        
                        from your app                              
                                                                   
                        [ _ ][ _ ][ _ ]                           
                        [ _ ][ _ ][ _ ]                           
                                                                   
                        [    Verify    ]                           
                                                                   
                        [  Resend code  ]                         
                                                                   
                                          
                                                                     

```

### 4.3 Auth Context Provider

```

                         AUTH CONTEXT                                

                                                                     
     
    AuthContext                                                    
                                                                  
    State:                                                        
    - user: AuthUser | null                                       
    - isAuthenticated: boolean                                     
    - isLoading: boolean                                          
    - mfaRequired: boolean                                        
    - mfaToken: string | null                                     
                                                                  
    Methods:                                                      
    - login(credentials): Promise<void>                           
    - register(data): Promise<void>                               
    - logout(): Promise<void>                                      
    - verifyMfa(code): Promise<void>                              
    - forgotPassword(email): Promise<void>                         
    - resetPassword(token, password): Promise<void>               
                                                                  
    Provides useAuth() hook globally                               
     
                                                                     

```

---

## 5. Domain Layer

### 5.1 Types (`domain/auth/auth-types.ts`)

```typescript
// User types
export interface AuthUser {
  id: string;
  email: string;
  name: string;
  mfaEnabled: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: Date;
}

// Login types
export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: AuthUser;
  tokens: AuthTokens;
  mfaRequired: boolean;
  mfaToken?: string;
}

// Registration types
export interface RegisterData {
  email: string;
  password: string;
  name: string;
}

// MFA types
export interface MfaVerification {
  code: string;
  mfaToken: string;
}

export interface MfaSetup {
  secret: string;
  qrCodeUrl: string;
  backupCodes: string[];
}

// Password reset types
export interface ForgotPasswordData {
  email: string;
}

export interface ResetPasswordData {
  token: string;
  password: string;
  passwordConfirm: string;
}

// Auth state
export interface AuthState {
  user: AuthUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}
```

### 5.2 Transforms (`domain/auth/auth-transforms.ts`)

```typescript
import type { AuthUser, AuthTokens, LoginResponse } from './types';

interface RawUser {
  id: string;
  email: string;
  name: string;
  mfa_enabled: boolean;
  created_at: string;
  updated_at: string;
}

interface RawTokens {
  access_token: string;
  refresh_token: string;
  expires_at: string;
}

export const userFromRaw = (raw: RawUser): AuthUser => ({
  id: raw.id,
  email: raw.email,
  name: raw.name,
  mfaEnabled: raw.mfa_enabled,
  createdAt: new Date(raw.created_at),
  updatedAt: new Date(raw.updated_at),
});

export const tokensFromRaw = (raw: RawTokens): AuthTokens => ({
  accessToken: raw.access_token,
  refreshToken: raw.refresh_token,
  expiresAt: new Date(raw.expires_at),
});

export const loginResponseFromRaw = (raw: any): LoginResponse => ({
  user: userFromRaw(raw.user),
  tokens: tokensFromRaw(raw.tokens),
  mfaRequired: raw.mfa_required ?? false,
  mfaToken: raw.mfa_token,
});
```

### 5.3 Validation (`domain/auth/auth-validation.ts`)

```typescript
import type { LoginCredentials, RegisterData, ResetPasswordData } from './types';

export const validateEmail = (email: string): string | null => {
  if (!email) return 'Email is required';
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    return 'Invalid email format';
  }
  return null;
};

export const validatePassword = (password: string): string | null => {
  if (!password) return 'Password is required';
  if (password.length < 8) return 'Password must be at least 8 characters';
  if (!/[A-Z]/.test(password)) return 'Password must contain uppercase';
  if (!/[a-z]/.test(password)) return 'Password must contain lowercase';
  if (!/[0-9]/.test(password)) return 'Password must contain number';
  return null;
};

export const validateLoginCredentials = (
  credentials: LoginCredentials
): Record<string, string> => {
  const errors: Record<string, string> = {};
  
  const emailError = validateEmail(credentials.email);
  if (emailError) errors.email = emailError;
  
  if (!credentials.password) errors.password = 'Password is required';
  
  return errors;
};

export const validateRegisterData = (
  data: RegisterData
): Record<string, string> => {
  const errors: Record<string, string> = {};
  
  const emailError = validateEmail(data.email);
  if (emailError) errors.email = emailError;
  
  const passwordError = validatePassword(data.password);
  if (passwordError) errors.password = passwordError;
  
  if (data.password !== data.passwordConfirm) {
    errors.passwordConfirm = 'Passwords do not match';
  }
  
  if (!data.name || data.name.length < 2) {
    errors.name = 'Name must be at least 2 characters';
  }
  
  return errors;
};

export const validateMfaCode = (code: string): string | null => {
  if (!code) return 'Code is required';
  if (!/^\d{6}$/.test(code)) return 'Code must be 6 digits';
  return null;
};

export const validateResetPassword = (
  data: ResetPasswordData
): Record<string, string> => {
  const errors: Record<string, string> = {};
  
  const passwordError = validatePassword(data.password);
  if (passwordError) errors.password = passwordError;
  
  if (data.password !== data.passwordConfirm) {
    errors.passwordConfirm = 'Passwords do not match';
  }
  
  return errors;
};
```

---

## 6. Data Layer

### 6.1 REST Endpoints (`lib/api/rest/auth-rest.ts`)

```typescript
const BASE = '/v1/auth';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponseDto {
  user: any;
  tokens: any;
  mfa_required: boolean;
  mfa_token?: string;
}

export async function login(credentials: LoginRequest): Promise<LoginResponseDto> {
  const res = await fetch(`${BASE}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(credentials),
  });
  
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message || 'Login failed');
  }
  
  return res.json();
}

export async function register(data: {
  email: string;
  password: string;
  name: string;
}): Promise<LoginResponseDto> {
  const res = await fetch(`${BASE}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(data),
  });
  
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message || 'Registration failed');
  }
  
  return res.json();
}

export async function logout(): Promise<void> {
  const res = await fetch(`${BASE}/logout`, {
    method: 'POST',
    credentials: 'include',
  });
  
  if (!res.ok) {
    throw new Error('Logout failed');
  }
}

export async function verifyMfa(mfaToken: string, code: string): Promise<LoginResponseDto> {
  const res = await fetch(`${BASE}/mfa/verify`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ mfa_token: mfaToken, code }),
  });
  
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message || 'MFA verification failed');
  }
  
  return res.json();
}

export async function forgotPassword(email: string): Promise<void> {
  const res = await fetch(`${BASE}/forgot-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  });
  
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message || 'Request failed');
  }
}

export async function resetPassword(token: string, password: string): Promise<void> {
  const res = await fetch(`${BASE}/reset-password`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, password }),
  });
  
  if (!res.ok) {
    const error = await res.json();
    throw new Error(error.message || 'Reset failed');
  }
}

export async function getCurrentUser(): Promise<any> {
  const res = await fetch(`${BASE}/me`, {
    credentials: 'include',
  });
  
  if (res.status === 401) return null;
  
  if (!res.ok) {
    throw new Error('Failed to get user');
  }
  
  return res.json();
}

export async function refreshToken(): Promise<any> {
  const res = await fetch(`${BASE}/refresh`, {
    method: 'POST',
    credentials: 'include',
  });
  
  if (!res.ok) {
    throw new Error('Token refresh failed');
  }
  
  return res.json();
}
```

### 6.2 Token Storage Strategy

```typescript
// Token storage using httpOnly cookies (recommended)
// or secure storage as fallback

interface TokenStorage {
  getAccessToken(): string | null;
  getRefreshToken(): string | null;
  setTokens(tokens: { accessToken: string; refreshToken: string }): void;
  clearTokens(): void;
}

// Strategy 1: HttpOnly Cookies (handled by server)
// Client uses credentials: 'include' on all requests

// Strategy 2: Memory-only (for SPA)
// Tokens stored in memory, lost on page refresh

// Strategy 3: Secure Storage (least recommended)
// Can be XSS attacked if not careful
```

---

## 7. Presentation Layer - Hooks

### 7.1 Auth Hooks (`hooks/auth/`)

| File | Status | Purpose |
|------|--------|---------|
| `use-login.ts` | **NEW** | Login form logic |
| `use-register.ts` | **NEW** | Registration form logic |
| `use-logout.ts` | **NEW** | Logout logic |
| `use-mfa.ts` | **NEW** | MFA verification logic |
| `use-password-reset.ts` | **NEW** | Forgot/reset password logic |
| `use-session.ts` | **NEW** | Session management |
| `use-auth-callback.ts` | **NEW** | OAuth callback handler |
| `index.ts` | **NEW** | Barrel export |

### 7.2 Hook Implementations

```typescript
// hooks/auth/use-login.ts
import { useState, useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { login as loginApi } from '@/lib/api/rest/auth';
import { loginFromRaw } from '@/domain/auth';
import { validateLoginCredentials } from '@/domain/auth/validation';
import { useAuth } from '@/context/auth-context';

export function useLogin() {
  const navigate = useNavigate();
  const { setAuth } = useAuth();
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [globalError, setGlobalError] = useState<string | null>(null);

  const validate = useCallback(() => {
    const errs = validateLoginCredentials({ email, password });
    setErrors(errs);
    return Object.keys(errs).length === 0;
  }, [email, password]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setGlobalError(null);
    
    if (!validate()) return;
    
    setIsLoading(true);
    try {
      const response = await loginApi({ email, password });
      const data = loginFromRaw(response);
      
      if (data.mfaRequired) {
        // Navigate to MFA verification
        navigate({ to: '/auth/mfa', search: { mfaToken: data.mfaToken } });
      } else {
        setAuth({ user: data.user, tokens: data.tokens });
        navigate({ to: '/dashboard' });
      }
    } catch (err) {
      setGlobalError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setIsLoading(false);
    }
  }, [email, password, validate, navigate, setAuth]);

  return {
    email,
    setEmail,
    password,
    setPassword,
    isLoading,
    errors,
    globalError,
    handleSubmit,
  };
}
```

```typescript
// hooks/auth/use-logout.ts
import { useCallback } from 'react';
import { useNavigate } from '@tanstack/react-router';
import { logout as logoutApi } from '@/lib/api/rest/auth';
import { useAuth } from '@/context/auth-context';

export function useLogout() {
  const navigate = useNavigate();
  const { clearAuth } = useAuth();

  const handleLogout = useCallback(async () => {
    try {
      await logoutApi();
    } catch (err) {
      // Ignore logout API errors, clear local state anyway
    } finally {
      clearAuth();
      navigate({ to: '/auth/login' });
    }
  }, [navigate, clearAuth]);

  return { handleLogout };
}
```

```typescript
// hooks/auth/use-mfa.ts
import { useState, useCallback } from 'react';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { verifyMfa as verifyMfaApi } from '@/lib/api/rest/auth';
import { loginFromRaw } from '@/domain/auth';
import { validateMfaCode } from '@/domain/auth/validation';
import { useAuth } from '@/context/auth-context';

export function useMfa() {
  const navigate = useNavigate();
  const search = useSearch({ from: '/auth/mfa' });
  const { setAuth } = useAuth();

  const [code, setCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleVerify = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const codeError = validateMfaCode(code);
    if (codeError) {
      setError(codeError);
      return;
    }

    setIsLoading(true);
    try {
      const response = await verifyMfaApi(search.mfaToken, code);
      const data = loginFromRaw(response);
      
      setAuth({ user: data.user, tokens: data.tokens });
      navigate({ to: '/dashboard' });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Verification failed');
    } finally {
      setIsLoading(false);
    }
  }, [code, search.mfaToken, navigate, setAuth]);

  return {
    code,
    setCode,
    isLoading,
    error,
    handleVerify,
  };
}
```

```typescript
// hooks/auth/use-session.ts
import { useEffect, useCallback, useRef } from 'react';
import { refreshToken as refreshTokenApi } from '@/lib/api/rest/auth';
import { useAuth } from '@/context/auth-context';

const SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes
const TOKEN_REFRESH_BUFFER = 5 * 60 * 1000; // Refresh 5 min before expiry

export function useSession() {
  const { user, tokens, setAuth, clearAuth } = useAuth();
  const refreshTimeoutRef = useRef<NodeJS.Timeout>();

  const refreshSession = useCallback(async () => {
    try {
      const response = await refreshTokenApi();
      const newTokens = {
        accessToken: response.access_token,
        refreshToken: response.refresh_token,
        expiresAt: new Date(response.expires_at),
      };
      setAuth({ user, tokens: newTokens });
    } catch (err) {
      clearAuth();
    }
  }, [user, setAuth, clearAuth]);

  const scheduleRefresh = useCallback(() => {
    if (refreshTimeoutRef.current) {
      clearTimeout(refreshTimeoutRef.current);
    }

    if (!tokens?.expiresAt) return;

    const now = Date.now();
    const expiry = tokens.expiresAt.getTime();
    const refreshIn = expiry - TOKEN_REFRESH_BUFFER - now;

    if (refreshIn > 0) {
      refreshTimeoutRef.current = setTimeout(refreshSession, refreshIn);
    }
  }, [tokens, refreshSession]);

  useEffect(() => {
    scheduleRefresh();
    return () => {
      if (refreshTimeoutRef.current) {
        clearTimeout(refreshTimeoutRef.current);
      }
    };
  }, [scheduleRefresh]);

  // Activity detection for session timeout
  useEffect(() => {
    let activityTimeout: NodeJS.Timeout;

    const resetTimer = () => {
      if (activityTimeout) clearTimeout(activityTimeout);
      activityTimeout = setTimeout(() => {
        clearAuth();
      }, SESSION_TIMEOUT);
    };

    const events = ['mousedown', 'keydown', 'scroll', 'touchstart'];
    events.forEach(event => {
      window.addEventListener(event, resetTimer);
    });

    resetTimer();

    return () => {
      events.forEach(event => {
        window.removeEventListener(event, resetTimer);
      });
      if (activityTimeout) clearTimeout(activityTimeout);
    };
  }, [clearAuth]);

  return { refreshSession };
}
```

---

## 8. UI Layer - Components

### 8.1 Shared Components (`components/shared/`)

| File | Status | Purpose |
|------|--------|---------|
| `auth-card.tsx` | **NEW** | Centered card wrapper |
| `auth-input.tsx` | **NEW** | Email/password input |
| `auth-button.tsx` | **NEW** | Submit button |
| `auth-error.tsx` | **NEW** | Error message |
| `auth-links.tsx` | **NEW** | Secondary links |
| `mfa-input.tsx` | **NEW** | 6-digit TOTP input |
| `password-strength.tsx` | **NEW** | Password strength |
| `index.ts` | **NEW** | Barrel export |

### 8.2 Auth Components (`components/auth/`)

| File | Status | Purpose |
|------|--------|---------|
| `login-form.tsx` | **NEW** | Login form |
| `register-form.tsx` | **NEW** | Registration form |
| `mfa-form.tsx` | **NEW** | MFA verification |
| `forgot-password-form.tsx` | **NEW** | Forgot password |
| `reset-password-form.tsx` | **NEW** | Reset password |
| `auth-layout.tsx` | **NEW** | Auth page layout |
| `protected-route.tsx` | **NEW** | Route guard |
| `index.ts` | **NEW** | Barrel export |

### 8.3 Component Implementations

```typescript
// components/shared/auth-card.tsx
interface AuthCardProps {
  children: React.ReactNode;
  title: string;
  subtitle?: string;
}

export function AuthCard({ children, title, subtitle }: AuthCardProps) {
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="bg-card rounded-lg border shadow-sm p-8">
          <div className="text-center mb-8">
            <VyzorixLogo className="h-12 w-12 mx-auto mb-4" />
            <h1 className="text-2xl font-semibold">{title}</h1>
            {subtitle && (
              <p className="text-muted-foreground mt-2">{subtitle}</p>
            )}
          </div>
          {children}
        </div>
      </div>
    </div>
  );
}
```

```typescript
// components/shared/mfa-input.tsx
import { useRef, useState, useEffect } from 'react';
import { Input } from '@/components/ui/input';

interface MfaInputProps {
  value: string;
  onChange: (value: string) => void;
  length?: number;
  disabled?: boolean;
}

export function MfaInput({ value, onChange, length = 6, disabled }: MfaInputProps) {
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  const digits = value.padEnd(length, '').slice(0, length).split('');

  useEffect(() => {
    // Focus first empty input on mount
    const firstEmpty = digits.findIndex(d => !d);
    const focusIndex = firstEmpty === -1 ? length - 1 : firstEmpty;
    inputRefs.current[focusIndex]?.focus();
  }, []);

  const handleChange = (index: number, char: string) => {
    if (!/^\d?$/.test(char)) return;
    
    const newDigits = [...digits];
    newDigits[index] = char;
    onChange(newDigits.join(''));
    
    // Auto-advance to next input
    if (char && index < length - 1) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent) => {
    if (e.key === 'Backspace' && !digits[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, length);
    onChange(pasted);
    inputRefs.current[pasted.length - 1]?.focus();
  };

  return (
    <div className="flex gap-2 justify-center" onPaste={handlePaste}>
      {digits.map((digit, index) => (
        <Input
          key={index}
          ref={el => { inputRefs.current[index] = el; }}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={digit}
          onChange={e => handleChange(index, e.target.value)}
          onKeyDown={e => handleKeyDown(index, e)}
          disabled={disabled}
          className="w-12 h-14 text-center text-2xl font-mono"
        />
      ))}
    </div>
  );
}
```

```typescript
// components/auth/login-form.tsx
import { AuthCard } from '@/components/shared/auth-card';
import { AuthInput } from '@/components/shared/auth-input';
import { AuthButton } from '@/components/shared/auth-button';
import { AuthError } from '@/components/shared/auth-error';
import { AuthLinks } from '@/components/shared/auth-links';
import { useLogin } from '@/hooks/auth/use-login';

export function LoginForm() {
  const {
    email,
    setEmail,
    password,
    setPassword,
    isLoading,
    errors,
    globalError,
    handleSubmit,
  } = useLogin();

  return (
    <AuthCard title="Welcome back" subtitle="Sign in to your account">
      <form onSubmit={handleSubmit} className="space-y-4">
        <AuthError message={globalError} />
        
        <AuthInput
          label="Email"
          type="email"
          value={email}
          onChange={setEmail}
          error={errors.email}
          placeholder="you@example.com"
        />
        
        <AuthInput
          label="Password"
          type="password"
          value={password}
          onChange={setPassword}
          error={errors.password}
          placeholder="••••••••"
        />
        
        <AuthButton type="submit" isLoading={isLoading}>
          Sign in
        </AuthButton>
        
        <AuthLinks
          leftLabel="Forgot password?"
          leftHref="/auth/forgot"
          rightLabel="Don't have an account?"
          rightHref="/auth/register"
        />
      </form>
      
      <div className="mt-6">
        <div className="relative">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center text-xs uppercase">
            <span className="bg-card px-2 text-muted-foreground">
              Or continue with
            </span>
          </div>
        </div>
        
        <button
          type="button"
          className="mt-4 w-full flex items-center justify-center gap-3 px-4 py-2 border rounded-lg hover:bg-muted transition-colors"
        >
          <GoogleIcon className="h-5 w-5" />
          Sign in with Google
        </button>
      </div>
    </AuthCard>
  );
}
```

```typescript
// components/auth/protected-route.tsx
import { Navigate, useLocation } from '@tanstack/react-router';
import { useAuth } from '@/context/auth-context';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export function ProtectedRoute({ children }: ProtectedRouteProps) {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <Navigate
        to="/auth/login"
        search={{ redirect: location.href }}
        replace
      />
    );
  }

  return <>{children}</>;
}
```

---

## 9. File Changes Summary

### 9.1 Total File Count

| Category | New Files | Modified Files |
|----------|-----------|----------------|
| Domain Layer | 4 | 0 |
| Data Layer (REST) | 1 | 0 |
| Presentation Layer | 7 | 0 |
| UI Layer (Shared) | 7 | 0 |
| UI Layer (Auth) | 7 | 0 |
| Context | 1 | 0 |
| Routes | 4 | 1 |
| **TOTAL** | **31** | **1** |

### 9.2 All Files Listed

#### Domain Layer (4 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `domain/auth/auth-types.ts` | **NEW** | AuthUser, LoginCredentials, MFA types |
| `domain/auth/auth-transforms.ts` | **NEW** | userFromRaw(), loginFromRaw() |
| `domain/auth/auth-validation.ts` | **NEW** | validateEmail(), validatePassword(), etc. |
| `domain/common/error.ts` | **NEW** | Auth error types (or reuse existing) |

#### Data Layer - REST (1 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `lib/api/rest/auth-rest.ts` | **NEW** | All auth REST endpoints |

#### Presentation Layer (7 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `hooks/auth/use-login.ts` | **NEW** | Login hook |
| `hooks/auth/use-register.ts` | **NEW** | Registration hook |
| `hooks/auth/use-logout.ts` | **NEW** | Logout hook |
| `hooks/auth/use-mfa.ts` | **NEW** | MFA verification hook |
| `hooks/auth/use-password-reset.ts` | **NEW** | Forgot/reset password hooks |
| `hooks/auth/use-session.ts` | **NEW** | Session management |
| `hooks/auth/use-auth-callback.ts` | **NEW** | OAuth callback hook |
| `hooks/auth/index.ts` | **NEW** | Barrel export |

#### UI Layer - Shared (7 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/shared/auth-card.tsx` | **NEW** | Centered card wrapper |
| `components/shared/auth-input.tsx` | **NEW** | Form input |
| `components/shared/auth-button.tsx` | **NEW** | Submit button |
| `components/shared/auth-error.tsx` | **NEW** | Error display |
| `components/shared/auth-links.tsx` | **NEW** | Secondary links |
| `components/shared/mfa-input.tsx` | **NEW** | 6-digit MFA input |
| `components/shared/password-strength.tsx` | **NEW** | Password strength |
| `components/shared/index.ts` | **NEW** | Barrel export (update) |

#### UI Layer - Auth (7 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `components/auth/login-form.tsx` | **NEW** | Login form |
| `components/auth/register-form.tsx` | **NEW** | Registration form |
| `components/auth/mfa-form.tsx` | **NEW** | MFA verification |
| `components/auth/forgot-password-form.tsx` | **NEW** | Forgot password |
| `components/auth/reset-password-form.tsx` | **NEW** | Reset password |
| `components/auth/auth-layout.tsx` | **NEW** | Auth page layout |
| `components/auth/protected-route.tsx` | **NEW** | Route guard |
| `components/auth/index.ts` | **NEW** | Barrel export |

#### Context (1 NEW)

| File | Status | Purpose |
|------|--------|---------|
| `context/auth-context.tsx` | **NEW** | Global auth state provider |

#### Routes (4 NEW, 1 MODIFIED)

| File | Status | Purpose |
|------|--------|---------|
| `routes/auth.login.tsx` | **MODIFIED** | Update to use LoginForm |
| `routes/auth.register.tsx` | **NEW** | `/auth/register` |
| `routes/auth.forgot.tsx` | **NEW** | `/auth/forgot` |
| `routes/auth.reset.tsx` | **NEW** | `/auth/reset` |
| `routes/auth.callback.tsx` | **NEW** | `/auth/callback` (OAuth) |

---

## 10. Implementation Order

### Phase 1: Domain Layer (Day 1)
1. Create `domain/auth/auth-types.ts` with all auth types
2. Create `domain/auth/auth-transforms.ts` with fromRaw functions
3. Create `domain/auth/auth-validation.ts` with validation functions

### Phase 2: Data Layer (Day 1)
1. Create `lib/api/rest/auth-rest.ts` with REST endpoints
2. Test token storage strategy

### Phase 3: Context (Day 1)
1. Create `context/auth-context.tsx`
2. Implement AuthProvider with all methods
3. Test global auth state

### Phase 4: Presentation Layer - Hooks (Day 2)
1. Create `hooks/auth/use-login.ts`
2. Create `hooks/auth/use-logout.ts`
3. Create `hooks/auth/use-session.ts`
4. Create `hooks/auth/use-mfa.ts`
5. Create remaining hooks

### Phase 5: UI Layer - Shared Components (Day 2)
1. Create shared components (auth-card, auth-input, etc.)
2. Create MFA input component
3. Create password strength component

### Phase 6: UI Layer - Auth Components (Day 2-3)
1. Create `login-form.tsx`
2. Create `register-form.tsx`
3. Create `mfa-form.tsx`
4. Create `forgot-password-form.tsx`
5. Create `reset-password-form.tsx`
6. Create `protected-route.tsx`

### Phase 7: Routes (Day 3)
1. Update `auth.login.tsx`
2. Create `auth.register.tsx`
3. Create `auth.forgot.tsx`
4. Create `auth.reset.tsx`
5. Create `auth.callback.tsx`

### Phase 8: Integration (Day 3-4)
1. Wire AuthProvider at app root
2. Add ProtectedRoute to dashboard
3. Test complete auth flows
4. Add loading/error states
5. Test session timeout

### Phase 9: Polish (Day 4)
1. Add OAuth Google flow
2. Add MFA enrollment flow
3. Add "Remember me" option
4. Test edge cases

---

## Appendix: API Contract Reference

### Login
```
POST /v1/auth/login
Request: { email, password }
Response: { user, tokens, mfa_required, mfa_token? }
```

### Register
```
POST /v1/auth/register
Request: { email, password, name }
Response: { user, tokens }
```

### MFA Verify
```
POST /v1/auth/mfa/verify
Request: { mfa_token, code }
Response: { user, tokens }
```

### Forgot Password
```
POST /v1/auth/forgot-password
Request: { email }
Response: { success: true }
```

### Reset Password
```
POST /v1/auth/reset-password
Request: { token, password }
Response: { success: true }
```

### Get Current User
```
GET /v1/auth/me
Response: { user } | 401
```

### Logout
```
POST /v1/auth/logout
Response: { success: true }
```

### Refresh Token
```
POST /v1/auth/refresh
Response: { access_token, refresh_token, expires_at }
```

---

*Document Version: 1.0*  
*Status: Ready for Implementation*  
*Architecture: Layered (Following FRONTEND_ARCHITECTURE.md)*
