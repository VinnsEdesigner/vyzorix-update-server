/**
 * login.tsx - Operator login page.
 *
 * Uses the wolf background layout and LoginForm component.
 */

import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { type ReactNode, useEffect, useState } from "react";
import { toast } from "sonner";

import wolfImage from "@/assets/images/black_wolf_evening_1781264516831.jpg";
import AuthLayout from "@/components/layout/AuthLayout";
import LoginForm from "@/components/auth/LoginForm";
import { loginOperator } from "@/lib/clients/authClient";
import { initiateSSO, type SSOProvider } from "@/lib/clients/ssoClient";
import { getFullHydratedState } from "@/lib/server/state-injector";

const LoginPage = (): ReactNode => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  // SSR-aware authentication check
  useEffect(() => {
    const checkAuth = async (): Promise<void> => {
      // 1. Check server-injected state first (SSR hydration)
      if (typeof window !== "undefined") {
        const globalState = getFullHydratedState();
        if (globalState?.isAuthenticated) {
          navigate({ to: "/dashboard", replace: true });
          return;
        }
      }

      // 2. Fallback: Fetch from API (client-side navigation without SSR state)
      try {
        const res = await fetch("/v1/auth/me", { credentials: "include" });
        if (res.ok) {
          navigate({ to: "/dashboard", replace: true });
        }
      } catch {
        // Not authenticated - stay on login page
      }
    };

    checkAuth();
  }, [navigate]);

  const handleLogin = async (ident: string, pass: string): Promise<void> => {
    setIsSubmitting(true);
    try {
      await loginOperator(ident, pass);
      toast.success("Welcome back!");
      navigate({ to: "/dashboard", replace: true });
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Login failed";
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSSO = (provider: SSOProvider): void => {
    initiateSSO(provider);
  };

  const handleForgotPassword = (): void => {
    navigate({ to: "/forgot-password" });
  };

  return (
    <div className="relative min-h-screen w-full overflow-hidden">
      {/* Wolf background image - fixed, covers viewport */}
      <div
        className="fixed inset-0 bg-cover bg-center bg-no-repeat"
        style={{ backgroundImage: `url(${wolfImage})` }}
        aria-hidden="true"
      />

      {/* Dark gradient overlay for readability */}
      <div
        className="fixed inset-0 bg-gradient-to-br from-slate-950/80 via-slate-950/70 to-slate-950/85"
        aria-hidden="true"
      />

      {/* Content container - centered */}
      <div className="relative z-10 flex min-h-screen items-center justify-center px-4 py-8">
        <div className="w-full max-w-md">
          <LoginForm
            onLogin={handleLogin}
            onSSO={handleSSO}
            onForgotPassword={handleForgotPassword}
            isSubmitting={isSubmitting}
          />
        </div>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/login")({
  head: () => ({ meta: [{ title: "Sign in — Vyzorix" }] }),
  component: () => (
    <AuthLayout>
      <LoginPage />
    </AuthLayout>
  ),
});
